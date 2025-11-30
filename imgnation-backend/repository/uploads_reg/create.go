package uploadsReg

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoUploadsRegistry) CreateFileMetadata(
	ctx context.Context,
	metadata *types.NewFileMetadataOpts,
	uploadRecord *types.NewUploadRecordOpts,
) error {
	now := time.Now().UTC()
	// Create base metadata from options
	fileMeta := &types.FileMetadata{
		ID:            metadata.ID,
		Blake3sum:     metadata.Blake3sum,
		FileSizeBytes: metadata.FileSizeBytes,
		Variants:      metadata.Variants,
		CreatedAt:     now,
		Uploads:       []types.UploadRecord{},
		Reactions: map[string]int64{
			"view":    0,
			"like":    0,
			"dislike": 0,
		},
	}
	// Add initial upload record if provided
	if uploadRecord != nil {
		fileMeta.Uploads = []types.UploadRecord{{
			UserID:      uploadRecord.UserID,
			FileName:    uploadRecord.FileName,
			Access:      uploadRecord.Access,
			Tags:        uploadRecord.Tags,
			Description: uploadRecord.Description,
			UploadedAt:  now,
		}}
	}

	_, err := r.uploadsColl.InsertOne(ctx, fileMeta, options.InsertOne().
		SetBypassDocumentValidation(false).
		SetComment("CreateFileMetadata"),
	)

	// Handle errors
	switch {
	case mongo.IsDuplicateKeyError(err):
		return werr.Wrapf(xerr.ErrEntityExists,
			"file metadata with provided file hash or ID already exists. Hash: %s",
			base64.StdEncoding.EncodeToString(metadata.Blake3sum),
		)
	case errors.Is(err, context.DeadlineExceeded):
		return werr.Wrapf(xerr.ErrDeadlineExceeded, "database operation timed out")
	default:
		return err
	}
}

func (r *MongoUploadsRegistry) AddUploadRecord(
	ctx context.Context,
	fileID uuid.UUID,
	opts *types.NewUploadRecordOpts,
) error {
	exists, err := r.checkUploadExists(ctx, bson.M{
		"_id":             fileID,
		"uploads.user_id": opts.UserID,
	})
	if err != nil {
		return err
	}
	if exists {
		return werr.Wrapf(xerr.ErrEntityExists,
			"upload record for file ID %s from UserID %s already exists",
			fileID.String(), opts.UserID.String())
	}
	return r.addUploadRecord(ctx, bson.M{
		"_id":             fileID,
		"uploads.user_id": bson.M{"$ne": opts.UserID},
	}, opts)
}

// Returns xerr.ErrEntityNotFound for both "upload record already exists"
// since we search for files with hash without that user uploading them
// and "file wasn't found". But for app logic that's good enough error handling
func (r *MongoUploadsRegistry) AddUploadRecordByHash(
	ctx context.Context,
	hash []byte,
	opts *types.NewUploadRecordOpts,
) error {
	return r.addUploadRecord(ctx, bson.M{
		"hash":            hash,
		"uploads.user_id": bson.M{"$ne": opts.UserID},
	}, opts)
}

func (r *MongoUploadsRegistry) UpdateFileVariants(
	ctx context.Context,
	fileID uuid.UUID,
	addVariants []types.FileVariantInfo,
	removeVariants []string,
) error {
	// Build the update operations
	update := bson.M{}

	// Add variants if any are provided
	if len(addVariants) > 0 {
		update["$addToSet"] = bson.M{
			"variants": bson.M{
				"$each": addVariants,
			},
		}
	}
	// Remove variants if any are provided
	if len(removeVariants) > 0 {
		update["$pull"] = bson.M{
			"variants": bson.M{
				"name": bson.M{
					"$in": removeVariants,
				},
			},
		}
	}

	// If no operations to perform, return early
	if len(update) == 0 {
		return nil
	}

	// Execute the update operation
	result, err := r.uploadsColl.UpdateOne(
		ctx,
		bson.M{"_id": fileID},
		update,
		options.UpdateOne().
			SetBypassDocumentValidation(false).
			SetComment("UpdateFileVariants"),
	)

	// Handle errors
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return werr.Wrapf(xerr.ErrDeadlineExceeded, "database operation timed out")
	case result.MatchedCount == 0:
		return werr.Wrapf(xerr.ErrEntityNotFound, "file metadata not found for ID: %s", fileID.String())
	default:
		return werr.Wrapf(err, "failed to update file variants for file ID: %s", fileID.String())
	}
}
