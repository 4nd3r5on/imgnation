package reactionsRepo

import (
	"context"
	"errors"
	"time"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
)

// Reaction represents a user's reaction to an upload
type Reaction struct {
	ID        uuid.UUID `json:"id" bson:"_id,omitempty"`
	UploadID  uuid.UUID `json:"upload_id" bson:"upload_id"`
	UserID    uuid.UUID `json:"user_id" bson:"user_id"`
	Reaction  string    `json:"reaction" bson:"reaction"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type ReactionOpts struct {
	UploadID uuid.UUID `json:"upload_id" bson:"upload_id"`
	UserID   uuid.UUID `json:"user_id" bson:"user_id"`
	Reaction string    `json:"reaction" bson:"reaction"`
}

type ToggleReactionResult struct {
	UploadID uuid.UUID `json:"upload_id" bson:"upload_id"`
	UserID   uuid.UUID `json:"user_id" bson:"user_id"`
	Reaction string    `json:"reaction" bson:"reaction"`
	IsSet    bool      `json:"is_set" bson:"is_set"`
}

// UploadReactionSummary represents reaction counts for a specific upload
type UploadReactionSummary struct {
	UploadID uuid.UUID      `json:"upload_id" bson:"upload_id"`
	Counts   map[string]int `json:"reaction_counts" bson:"reaction_counts"`
}

// CachedReactionsRepo wraps the ReactionsRepo with caching
type CachedReactionsRepo struct {
	repo  ReactionsRepo
	cache ReactionsCache
}

// NewCachedReactionsRepo creates a new cached reactions repository
func NewCachedReactionsRepo(repo ReactionsRepo, cache ReactionsCache) *CachedReactionsRepo {
	return &CachedReactionsRepo{
		repo:  repo,
		cache: cache,
	}
}

// ToggleReactions uses cache for immediate response, actual DB operations happen during flush
func (c *CachedReactionsRepo) ToggleReactions(ctx context.Context, opts []ReactionOpts) ([]ToggleReactionResult, error) {
	results := make([]ToggleReactionResult, len(opts))

	for i, opt := range opts {
		result, err := c.cache.ToggleReaction(ctx, opt, c.repo)
		if err != nil {
			if errors.Is(err, xerr.ErrEntityNotFound) {
				continue
			}
			return nil, err
		}
		results[i] = result
	}

	return results, nil
}

// FlushToDatabase flushes all pending cache operations to the database
func (c *CachedReactionsRepo) FlushToDatabase(ctx context.Context) error {
	pendingOpts, err := c.cache.FlushPendingReactions()
	if err != nil {
		return err
	}

	if len(pendingOpts) == 0 {
		return nil // Nothing to flush
	}

	// Execute the actual database operations
	_, err = c.repo.ToggleReactions(ctx, pendingOpts)
	return err
}

type ReactionsRepo interface {
	GetReactionState(ctx context.Context, opts ReactionOpts) (isSet bool, err error)
	AddReaction(ctx context.Context, opts ReactionOpts) error
	ToggleReaction(ctx context.Context, opts ReactionOpts) (ToggleReactionResult, error)
	ToggleReactions(ctx context.Context, opts []ReactionOpts) ([]ToggleReactionResult, error)
	GetUserUploadReactions(ctx context.Context, uploadID, userID uuid.UUID) ([]string, error)
	GetUserReactions(ctx context.Context, userID uuid.UUID, paginationParams types.PaginationParams) ([]Reaction, error)
	GetUploadReactions(ctx context.Context, uploadID uuid.UUID) (UploadReactionSummary, error)
	GetUploadsReactions(ctx context.Context, uploadIDs []uuid.UUID) ([]UploadReactionSummary, error)
	RemoveAllUserReactions(ctx context.Context, userID uuid.UUID) error
	RemoveAllUploadReactions(ctx context.Context, uploadID uuid.UUID) error
	RemoveUserUploadReactions(ctx context.Context, uploadID, userID uuid.UUID) error
}

type GetCurrentReactionState interface {
	GetReactionState(ctx context.Context, opts ReactionOpts) (isSet bool, err error)
}

type ReactionsCache interface {
	ToggleReaction(
		ctx context.Context,
		opts ReactionOpts,
		getCurrentState GetCurrentReactionState,
	) (ToggleReactionResult, error)
	FlushPendingReactions() ([]ReactionOpts, error)
}
