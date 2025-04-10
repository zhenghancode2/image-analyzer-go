package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateUniqueID 生成一个唯一的标识符
func GenerateUniqueID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
