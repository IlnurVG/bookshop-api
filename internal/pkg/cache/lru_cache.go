package cache

import (
	"container/list"
	"sync"
)

// LRUCache represents the Least Recently Used cache structure
type LRUCache struct {
	capacity int                      // Maximum capacity of the cache
	cache    map[string]*list.Element // Hash map to store cache entries
	list     *list.List               // Doubly linked list to maintain order of usage
	mu       sync.RWMutex             // Mutex to ensure thread safety (using RWMutex for better performance)
}

// cacheEntry represents a key-value pair stored in the cache
type cacheEntry struct {
	key   string
	value interface{}
}

// NewLRUCache creates a new LRU cache with the given capacity
func NewLRUCache(capacity int) *LRUCache {
	// Protection against invalid capacity
	if capacity <= 0 {
		capacity = 100 // Set default value
	}

	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves a value from the cache by key.
// Returns the value and true if the key is found, otherwise nil and false.
// Thread-safe method.
func (lru *LRUCache) Get(key string) (interface{}, bool) {
	lru.mu.RLock()
	elem, exists := lru.cache[key]
	lru.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// If the key exists, move the element to the front of the list (most recently used)
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// Check again if the element exists (it could have been removed between unlocks)
	if elem, exists = lru.cache[key]; !exists {
		return nil, false
	}

	lru.list.MoveToFront(elem)
	return elem.Value.(*cacheEntry).value, true
}

// Put adds or updates a key-value pair in the cache.
// Thread-safe method.
func (lru *LRUCache) Put(key string, value interface{}) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// If the key already exists, update its value and move it to the front
	if elem, exists := lru.cache[key]; exists {
		lru.list.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	// If the cache is full, evict the least recently used item
	if lru.list.Len() >= lru.capacity {
		// Get the least recently used item (back of the list)
		lastElem := lru.list.Back()
		if lastElem != nil {
			// Remove the item from the cache and the list
			delete(lru.cache, lastElem.Value.(*cacheEntry).key)
			lru.list.Remove(lastElem)
		}
	}

	// Add the new item to the cache and the front of the list
	newEntry := &cacheEntry{key, value}
	newElem := lru.list.PushFront(newEntry)
	lru.cache[key] = newElem
}

// Remove deletes an item from the cache by key.
// Returns true if the item was removed, false if it was not found.
// Thread-safe method.
func (lru *LRUCache) Remove(key string) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, exists := lru.cache[key]; exists {
		lru.list.Remove(elem)
		delete(lru.cache, key)
		return true
	}
	return false
}

// Size returns the current size of the cache.
// Thread-safe method.
func (lru *LRUCache) Size() int {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	return lru.list.Len()
}

// Clear empties the entire cache.
// Thread-safe method.
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list = list.New()
	lru.cache = make(map[string]*list.Element)
}

// Contains checks if a key exists in the cache.
// Thread-safe method.
func (lru *LRUCache) Contains(key string) bool {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	_, exists := lru.cache[key]
	return exists
}

// Keys returns all keys stored in the cache.
// The order of the keys is arbitrary.
// Thread-safe method.
func (lru *LRUCache) Keys() []string {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	keys := make([]string, 0, len(lru.cache))
	for key := range lru.cache {
		keys = append(keys, key)
	}
	return keys
}

// Capacity returns the maximum capacity of the cache.
func (lru *LRUCache) Capacity() int {
	// Capacity is immutable after creation, so no locking needed
	return lru.capacity
}
