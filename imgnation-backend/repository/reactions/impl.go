package reactionsRepo

import (
	"context"
	"time"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoReactionsRepo struct {
	coll *mongo.Collection
}

func NewReactionsRepo(ctx context.Context, db *mongo.Database) (*MongoReactionsRepo, error) {
	collection := db.Collection("reactions")
	repo := &MongoReactionsRepo{
		coll: collection,
	}
	err := repo.CreateIndexes(ctx)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to create user repository indexes")
	}
	return repo, nil
}

// CreateIndexes creates the necessary indexes for the reactions collection
func (r *MongoReactionsRepo) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "upload_id", Value: 1},
				{Key: "user_id", Value: 1},
				{Key: "reaction", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "upload_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}

	_, err := r.coll.Indexes().CreateMany(ctx, indexes)
	return err
}

// GetReactionState checks if a specific reaction is set by a user for an upload
func (r *MongoReactionsRepo) GetReactionState(ctx context.Context, opts ReactionOpts) (isSet bool, err error) {
	filter := bson.M{
		"upload_id": opts.UploadID,
		"user_id":   opts.UserID,
		"reaction":  opts.Reaction,
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, werr.Wrapf(err, "failed to check reaction state")
	}

	return count > 0, nil
}

// AddReaction adds a new reaction (will fail if reaction already exists due to unique index)
func (r *MongoReactionsRepo) AddReaction(ctx context.Context, opts ReactionOpts) error {
	reaction := Reaction{
		ID:        uuid.New(),
		UploadID:  opts.UploadID,
		UserID:    opts.UserID,
		Reaction:  opts.Reaction,
		CreatedAt: time.Now(),
	}

	_, err := r.coll.InsertOne(ctx, reaction)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return werr.Wrapf(xerr.ErrEntityExists, "reaction exists")
		}
		return werr.Wrapf(err, "failed to add reaction")
	}

	return nil
}

// ToggleReactions toggles reactions for multiple uploads
func (r *MongoReactionsRepo) ToggleReactions(ctx context.Context, opts []ReactionOpts) ([]ToggleReactionResult, error) {
	results := make([]ToggleReactionResult, 0, len(opts))

	for _, opt := range opts {
		result, err := r.ToggleReaction(ctx, opt)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// toggleSingleReaction handles toggling a single reaction
func (r *MongoReactionsRepo) ToggleReaction(ctx context.Context, opt ReactionOpts) (ToggleReactionResult, error) {
	filter := bson.M{
		"upload_id": opt.UploadID,
		"user_id":   opt.UserID,
		"reaction":  opt.Reaction,
	}

	// Check if reaction exists
	var existing Reaction
	err := r.coll.FindOne(ctx, filter).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		// Reaction doesn't exist, create it
		reaction := Reaction{
			ID:        uuid.New(),
			UploadID:  opt.UploadID,
			UserID:    opt.UserID,
			Reaction:  opt.Reaction,
			CreatedAt: time.Now(),
		}

		_, err := r.coll.InsertOne(ctx, reaction)
		if err != nil {
			return ToggleReactionResult{}, err
		}

		return ToggleReactionResult{
			UploadID: opt.UploadID,
			UserID:   opt.UserID,
			Reaction: opt.Reaction,
			IsSet:    true,
		}, nil
	} else if err != nil {
		return ToggleReactionResult{}, err
	}

	// Reaction exists, remove it
	_, err = r.coll.DeleteOne(ctx, filter)
	if err != nil {
		return ToggleReactionResult{}, err
	}

	return ToggleReactionResult{
		UploadID: opt.UploadID,
		UserID:   opt.UserID,
		Reaction: opt.Reaction,
		IsSet:    false,
	}, nil
}

// GetUserUploadReactions gets all reactions a user has made to a specific upload
func (r *MongoReactionsRepo) GetUserUploadReactions(ctx context.Context, uploadID, userID uuid.UUID) ([]string, error) {
	filter := bson.M{
		"upload_id": uploadID,
		"user_id":   userID,
	}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reactions []Reaction
	if err := cursor.All(ctx, &reactions); err != nil {
		return nil, err
	}

	result := make([]string, len(reactions))
	for i, reaction := range reactions {
		result[i] = reaction.Reaction
	}

	return result, nil
}

// GetUserReactions gets all reactions made by a specific user with pagination
func (r *MongoReactionsRepo) GetUserReactions(ctx context.Context, userID uuid.UUID, paginationParams types.PaginationParams) ([]Reaction, error) {
	filter := bson.M{"user_id": userID}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(paginationParams.PageSize)).
		SetSkip(int64((paginationParams.Page - 1) * paginationParams.PageSize))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reactions []Reaction
	if err := cursor.All(ctx, &reactions); err != nil {
		return nil, err
	}

	return reactions, nil
}

// GetUploadReactions gets reaction summary for a specific upload
func (r *MongoReactionsRepo) GetUploadReactions(ctx context.Context, uploadID uuid.UUID) (UploadReactionSummary, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{"upload_id": uploadID},
		},
		{
			"$group": bson.M{
				"_id":   "$reaction",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return UploadReactionSummary{}, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return UploadReactionSummary{}, err
		}
		counts[result.ID] = result.Count
	}

	if err := cursor.Err(); err != nil {
		return UploadReactionSummary{}, err
	}

	return UploadReactionSummary{
		UploadID: uploadID,
		Counts:   counts,
	}, nil
}

// GetUploadsReactions gets reaction summaries for multiple uploads
func (r *MongoReactionsRepo) GetUploadsReactions(ctx context.Context, uploadIDs []uuid.UUID) ([]UploadReactionSummary, error) {
	if len(uploadIDs) == 0 {
		return []UploadReactionSummary{}, nil
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{"upload_id": bson.M{"$in": uploadIDs}},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"upload_id": "$upload_id",
					"reaction":  "$reaction",
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id.upload_id",
				"reactions": bson.M{
					"$push": bson.M{
						"reaction": "$_id.reaction",
						"count":    "$count",
					},
				},
			},
		},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	uploadReactionsMap := make(map[uuid.UUID]map[string]int)

	for cursor.Next(ctx) {
		var result struct {
			ID        uuid.UUID `bson:"_id"`
			Reactions []struct {
				Reaction string `bson:"reaction"`
				Count    int    `bson:"count"`
			} `bson:"reactions"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		counts := make(map[string]int)
		for _, r := range result.Reactions {
			counts[r.Reaction] = r.Count
		}
		uploadReactionsMap[result.ID] = counts
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Ensure all requested uploads are included in the result
	results := make([]UploadReactionSummary, len(uploadIDs))
	for i, uploadID := range uploadIDs {
		counts := uploadReactionsMap[uploadID]
		if counts == nil {
			counts = make(map[string]int)
		}
		results[i] = UploadReactionSummary{
			UploadID: uploadID,
			Counts:   counts,
		}
	}

	return results, nil
}

// RemoveAllUserReactions removes all reactions made by a specific user
func (r *MongoReactionsRepo) RemoveAllUserReactions(ctx context.Context, userID uuid.UUID) error {
	filter := bson.M{"user_id": userID}
	_, err := r.coll.DeleteMany(ctx, filter)
	return err
}

// RemoveAllUploadReactions removes all reactions for a specific upload
func (r *MongoReactionsRepo) RemoveAllUploadReactions(ctx context.Context, uploadID uuid.UUID) error {
	filter := bson.M{"upload_id": uploadID}
	_, err := r.coll.DeleteMany(ctx, filter)
	return err
}

// RemoveUserUploadReactions removes all reactions by a specific user for a specific upload
func (r *MongoReactionsRepo) RemoveUserUploadReactions(ctx context.Context, uploadID, userID uuid.UUID) error {
	filter := bson.M{
		"upload_id": uploadID,
		"user_id":   userID,
	}
	_, err := r.coll.DeleteMany(ctx, filter)
	return err
}
