package handler

import "image-analyzer-go/pkg/analyze"

type AnalysisRequest struct {
	ImageRef string                  `json:"image_ref" binding:"required"`
	Options  *analyze.AnalyzeOptions `json:"options"`
	Format   string                  `json:"format" binding:"oneof=json yaml"`
}
