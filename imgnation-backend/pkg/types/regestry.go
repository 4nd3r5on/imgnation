package types

import (
	"time"

	"github.com/google/uuid"
)

type (
	FileStatus        string
	ChunkType         string
	StreamingProtocol string
	AccessLevel       int8
)

const (
	VariantStatusFailed  FileStatus = "failed"
	VariantStatusSuccess FileStatus = "success"

	ChunkTypeInit    ChunkType = "init"
	ChunkTypeSegment ChunkType = "segment"

	ProtocolMPEGDash StreamingProtocol = "mpeg-dash"
	// Not supported yet
	ProtocolHLS  StreamingProtocol = "hls"
	ProtocolWebM StreamingProtocol = "webm"
)

const (
	// Don't even need to log in to see
	AccessLevelAnyone AccessLevel = iota
	// Just log in and u can see the file
	AccessLevelAuthorized
	// Must have given permission to see the file
	AccessLevelPrivate
	// Only admins can see
	AccessLevelAdmin
)

var AccessLevelMap = map[string]AccessLevel{
	"anyone":     AccessLevelAnyone,
	"authorized": AccessLevelAuthorized,
	"private":    AccessLevelPrivate,
	"admin":      AccessLevelAdmin,
}

var AccessLevelToStrMap = map[AccessLevel]string{
	AccessLevelAnyone:     "anyone",
	AccessLevelAuthorized: "authorized",
	AccessLevelPrivate:    "private",
	AccessLevelAdmin:      "admin",
}

// ==================================================
//                 CREATE/NEW TYPES
// ==================================================

// NewFileMetadataOpts options for creating new file metadata
type NewFileMetadataOpts struct {
	ID            uuid.UUID         `json:"id"`
	Blake3sum     []byte            `json:"hash"`
	FileSizeBytes int64             `json:"size"`
	Variants      []FileVariantInfo `json:"variants,omitempty"`
}

// NewUploadRecordOpts options for creating new upload record
type NewUploadRecordOpts struct {
	UserID      uuid.UUID     `json:"user_id"`
	FileName    string        `json:"filename"`
	Access      AccessControl `json:"access"`
	Tags        []string      `json:"tags,omitempty"`
	Description string        `json:"description,omitempty"`
}

// ==================================================
//                 UPDATE TYPES
// ==================================================

// PartialUploadRecordUpdate for updating upload records
type PartialUploadRecordUpdate struct {
	Access      *AccessControl `json:"access,omitempty"`
	Description *string        `json:"description,omitempty"`
	AddTags     []string       `json:"add_tags,omitempty"`
	RemoveTags  []string       `json:"remove_tags,omitempty"`
}

// ==================================================
//                 RESULT TYPES
// ==================================================

// UploadsSearchResults results from upload search
type UploadsSearchResults = SearchResult[FileMetadataPublic]

// FileMetadataPublic public view of file metadata (filtered for access control)
type FileMetadataPublic struct {
	ID            uuid.UUID         `bson:"_id" json:"id"`
	FileSizeBytes int64             `bson:"size" json:"size"`
	Variants      []FileVariantInfo `bson:"variants" json:"variants"`
	Uploads       []UploadRecord    `bson:"uploads" json:"uploads,omitempty"` // Only the upload record user has access to
	Reactions     map[string]int64  `bson:"reactions" json:"reactions"`
	CreatedAt     time.Time         `bson:"created_at" json:"created_at"`
}

// ==================================================
//                 STATISTICS TYPES
// ==================================================

// UserStatistics statistics for a specific user
type UserStatistics struct {
	UserID           uuid.UUID        `json:"user_id"`
	TotalUploads     int64            `json:"total_uploads"`
	TotalSize        int64            `json:"total_size"`
	PublicUploads    int64            `json:"public_uploads"`
	PrivateUploads   int64            `json:"private_uploads"`
	ContentTypeStats map[string]int64 `json:"content_type_stats"`
	TagStats         map[string]int64 `json:"tag_stats"`
	AverageFileSize  float64          `json:"average_file_size"`
	FirstUploadAt    *time.Time       `json:"first_upload_at,omitempty"`
	LastUploadAt     *time.Time       `json:"last_upload_at,omitempty"`
}

// ==================================================
//             STORED IN MONGO TYPES
// ==================================================

// StreamingInfo provides info related to streaming file
// manifest_<VariantName>_<FileID>
type StreamingInfo struct {
	Protocol            StreamingProtocol `bson:"protocol" json:"protocol"` // "mpeg-dash", "hls", etc.
	ManifestContentType string            `bson:"manifest_content_type,omitempty" json:"manifest_content_type,omitempty"`
	HasManifestFile     bool              `bson:"has_manifest_file" json:"has_manifest_file"`
	DurationSec         float64           `bson:"duration_sec,omitempty" json:"duration_sec,omitempty"`
	Bitrate             int               `bson:"bitrate,omitempty" json:"bitrate,omitempty"` // Average bitrate
}

// FileVariantInfo provides info on file variants created from upload saved in storage
// If IsChunked=false -- file should be stored by <file_id>_<variant_name>
// If IsChunked=true -- file chunks should be provided
type FileVariantInfo struct {
	// Regular data
	Name        string         `bson:"name" json:"name"`
	ContentType string         `bson:"content_type" json:"content_type"`
	SizeBytes   int64          `bson:"size" json:"size"`
	Metadata    map[string]any `bson:"metadata" json:"metadata"`
	// Chunked variant data
	IsChunked  bool `bson:"is_chunked" json:"is_chunked"`
	ChunkCount *int `bson:"chunk_count,omitempty" json:"chunk_count,omitempty"`
	// Chunks itself are stored in a separate collection for scalability reasons

	// Streaming data
	StreamingInfo *StreamingInfo `bson:"streaming_info,omitempty" json:"streaming_info,omitempty"`

	// Error handling
	Status FileStatus `bson:"status" json:"status"` // e.g., "success", "failed"
	Error  string     `bson:"fail_msg" json:"fail_msg,omitempty"`
}

type AccessControl struct {
	AllowedUsers []uuid.UUID `bson:"allowed_users,omitempty" json:"allowed_users,omitempty"`
	Level        AccessLevel `bson:"level" json:"level"`
}

// UploadRecord provides info about file uploads and embedded inside the file metadata
type UploadRecord struct {
	UserID      uuid.UUID     `bson:"user_id" json:"user_id"`
	FileName    string        `bson:"filename" json:"filename"`
	Access      AccessControl `bson:"access" json:"access"`
	UploadedAt  time.Time     `bson:"uploaded_at" json:"uploaded_at"`
	Tags        []string      `bson:"tags" json:"tags"`
	Description string        `bson:"description" json:"description"`
}

// ==================================================
//             TOP LEVEL MONGO DOCUMENTS
// ==================================================

type FileMetadata struct {
	ID            uuid.UUID         `bson:"_id" json:"id"`
	Blake3sum     []byte            `bson:"hash" json:"hash"` // of the original file sent to server to detect duplicates
	FileSizeBytes int64             `bson:"size" json:"size"`
	Variants      []FileVariantInfo `bson:"variants" json:"variants"`
	// uploads are embedded into metadata mostly because filename is taked from uploads
	// and there are many other usecases where we need to get those two at the same time
	// so it's easier to go with
	Uploads   []UploadRecord   `bson:"uploads" json:"uploads"`
	Reactions map[string]int64 `bson:"reactions" json:"reactions"`
	CreatedAt time.Time        `bson:"created_at" json:"created_at"`
}

// ChunksInfo is stored in a separate collection from other file metadata for scalability reasons
// Storage key format: <FileID>_<VariantName>_<ChunkType>_<ChunkIndex>
type ChunkInfo struct {
	ID          uuid.UUID `bson:"_id" json:"id"`
	FileID      uuid.UUID `bson:"file_id" json:"file_id"`
	VariantName string    `bson:"variant_name" json:"variant_name"`
	ChunkIndex  int64     `bson:"chunk_index" json:"chunk_index"` // 0-based index for ordering
	SizeBytes   int64     `bson:"size" json:"size"`
	ChunkType   ChunkType `bson:"chunk_type,omitempty" json:"chunk_type,omitempty"` // e.g., "manifest", "init", "segment", "data"
}
