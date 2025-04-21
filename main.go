package main

import (
	"flag"
	"fmt"
	"os"

	"image-analyzer-go/cmd"
	"image-analyzer-go/pkg/config"
	"image-analyzer-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("f", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			cfg = config.DefaultConfig()
		} else {
			fmt.Println("加载配置失败", err)
			os.Exit(1)
		}
	}

	// 初始化日志系统
	if err := logger.Init(cfg); err != nil {
		fmt.Println("初始化日志系统失败", err)
		os.Exit(1)
	}
	defer logger.Logger.Sync()

	// 确保目录存在
	if err := cfg.EnsureDirs(); err != nil {
		logger.Fatal("目录不存在", logger.WithError(err))
	}

	// 设置 gin 运行模式（从配置读取）
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 执行命令
	if err := cmd.Execute(); err != nil {
		logger.Fatal("执行失败", logger.WithError(err))
	}
}
