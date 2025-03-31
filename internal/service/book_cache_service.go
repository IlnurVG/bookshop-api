package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/pkg/cache"
	"github.com/bookshop/api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// BookCacheService provides caching functionality for books
type BookCacheService struct {
	bookRepo    repositories.BookRepository
	cache       *cache.BookCache // L1 cache (in-memory LRU)
	redisClient *redis.Client    // L2 cache (Redis)
	logger      logger.Logger
}

// NewBookCacheService creates a new book cache service
func NewBookCacheService(
	bookRepo repositories.BookRepository,
	redisClient *redis.Client,
	logger logger.Logger,
) *BookCacheService {
	// Create in-memory LRU cache with capacity of 1000 books, 5-minute TTL and 1-minute cleanup interval
	bookCache := cache.NewBookCache(1000, 5*time.Minute, 1*time.Minute, logger)

	return &BookCacheService{
		bookRepo:    bookRepo,
		cache:       bookCache,
		redisClient: redisClient,
		logger:      logger,
	}
}

// GetBook retrieves a book by ID using caching
func (s *BookCacheService) GetBook(ctx context.Context, bookID int) (*models.Book, error) {
	bookIDStr := strconv.Itoa(bookID)

	// Try to get from L1 cache first (in-memory LRU)
	cachedBook, found := s.cache.Get(bookIDStr)
	if found {
		s.logger.Debug("Book found in L1 cache", "bookID", bookID)
		// Convert to domain model
		return s.convertCachedBook(cachedBook)
	}

	// Try to get from L2 cache (Redis)
	redisKey := fmt.Sprintf("book:%d", bookID)
	data, err := s.redisClient.Get(ctx, redisKey).Bytes()
	if err == nil && len(data) > 0 {
		s.logger.Debug("Book found in L2 cache (Redis)", "bookID", bookID)

		// Deserialize book
		cachedBook, err := s.cache.DeserializeBook(data)
		if err == nil {
			// Store in L1 cache for future requests
			s.cache.Set(bookIDStr, cachedBook)
			return s.convertCachedBook(cachedBook)
		}
		s.logger.Debug("Failed to deserialize book from Redis", "error", err)
	}

	// Not found in cache, fetch from database
	s.logger.Debug("Book not found in cache, fetching from database", "bookID", bookID)
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch book: %w", err)
	}

	// Store in caches
	s.storeInCaches(ctx, book)

	return book, nil
}

// GetBooks retrieves multiple books by their IDs using caching
func (s *BookCacheService) GetBooks(ctx context.Context, bookIDs []int) ([]*models.Book, error) {
	results := make([]*models.Book, 0, len(bookIDs))
	missingIDs := make([]int, 0)

	// Convert IDs to strings for cache lookup
	bookIDStrs := make([]string, len(bookIDs))
	for i, id := range bookIDs {
		bookIDStrs[i] = strconv.Itoa(id)
	}

	// Try to get from L1 cache
	cachedBooks := s.cache.GetMultiple(bookIDStrs)

	// Check which books we got from cache and which we need to fetch
	for i, id := range bookIDs {
		idStr := bookIDStrs[i]
		if cachedBook, found := cachedBooks[idStr]; found {
			// Convert to domain model
			book, err := s.convertCachedBook(cachedBook)
			if err == nil {
				results = append(results, book)
				continue
			}
		}
		// Not found in L1 cache, will need to fetch
		missingIDs = append(missingIDs, id)
	}

	if len(missingIDs) == 0 {
		// All books were found in L1 cache
		s.logger.Debug("All books found in L1 cache", "count", len(results))
		return results, nil
	}

	// Fetch missing books from database
	s.logger.Debug("Fetching missing books from database", "count", len(missingIDs))
	booksResult, err := s.bookRepo.GetBooksByIDs(ctx, missingIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch books: %w", err)
	}

	// Store fetched books in caches and add to results
	for i := range booksResult {
		book := &booksResult[i] // Get pointer to book in slice
		s.storeInCaches(ctx, book)
		results = append(results, book)
	}

	return results, nil
}

// InvalidateBook removes a book from all caches
func (s *BookCacheService) InvalidateBook(ctx context.Context, bookID int) {
	bookIDStr := strconv.Itoa(bookID)

	// Remove from L1 cache
	s.cache.Delete(bookIDStr)

	// Remove from L2 cache (Redis)
	redisKey := fmt.Sprintf("book:%d", bookID)
	if err := s.redisClient.Del(ctx, redisKey).Err(); err != nil {
		s.logger.Debug("Failed to delete book from Redis", "bookID", bookID, "error", err)
	}

	s.logger.Debug("Book cache invalidated", "bookID", bookID)
}

// UpdateBookCache updates a book in the cache when it's updated in the database
func (s *BookCacheService) UpdateBookCache(ctx context.Context, book *models.Book) {
	s.storeInCaches(ctx, book)
	s.logger.Debug("Book cache updated", "bookID", book.ID)
}

// storeInCaches stores a book in both L1 and L2 caches
func (s *BookCacheService) storeInCaches(ctx context.Context, book *models.Book) {
	// Convert domain model to cache model
	cachedBook := &cache.Book{
		ID:          book.ID,
		Title:       book.Title,
		Author:      book.Author,
		Price:       book.Price,
		Quantity:    book.Stock, // Map Stock to Quantity in cache model
		LastUpdated: book.UpdatedAt,
	}

	// Store in L1 cache
	bookIDStr := strconv.Itoa(book.ID)
	s.cache.Set(bookIDStr, cachedBook)

	// Store in L2 cache (Redis)
	redisKey := fmt.Sprintf("book:%d", book.ID)
	data, err := s.cache.SerializeBook(cachedBook)
	if err != nil {
		s.logger.Debug("Failed to serialize book for Redis", "error", err)
		return
	}

	// Set in Redis with 30-minute TTL
	if err := s.redisClient.Set(ctx, redisKey, data, 30*time.Minute).Err(); err != nil {
		s.logger.Debug("Failed to store book in Redis", "error", err)
	}
}

// convertCachedBook converts a cached book to a domain model
func (s *BookCacheService) convertCachedBook(cachedBook *cache.Book) (*models.Book, error) {
	return &models.Book{
		ID:        cachedBook.ID,
		Title:     cachedBook.Title,
		Author:    cachedBook.Author,
		Price:     cachedBook.Price,
		Stock:     cachedBook.Quantity, // Map Quantity to Stock in domain model
		CreatedAt: time.Time{},         // This field may not be available in the cache
		UpdatedAt: cachedBook.LastUpdated,
	}, nil
}

// Shutdown properly shuts down the cache service
func (s *BookCacheService) Shutdown() {
	s.cache.Shutdown()
	s.logger.Debug("Book cache service shut down")
}
