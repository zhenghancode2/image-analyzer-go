package cmd

import (
	"fmt"
	"net/http"
	"time"

	"image-analyzer-go/pkg/config"
	"image-analyzer-go/pkg/logger"
	"image-analyzer-go/pkg/router"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 API 服务器",
	Long:  `启动一个 HTTP 服务器，提供镜像分析 API。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令上下获取配置
		cfg := GetConfig(cmd.Context())
		return runServer(cfg)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func runServer(cfg *config.Config) error {
	// 根据环境设置 Gin 模式
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// 设置请求大小限制
	r.MaxMultipartMemory = cfg.Server.MaxRequestSize
	// 配置请求日志记录
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %s | %d | %s | %s | %s %s\n",
				param.TimeStamp.Format(time.RFC3339),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
			)
		},
		Output:    gin.DefaultWriter,
		SkipPaths: []string{"/readiness"}, // 跳过健康检查路径的日志
	}))
	// 使用路由器设置路由
	router.SetupRouters(r.Group("/api/v1"), cfg)
	// 可读性检查
	r.GET("/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 使用更优雅的方式获取服务地址
	addr := cfg.GetServerAddress()
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	logger.Info("启动服务器", logger.WithString("addr", addr))
	return httpServer.ListenAndServe()

}
