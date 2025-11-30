package upload

import (
	"strings"

	"imgnation-backend/core/process"
	"imgnation-backend/core/process/image"
	"imgnation-backend/core/process/video"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/safeblock-dev/werr"
)

func InitProcessors() map[string]process.ProcessFunc {
	processors := make(map[string]process.ProcessFunc)
	for contentType := range video.SupportedVideoFormats {
		processors[contentType] = video.ProcessVideo
	}
	for contentType := range image.SupportedImageFormats {
		processors[contentType] = image.ProcessImage
	}
	return processors
}

func GetProcessFunc(
	allowAny bool,
	allowed map[string]struct{},
	processors map[string]process.ProcessFunc,
	opts *types.ProcessFileOpts,
) (process.ProcessFunc, error) {
	if opts.DetectedContentType == "" {
		return nil, werr.Wrapf(xerr.ErrInvalidArgument, "content type not detected")
	}
	processor, exists := processors[opts.DetectedContentType]
	if !exists {
		return nil, werr.Wrapf(xerr.ErrPermissionDenied, "no processor available for content type: %s", opts.DetectedContentType)
	}

	if !allowAny {
		if _, allowed := allowed[opts.DetectedContentType]; !allowed {
			return nil, werr.Wrapf(xerr.ErrPermissionDenied, "content type not allowed: %s", opts.DetectedContentType)
		}
	}
	return processor, nil
}

// GetSupportedTypesByCategory returns supported types by category
func GetSupportedProcessorsByCategory(processors map[string]process.ProcessFunc) map[string][]string {
	categories := map[string][]string{
		"image": {},
		"video": {},
	}
	for contentType := range processors {
		if strings.HasPrefix(contentType, "image/") {
			categories["image"] = append(categories["image"], contentType)
		} else if strings.HasPrefix(contentType, "video/") {
			categories["video"] = append(categories["video"], contentType)
		}
	}
	return categories
}
