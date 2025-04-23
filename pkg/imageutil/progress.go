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
				switch bar.Event {
				case types.ProgressEventNewArtifact:
					logger.Info("开始下载新制品", logger.WithString("digest", bar.Artifact.Digest.String()))
				case types.ProgressEventRead:
					logger.Info("正在读取制品", logger.WithString("digest", bar.Artifact.Digest.String()))
				case types.ProgressEventDone:
					logger.Info("制品下载完成", logger.WithString("digest", bar.Artifact.Digest.String()))
				case types.ProgressEventSkipped:
					logger.Info("制品已跳过", logger.WithString("digest", bar.Artifact.Digest.String()))
				}
				lastUpdate = now
			}
		}
	}()
	return ch
}
