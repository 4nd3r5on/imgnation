package upload

import (
	"context"
	"errors"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/types"
	"imgnation-backend/repository/storage"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

// Removes file-related uploads from storage
func CleanUp(
	ctx context.Context,
	s storage.IStorage,
	fileID uuid.UUID,
	variants []types.FileVariantInfo,
	chunks []types.ChunkInfo,
	keepVariants []string,
) error {
	keepVariantsMap := make(map[string]struct{})
	for _, variant := range keepVariants {
		keepVariantsMap[variant] = struct{}{}
	}

	var keysToDelete []string
	for _, variant := range variants {
		// Skip the original variant
		if _, next := keepVariantsMap[variant.Name]; next {
			continue
		}
		// Handle regular (non-chunked) files
		variantKey := process.MakeKeyFileVariant(fileID, variant.Name)
		keysToDelete = append(keysToDelete, variantKey)
		if variant.StreamingInfo != nil && variant.StreamingInfo.HasManifestFile {
			manifestKey := process.MakeKeyManifest(fileID, variant.Name)
			keysToDelete = append(keysToDelete, manifestKey)
		}
	}
	for _, chunk := range chunks {
		if _, next := keepVariantsMap[chunk.VariantName]; next {
			continue
		}
		chunkKey := process.MakeKeyChunk(fileID, chunk.VariantName, chunk.ChunkType, chunk.ChunkIndex)
		keysToDelete = append(keysToDelete, chunkKey)
	}
	if len(keysToDelete) > 0 {
		if errs := s.Delete(ctx, true, keysToDelete...); len(errs) != 0 {
			return werr.Wrapf(errors.Join(errs...), "failed to delete from storage")
		}
	}
	return nil
}
