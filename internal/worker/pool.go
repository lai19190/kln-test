package worker

import (
	"context"
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"kln-test/internal/config"
)

// Job represents a unit of work to be processed
type Job[T any] struct {
	ID      string
	Payload T
	Process func(context.Context, T) error
}

// Pool manages a pool of workers and a job queue
type Pool[T any] struct {
	workers    int
	jobs       chan Job[T]
	wg         sync.WaitGroup
	cfg        *config.Config
	ctx        context.Context
	cancelFunc context.CancelFunc
	mu         sync.RWMutex
}

// NewPool creates a new worker pool
func NewPool[T any](cfg *config.Config) *Pool[T] {
	p := &Pool[T]{
		cfg: cfg,
	}
	p.Start()

	return p
}

// watchConfig monitors configuration changes and scales the pool accordingly
func (p *Pool[T]) watchConfig() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.scalePoolIfNeeded()
		}
	}
}

// scalePoolIfNeeded adjusts the pool size based on the current configuration
func (p *Pool[T]) scalePoolIfNeeded() {
	workerConfig := p.cfg.GetWorkerConfig()
	// resize pool if needed
	if workerConfig.QueueSize != cap(p.jobs) || workerConfig.PoolSize != p.workers {
		p.mu.Lock()
		defer p.mu.Unlock()
		log.Printf("Resizing worker pool from %d to %d", p.workers, workerConfig.PoolSize)
		log.Printf("Resizing worker queue from %d to %d", cap(p.jobs), workerConfig.QueueSize)
		p.Shutdown()
		p.Start()
	}
}

// Submit adds a job to the queue
func (p *Pool[T]) Submit(job Job[T]) error {
	p.scalePoolIfNeeded()

	select {
	case p.jobs <- job:
		return nil
	default:
		return errors.New("job queue is full")
	}
}

// Start start the worker pool
func (p *Pool[T]) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	workerConfig := p.cfg.GetWorkerConfig()

	p.workers = workerConfig.PoolSize
	p.ctx = ctx
	p.cancelFunc = cancel
	p.jobs = make(chan Job[T], workerConfig.QueueSize)

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.startWorker(i)
	}

	go p.watchConfig()
}

// Shutdown gracefully shuts down the worker pool
func (p *Pool[T]) Shutdown() {
	p.cancelFunc()
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool[T]) startWorker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.processJobWithRetry(id, job)
		}
	}
}

func (p *Pool[T]) processJobWithRetry(workerID int, job Job[T]) {
	p.mu.RLock()
	retry := p.cfg.GetWorkerConfig().Retry
	p.mu.RUnlock()

	timeout := time.Duration(retry.InitialTimeout) * time.Second
	maxTimeOut := time.Duration(retry.MaxTimeout) * time.Second

	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		log.Printf("Worker %d processing job %s (attempt %d/%d)", workerID, job.ID, attempt, retry.MaxAttempts)

		// Create a context with timeout for this attempt
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Run the job with timeout
		done := make(chan error, 1)
		go func() {
			done <- job.Process(ctx, job.Payload)
		}()

		// Wait for job completion or timeout
		select {
		case err := <-done:
			cancel()
			if err == nil {
				log.Printf("Worker %d successfully completed job %s", workerID, job.ID)
				return
			}
			log.Printf("Worker %d failed job %s: %v", workerID, job.ID, err)
		case <-ctx.Done():
			cancel()
			log.Printf("Worker %d still processing job %s (attempt %d/%d)", workerID, job.ID, attempt, retry.MaxAttempts)
		}

		// Calculate next timeout with exponential backoff
		backoffTime := timeout * time.Duration(math.Pow(2, float64(attempt-1)))
		if backoffTime > maxTimeOut {
			timeout = maxTimeOut
		} else {
			timeout = backoffTime
		}

		// If this was the last attempt, log failure
		if attempt == retry.MaxAttempts {
			log.Printf("Worker %d gave up on job %s after %d attempts", workerID, job.ID, retry.MaxAttempts)
			return
		}

		// Wait before retrying
		time.Sleep(timeout)
	}
}
