package types

import (
	"context"

	"github.com/google/uuid"
)

// UploadsReg stands for Uploads Registry
// Provides functionality for storing, searching and removing information about files uploads.
type UploadsReg interface {
	// FILE METADATA OPERATIONS

	CreateFileMetadata(
		ctx context.Context,
		metadata *NewFileMetadataOpts,
		uploadRecord *NewUploadRecordOpts,
	) error
	GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*FileMetadata, error)
	GetFileMetadataByHash(ctx context.Context, hash []byte) (*FileMetadata, error)
	GetFileMetadataPublic(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*FileMetadataPublic, error)
	DeleteFileMetadata(ctx context.Context, fileID uuid.UUID) error

	// FILE VARIANTS OPERATIONS

	// UpdateFileVariants updates file metadata by adding/removing file variants
	// addVariants provides info about added file variants.
	// removeVariants provides array of variant names to be removed.
	// If name was not found -- no errors will be returned, variant name will be just skipped
	UpdateFileVariants(
		ctx context.Context,
		fileID uuid.UUID,
		addVariants []FileVariantInfo,
		removeVariants []string,
	) error
	GetFileVariants(ctx context.Context, fileID uuid.UUID) ([]FileVariantInfo, error)
	GetFileVariant(ctx context.Context, fileID uuid.UUID, variantName string) (*FileVariantInfo, error)
	DeleteFileVariant(ctx context.Context, fileID uuid.UUID, variantName string) error

	// UPLOAD RECORDS OPERATIONS

	// AddUploadRecord adds upload record to file metadata
	// if record from a userID exists -- returns xerr.EntityAlreadyExists
	AddUploadRecord(ctx context.Context, fileID uuid.UUID, opts *NewUploadRecordOpts) error

	// AddUploadRecordByHash is used when file already has been uploaded and we just need to say
	// that other user also uploaded something with the same hash
	// if record from a userID exists -- returns xerr.EntityAlreadyExists
	AddUploadRecordByHash(ctx context.Context, hash []byte, opts *NewUploadRecordOpts) error

	GetUploadRecord(ctx context.Context, userID, fileID uuid.UUID) (*UploadRecord, error)
	GetUploadRecords(ctx context.Context, fileID uuid.UUID) ([]UploadRecord, error)
	GetUserUploadRecords(ctx context.Context, userID uuid.UUID, paginationParams *PaginationParams) ([]UploadRecord, error)
	UpdateUploadRecord(ctx context.Context, userID, fileID uuid.UUID, update *PartialUploadRecordUpdate) error
	BatchUpdateUploadRecords(ctx context.Context, update *PartialUploadRecordUpdate) error
	DeleteUploadRecords(ctx context.Context, userID uuid.UUID, fileID ...uuid.UUID) error
	DeleteAllUserUploadRecords(ctx context.Context, userID uuid.UUID) error

	// ==================================================
	//                SEARCH OPERATIONS
	// ==================================================

	SearchUploads(
		ctx context.Context,
		searchParams *SearchUploadsParams,
		paginationParams *PaginationParams,
		userID uuid.UUID, // User making request
		privacyFilter bool,
	) (*UploadsSearchResults, error)

	// ==================================================
	//                STATISTICS OPERATIONS
	// ==================================================

	AddReactionCount(ctx context.Context, uploadID uuid.UUID, reaction string, num int64) (newCount int64, err error)
	// GetUserStatistics retrieves statistics for a specific user
	GetUserStatistics(ctx context.Context, userID uuid.UUID) (*UserStatistics, error)

	// TODO
	// ==================================================
	//                INTEGRITY & VALIDATION
	// ==================================================

	// ValidateFileIntegrity checks file integrity
	// ValidateFileIntegrity(ctx context.Context, check *FileIntegrityCheck) (*IntegrityCheckResult, error)

	// ValidateUploadRecord validates upload record data
	// ValidateUploadRecord(ctx context.Context, record *NewUploadRecordOpts) (*ValidationResult, error)
}

// ChunksReg provides functionality for managing file chunks
type ChunksReg interface {
	// CHUNK OPERATIONS

	AddChunks(ctx context.Context, chunks []ChunkInfo) error
	GetChunk(ctx context.Context, chunkID uuid.UUID) (*ChunkInfo, error)
	// GetFileChunks retrieves chunks for a file variant
	// if nil -- ignore filter
	GetFileChunks(ctx context.Context, fileID uuid.UUID, variant *string, t *ChunkType) ([]ChunkInfo, error)
	// GetChunkRange retrieves chunks within a specific index range
	GetChunkRange(ctx context.Context, fileID uuid.UUID, variantName string, startIndex, endIndex int) ([]ChunkInfo, error)
	CountFileChunks(ctx context.Context, fileID uuid.UUID) (int64, error)
	// CountVariantChunks counts chunks for a specific file variant
	CountVariantChunks(ctx context.Context, fileID uuid.UUID, variantName string) (int64, error)

	// CHUNK REMOVAL

	RemoveFileChunks(ctx context.Context, fileID uuid.UUID) error
	RemoveVariantChunks(ctx context.Context, fileID uuid.UUID, variantName string) error
	RemoveAllChunks(ctx context.Context) error
	// GetChunkGaps(ctx context.Context, fileID uuid.UUID, variantName string) ([]int, error)

	// MAINTENANCE OPERATIONS

	// RemoveChunks removes multiple chunks by IDs
	RemoveChunks(ctx context.Context, chunkIDs []uuid.UUID) error
}
