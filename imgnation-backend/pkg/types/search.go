package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	DefaultOrderAsc = false
)

type PaginationParams struct {
	Page     int
	PageSize int
	OrderAsc *bool // "asc" -- true / "desc" -- false. in query params. nil if not provided
	OrderBy  string
}

// Results structure with pagination info
type SearchResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	OrderAsc bool  `json:"order_asc"`
}

// ==================================================
//                 SEARCH/FILTER TYPES
// ==================================================

type SearchUploadsParams struct {
	UserID         uuid.UUID // Filter by specific user
	Tags           []string  // Match all of the tags. If no tags provided -- select any tags
	MinAccessLevel AccessLevel
	BeforeDate     *time.Time
	AfterDate      *time.Time
}

/*
// SearchUploadsParams parameters for searching uploads
type SearchUploadsParams struct {
	// Text search
	Query       string   `json:"query,omitempty"`        // Search in filename, description, tags
	Tags        []string `json:"tags,omitempty"`         // Filter by tags (AND operation)
	ContentType string   `json:"content_type,omitempty"` // Filter by content type

	// Size filters
	MinSize *int64 `json:"min_size,omitempty"`
	MaxSize *int64 `json:"max_size,omitempty"`

	// Time filters
	UploadedAfter  *time.Time `json:"uploaded_after,omitempty"`
	UploadedBefore *time.Time `json:"uploaded_before,omitempty"`

	// Access filters
	PublicOnly     bool `json:"public_only,omitempty"`
	IncludePrivate bool `json:"include_private,omitempty"` // Only for authorized users

	// Streaming filters
	Protocol        StreamingProtocol `json:"protocol,omitempty"`          // Filter by streaming protocol
	HasStreaming    *bool             `json:"has_streaming,omitempty"`     // Filter files with/without streaming
	IsChunked       *bool             `json:"is_chunked,omitempty"`        // Filter chunked/non-chunked files
	HasManifestFile *bool             `json:"has_manifest_file,omitempty"` // Filter by manifest presence

	// Duration filters (for streaming content)
	MinDuration *float64 `json:"min_duration,omitempty"`
	MaxDuration *float64 `json:"max_duration,omitempty"`

	// Bitrate filters
	MinBitrate *int `json:"min_bitrate,omitempty"`
	MaxBitrate *int `json:"max_bitrate,omitempty"`

	// User filters
	UserID  *uuid.UUID  `json:"user_id,omitempty"`  // Filter by uploader
	UserIDs []uuid.UUID `json:"user_ids,omitempty"` // Filter by multiple uploaders
	FileIDs []uuid.UUID `json:"file_ids,omitempty"` // Filter by specific file IDs

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // "uploaded_at", "size", "filename", "duration"
	SortOrder string `json:"sort_order,omitempty"` // "asc", "desc"
}
*/
