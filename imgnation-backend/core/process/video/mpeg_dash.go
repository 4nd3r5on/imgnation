package video

import (
	"context"
	"fmt"
	"os"

	"imgnation-backend/pkg/types"
	videoUtils "imgnation-backend/pkg/utils/video"
	"imgnation-backend/pkg/xerr"

	"github.com/safeblock-dev/werr"
)

const (
	VariantNameMpegDash320p  = "mpeg_dash_320p"
	VariantNameMpegDash720p  = "mpeg_dash_720p"
	VariantNameMpegDash1080p = "mpeg_dash_1080p"
)

// TODO
func CreateVideoMpegDashCompressed(
	ctx context.Context,
	inputPath string,
	opts *types.ProcessFileOpts,
	variantOpts *VideoVariantOpts,
) error {
	// Create temporary output file
	outputExt := getFileExtensionFromContentType(variantOpts.OutputFormat)
	outputFile, err := os.CreateTemp("", fmt.Sprintf("compressed_%s_*%s", variantOpts.VariantName, outputExt))
	if err != nil {
		return werr.Wrapf(err, "failed to create temp compressed file")
	}
	defer os.Remove(outputFile.Name())
	defer outputFile.Close()

	// Build compression options
	compressOpts := videoUtils.CompressionOptions{
		OutputPath:   outputFile.Name(),
		Resolution:   variantOpts.Resolution,
		CRF:          variantOpts.CRF,
		MaxBitrate:   variantOpts.MaxBitrate,
		OutputFormat: getFormatFromContentType(variantOpts.OutputFormat),
	}

	// Compress video using ffmpeg
	if err := videoUtils.CompressVideo(ctx, inputPath, compressOpts); err != nil {
		return werr.Wrapf(err, "failed to compress video")
	}

	// Get file size
	_, err = outputFile.Stat()
	if err != nil {
		return werr.Wrapf(err, "failed to stat compressed file")
	}

	// Create reader from output file
	if _, err := outputFile.Seek(0, 0); err != nil {
		return werr.Wrapf(err, "failed to seek compressed file")
	}

	return werr.Wrapf(xerr.ErrNotImplemented, "Function CreateVideoMpegDashCompressed is not yet implemented")
}

// Helper functions
func getFileExtensionFromContentType(contentType string) string {
	switch contentType {
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "image/jpeg":
		return ".jpg"
	default:
		return ".tmp"
	}
}

func getFormatFromContentType(contentType string) string {
	switch contentType {
	case "video/mp4":
		return "mp4"
	case "video/quicktime":
		return "mov"
	case "video/x-msvideo":
		return "avi"
	default:
		return "mp4" // default fallback
	}
}
