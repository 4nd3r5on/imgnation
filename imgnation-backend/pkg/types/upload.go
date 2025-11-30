package types

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type NewFileOpts[MetadataT any] struct {
	AuthPayload  *AccessClaims
	UploadID     uuid.UUID
	FileMetadata MetadataT
	// Analyzed
	DetectedContentType string
	SizeBytes           int64
	Blake3Sum           []byte
	// might be "" if reader is just a buffer, check first
	FilePath string
	Reader   io.ReadSeeker
}

// Used to initialize VariantHandler
type VariantOpts struct {
	// Usually not used directly on the output (used size from reading)
	// But might be used for pre-allocating resources
	SizeBytes     int64
	ContentType   string
	Metadata      map[string]any
	StreamingInfo *StreamingInfo
	IsChunked     bool
}

type ProcessFileOpts struct {
	UploadID            uuid.UUID
	HeaderContentType   string
	DetectedContentType string
	SizeBytes           int64
	Blake3Sum           []byte
	// might be "" if reader is just a buffer, check before using
	// path to file (most likely temporary) related to the reader
	FilePath string
	Reader   io.ReadSeeker
}

type RepositoryHelpers[MetadataT any] interface {
	// CheckUploaded is used to check maybe file already was uploaded before we start processing it
	// in case of file by that hash being found -- returns it's ID
	// this functions also should record file being uploaded by itself if it's an intended behaviour
	CheckUploaded(
		ctx context.Context,
		opts *NewFileOpts[MetadataT],
	) (uuid.UUID, error)

	// RecordNewFileFunc is used to create a record in some DB about file being uploaded
	RecordNewFile(
		ctx context.Context,
		opts *NewFileOpts[MetadataT],
		variants []FileVariantInfo,
		chunks []ChunkInfo,
	) error

	AddVariants(
		ctx context.Context,
		uploadID uuid.UUID,
		opts []FileVariantInfo,
		chunks []ChunkInfo,
	) error
}
