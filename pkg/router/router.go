package router

import (
	"image-analyzer-go/pkg/config"
	"image-analyzer-go/pkg/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouters(rg *gin.RouterGroup, cfg *config.Config) {
	imgHandler := handler.NewAnalyzeImage(cfg)
	rg.POST("/analyze", imgHandler.HandleAnalyze)
}
