package analyze

import (
	"os"
	"path/filepath"

	"image-analyzer-go/pkg/logger"
)

func CheckOSInfo(root string) string {
	// 只有运行的容器里/etc/os-release才会被建立软链，这里静态的直接取/usr/lib/os-release
	path := filepath.Join(root, "usr", "lib", "os-release")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("读取操作系统信息失败", logger.WithError(err))
		return ""
	}
	return string(data)
}
