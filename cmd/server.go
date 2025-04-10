package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"image-analyzer-go/pkg/analyze"
	"image-analyzer-go/pkg/imageutil"
	"image-analyzer-go/pkg/logger"
	"image-analyzer-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 API 服务器",
	Long:  `启动一个 HTTP 服务器，提供镜像分析 API。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

type AnalysisRequest struct {
	ImageRef string                  `json:"image_ref" binding:"required"`
	Options  *analyze.AnalyzeOptions `json:"options"`
	Format   string                  `json:"format" binding:"oneof=json yaml"`
}

func runServer() error {
	r := gin.Default()

	// 设置请求大小限制
	r.MaxMultipartMemory = cfg.Server.MaxRequestSize

	// 设置超时中间件
	r.Use(func(c *gin.Context) {
		c.Set("timeout", cfg.Server.ReadTimeout)
		c.Next()
	})

	r.POST("/analyze", handleAnalyze)
	r.GET("/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("启动服务器", logger.WithString("addr", addr))
	return r.Run(addr)
}

func handleAnalyze(c *gin.Context) {
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Options == nil {
		req.Options = &analyze.AnalyzeOptions{
			CheckOSInfo:         cfg.Analyze.CheckOSInfo,
			CheckPythonPackages: cfg.Analyze.CheckPythonPackages,
			CheckCommonTools:    cfg.Analyze.CheckCommonTools,
			SpecificCommands:    cfg.Analyze.SpecificCommands,
		}
	}

	ctx := context.Background()
	layersDir, imgCfg, err := imageutil.PullAndExtract(ctx, req.ImageRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("提取镜像失败: %v", err)})
		return
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

	if req.Options.CheckOSInfo {
		summary.OSInfo = analyze.CheckOSInfo(layersDir)
	}
	if req.Options.CheckPythonPackages {
		summary.PythonPackages = analyze.ListPythonPackages(layersDir)
	}
	if req.Options.CheckCommonTools {
		summary.Tools = analyze.CheckCommonTools(layersDir)
	}

	var response []byte
	var marshalErr error

	switch req.Format {
	case "yaml":
		response, marshalErr = yaml.Marshal(summary)
	default:
		response, marshalErr = json.MarshalIndent(summary, "", "  ")
	}

	if marshalErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("生成报告失败: %v", marshalErr)})
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(response))
}
