package utils

import (
	"fmt"
)

// WrapError 包装错误，添加上下文信息
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}
