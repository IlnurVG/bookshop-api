package cache

import (
	"sync"
	"time"
)

// Profile represents a user profile with orders
type Profile struct {
	UUID   string
	Name   string
	Orders []*Order
}

// Order represents a user order
type Order struct {
	UUID      string
	Value     any
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProfileCache represents a fast in-memory cache for user profiles
// Serves as an L1 cache before Redis
type ProfileCache struct {
	data          map[string]*cacheItem
	mu            sync.RWMutex
	ttl           time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// cacheItem represents a cached item with expiration time
type cacheItem struct {
	profile    *Profile
	expiration time.Time
}

// NewProfileCache creates a new profile cache with the specified TTL
func NewProfileCache(ttl time.Duration, cleanupInterval time.Duration) *ProfileCache {
	cache := &ProfileCache{
		data:          make(map[string]*cacheItem),
		ttl:           ttl,
		cleanupTicker: time.NewTicker(cleanupInterval),
		stopCleanup:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.startCleanup()

	return cache
}

// startCleanup periodically cleans up expired cache items
func (c *ProfileCache) startCleanup() {
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

// Stop stops the automatic cleanup process
func (c *ProfileCache) Stop() {
	close(c.stopCleanup)
}

// Cleanup removes expired items from the cache
func (c *ProfileCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if item.expiration.Before(now) {
			delete(c.data, key)
		}
	}
}

// Set adds or updates a profile in the cache
func (c *ProfileCache) Set(profile *Profile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[profile.UUID] = &cacheItem{
		profile:    profile,
		expiration: time.Now().Add(c.ttl),
	}
}

// Get retrieves a profile from the cache
func (c *ProfileCache) Get(uuid string) *Profile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[uuid]
	if !exists {
		return nil
	}

	// Check if the item has expired
	if time.Now().After(item.expiration) {
		return nil
	}

	// Update expiration time
	item.expiration = time.Now().Add(c.ttl)

	return item.profile
}

// Delete removes a profile from the cache
func (c *ProfileCache) Delete(uuid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, uuid)
}

// AddOrder adds an order to a user's profile and updates the cache
func (c *ProfileCache) AddOrder(userUUID string, order *Order) *Profile {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[userUUID]
	if !exists || time.Now().After(item.expiration) {
		return nil
	}

	// Add order to profile
	item.profile.Orders = append(item.profile.Orders, order)

	// Reset expiration
	item.expiration = time.Now().Add(c.ttl)

	return item.profile
}

// UpdateOrder updates an order in a user's profile
func (c *ProfileCache) UpdateOrder(userUUID string, orderUUID string, newValue any) *Profile {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[userUUID]
	if !exists || time.Now().After(item.expiration) {
		return nil
	}

	// Find and update the order
	for _, order := range item.profile.Orders {
		if order.UUID == orderUUID {
			order.Value = newValue
			order.UpdatedAt = time.Now()
			break
		}
	}

	// Reset expiration
	item.expiration = time.Now().Add(c.ttl)

	return item.profile
}

// RemoveOrder removes an order from a user's profile
func (c *ProfileCache) RemoveOrder(userUUID string, orderUUID string) *Profile {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[userUUID]
	if !exists || time.Now().After(item.expiration) {
		return nil
	}

	// Remove order
	for i, order := range item.profile.Orders {
		if order.UUID == orderUUID {
			// Remove this order from the slice
			item.profile.Orders = append(item.profile.Orders[:i], item.profile.Orders[i+1:]...)
			break
		}
	}

	// Reset expiration
	item.expiration = time.Now().Add(c.ttl)

	return item.profile
}
