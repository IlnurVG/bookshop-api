package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/pkg/workerpool"
	"github.com/bookshop/api/pkg/logger"
)

// OrderProcessRequest contains data for asynchronous order processing
type OrderProcessRequest struct {
	UserID    int
	CartItems []models.CartItem
	Order     *models.Order
}

// OrderProcessor processes orders asynchronously using a worker pool
type OrderProcessor struct {
	orderRepo  repositories.OrderRepository
	bookRepo   repositories.BookRepository
	cartRepo   repositories.CartRepository
	workerPool *workerpool.WorkerPool
	logger     logger.Logger
	mu         sync.Mutex
	results    map[int]chan error // Map to store processing results for each order
}

// NewOrderProcessor creates a new asynchronous order processor
func NewOrderProcessor(
	orderRepo repositories.OrderRepository,
	bookRepo repositories.BookRepository,
	cartRepo repositories.CartRepository,
	logger logger.Logger,
	numWorkers int,
) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:  orderRepo,
		bookRepo:   bookRepo,
		cartRepo:   cartRepo,
		workerPool: workerpool.New(numWorkers),
		logger:     logger,
		results:    make(map[int]chan error),
	}
}

// ProcessOrder sends an order for asynchronous processing
func (p *OrderProcessor) ProcessOrder(ctx context.Context, request OrderProcessRequest) chan error {
	resultCh := make(chan error, 1)

	// Save result channel
	p.mu.Lock()
	p.results[request.Order.ID] = resultCh
	p.mu.Unlock()

	// Submit task to worker pool
	p.workerPool.Submit(func(ctx context.Context) error {
		err := p.processOrderTask(ctx, request)

		// Send result to channel
		resultCh <- err
		close(resultCh)

		// Remove channel from results map
		p.mu.Lock()
		delete(p.results, request.Order.ID)
		p.mu.Unlock()

		return err
	})

	return resultCh
}

// processOrderTask processes an order asynchronously
func (p *OrderProcessor) processOrderTask(ctx context.Context, request OrderProcessRequest) error {
	p.logger.Debug("Starting asynchronous order processing", "orderID", request.Order.ID)

	// Add delay to simulate long processing
	time.Sleep(100 * time.Millisecond)

	// Check book availability and reserve them
	bookIDs := make([]int, 0, len(request.CartItems))
	for _, item := range request.CartItems {
		bookIDs = append(bookIDs, item.BookID)
	}

	if err := p.bookRepo.ReserveBooks(ctx, bookIDs); err != nil {
		p.logger.Error("Error reserving books", "error", err, "orderID", request.Order.ID)
		return fmt.Errorf("error reserving books: %w", err)
	}

	// Create order in database
	if err := p.orderRepo.Create(ctx, request.Order); err != nil {
		// In case of error, release reserved books
		p.bookRepo.ReleaseBooks(ctx, bookIDs)
		p.logger.Error("Error creating order", "error", err, "orderID", request.Order.ID)
		return fmt.Errorf("error creating order: %w", err)
	}

	// Clear cart
	if err := p.cartRepo.ClearCart(ctx, request.UserID); err != nil {
		p.logger.Error("Error clearing cart", "error", err, "userID", request.UserID)
		// We don't return an error here because the order has already been created
	}

	p.logger.Debug("Completed asynchronous order processing", "orderID", request.Order.ID)
	return nil
}

// GetResult returns the channel with the order processing result
func (p *OrderProcessor) GetResult(orderID int) (chan error, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	result, exists := p.results[orderID]
	return result, exists
}

// Shutdown stops the worker pool and waits for all tasks to complete
func (p *OrderProcessor) Shutdown() {
	p.workerPool.Shutdown()
	p.logger.Info("Order processor stopped")
}
