package uploads

import (
	"context"
	"slices"
	"time"

	"imgnation-backend/config"
	"imgnation-backend/core/users"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

type UploadRecordPublic struct {
	UserID      uuid.UUID `json:"user_id"`
	FileName    string    `json:"filename"`
	Tags        []string  `json:"tags"`
	Description string    `json:"description"`
	AccessLevel string    `json:"access_level"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type FileMetadataPublic struct {
	ID            uuid.UUID               `json:"id"`
	FileSizeBytes int64                   `json:"size"`
	Variants      []types.FileVariantInfo `json:"variants"`
	Description   string                  `json:"description"`
	Uploads       []UploadRecordPublic    `json:"uploads"`
	AllTags       []string                `json:"all_tags"`
	CreatedAt     time.Time               `json:"created_at"`
	Reactions     map[string]int64        `json:"reactions"`
	YourReactions []string                `json:"your_reactions"`
}

func GetFileMetadata(
	ctx context.Context,
	repo *repository.Repo,
	cfg *config.ServerCfg,
	authData *types.AccessClaims,
	fileID uuid.UUID,
) (*FileMetadataPublic, error) {
	userID := uuid.Nil

	metadata, err := repo.UploadsReg.GetFileMetadata(ctx, fileID)
	if err != nil {
		return nil, err
	}
	reactions := []string{}
	if authData != nil {
		userID = authData.UserID
		reactions, err = repo.ReactionsRepo.GetUserUploadReactions(ctx, fileID, userID)
		if err != nil {
			return nil, err
		}
	}

	tagsMap := map[string]struct{}{}
	publicUploads := []UploadRecordPublic{}
	for _, upload := range metadata.Uploads {
		if (upload.Access.Level == types.AccessLevelAdmin &&
			!slices.Contains(authData.Roles, users.RoleAdmin)) ||
			(upload.Access.Level == types.AccessLevelPrivate &&
				!slices.Contains(upload.Access.AllowedUsers, userID)) {
			continue
		}
		publicUploads = append(publicUploads, UploadRecordPublic{
			UserID:      upload.UserID,
			FileName:    upload.FileName,
			Tags:        upload.Tags,
			Description: upload.Description,
			AccessLevel: types.AccessLevelToStrMap[upload.Access.Level],
			UploadedAt:  upload.UploadedAt,
		})
		for _, tag := range upload.Tags {
			tagsMap[tag] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		tags = append(tags, tag)
	}

	if len(publicUploads) == 0 {
		return nil, werr.Wrapf(xerr.ErrPermissionDenied, "cannot access this upload")
	}

	return &FileMetadataPublic{
		ID:            metadata.ID,
		FileSizeBytes: metadata.FileSizeBytes,
		AllTags:       tags,
		Variants:      metadata.Variants,
		Uploads:       publicUploads,
		Reactions:     metadata.Reactions,
		YourReactions: reactions,
		CreatedAt:     metadata.CreatedAt,
	}, nil
}
