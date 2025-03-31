package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bookshop/api/pkg/logger"
)

// Book represents a book model for caching
type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	LastUpdated time.Time `json:"last_updated"`
}

// BookCache provides caching functionality for books
type BookCache struct {
	cache       *LRUCache
	ttl         time.Duration
	logger      logger.Logger
	cleanupTick *time.Ticker
	stopChan    chan struct{}
}

// bookCacheItem represents a cached book with expiration time
type bookCacheItem struct {
	book       *Book
	expiration time.Time
}

// NewBookCache creates a new book cache
// capacity - maximum number of books to cache
// ttl - time-to-live for each book entry
// cleanupInterval - how often to check for expired entries
func NewBookCache(capacity int, ttl time.Duration, cleanupInterval time.Duration, logger logger.Logger) *BookCache {
	cache := &BookCache{
		cache:       NewLRUCache(capacity),
		ttl:         ttl,
		logger:      logger,
		cleanupTick: time.NewTicker(cleanupInterval),
		stopChan:    make(chan struct{}),
	}

	// Start cleanup routine
	go cache.startCleanup()

	return cache
}

// startCleanup periodically cleans up expired cache entries
func (c *BookCache) startCleanup() {
	for {
		select {
		case <-c.cleanupTick.C:
			c.cleanup()
		case <-c.stopChan:
			c.cleanupTick.Stop()
			return
		}
	}
}

// cleanup removes expired entries from the cache
func (c *BookCache) cleanup() {
	now := time.Now()
	keys := c.cache.Keys()

	evicted := 0
	for _, key := range keys {
		if item, found := c.cache.Get(key); found {
			if cacheItem, ok := item.(*bookCacheItem); ok {
				if cacheItem.expiration.Before(now) {
					c.cache.Remove(key)
					evicted++
				}
			}
		}
	}

	if evicted > 0 {
		c.logger.Debug("Cleaned up expired book cache entries", "evicted", evicted)
	}
}

// Get retrieves a book from the cache
func (c *BookCache) Get(bookID string) (*Book, bool) {
	item, exists := c.cache.Get(bookID)
	if !exists {
		return nil, false
	}

	cacheItem, ok := item.(*bookCacheItem)
	if !ok {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cacheItem.expiration) {
		c.cache.Remove(bookID)
		return nil, false
	}

	// Update expiration and refresh position in LRU
	cacheItem.expiration = time.Now().Add(c.ttl)
	c.cache.Put(bookID, cacheItem)

	return cacheItem.book, true
}

// Set adds or updates a book in the cache
func (c *BookCache) Set(bookID string, book *Book) {
	// Update the last updated timestamp
	book.LastUpdated = time.Now()

	item := &bookCacheItem{
		book:       book,
		expiration: time.Now().Add(c.ttl),
	}
	c.cache.Put(bookID, item)
}

// Delete removes a book from the cache
func (c *BookCache) Delete(bookID string) {
	c.cache.Remove(bookID)
}

// GetMultiple retrieves multiple books from the cache
func (c *BookCache) GetMultiple(bookIDs []string) map[string]*Book {
	result := make(map[string]*Book)

	for _, id := range bookIDs {
		if book, found := c.Get(id); found {
			result[id] = book
		}
	}

	return result
}

// SetMultiple adds or updates multiple books in the cache
func (c *BookCache) SetMultiple(books map[string]*Book) {
	for id, book := range books {
		c.Set(id, book)
	}
}

// DeleteMultiple removes multiple books from the cache
func (c *BookCache) DeleteMultiple(bookIDs []string) {
	for _, id := range bookIDs {
		c.Delete(id)
	}
}

// UpdateQuantity updates a book's quantity in the cache if it exists
func (c *BookCache) UpdateQuantity(bookID string, newQuantity int) bool {
	book, found := c.Get(bookID)
	if !found {
		return false
	}

	book.Quantity = newQuantity
	book.LastUpdated = time.Now()
	c.Set(bookID, book)

	return true
}

// Size returns the current number of books in the cache
func (c *BookCache) Size() int {
	return c.cache.Size()
}

// Shutdown stops the cache cleanup goroutine
func (c *BookCache) Shutdown() {
	close(c.stopChan)
	c.logger.Debug("Book cache shut down")
}

// SerializeBook serializes a book to JSON format
func (c *BookCache) SerializeBook(book *Book) ([]byte, error) {
	return json.Marshal(book)
}

// DeserializeBook deserializes a book from JSON format
func (c *BookCache) DeserializeBook(data []byte) (*Book, error) {
	var book Book
	err := json.Unmarshal(data, &book)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize book: %w", err)
	}
	return &book, nil
}
