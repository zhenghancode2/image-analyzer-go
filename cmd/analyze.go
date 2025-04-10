package cmd

import (
	"context"
	"encoding/json"

	"image-analyzer-go/pkg/analyze"
	"image-analyzer-go/pkg/imageutil"
	"image-analyzer-go/pkg/logger"
	"image-analyzer-go/pkg/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outputFile string
	imageRef   string
	format     string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [image-reference]",
	Short: "分析指定的容器镜像",
	Long:  `分析指定的容器镜像，提取其详细信息并生成报告。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef = args[0]
		return runAnalysis()
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "report.json", "输出报告的文件路径")
	analyzeCmd.Flags().StringVarP(&format, "format", "f", "json", "输出格式 (json 或 yaml)")
	analyzeCmd.Flags().BoolVar(&cfg.Analyze.CheckOSInfo, "check-os", true, "是否检查系统信息")
	analyzeCmd.Flags().BoolVar(&cfg.Analyze.CheckPythonPackages, "check-python", true, "是否检查 Python 包")
	analyzeCmd.Flags().BoolVar(&cfg.Analyze.CheckCommonTools, "check-tools", true, "是否检查常用工具")
	analyzeCmd.Flags().StringSliceVar(&cfg.Analyze.SpecificCommands, "commands", []string{}, "要检查的特定命令列表")
}

func runAnalysis() error {
	ctx := context.Background()

	layersDir, imgCfg, err := imageutil.PullAndExtract(ctx, imageRef)
	if err != nil {
		return utils.WrapError(err, "提取镜像失败")
	}
	defer func() {
		if cleanupErr := utils.CleanupTempDir(layersDir); cleanupErr != nil {
			logger.Warn("清理临时目录失败", logger.WithString("dir", layersDir), logger.WithError(cleanupErr))
		}
	}()

	summary := analyze.Summary{
		Architecture: imgCfg.Architecture,
		OS:           imgCfg.OS,
		Env:          imgCfg.Config.Env,
	}

	if cfg.Analyze.CheckOSInfo {
		summary.OSInfo = analyze.CheckOSInfo(layersDir)
	}
	if cfg.Analyze.CheckPythonPackages {
		summary.PythonPackages = analyze.ListPythonPackages(layersDir)
	}
	if cfg.Analyze.CheckCommonTools {
		summary.Tools = analyze.CheckCommonTools(layersDir)
	}

	var output []byte
	var marshalErr error

	switch format {
	case "yaml":
		output, marshalErr = yaml.Marshal(summary)
	default:
		output, marshalErr = json.MarshalIndent(summary, "", "  ")
	}

	if marshalErr != nil {
		return utils.WrapError(marshalErr, "生成报告失败")
	}

	if err := utils.WriteFile(outputFile, output, 0644); err != nil {
		return err
	}

	logger.Info("分析报告已保存", logger.WithString("file", outputFile))
	return nil
}
