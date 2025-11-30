package image

import (
	"image"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
)

type ImageVariantOpts struct {
	VariantName  string
	Format       string
	Quality      int
	ShouldCreate func(img image.Image, opts *types.ProcessFileOpts) bool
}

type (
	Format       = process.Format[ImageVariantOpts]
	VariantGroup = process.VariantGroup[ImageVariantOpts]
)

var SupportedImageFormats = map[string]Format{
	"image/jpeg": {
		Name:   "jpeg",
		Format: "image/jpeg",
		CreateVariants: []VariantGroup{
			{
				Name:     VariantNameOriginal,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						Format:       "image/jpeg",
						Quality:      0,
						ShouldCreate: func(_ image.Image, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameThumbnail,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						Format:       "image/jpeg",
						Quality:      70,
						ShouldCreate: shouldCreateThumbnail,
					},
				},
			},
			{
				Name:     VariantNameCompressed,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameCompressed,
						Format:       "image/jpeg",
						Quality:      80,
						ShouldCreate: func(_ image.Image, opts *types.ProcessFileOpts) bool { return opts.SizeBytes > 100<<10 },
					},
				},
			},
		},
	},
	"image/png": {
		Name:   "png",
		Format: "image/png",
		CreateVariants: []VariantGroup{
			{
				Name:     VariantNameOriginal,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						Format:       "image/png",
						Quality:      0,
						ShouldCreate: func(_ image.Image, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameThumbnail,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						Format:       "image/png",
						Quality:      0,
						ShouldCreate: shouldCreateThumbnail,
					},
				},
			},
			{
				Name:     VariantNameCompressed,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameCompressed,
						Format:       "image/jpeg", // convert large PNGs to JPEG
						Quality:      70,
						ShouldCreate: func(_ image.Image, opts *types.ProcessFileOpts) bool { return opts.SizeBytes > 200<<10 },
					},
				},
			},
		},
	},
	"image/webp": {
		Name:   "webp",
		Format: "image/webp",
		CreateVariants: []VariantGroup{
			{
				Name:     VariantNameOriginal,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						Format:       "image/webp",
						Quality:      0,
						ShouldCreate: func(_ image.Image, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameThumbnail,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						Format:       "image/webp",
						Quality:      70,
						ShouldCreate: shouldCreateThumbnail,
					},
				},
			},
			{
				Name:     VariantNameCompressed,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameCompressed,
						Format:       "image/webp",
						Quality:      80,
						ShouldCreate: func(_ image.Image, opts *types.ProcessFileOpts) bool { return opts.SizeBytes > 100<<10 },
					},
				},
			},
		},
	},
	"image/gif": {
		Name:   "gif",
		Format: "image/gif",
		CreateVariants: []VariantGroup{
			{
				Name:     VariantNameOriginal,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameOriginal,
						Format:       "image/gif",
						Quality:      0,
						ShouldCreate: func(_ image.Image, _ *types.ProcessFileOpts) bool { return true },
					},
				},
			},
			{
				Name:     VariantNameThumbnail,
				Strategy: process.StrategySequential,
				Variants: []ImageVariantOpts{
					{
						VariantName:  VariantNameThumbnail,
						Format:       "image/jpeg", // convert GIF thumbnails to JPEG
						Quality:      70,
						ShouldCreate: shouldCreateThumbnail,
					},
				},
			},
		},
	},
}

func shouldCreateThumbnail(img image.Image, _ *types.ProcessFileOpts) bool {
	return img.Bounds().Max.X > ThumbnailMaxWidth
}
