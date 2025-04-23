package imageutil

import (
	"image-analyzer-go/pkg/logger"
	"time"

	"github.com/containers/image/v5/types"
)

// NewProgressHandler 创建一个新的进度处理器
func NewProgressHandler() chan types.ProgressProperties {
	ch := make(chan types.ProgressProperties)
	go func() {
		lastUpdate := time.Now()
		for bar := range ch {
			now := time.Now()
			if now.Sub(lastUpdate) >= time.Second {
				logger.Info("下载进度", logger.WithAny("event", bar.Event))
				lastUpdate = now
			}
		}
	}()
	return ch
}
