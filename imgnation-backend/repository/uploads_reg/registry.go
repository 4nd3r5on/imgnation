package uploadsReg

import (
	"context"

	"imgnation-backend/pkg/types"

	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoUploadsRegistry struct {
	types.UploadsReg
	uploadsColl *mongo.Collection
}

func NewMongoUploadsRegistry(ctx context.Context, db *mongo.Database, createIndexes bool) (*MongoUploadsRegistry, error) {
	reg := &MongoUploadsRegistry{uploadsColl: db.Collection("uploads")}

	if createIndexes {
		if err := reg.ensureIndexes(ctx); err != nil {
			return nil, werr.Wrapf(err, "failed to create indexes")
		}
	}
	return reg, nil
}
