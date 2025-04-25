package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"image-analyzer-go/cmd"
	"image-analyzer-go/pkg/config"
	"image-analyzer-go/pkg/logger"
)

func main() {
	runApp()
}

func runApp() {
	// 加载默认配置
	cfg := config.DefaultConfig()

	// 解析命令行参数
	configPath := flag.String("f", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	loadedCfg, err := config.LoadConfig(*configPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			loadedCfg = cfg
		} else {
			fmt.Println("加载配置失败", err)
			os.Exit(1)
		}
	}
	cfg = loadedCfg

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
	// 设置命令上下文
	cmd.SetContext(context.Background())
	// 将配置设置到命令上下文中
	cmd.SetConfig(cfg)
	// 执行命令
	if err := cmd.Execute(); err != nil {
		logger.Fatal("执行失败", logger.WithError(err))
	}
}
