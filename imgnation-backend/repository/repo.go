package repository

import (
	"context"

	"imgnation-backend/pkg/types"
	"imgnation-backend/repository/storage"

	chunksRepo "imgnation-backend/repository/chunks"
	reactionsRepo "imgnation-backend/repository/reactions"

	uploadsReg "imgnation-backend/repository/uploads_reg"
	usersRepo "imgnation-backend/repository/users"

	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const UPLODAS_BUCKET = "uploads"

type Repo struct {
	UsersRepo      usersRepo.UsersRepo
	UploadsReg     types.UploadsReg
	ChunksReg      types.ChunksReg
	UploadsStorage storage.IStorage
	ReactionsRepo  reactionsRepo.ReactionsRepo
	ReactionsCache reactionsRepo.ReactionsCache
}

func NewRepo(ctx context.Context, mongodb *mongo.Database, minio *minio.Client) (*Repo, error) {
	imgStorage := storage.NewStorage(minio, UPLODAS_BUCKET)
	err := imgStorage.SetupBucket(ctx)
	if err != nil {
		return nil, err
	}
	uploadsReg, err := uploadsReg.NewMongoUploadsRegistry(ctx, mongodb, true)
	if err != nil {
		return nil, err
	}
	users, err := usersRepo.NewUsersRepository(ctx, mongodb)
	if err != nil {
		return nil, err
	}
	chunksReg, err := chunksRepo.NewChunksRepository(ctx, mongodb)
	if err != nil {
		return nil, err
	}
	reactions, err := reactionsRepo.NewReactionsRepo(ctx, mongodb)
	if err != nil {
		return nil, err
	}
	reactionsCache := reactionsRepo.NewInMemoryReactionsCache()
	return &Repo{
		UploadsReg:     uploadsReg,
		UsersRepo:      users,
		UploadsStorage: imgStorage,
		ChunksReg:      chunksReg,
		ReactionsRepo:  reactions,
		ReactionsCache: reactionsCache,
	}, nil
}
