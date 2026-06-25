// package tasks provides a simple worker pool implementation for processing background jobs asynchronously. It defines a Job struct representing a unit of work, and a WorkerPool struct that manages a fixed number of worker goroutines to process jobs from a channel. The WorkerPool supports starting, enqueuing jobs, and graceful shutdown. This allows the application to offload time-consuming tasks such as sending emails, processing payments, or performing database maintenance without blocking the main request handling flow.
package asynq

import (
	"context"
	"sync"

	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Job represents a unit of work.
type Job struct {
	ID      string
	Payload interface{}
	Handler func(payload interface{}) error
}

// WorkerPool manages a fixed pool of workers processing jobs.
type WorkerPool struct {
	jobQueue   chan Job
	workers    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
	stopOnce   sync.Once
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		jobQueue:   make(chan Job, queueSize),
		workers:    workers,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// Start launches the worker goroutines.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	utils.Log.Infof("Worker pool started with %d workers", wp.workers)
}

// worker processes jobs from the queue.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for {
		select {
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			utils.Log.Debugf("Worker %d processing job %s", id, job.ID)
			if err := job.Handler(job.Payload); err != nil {
				utils.Log.WithError(err).Errorf("Worker %d: job %s failed", id, job.ID)
				// Optional: retry logic could be added here
			} else {
				utils.Log.Debugf("Worker %d completed job %s", id, job.ID)
			}
		case <-wp.ctx.Done():
			return
		}
	}
}

// Enqueue adds a job to the queue.
func (wp *WorkerPool) Enqueue(job Job) {
	select {
	case wp.jobQueue <- job:
		// job queued
	default:
		utils.Log.Warnf("Job queue full, dropping job %s", job.ID)
		// Could implement retry or persist to database
	}
}

// Stop gracefully shuts down the worker pool.
func (wp *WorkerPool) Stop() {
	wp.stopOnce.Do(func() {
		wp.cancelFunc()
		close(wp.jobQueue)
		wp.wg.Wait()
		utils.Log.Info("Worker pool stopped")
	})
}
