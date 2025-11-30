package reactionsRepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type reactionKey struct {
	UploadID uuid.UUID
	UserID   uuid.UUID
	Reaction string
}

func (r reactionKey) String() string {
	return fmt.Sprintf("%s:%s:%s", r.UploadID.String(), r.UserID.String(), r.Reaction)
}

// ReactionCacheEntry stores both the current state and flush flag
type ReactionCacheEntry struct {
	CurrentState bool // Current state of the reaction (true = set, false = not set)
	ShouldFlush  bool // Whether to include in flush (true = odd toggles, false = even toggles)
}

// InMemoryReactionsCache implements ReactionsCash interface
type InMemoryReactionsCache struct {
	mu    sync.RWMutex
	cache map[reactionKey]ReactionCacheEntry
}

// NewInMemoryReactionsCache creates a new in-memory reactions cache
func NewInMemoryReactionsCache() *InMemoryReactionsCache {
	return &InMemoryReactionsCache{
		cache: make(map[reactionKey]ReactionCacheEntry),
	}
}

// ToggleReaction handles reaction toggling with caching logic
func (c *InMemoryReactionsCache) ToggleReaction(
	ctx context.Context,
	opts ReactionOpts,
	getCurrentState GetCurrentReactionState,
) (ToggleReactionResult, error) {
	key := reactionKey{
		UploadID: opts.UploadID,
		UserID:   opts.UserID,
		Reaction: opts.Reaction,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, existsInCache := c.cache[key]

	if existsInCache {
		// Toggle both the current state and the flush flag
		newEntry := ReactionCacheEntry{
			CurrentState: !entry.CurrentState,
			ShouldFlush:  !entry.ShouldFlush,
		}
		c.cache[key] = newEntry

		return ToggleReactionResult{
			UploadID: opts.UploadID,
			UserID:   opts.UserID,
			Reaction: opts.Reaction,
			IsSet:    newEntry.CurrentState,
		}, nil
	} else {
		// First time - get initial state from database (only time we hit DB)
		initialState, err := getCurrentState.GetReactionState(ctx, opts)
		if err != nil {
			return ToggleReactionResult{}, err
		}

		// Create cache entry with toggled state and set to flush
		newEntry := ReactionCacheEntry{
			CurrentState: !initialState,
			ShouldFlush:  true,
		}
		c.cache[key] = newEntry

		return ToggleReactionResult{
			UploadID: opts.UploadID,
			UserID:   opts.UserID,
			Reaction: opts.Reaction,
			IsSet:    newEntry.CurrentState,
		}, nil
	}
}

// FlushPendingReactions returns reactions that should be flushed and clears entire cache
func (c *InMemoryReactionsCache) FlushPendingReactions() ([]ReactionOpts, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) == 0 {
		return []ReactionOpts{}, nil
	}

	// Collect reactions that should be flushed (ShouldFlush = true)
	results := make([]ReactionOpts, 0)

	for key, entry := range c.cache {
		if entry.ShouldFlush {
			results = append(results, ReactionOpts{
				UploadID: key.UploadID,
				UserID:   key.UserID,
				Reaction: key.Reaction,
			})
		}
	}

	// Clear entire cache
	c.cache = make(map[reactionKey]ReactionCacheEntry)

	return results, nil
}

// GetPendingCount returns the number of reactions that will be flushed
func (c *InMemoryReactionsCache) GetPendingCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, entry := range c.cache {
		if entry.ShouldFlush {
			count++
		}
	}
	return count
}

// HasPendingReactions checks if there are any reactions that will be flushed
func (c *InMemoryReactionsCache) HasPendingReactions() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, entry := range c.cache {
		if entry.ShouldFlush {
			return true
		}
	}
	return false
}

// GetCurrentState returns the current cached state of a reaction (useful for debugging)
func (c *InMemoryReactionsCache) GetCurrentState(opts ReactionOpts) (bool, bool) {
	key := reactionKey{
		UploadID: opts.UploadID,
		UserID:   opts.UserID,
		Reaction: opts.Reaction,
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return false, false // state, exists
	}
	return entry.CurrentState, true
}
