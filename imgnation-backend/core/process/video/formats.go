package video

import (
	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
	videoUtils "imgnation-backend/pkg/utils/video"
)

type VideoVariantOpts struct {
	VariantName  string
	OutputFormat string
	Resolution   string
	CRF          int // Constant Rate Factor (quality)
	MaxBitrate   int // in kbps
	ShouldCreate func(info *videoUtils.VideoInfo, opts *types.ProcessFileOpts) bool
}

type (
	Format       = process.Format[VideoVariantOpts]
	VariantGroup = process.VariantGroup[VideoVariantOpts]
)

var SupportedVideoFormats = map[string]Format{
	"video/mp4": {
		Name:   "mp4",
		Format: "video/mp4",
		CreateVariants: []VariantGroup{
			{
				Name:     VariantNameOriginal,
				Strategy: process.StrategySequential,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						OutputFormat: "video/mp4",
						ShouldCreate: func(_ *videoUtils.VideoInfo, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameThumbnail,
				Strategy: process.StrategySequential,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						OutputFormat: "image/jpeg",
						ShouldCreate: func(_ *videoUtils.VideoInfo, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameMpegDash1080p,
				Strategy: process.StrategyMultivariant,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameMpegDash320p,
						OutputFormat: "video/mp4",
						Resolution:   "320p",
						CRF:          28,
						MaxBitrate:   500,
						ShouldCreate: shouldCreateCompressed320p,
					},
					{
						VariantName:  VariantNameMpegDash720p,
						OutputFormat: "video/mp4",
						Resolution:   "720p",
						CRF:          23,
						MaxBitrate:   2500,
						ShouldCreate: shouldCreateCompressed720p,
					},
					{
						VariantName:  VariantNameMpegDash1080p,
						OutputFormat: "video/mp4",
						Resolution:   "1080p",
						CRF:          23,
						MaxBitrate:   5000,
						ShouldCreate: shouldCreateCompressed1080p,
					},
				},
			},
		},
	},
	"video/quicktime": {
		Name:   "mov",
		Format: "video/quicktime",
		CreateVariants: []VariantGroup{
			{
				Name:     VariantNameOriginal,
				Strategy: process.StrategySequential,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						OutputFormat: "video/quicktime",
						ShouldCreate: func(_ *videoUtils.VideoInfo, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameThumbnail,
				Strategy: process.StrategySequential,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						OutputFormat: "image/jpeg",
						ShouldCreate: func(_ *videoUtils.VideoInfo, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     "compression",
				Strategy: process.StrategyMultivariant,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameMpegDash320p,
						OutputFormat: "video/mp4",
						Resolution:   "320p",
						CRF:          28,
						MaxBitrate:   500,
						ShouldCreate: shouldCreateCompressed320p,
					},
					{
						VariantName:  VariantNameMpegDash720p,
						OutputFormat: "video/mp4",
						Resolution:   "720p",
						CRF:          23,
						MaxBitrate:   2500,
						ShouldCreate: shouldCreateCompressed720p,
					},
					{
						VariantName:  VariantNameMpegDash1080p,
						OutputFormat: "video/mp4",
						Resolution:   "1080p",
						CRF:          23,
						MaxBitrate:   5000,
						ShouldCreate: shouldCreateCompressed1080p,
					},
				},
			},
		},
	},
	"video/x-msvideo": {
		Name:   "avi",
		Format: "video/x-msvideo",
		CreateVariants: []VariantGroup{
			{
				Name:     "original",
				Strategy: process.StrategySequential,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						OutputFormat: "video/x-msvideo",
						ShouldCreate: func(_ *videoUtils.VideoInfo, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     "thumbnails",
				Strategy: process.StrategySequential,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						OutputFormat: "image/jpeg",
						ShouldCreate: func(_ *videoUtils.VideoInfo, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     "compression",
				Strategy: process.StrategyMultivariant,
				Variants: []VideoVariantOpts{
					{
						VariantName:  VariantNameMpegDash720p,
						OutputFormat: "video/mp4",
						Resolution:   "720p",
						CRF:          23,
						MaxBitrate:   2500,
						ShouldCreate: shouldCreateCompressed720p,
					},
					{
						VariantName:  VariantNameMpegDash1080p,
						OutputFormat: "video/mp4",
						Resolution:   "1080p",
						CRF:          23,
						MaxBitrate:   5000,
						ShouldCreate: shouldCreateCompressed1080p,
					},
				},
			},
		},
	},
}

// Conditional creation functions
func shouldCreateCompressed320p(info *videoUtils.VideoInfo, opts *types.ProcessFileOpts) bool {
	// Create 320p for all videos - useful for previews and low bandwidth
	return true
}

func shouldCreateCompressed720p(info *videoUtils.VideoInfo, opts *types.ProcessFileOpts) bool {
	// Create 720p if original is higher resolution or file is large
	return info.Height > 720 || opts.SizeBytes > 50<<20 // 50MB
}

func shouldCreateCompressed1080p(info *videoUtils.VideoInfo, opts *types.ProcessFileOpts) bool {
	// Create 1080p only if original is higher resolution
	return info.Height > 1080
}
