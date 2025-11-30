package uploadsReg

import (
	"context"
	"errors"
	"maps"
	"time"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoUploadsRegistry) checkUploadExists(ctx context.Context, filter bson.M) (exists bool, err error) {
	count, err := r.uploadsColl.CountDocuments(ctx, filter)
	if err != nil {
		return false, werr.Wrapf(err, "failed to check for duplicate upload")
	}
	return count > 0, nil
}

func (r *MongoUploadsRegistry) addUploadRecord(
	ctx context.Context,
	filter bson.M,
	opts *types.NewUploadRecordOpts,
) error {
	uploadedAt := time.Now().UTC()

	duplicateCheckFilter := bson.M{"uploads.user_id": opts.UserID}
	maps.Copy(duplicateCheckFilter, filter)

	newUploadRecord := types.UploadRecord{
		UserID:      opts.UserID,
		FileName:    opts.FileName,
		Access:      opts.Access,
		Tags:        opts.Tags,
		Description: opts.Description,
		UploadedAt:  uploadedAt,
	}

	update := bson.M{"$push": bson.M{"uploads": newUploadRecord}}

	res, err := r.uploadsColl.UpdateOne(
		ctx, filter, update,
		options.UpdateOne().SetUpsert(false),
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return werr.Wrapf(xerr.ErrDeadlineExceeded,
				"database operation timed out")
		}
		return werr.Wrapf(err, "failed to add upload record")
	}

	if res.ModifiedCount < 1 {
		return werr.Wrapf(xerr.ErrEntityNotFound, "file wasn't found")
	}
	return nil
}

func (r *MongoUploadsRegistry) getFileMetadata(ctx context.Context, filter bson.M) (*types.FileMetadata, error) {
	var fileMeta types.FileMetadata
	err := r.uploadsColl.FindOne(ctx, filter, options.FindOne().
		SetComment("GetFileMetadata"),
	).Decode(&fileMeta)

	// Handle errors
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return nil, werr.Wrapf(xerr.ErrEntityNotFound,
			"file metadata not found with provided filter")
	case errors.Is(err, context.DeadlineExceeded):
		return nil, werr.Wrapf(xerr.ErrDeadlineExceeded, "database operation timed out")
	case err != nil:
		return nil, err
	}
	return &fileMeta, nil
}

// buildFacetFilter creates the facet stage for pagination
func buildFacetFilter(paginationParams *types.PaginationParams) bson.D {
	// Sort stage
	sortOrder := 1 // ascending
	if paginationParams.OrderAsc != nil && !*paginationParams.OrderAsc {
		sortOrder = -1 // descending
	}

	return bson.D{
		{"$facet", bson.M{
			"data": []bson.M{
				{"$sort": bson.M{"created_at": sortOrder}},
				{"$skip": int64((paginationParams.Page - 1) * paginationParams.PageSize)},
				{"$limit": int64(paginationParams.PageSize)},
			},
			"total": []bson.M{
				{"$count": "count"},
			},
		}},
	}
}

func buildFileReadAccessFilter(userID uuid.UUID) bson.D {
	orConditions := []bson.M{
		{"uploads.access.level": types.AccessLevelAnyone},
	}
	if userID != uuid.Nil {
		userConditions := []bson.M{
			{"uploads.access.level": types.AccessLevelAuthorized},
			{"uploads.user_id": userID},
			{
				"$and": []bson.M{
					{"uploads.access.level": types.AccessLevelPrivate},
					{"uploads.access.allowed_users": userID},
				},
			},
		}
		orConditions = append(orConditions, userConditions...)
	}
	return bson.D{{Key: "$or", Value: orConditions}}
}
