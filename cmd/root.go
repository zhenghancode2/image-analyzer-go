package cmd

import (
	"context"
	"image-analyzer-go/pkg/config"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "image-analyzer",
		Short: "一个的容器镜像分析工具",
		Long: `image-analyzer 是一个用于分析容器镜像的 CLI 工具。
它可以提取镜像的详细信息，包括：
- 基础系统信息
- 已安装的软件包
- Python 包依赖
- 系统工具等`,
	}
)

func SetContext(ctx context.Context) {
	rootCmd.SetContext(ctx)
}

// SetConfig 在根命令上下文中设置配置
func SetConfig(cfg *config.Config) {
	rootCmd.SetContext(WithConfig(rootCmd.Context(), cfg))
}

func Execute() error {
	return rootCmd.Execute()
}
