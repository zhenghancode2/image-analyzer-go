package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"image-analyzer-go/pkg/logger"

	"github.com/avast/retry-go/v4"
)

// EnsureDirExists 确保目录存在，如果不存在则创建
func EnsureDirExists(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	return nil
}

// CreateTempDir 创建一个带有前缀的临时目录
func CreateTempDir(prefix string) (string, error) {
	uniqueID, err := GenerateUniqueID()
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("%s-%s", prefix, uniqueID))
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	logger.Info("创建临时目录", logger.WithString("dir", tmpDir), logger.WithString("prefix", prefix))
	return tmpDir, nil
}

// CleanupTempDir 清理临时目录，带有重试机制
func CleanupTempDir(dir string) error {
	if dir == "" {
		return nil
	}

	// 定义重试策略
	opts := []retry.Option{
		retry.Attempts(3),                   // 最多重试3次
		retry.Delay(100 * time.Millisecond), // 初始延迟100ms
		retry.MaxDelay(1 * time.Second),     // 最大延迟1s
		retry.OnRetry(func(n uint, err error) {
			logger.Warn("清理目录失败，准备重试",
				logger.WithString("dir", dir),
				logger.WithInt("attempt", int(n+1)),
				logger.WithError(err))
		}),
	}

	// 使用重试机制执行删除操作
	err := retry.Do(func() error {
		if err := os.RemoveAll(dir); err != nil {
			// 检查是否是权限错误，如果是则立即返回错误
			if os.IsPermission(err) {
				return retry.Unrecoverable(fmt.Errorf("权限不足，无法删除目录 %s: %w", dir, err))
			}
			return fmt.Errorf("删除目录 %s 失败: %w", dir, err)
		}
		return nil
	}, opts...)

	if err != nil {
		logger.Error("清理目录失败，已重试3次",
			logger.WithString("dir", dir),
			logger.WithError(err))
		return fmt.Errorf("清理目录 %s 失败 (已重试3次): %w", dir, err)
	}

	logger.Info("清理目录成功", logger.WithString("dir", dir))
	return nil
}

// WriteFile 写入文件，确保目录存在
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := EnsureDirExists(filepath.Dir(path)); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	logger.Info("写入文件成功", logger.WithString("path", path))
	return nil
}

// 指定路径创建路径
func CreatePath(path string) error {
	if err := EnsureDirExists(path); err != nil {
		return fmt.Errorf("创建路径失败: %w", err)
	}
	logger.Info("创建路径成功", logger.WithString("path", path))
	return nil
}

// EnsureAbsPath 确保路径是绝对路径，如果是相对路径则转换为绝对路径
func EnsureAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	absPath := filepath.Join(currentDir, path)
	logger.Info("将相对路径转换为绝对路径",
		logger.WithString("relative_path", path),
		logger.WithString("absolute_path", absPath))

	return absPath, nil
}
