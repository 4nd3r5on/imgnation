package upload

import (
	"context"
	"errors"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

// TODO: Fix

type RegistryHelpers struct {
	reg       types.UploadsReg
	chunksReg types.ChunksReg
}

// Checks if file already has been uploaded.
// If yes -- just adds an upload record from a user
func (rh *RegistryHelpers) CheckUploaded(
	ctx context.Context,
	opts *types.NewFileOpts[*UploadMetadata],
) (uploadID uuid.UUID, err error) {
	userID := uuid.Nil
	if opts.AuthPayload != nil {
		userID = opts.AuthPayload.UserID
	}
	metadata, err := rh.reg.GetFileMetadataByHash(ctx, opts.Blake3Sum)
	if err != nil {
		if errors.Is(err, xerr.ErrEntityNotFound) {
			return uuid.Nil, nil
		}
		return uuid.Nil, werr.Wrapf(err, "failed to get file metadata by hash")
	}
	err = rh.reg.AddUploadRecordByHash(ctx, opts.Blake3Sum, &types.NewUploadRecordOpts{
		UserID:   userID,
		FileName: opts.FileMetadata.FileName,
		Tags:     opts.FileMetadata.Tags,
		Access:   opts.FileMetadata.Access,
	})
	if err != nil && !errors.Is(err, xerr.ErrEntityExists) {
		return uuid.Nil, werr.Wrapf(err, "failed to add an upload record")
	}
	return metadata.ID, nil
}

func (rh *RegistryHelpers) RecordNewFile(
	ctx context.Context,
	opts *types.NewFileOpts[*UploadMetadata],
	variants []types.FileVariantInfo,
	chunks []types.ChunkInfo,
) (err error) {
	userID := uuid.Nil
	if opts.AuthPayload != nil {
		userID = opts.AuthPayload.UserID
	}
	err = rh.reg.CreateFileMetadata(ctx,
		&types.NewFileMetadataOpts{
			ID:            opts.UploadID,
			Blake3sum:     opts.Blake3Sum,
			FileSizeBytes: opts.SizeBytes,
			Variants:      variants,
		},
		&types.NewUploadRecordOpts{
			UserID:      userID,
			FileName:    opts.FileMetadata.FileName,
			Tags:        opts.FileMetadata.Tags,
			Access:      opts.FileMetadata.Access,
			Description: opts.FileMetadata.Description,
		},
	)
	if err != nil {
		return werr.Wrapf(err, "failed to create new file upload metadata")
	}

	if len(chunks) != 0 {
		err := rh.chunksReg.AddChunks(ctx, chunks)
		if err != nil {
			return werr.Wrapf(err, "failed to add chunks into repository")
		}
	}
	return nil
}

func (rh *RegistryHelpers) AddVariants(
	ctx context.Context,
	uploadID uuid.UUID,
	opts []types.FileVariantInfo,
	chunks []types.ChunkInfo,
) error {
	err := rh.reg.UpdateFileVariants(ctx, uploadID, opts, []string{})
	if err != nil {
		return werr.Wrapf(err, "failed to create new file upload variants")
	}
	if len(chunks) != 0 {
		err := rh.chunksReg.AddChunks(ctx, chunks)
		if err != nil {
			return werr.Wrapf(err, "failed to add chunks into repository")
		}
	}
	return nil
}

func NewRegistryHelpers(repo *repository.Repo) *RegistryHelpers {
	return &RegistryHelpers{
		reg:       repo.UploadsReg,
		chunksReg: repo.ChunksReg,
	}
}
