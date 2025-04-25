package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"image-analyzer-go/pkg/analyze"
	"image-analyzer-go/pkg/config"
	"image-analyzer-go/pkg/imageutil"
	"image-analyzer-go/pkg/logger"
	"image-analyzer-go/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type AnalyzeImage struct {
	cfg *config.Config
}

func NewAnalyzeImage(cfg *config.Config) *AnalyzeImage {
	return &AnalyzeImage{cfg: cfg}
}

func (a *AnalyzeImage) HandleAnalyze(c *gin.Context) {
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Options == nil {
		req.Options = &analyze.AnalyzeOptions{
			CheckOSInfo:         a.cfg.Analyze.CheckOSInfo,
			CheckPythonPackages: a.cfg.Analyze.CheckPythonPackages,
			CheckCommonTools:    a.cfg.Analyze.CheckCommonTools,
			SpecificCommands:    a.cfg.Analyze.SpecificCommands,
		}
	}

	ctx := context.Background()
	layersDir, imgCfg, err := imageutil.PullAndExtract(ctx, req.ImageRef, a.cfg.GetUnpackDir())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("提取镜像失败: %v", err)})
		return
	}
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
		c.Header("Content-Type", "application/x-yaml")
	case "json", "":
		response, marshalErr = json.MarshalIndent(summary, "", "  ")
		c.Header("Content-Type", "application/json")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的输出格式: " + req.Format})
		return
	}

	if marshalErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("生成报告失败: %v", marshalErr)})
		return
	}

	c.String(http.StatusOK, string(response))
}
