package queue

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// Task represents a unit of work
type Task interface {
	ID() string
	Process(ctx context.Context) error
}

// WorkerPool manages a pool of workers processing tasks
type WorkerPool struct {
	numWorkers     int
	taskChan       chan Task
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	processedCount int64
	errorCount     int64
	mu             sync.RWMutex
	errors         []error
	maxErrors      int
}

// WorkerPoolOptions configures the worker pool
type WorkerPoolOptions struct {
	NumWorkers     int
	BufferSize     int
	MaxErrors      int
	WorkerTimeout  time.Duration
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(opts WorkerPoolOptions) *WorkerPool {
	if opts.NumWorkers <= 0 {
		opts.NumWorkers = runtime.NumCPU()
	}
	if opts.BufferSize <= 0 {
		opts.BufferSize = opts.NumWorkers * 2
	}
	if opts.MaxErrors <= 0 {
		opts.MaxErrors = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		numWorkers: opts.NumWorkers,
		taskChan:   make(chan Task, opts.BufferSize),
		ctx:        ctx,
		cancel:     cancel,
		errors:     make([]error, 0),
		maxErrors:  opts.MaxErrors,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	log.Info().Int("workers", wp.numWorkers).Msg("starting worker pool")

	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// EnqueueTask adds a task to the queue
func (wp *WorkerPool) EnqueueTask(task Task) error {
	select {
	case wp.taskChan <- task:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Stop gracefully stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.taskChan)
	wp.wg.Wait()
	wp.cancel()

	log.Info().
		Int64("processed", atomic.LoadInt64(&wp.processedCount)).
		Int64("errors", atomic.LoadInt64(&wp.errorCount)).
		Msg("worker pool stopped")
}

// WaitForCompletion waits for all tasks to complete or context to be cancelled
func (wp *WorkerPool) WaitForCompletion(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Debug().Msg("all workers completed")
	case <-ctx.Done():
		log.Warn().Msg("worker pool cancelled")
		wp.cancel()
	}
}

// GetStats returns processing statistics
func (wp *WorkerPool) GetStats() (processed int64, errors int64) {
	return atomic.LoadInt64(&wp.processedCount), atomic.LoadInt64(&wp.errorCount)
}

// GetErrors returns all errors encountered
func (wp *WorkerPool) GetErrors() []error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	errors := make([]error, len(wp.errors))
	copy(errors, wp.errors)
	return errors
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(workerID int) {
	defer wp.wg.Done()

	log.Debug().Int("worker_id", workerID).Msg("worker started")

	for {
		select {
		case task, ok := <-wp.taskChan:
			if !ok {
				log.Debug().Int("worker_id", workerID).Msg("worker stopped - channel closed")
				return
			}

			wp.processTask(workerID, task)

		case <-wp.ctx.Done():
			log.Debug().Int("worker_id", workerID).Msg("worker stopped - context cancelled")
			return
		}
	}
}

// processTask processes a single task
func (wp *WorkerPool) processTask(workerID int, task Task) {
	start := time.Now()

	log.Debug().
		Int("worker_id", workerID).
		Str("task_id", task.ID()).
		Msg("processing task")

	err := task.Process(wp.ctx)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&wp.errorCount, 1)
		wp.recordError(fmt.Errorf("task %s failed: %w", task.ID(), err))

		log.Error().
			Err(err).
			Int("worker_id", workerID).
			Str("task_id", task.ID()).
			Dur("duration", duration).
			Msg("task failed")
	} else {
		log.Debug().
			Int("worker_id", workerID).
			Str("task_id", task.ID()).
			Dur("duration", duration).
			Msg("task completed")
	}

	atomic.AddInt64(&wp.processedCount, 1)

	// Log progress periodically
	processed := atomic.LoadInt64(&wp.processedCount)
	if processed%10 == 0 {
		log.Info().
			Int64("processed", processed).
			Int64("errors", atomic.LoadInt64(&wp.errorCount)).
			Msg("processing progress")
	}
}

// recordError records an error, with a maximum limit
func (wp *WorkerPool) recordError(err error) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.errors) < wp.maxErrors {
		wp.errors = append(wp.errors, err)
	}
}

// UserProcessingTask represents a task to process a single user
type UserProcessingTask struct {
	Username string
	Processor func(ctx context.Context, username string) error
}

// ID returns the task ID
func (t *UserProcessingTask) ID() string {
	return fmt.Sprintf("user_processing:%s", t.Username)
}

// Process processes the user
func (t *UserProcessingTask) Process(ctx context.Context) error {
	if t.Processor == nil {
		return fmt.Errorf("no processor function provided")
	}
	return t.Processor(ctx, t.Username)
}

// RateLimitedWorkerPool extends WorkerPool with rate limiting
type RateLimitedWorkerPool struct {
	*WorkerPool
	rateLimiter chan struct{}
	rateLimit   time.Duration
}

// NewRateLimitedWorkerPool creates a worker pool with rate limiting
func NewRateLimitedWorkerPool(opts WorkerPoolOptions, requestsPerSecond int) *RateLimitedWorkerPool {
	pool := NewWorkerPool(opts)

	rateLimit := time.Second / time.Duration(requestsPerSecond)
	rateLimiter := make(chan struct{}, requestsPerSecond)

	// Fill the rate limiter
	for i := 0; i < requestsPerSecond; i++ {
		rateLimiter <- struct{}{}
	}

	rlwp := &RateLimitedWorkerPool{
		WorkerPool:  pool,
		rateLimiter: rateLimiter,
		rateLimit:   rateLimit,
	}

	// Start the rate limiter refiller
	go rlwp.refillRateLimiter()

	return rlwp
}

// refillRateLimiter refills the rate limiter tokens
func (rlwp *RateLimitedWorkerPool) refillRateLimiter() {
	ticker := time.NewTicker(rlwp.rateLimit)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case rlwp.rateLimiter <- struct{}{}:
				// Token added
			default:
				// Rate limiter full, skip
			}
		case <-rlwp.ctx.Done():
			return
		}
	}
}

// WaitForRateLimit waits for a rate limit token
func (rlwp *RateLimitedWorkerPool) WaitForRateLimit(ctx context.Context) error {
	select {
	case <-rlwp.rateLimiter:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}