package image

import (
	"bytes"
	"context"
	"image"
	"io"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository/storage"

	"github.com/disintegration/imaging"
	"github.com/safeblock-dev/werr"
)

const ThumbnailMaxWidth = 256 // will maintain original ratio

const (
	VariantNameOriginal   = ""
	VariantNameThumbnail  = "thumbnail"
	VariantNameCompressed = "compressed"
)

var VariantFuncs = map[string]CreateVariantsFunc{
	VariantNameOriginal:   CreateOriginalVariant,
	VariantNameThumbnail:  CreateThumbnailVariant,
	VariantNameCompressed: CreateCompressedVariant,
}

func CreateOriginalVariant(
	ctx context.Context,
	s storage.IStorage,
	img image.Image,
	opts *types.ProcessFileOpts,
	variantOpts []ImageVariantOpts,
) ([]types.FileVariantInfo, []types.ChunkInfo, error) {
	if len(variantOpts) != 1 {
		return nil, nil, werr.Wrapf(
			xerr.ErrOutOfRange,
			"expected one variant argument. Got: %d", len(variantOpts),
		)
	}
	variant := variantOpts[0]

	_, err := opts.Reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to seek reader")
	}
	key := process.MakeKeyFileVariant(opts.UploadID, variant.VariantName)
	err = s.Store(ctx, storage.StoreObj{
		Key:         key,
		Size:        opts.SizeBytes,
		ContentType: variant.Format,
		Content:     opts.Reader,
	})
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to store object key %s", key)
	}

	bounds := img.Bounds()
	return []types.FileVariantInfo{{
		Name:        variant.VariantName,
		ContentType: variant.Format,
		SizeBytes:   opts.SizeBytes,
		Metadata: map[string]any{
			"height": bounds.Max.Y,
			"width":  bounds.Max.X,
		},
	}}, nil, nil
}

func CreateThumbnailVariant(
	ctx context.Context,
	s storage.IStorage,
	img image.Image,
	opts *types.ProcessFileOpts,
	variantOpts []ImageVariantOpts,
) ([]types.FileVariantInfo, []types.ChunkInfo, error) {
	if len(variantOpts) != 1 {
		return nil, nil, werr.Wrapf(
			xerr.ErrOutOfRange,
			"expected one variant argument. Got: %d", len(variantOpts),
		)
	}
	variant := variantOpts[0]

	thumbnailImg := imaging.Resize(img, ThumbnailMaxWidth, 0, imaging.Lanczos)

	buf, err := encodeImage(thumbnailImg, variant.Format, variant.Quality)
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to encode thumbnail")
	}

	thumbnailImgSize := int64(buf.Len())
	outputFile := bytes.NewReader(buf.Bytes())

	key := process.MakeKeyFileVariant(opts.UploadID, variant.VariantName)
	err = s.Store(ctx, storage.StoreObj{
		Key:         key,
		Size:        thumbnailImgSize,
		ContentType: variant.Format,
		Content:     outputFile,
	})
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to store object key %s", key)
	}

	thumbnailBounds := thumbnailImg.Bounds()
	return []types.FileVariantInfo{{
		Name:        variant.VariantName,
		ContentType: variant.Format,
		SizeBytes:   thumbnailImgSize,
		Metadata: map[string]any{
			"height": thumbnailBounds.Max.Y,
			"width":  thumbnailBounds.Max.X,
		},
	}}, nil, nil
}

func CreateCompressedVariant(
	ctx context.Context,
	s storage.IStorage,
	img image.Image,
	opts *types.ProcessFileOpts,
	variantOpts []ImageVariantOpts,
) ([]types.FileVariantInfo, []types.ChunkInfo, error) {
	if len(variantOpts) != 1 {
		return nil, nil, werr.Wrapf(
			xerr.ErrOutOfRange,
			"expected one variant argument. Got: %d", len(variantOpts),
		)
	}
	variant := variantOpts[0]

	buf, err := encodeImage(img, variant.Format, variant.Quality)
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to encode thumbnail")
	}
	thumbnailImgSize := int64(buf.Len())
	outputFile := bytes.NewReader(buf.Bytes())

	key := process.MakeKeyFileVariant(opts.UploadID, variant.VariantName)
	err = s.Store(ctx, storage.StoreObj{
		Key:         key,
		Size:        thumbnailImgSize,
		ContentType: variant.Format,
		Content:     outputFile,
	})
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to store object key %s", key)
	}

	bounds := img.Bounds()
	return []types.FileVariantInfo{{
		Name:        variant.VariantName,
		ContentType: variant.Format,
		SizeBytes:   thumbnailImgSize,
		Metadata: map[string]any{
			"height": bounds.Max.Y,
			"width":  bounds.Max.X,
		},
	}}, nil, nil
}
