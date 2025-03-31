package workerpool

import (
	"context"
	"sync"
)

// Task represents a job to be executed by the worker pool
type Task func(ctx context.Context) error

// WorkerPool represents a pool of workers that can execute tasks
type WorkerPool struct {
	tasks    chan Task
	wg       sync.WaitGroup
	ctx      context.Context
	cancelFn context.CancelFunc
}

// New creates a new worker pool with the specified number of workers
func New(numWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		tasks:    make(chan Task, numWorkers*10), // Buffer size is 10x the number of workers
		ctx:      ctx,
		cancelFn: cancel,
	}

	// Start the workers
	for i := 0; i < numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}

	return wp
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case task, ok := <-wp.tasks:
			if !ok {
				// Channel closed, exit
				return
			}
			// Execute the task
			_ = task(wp.ctx) // Ignoring errors for now
		case <-wp.ctx.Done():
			// Context cancelled, exit
			return
		}
	}
}

// Submit adds a task to the worker pool
func (wp *WorkerPool) Submit(task Task) {
	select {
	case wp.tasks <- task:
		// Task submitted successfully
	case <-wp.ctx.Done():
		// Worker pool is shutting down
	}
}

// Shutdown stops the worker pool and waits for all tasks to complete
func (wp *WorkerPool) Shutdown() {
	close(wp.tasks)
	wp.cancelFn()
	wp.wg.Wait()
}
