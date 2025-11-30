package video

import (
	"context"
	"fmt"
	"log"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
	videoUtils "imgnation-backend/pkg/utils/video"
	"imgnation-backend/repository/storage"

	"github.com/safeblock-dev/werr"
)

func NewVariantFunc(
	videoInfo *videoUtils.VideoInfo,
	opts *types.ProcessFileOpts,
	exec CreateVariantsFunc,
	variantOpts []VideoVariantOpts,
) process.VariantFunc {
	return func(ctx context.Context, storage storage.IStorage) ([]types.FileVariantInfo, []types.ChunkInfo, error) {
		return exec(ctx, storage, videoInfo, opts, variantOpts)
	}
}

func ProcessVideo(
	ctx context.Context,
	opts *types.ProcessFileOpts,
) ([]process.VariantFunc, error) {
	var err error
	videoInfo, err := videoUtils.GetVideoInfo(ctx, opts.FilePath)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to get video info")
	}
	formatOpts, exists := SupportedVideoFormats[opts.DetectedContentType]
	if !exists {
		return nil, fmt.Errorf("unsupported video format: %s", opts.DetectedContentType)
	}

	variantFuncs := []process.VariantFunc{}

	for _, variantGroup := range formatOpts.CreateVariants {
		processVariants, exists := VariantFuncs[variantGroup.Name]
		if !exists {
			log.Printf(
				"Warning: not found variant group processing func "+
					"for group name %s content type %s\n",
				variantGroup.Name, opts.DetectedContentType,
			)
			continue
		}

		// Create a group with only active variants
		var activeVariants []VideoVariantOpts
		for _, variant := range variantGroup.Variants {
			if variant.ShouldCreate(videoInfo, opts) {
				activeVariants = append(activeVariants, variant)
			}
		}
		if len(activeVariants) == 0 {
			continue
		}

		switch variantGroup.Strategy {
		case process.StrategySequential:
			// variants are processed one by one in a loop
			for _, variant := range activeVariants {
				variantFuncs = append(variantFuncs, NewVariantFunc(
					videoInfo,
					opts,
					processVariants,
					[]VideoVariantOpts{variant},
				))
			}
		case process.StrategyMultivariant:
			// variant options are provided in a batch
			variantFuncs = append(variantFuncs, NewVariantFunc(
				videoInfo,
				opts,
				processVariants,
				activeVariants,
			))
		}

	}
	return variantFuncs, nil
}
