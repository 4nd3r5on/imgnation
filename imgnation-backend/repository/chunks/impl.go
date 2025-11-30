package chunksRepo

import (
	"context"
	"fmt"

	"imgnation-backend/pkg/types"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ChunksRepository struct {
	collection *mongo.Collection
}

func NewChunksRepository(ctx context.Context, db *mongo.Database) (*ChunksRepository, error) {
	collection := db.Collection("chunks") // Fixed: was "users", should be "chunks"
	repo := &ChunksRepository{
		collection: collection,
	}
	err := repo.CreateIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunks repository indexes: %w", err)
	}
	return repo, nil
}

// CreateIndexes creates necessary indexes for the chunks collection
func (r *ChunksRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "file_id", Value: 1},
				{Key: "variant_name", Value: 1},
				{Key: "chunk_index", Value: 1},
			},
			Options: options.Index().SetName("file_variant_index_idx"),
		},
		{
			Keys:    bson.D{{Key: "file_id", Value: 1}},
			Options: options.Index().SetName("file_id_idx"),
		},
		{
			Keys: bson.D{
				{Key: "file_id", Value: 1},
				{Key: "variant_name", Value: 1},
			},
			Options: options.Index().SetName("file_variant_idx"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// AddChunks adds multiple chunks to the collection
func (r *ChunksRepository) AddChunks(ctx context.Context, chunks []types.ChunkInfo) error {
	if len(chunks) == 0 {
		return nil
	}

	docs := make([]any, len(chunks))
	for i, chunk := range chunks {
		docs[i] = chunk
	}
	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

// GetChunk retrieves a single chunk by ID
func (r *ChunksRepository) GetChunk(ctx context.Context, chunkID uuid.UUID) (*types.ChunkInfo, error) {
	var chunk types.ChunkInfo
	err := r.collection.FindOne(ctx, bson.M{"_id": chunkID}).Decode(&chunk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &chunk, nil
}

// GetFileChunks retrieves chunks for a file with optional variant and type filters
func (r *ChunksRepository) GetFileChunks(ctx context.Context, fileID uuid.UUID, variant *string, t *types.ChunkType) ([]types.ChunkInfo, error) {
	filter := bson.M{"file_id": fileID}

	if variant != nil {
		filter["variant_name"] = *variant
	}

	if t != nil {
		filter["chunk_type"] = *t
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "chunk_index", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chunks []types.ChunkInfo
	err = cursor.All(ctx, &chunks)
	return chunks, err
}

// GetChunkRange retrieves chunks within a specific index range
func (r *ChunksRepository) GetChunkRange(ctx context.Context, fileID uuid.UUID, variantName string, startIndex, endIndex int) ([]types.ChunkInfo, error) {
	filter := bson.M{
		"file_id":      fileID,
		"variant_name": variantName,
		"chunk_index": bson.M{
			"$gte": startIndex,
			"$lte": endIndex,
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "chunk_index", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chunks []types.ChunkInfo
	err = cursor.All(ctx, &chunks)
	return chunks, err
}

// CountFileChunks counts total chunks for a file
func (r *ChunksRepository) CountFileChunks(ctx context.Context, fileID uuid.UUID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"file_id": fileID})
}

// CountVariantChunks counts chunks for a specific file variant
func (r *ChunksRepository) CountVariantChunks(ctx context.Context, fileID uuid.UUID, variantName string) (int64, error) {
	filter := bson.M{
		"file_id":      fileID,
		"variant_name": variantName,
	}
	return r.collection.CountDocuments(ctx, filter)
}

// RemoveFileChunks removes all chunks for a file
func (r *ChunksRepository) RemoveFileChunks(ctx context.Context, fileID uuid.UUID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"file_id": fileID})
	return err
}

// RemoveVariantChunks removes chunks for a specific file variant
func (r *ChunksRepository) RemoveVariantChunks(ctx context.Context, fileID uuid.UUID, variantName string) error {
	filter := bson.M{
		"file_id":      fileID,
		"variant_name": variantName,
	}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

// RemoveAllChunks removes all chunks from the collection
func (r *ChunksRepository) RemoveAllChunks(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{})
	return err
}

// RemoveChunks removes multiple chunks by their IDs
func (r *ChunksRepository) RemoveChunks(ctx context.Context, chunkIDs []uuid.UUID) error {
	if len(chunkIDs) == 0 {
		return nil
	}

	filter := bson.M{"_id": bson.M{"$in": chunkIDs}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}
