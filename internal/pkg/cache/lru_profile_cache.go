package cache

import (
	"time"
)

// LRUProfileCache represents a fast in-memory LRU cache for user profiles
// This implementation uses the LRUCache for better memory management and performance
type LRUProfileCache struct {
	cache         *LRUCache
	ttl           time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// profileCacheItem represents a cached profile with expiration time
type profileCacheItem struct {
	profile    *Profile
	expiration time.Time
}

// NewLRUProfileCache creates a new LRU-based profile cache
// capacity - maximum number of profiles to store
// ttl - time-to-live for each profile
// cleanupInterval - how often to check for expired items
func NewLRUProfileCache(capacity int, ttl time.Duration, cleanupInterval time.Duration) *LRUProfileCache {
	cache := &LRUProfileCache{
		cache:         NewLRUCache(capacity),
		ttl:           ttl,
		cleanupTicker: time.NewTicker(cleanupInterval),
		stopCleanup:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.startCleanup()

	return cache
}

// startCleanup periodically cleans up expired cache items
func (c *LRUProfileCache) startCleanup() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.Cleanup()
		case <-c.stopCleanup:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// Stop terminates the cleanup goroutine
func (c *LRUProfileCache) Stop() {
	close(c.stopCleanup)
}

// Shutdown is an alias for Stop for interface consistency
func (c *LRUProfileCache) Shutdown() {
	c.Stop()
}

// Cleanup removes expired items from the cache
func (c *LRUProfileCache) Cleanup() {
	now := time.Now()
	keys := c.cache.Keys()

	for _, key := range keys {
		if item, found := c.cache.Get(key); found {
			if cacheItem, ok := item.(*profileCacheItem); ok {
				if cacheItem.expiration.Before(now) {
					c.cache.Remove(key)
				}
			}
		}
	}
}

// Set adds or updates a profile in the cache
func (c *LRUProfileCache) Set(profile *Profile) {
	item := &profileCacheItem{
		profile:    profile,
		expiration: time.Now().Add(c.ttl),
	}
	c.cache.Put(profile.UUID, item)
}

// Get retrieves a profile from the cache
func (c *LRUProfileCache) Get(uuid string) *Profile {
	item, exists := c.cache.Get(uuid)
	if !exists {
		return nil
	}

	cacheItem, ok := item.(*profileCacheItem)
	if !ok {
		return nil
	}

	// Check if the item has expired
	if time.Now().After(cacheItem.expiration) {
		c.cache.Remove(uuid)
		return nil
	}

	// Update expiration time and refresh position in LRU
	cacheItem.expiration = time.Now().Add(c.ttl)
	c.cache.Put(uuid, cacheItem)

	return cacheItem.profile
}

// Delete removes a profile from the cache
func (c *LRUProfileCache) Delete(uuid string) {
	c.cache.Remove(uuid)
}

// AddOrder adds an order to a user's profile and updates the cache
func (c *LRUProfileCache) AddOrder(userUUID string, order *Order) *Profile {
	item, exists := c.cache.Get(userUUID)
	if !exists {
		return nil
	}

	cacheItem, ok := item.(*profileCacheItem)
	if !ok || time.Now().After(cacheItem.expiration) {
		c.cache.Remove(userUUID)
		return nil
	}

	// Add order to profile
	cacheItem.profile.Orders = append(cacheItem.profile.Orders, order)

	// Reset expiration and refresh position in LRU
	cacheItem.expiration = time.Now().Add(c.ttl)
	c.cache.Put(userUUID, cacheItem)

	return cacheItem.profile
}

// UpdateOrder updates an order in a user's profile
func (c *LRUProfileCache) UpdateOrder(userUUID string, orderUUID string, newValue any) *Profile {
	item, exists := c.cache.Get(userUUID)
	if !exists {
		return nil
	}

	cacheItem, ok := item.(*profileCacheItem)
	if !ok || time.Now().After(cacheItem.expiration) {
		c.cache.Remove(userUUID)
		return nil
	}

	// Find and update the order
	for _, order := range cacheItem.profile.Orders {
		if order.UUID == orderUUID {
			order.Value = newValue
			order.UpdatedAt = time.Now()
			break
		}
	}

	// Reset expiration and refresh position in LRU
	cacheItem.expiration = time.Now().Add(c.ttl)
	c.cache.Put(userUUID, cacheItem)

	return cacheItem.profile
}

// RemoveOrder removes an order from a user's profile
func (c *LRUProfileCache) RemoveOrder(userUUID string, orderUUID string) *Profile {
	item, exists := c.cache.Get(userUUID)
	if !exists {
		return nil
	}

	cacheItem, ok := item.(*profileCacheItem)
	if !ok || time.Now().After(cacheItem.expiration) {
		c.cache.Remove(userUUID)
		return nil
	}

	// Remove order
	for i, order := range cacheItem.profile.Orders {
		if order.UUID == orderUUID {
			// Remove this order from the slice
			cacheItem.profile.Orders = append(cacheItem.profile.Orders[:i], cacheItem.profile.Orders[i+1:]...)
			break
		}
	}

	// Reset expiration and refresh position in LRU
	cacheItem.expiration = time.Now().Add(c.ttl)
	c.cache.Put(userUUID, cacheItem)

	return cacheItem.profile
}

// Size returns the current number of profiles in the cache
func (c *LRUProfileCache) Size() int {
	return c.cache.Size()
}

// Clear empties the entire cache
func (c *LRUProfileCache) Clear() {
	c.cache.Clear()
}

// GetMaxCapacity returns the maximum capacity of the cache
func (c *LRUProfileCache) GetMaxCapacity() int {
	return c.cache.Capacity()
}
