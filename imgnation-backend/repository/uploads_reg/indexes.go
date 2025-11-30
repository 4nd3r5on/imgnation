package uploadsReg

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (m *MongoUploadsRegistry) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		// Unique file hash index
		{
			Keys: bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("hash_unique"),
		},
		// CreatedAt temporal index
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName("created_at_desc"),
		},
		// File variant name search index (case-insensitive)
		{
			Keys: bson.D{{Key: "variants.name", Value: 1}},
			Options: options.Index().
				SetCollation(&options.Collation{
					Locale:   "en",
					Strength: 2, // Case-insensitive
				}).
				SetName("variants_name_search"),
		},
		// User ID index for uploads
		{
			Keys: bson.D{{Key: "uploads.user_id", Value: 1}},
			Options: options.Index().
				SetName("user_uploads"),
		},
		// Upload filename and description text search index
		{
			Keys: bson.D{
				{Key: "uploads.filename", Value: "text"},
				{Key: "uploads.description", Value: "text"},
			},
			Options: options.Index().
				SetWeights(bson.M{
					"uploads.filename":    10,
					"uploads.description": 5,
				}).
				SetDefaultLanguage("english").
				SetName("filename_description_text_search"),
		},
		// Upload tags multikey index
		{
			Keys: bson.D{{Key: "uploads.tags", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"uploads.tags": bson.M{"$exists": true, "$type": "array"},
				}).
				SetName("tags_search"),
		},
		// Compound index for recent uploads with access level
		{
			Keys: bson.D{
				{Key: "uploads.uploaded_at", Value: -1},
				{Key: "uploads.access.level", Value: 1},
			},
			Options: options.Index().
				SetName("recent_public_uploads"),
		},
		// Compound index for user's recent uploads
		{
			Keys: bson.D{
				{Key: "uploads.user_id", Value: 1},
				{Key: "uploads.uploaded_at", Value: -1},
			},
			Options: options.Index().
				SetName("user_recent_uploads"),
		},
		// File size index
		{
			Keys: bson.D{{Key: "size", Value: 1}},
			Options: options.Index().
				SetName("file_size"),
		},
		// Status index for filtering
		{
			Keys: bson.D{{Key: "status", Value: 1}},
			Options: options.Index().
				SetName("status_filter"),
		},
		// Access level index (standalone)
		{
			Keys: bson.D{{Key: "uploads.access.level", Value: 1}},
			Options: options.Index().
				SetName("access_level"),
		},
		// Allowed users index for access control
		{
			Keys: bson.D{{Key: "uploads.access.allowed_users", Value: 1}},
			Options: options.Index().
				SetName("allowed_users"),
		},
		// === STATISTICS INDEXES ===
		// Users who liked files
		{
			Keys: bson.D{{Key: "statistics.users_like", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"statistics.users_like.0": bson.M{"$exists": true},
				}).
				SetName("users_liked"),
		},
		// Users who disliked files
		{
			Keys: bson.D{{Key: "statistics.users_dislike", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"statistics.users_dislike.0": bson.M{"$exists": true},
				}).
				SetName("users_disliked"),
		},
		// Users who viewed files
		{
			Keys: bson.D{{Key: "statistics.users_viewed", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"statistics.users_viewed.0": bson.M{"$exists": true},
				}).
				SetName("users_viewed"),
		},
		// === VARIANT CONTENT INDEXES ===
		// Variant content type
		{
			Keys: bson.D{{Key: "variants.content_type", Value: 1}},
			Options: options.Index().
				SetName("variants_content_type"),
		},
		// Chunked variants filter
		{
			Keys: bson.D{{Key: "variants.is_chunked", Value: 1}},
			Options: options.Index().
				SetName("variants_chunked"),
		},
		// Variant size
		{
			Keys: bson.D{{Key: "variants.size", Value: 1}},
			Options: options.Index().
				SetName("variants_size"),
		},
		// === STREAMING INDEXES ===
		// Streaming protocol
		{
			Keys: bson.D{{Key: "variants.streaming_info.protocol", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"variants.streaming_info": bson.M{"$exists": true, "$type": "object"},
				}).
				SetName("streaming_protocol"),
		},
		// Streaming duration
		{
			Keys: bson.D{{Key: "variants.streaming_info.duration_sec", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"variants.streaming_info.duration_sec": bson.M{
						"$exists": true,
						"$type":   "number",
					},
				}).
				SetName("streaming_duration"),
		},
		// Streaming bitrate
		{
			Keys: bson.D{{Key: "variants.streaming_info.bitrate", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"variants.streaming_info.bitrate": bson.M{
						"$exists": true,
						"$type":   "number",
					},
				}).
				SetName("streaming_bitrate"),
		},
		// === COMPOUND INDEXES FOR COMMON QUERIES ===
		// User + status compound
		{
			Keys: bson.D{
				{Key: "uploads.user_id", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().
				SetName("user_status"),
		},
		// Protocol + bitrate for streaming quality filtering
		{
			Keys: bson.D{
				{Key: "variants.streaming_info.protocol", Value: 1},
				{Key: "variants.streaming_info.bitrate", Value: -1},
			},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{
					"variants.streaming_info": bson.M{"$exists": true, "$type": "object"},
				}).
				SetName("streaming_protocol_bitrate"),
		},
		// Content type + size for variant filtering
		{
			Keys: bson.D{
				{Key: "variants.content_type", Value: 1},
				{Key: "variants.size", Value: -1},
			},
			Options: options.Index().
				SetName("variants_type_size"),
		},
		// Status + created_at for filtering recent files by status
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetName("status_recent"),
		},
	}

	// Create indexes with proper error handling
	createdIndexes, err := m.uploadsColl.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	// Log successful index creation
	log.Printf("Successfully created %d indexes for uploads collection", len(createdIndexes))
	return nil
}
