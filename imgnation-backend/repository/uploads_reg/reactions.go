package uploadsReg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AddReactionCount atomically adds/subtracts a reaction count for an upload
// Returns the new count after the operation
func (r *MongoUploadsRegistry) AddReactionCount(ctx context.Context, uploadID uuid.UUID, reaction string, num int64) (newCount int64, err error) {
	filter := bson.M{"_id": uploadID}

	// Use $inc to atomically increment the reaction count
	// MongoDB's $inc operator automatically creates the field if it doesn't exist (starting from 0)
	update := bson.M{
		"$inc": bson.M{
			fmt.Sprintf("reactions.%s", reaction): num,
		},
	}

	// Use FindOneAndUpdate with ReturnDocument set to After to get the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result struct {
		Reactions map[string]int64 `bson:"reactions"`
	}

	err = r.uploadsColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("upload with ID %s not found", uploadID.String())
		}
		return 0, fmt.Errorf("failed to update reaction count: %w", err)
	}

	// Get the new count from the result
	newCount = result.Reactions[reaction]

	// Handle edge case where count becomes negative or zero
	if newCount <= 0 {
		// If count is zero or negative, remove the field entirely to keep the document clean
		if newCount <= 0 {
			unsetUpdate := bson.M{
				"$unset": bson.M{
					fmt.Sprintf("reactions.%s", reaction): "",
				},
			}

			// Don't return error if this fails, just log it
			// The count operation was successful, cleanup is optional
			_, unsetErr := r.uploadsColl.UpdateOne(ctx, filter, unsetUpdate)
			if unsetErr != nil {
				// Log error but don't fail the operation
				// You might want to use your logging framework here
				fmt.Printf("Warning: failed to cleanup zero reaction count for upload %s, reaction %s: %v\n",
					uploadID.String(), reaction, unsetErr)
			}

			// Return 0 for removed reactions
			if newCount < 0 {
				newCount = 0
			}
		}
	}

	return newCount, nil
}
