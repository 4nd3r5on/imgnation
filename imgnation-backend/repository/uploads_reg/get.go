package uploadsReg

import (
	"context"
	"errors"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoUploadsRegistry) GetFileMetadata(ctx context.Context, id uuid.UUID) (*types.FileMetadata, error) {
	return r.getFileMetadata(ctx, bson.M{"_id": id})
}

func (r *MongoUploadsRegistry) GetFileMetadataByHash(ctx context.Context, hash []byte) (*types.FileMetadata, error) {
	return r.getFileMetadata(ctx, bson.M{"hash": hash})
}

func (r *MongoUploadsRegistry) getUploadRecords(ctx context.Context, filter bson.M) ([]types.UploadRecord, error) {
	cursor, err := r.uploadsColl.Find(ctx, filter, options.Find().
		SetComment("getUploadRecords"),
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return nil, werr.Wrapf(xerr.ErrDeadlineExceeded, "database operation timed out")
		default:
			return nil, err
		}
	}
	defer cursor.Close(ctx)

	var uploadRecords []types.UploadRecord

	for cursor.Next(ctx) {
		var fileMeta types.FileMetadata
		if err := cursor.Decode(&fileMeta); err != nil {
			return nil, werr.Wrapf(err, "failed to decode file metadata")
		}
		uploadRecords = append(uploadRecords, fileMeta.Uploads...)
	}

	if err := cursor.Err(); err != nil {
		return nil, werr.Wrapf(err, "cursor iteration failed")
	}

	return uploadRecords, nil
}

func (r *MongoUploadsRegistry) GetUploadRecords(ctx context.Context, uploadID uuid.UUID) ([]types.UploadRecord, error) {
	return r.getUploadRecords(ctx, bson.M{"_id": uploadID})
}
