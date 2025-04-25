package cmd

import (
	"context"
	"encoding/json"
	"errors"

	"image-analyzer-go/pkg/analyze"
	"image-analyzer-go/pkg/imageutil"
	"image-analyzer-go/pkg/logger"
	"image-analyzer-go/pkg/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outputFile          string
	imageRef            string
	format              string
	checkOSInfo         bool
	checkPythonPackages bool
	checkCommonTools    bool
	specificCommands    []string
	unpackDir           string
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
	analyzeCmd.Flags().BoolVar(&checkOSInfo, "check-os", true, "是否检查系统信息")
	analyzeCmd.Flags().BoolVar(&checkPythonPackages, "check-python", true, "是否检查 Python 包")
	analyzeCmd.Flags().BoolVar(&checkCommonTools, "check-tools", true, "是否检查常用工具")
	analyzeCmd.Flags().StringSliceVar(&specificCommands, "commands", []string{}, "要检查的特定命令列表")
	analyzeCmd.Flags().StringVarP(&unpackDir, "unpack-dir", "d", "images", "解压缩镜像的临时目录")
}

func runAnalysis() error {
	ctx := context.Background()

	layersDir, imgCfg, err := imageutil.PullAndExtract(ctx, imageRef, unpackDir)
	if err != nil {
		return utils.WrapError(err, "提取镜像失败")
	}
	// 优雅地清理临时目录
	defer func(dir string) {
		if dir == "" {
			return
		}
		if cleanupErr := utils.CleanupTempDir(dir); cleanupErr != nil {
			logger.Warn("清理临时目录失败", logger.WithString("dir", dir), logger.WithError(cleanupErr))
		}
	}(layersDir)

	summary := analyze.Summary{
		Architecture: imgCfg.Architecture,
		OS:           imgCfg.OS,
		Env:          imgCfg.Config.Env,
	}

	if checkOSInfo {
		summary.OSInfo = analyze.CheckOSInfo(layersDir)
	}
	if checkPythonPackages {
		summary.PythonPackages = analyze.ListPythonPackages(layersDir)
	}
	if checkCommonTools {
		summary.Tools = analyze.CheckCommonTools(layersDir)
	}

	var output []byte
	var marshalErr error

	switch format {
	case "yaml":
		output, marshalErr = yaml.Marshal(summary)
	case "json":
		output, marshalErr = json.MarshalIndent(summary, "", "  ")
	default:
		return errors.New("不支持的输出格式: " + format)
	}

	if marshalErr != nil {
		return utils.WrapError(marshalErr, "生成报告失败")
	}

	if err := utils.WriteFile(outputFile, output, 0644); err != nil {
		return utils.WrapError(err, "写入报告文件失败")
	}

	logger.Info("分析报告已保存", logger.WithString("file", outputFile))
	return nil
}
