package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"

	_ "golang.org/x/image/webp"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository/storage"

	"github.com/disintegration/imaging"
	"github.com/safeblock-dev/werr"
)

func encodeImage(img image.Image, format string, quality int) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	switch format {
	case "image/jpeg":
		opts := imaging.JPEGQuality(quality)
		if err := imaging.Encode(&buf, img, imaging.JPEG, opts); err != nil {
			return nil, err
		}
	case "image/png":
		if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
			return nil, err
		}
	default:
		return nil, werr.Wrapf(xerr.ErrInvalidArgument,
			"unsupported output format: %s", format)
	}

	return &buf, nil
}

// CreateVariantsFunc defines the function signature for creating image variants
type CreateVariantsFunc func(
	ctx context.Context,
	storage storage.IStorage,
	img image.Image,
	opts *types.ProcessFileOpts,
	variantOpts []ImageVariantOpts,
) ([]types.FileVariantInfo, []types.ChunkInfo, error)

// VariantFuncs maps variant group names to their processing functions

func NewVariantFunc(
	img image.Image,
	opts *types.ProcessFileOpts,
	exec CreateVariantsFunc,
	variantOpts []ImageVariantOpts,
) process.VariantFunc {
	return func(ctx context.Context, storage storage.IStorage) ([]types.FileVariantInfo, []types.ChunkInfo, error) {
		return exec(ctx, storage, img, opts, variantOpts)
	}
}

func ProcessImage(
	ctx context.Context,
	opts *types.ProcessFileOpts,
) ([]process.VariantFunc, error) {
	img, err := imaging.Decode(opts.Reader, imaging.AutoOrientation(true))
	if err != nil {
		return nil, werr.Wrapf(xerr.ErrInvalidArgument, "failed to decode image")
	}

	formatOpts, exists := SupportedImageFormats[opts.DetectedContentType]
	if !exists {
		return nil, fmt.Errorf("unsupported image format: %s", opts.DetectedContentType)
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
		var activeVariants []ImageVariantOpts
		for _, variant := range variantGroup.Variants {
			if variant.ShouldCreate(img, opts) {
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
					img,
					opts,
					processVariants,
					[]ImageVariantOpts{variant},
				))
			}
		case process.StrategyMultivariant:
			// variant options are provided in a batch
			variantFuncs = append(variantFuncs, NewVariantFunc(
				img,
				opts,
				processVariants,
				activeVariants,
			))
		}
	}

	return variantFuncs, nil
}
