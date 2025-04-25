package cmd

import (
	"context"
	"image-analyzer-go/pkg/config"
)

// configKey 是用于在上下文中存储配置的键
type configKey struct{}

// WithConfig 将配置添加到上下文中
func WithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, configKey{}, cfg)
}

// GetConfig 从上下文中获取配置
func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(configKey{}).(*config.Config); ok {
		return cfg
	}
	return config.DefaultConfig()
}
