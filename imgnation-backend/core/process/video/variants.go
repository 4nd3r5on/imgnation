package video

import (
	"context"
	"io"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
	videoUtils "imgnation-backend/pkg/utils/video"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository/storage"

	"github.com/safeblock-dev/werr"
)

const (
	VariantNameOriginal      = ""
	VariantNameThumbnail     = "thumbnail"
	VariantGroupNameMpegDash = "mpeg-dash"
)

const ThumbnailMaxWidth = 320

type CreateVariantsFunc func(
	ctx context.Context,
	storage storage.IStorage,
	videoInfo *videoUtils.VideoInfo,
	opts *types.ProcessFileOpts,
	variantOpts []VideoVariantOpts,
) ([]types.FileVariantInfo, []types.ChunkInfo, error)

// variant group : variant group process func
var VariantFuncs = map[string]CreateVariantsFunc{
	VariantNameOriginal: CreateOriginalVariant,
	// VariantNameThumbnail: CreateVideoThumbnail,
	// VariantGroupNameMpegDash: CreateVideoCompressed,
}

func CreateOriginalVariant(
	ctx context.Context,
	s storage.IStorage,
	videoInfo *videoUtils.VideoInfo,
	opts *types.ProcessFileOpts,
	variantOpts []VideoVariantOpts,
) ([]types.FileVariantInfo, []types.ChunkInfo, error) {
	if len(variantOpts) != 1 {
		return nil, nil, werr.Wrapf(
			xerr.ErrOutOfRange,
			"expected one variant argument. Got: %d", len(variantOpts),
		)
	}

	variant := variantOpts[0]

	// Reset reader to beginning
	_, err := opts.Reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to seek reader")
	}

	key := process.MakeKeyFileVariant(opts.UploadID, variant.VariantName)
	err = s.Store(ctx, storage.StoreObj{
		Key:         key,
		Size:        opts.SizeBytes,
		ContentType: variant.OutputFormat,
		Content:     opts.Reader,
	})
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to store object key %s", key)
	}

	return []types.FileVariantInfo{{
		Name:        variant.VariantName,
		ContentType: variant.OutputFormat,
		SizeBytes:   opts.SizeBytes,
		Metadata: map[string]any{
			"height":   videoInfo.Height,
			"width":    videoInfo.Width,
			"duration": videoInfo.Duration.Seconds(),
			"bitrate":  videoInfo.Bitrate,
			"format":   videoInfo.Format,
		},
	}}, nil, nil
}
