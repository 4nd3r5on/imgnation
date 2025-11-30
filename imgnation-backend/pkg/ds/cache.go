package ds

import (
	"context"
	"sync"
	"time"
)

// ICache defines generic cache behaviour
// Push uses provided key; values expire after ttl
// WaitRemoveExpired blocks until next expiry (or ctx done)
type ICache[K comparable, V any] interface {
	Push(key K, val V) (ok bool)
	Set(key K, val V)
	Pop(key K) (val V, exists bool)
	Get(key K) (val V, exists bool)
	GetPurge(key K) (val V, exists bool)
	Delete(key K) (exists bool)

	TimeUntilNextExpired() (time.Duration, K)
	WaitRemoveExpired(ctx context.Context) error
}

// CacheItem holds a value and its expiration
// no index needed for lazy expiration
type CacheItem[K comparable, V any] struct {
	Key        K
	Value      V
	Expiration time.Time
}

// Cache implements ICache with lazy expiration and fixed TTL
type Cache[K comparable, V any] struct {
	mu          *sync.Mutex
	newItemCond *sync.Cond
	items       map[K]*CacheItem[K, V]
	heap        []*keyExpirationPair[K]
	ttl         time.Duration
}

type keyExpirationPair[K comparable] struct {
	Key        K
	Expiration time.Time
}

// NewCache returns a Cache with entries expiring after ttl
func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	mu := &sync.Mutex{}
	c := &Cache[K, V]{
		mu:          mu,
		newItemCond: sync.NewCond(mu),
		items:       make(map[K]*CacheItem[K, V]),
		heap:        []*keyExpirationPair[K]{},
		ttl:         ttl,
	}
	return c
}

// Push inserts key=value into cache with TTL
func (c *Cache[K, V]) Push(key K, val V) (ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.items[key]; exists {
		return false
	}
	exp := time.Now().Add(c.ttl)
	item := &CacheItem[K, V]{Key: key, Value: val, Expiration: exp}
	c.items[key] = item
	c.heap = append(c.heap, &keyExpirationPair[K]{
		Key:        key,
		Expiration: exp,
	})
	c.newItemCond.Signal()
	return true
}

// Set inserts or updates key, resetting TTL
// old entries in heap are ignored on expiration
func (c *Cache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp := time.Now().Add(c.ttl)
	item := &CacheItem[K, V]{Key: key, Value: val, Expiration: exp}
	c.items[key] = item
	c.heap = append(c.heap, &keyExpirationPair[K]{
		Key:        key,
		Expiration: exp,
	})
	c.newItemCond.Signal()
}

// purgeExpired removes top expired or stale items
// must hold c.mu
// returns next element to expire
func (c *Cache[K, V]) purgeExpired() (key K, expireAt time.Time, exists bool) {
	for len(c.heap) > 0 {
		top := c.heap[0]
		if time.Until(top.Expiration) <= 0 {
			c.heap = c.heap[1:]
		}
		stored, ok := c.items[top.Key]
		if !ok || time.Until(stored.Expiration) <= 0 {
			if ok {
				delete(c.items, top.Key)
			}
			continue
		}
		return stored.Key, stored.Expiration, true
	}
	var zero K
	return zero, time.Now(), false
}

// LockPurgeExpired removes top expired or stale items
// returns next element to expire
func (c *Cache[K, V]) LockPurgeExpired() (key K, expireAt time.Time, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.purgeExpired()
}

// Get returns value if exists
// Unlike GetPurge might return expired items
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if it, ok := c.items[key]; ok {
		return it.Value, true
	}
	var zero V
	return zero, false
}

// Get returns value if exists and not expired
func (c *Cache[K, V]) GetPurge(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.purgeExpired()
	if it, ok := c.items[key]; ok {
		return it.Value, true
	}
	var zero V
	return zero, false
}

// Pop returns and deletes the item
func (c *Cache[K, V]) Pop(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if it, ok := c.items[key]; ok {
		val := it.Value
		delete(c.items, key)
		return val, true
	}
	var zero V
	return zero, false
}

// Delete removes item if present
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
		return true
	}
	return false
}

// TimeUntilNextExpired gives time until soonest expiration and its key
func (c *Cache[K, V]) TimeUntilNextExpired() (time.Duration, K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key, expireAt, exists := c.purgeExpired()
	if !exists {
		return -1, key
	}
	return time.Until(expireAt), key
}

// WaitRemoveExpired blocks until next expiration or ctx done
func (c *Cache[K, V]) WaitRemoveExpired(ctx context.Context) error {
	c.mu.Lock()
	for {
		key, expireAt, exists := c.purgeExpired()
		if !exists {
			c.newItemCond.Wait()
			continue
		}
		waitUntil := time.Until(expireAt)
		c.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitUntil):
			if item, ok := c.items[key]; ok && time.Until(item.Expiration) <= 0 {
				c.Delete(key)
			}
			return nil
		}
	}
}

func (c *Cache[K, V]) RunAutoRemoveExpired(ctx context.Context) (err error) {
	for err == nil {
		err = c.WaitRemoveExpired(ctx)
	}
	return err
}
