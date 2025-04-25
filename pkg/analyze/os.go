package analyze

import (
	"os"
	"path/filepath"

	"image-analyzer-go/pkg/logger"
)

func CheckOSInfo(root string) string {
	path := filepath.Join(root, "etc", "os-release")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("读取操作系统信息失败", logger.WithError(err))
		return ""
	}
	return string(data)
}
