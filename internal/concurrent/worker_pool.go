package concurrent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a task that can be executed by the worker pool
type Task interface {
	Execute() Result
	GetID() string
	GetPriority() int
}

// Result represents the result of a task execution
type Result interface {
	GetTaskID() string
	GetError() error
	GetData() interface{}
}

// WorkerPool manages a pool of workers for concurrent task execution
type WorkerPool struct {
	maxWorkers    int
	taskQueue     chan Task
	resultQueue   chan Result
	workers       []*Worker
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	started       int32
	stopped       int32
	
	// Metrics
	tasksSubmitted int64
	tasksCompleted int64
	tasksFailed    int64
	startTime      time.Time
}

// Worker represents a single worker in the pool
type Worker struct {
	id        int
	pool      *WorkerPool
	taskQueue chan Task
	ctx       context.Context
}

// PoolMetrics represents metrics for the worker pool
type PoolMetrics struct {
	TasksSubmitted   int64
	TasksCompleted   int64
	TasksFailed      int64
	ActiveWorkers    int
	QueueSize        int
	Uptime           time.Duration
	ThroughputPerSec float64
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		maxWorkers:  maxWorkers,
		taskQueue:   make(chan Task, maxWorkers*2), // Buffer for 2x workers
		resultQueue: make(chan Result, maxWorkers*2),
		ctx:         ctx,
		cancel:      cancel,
		startTime:   time.Now(),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() error {
	if !atomic.CompareAndSwapInt32(&wp.started, 0, 1) {
		return fmt.Errorf("worker pool already started")
	}

	// Create workers
	wp.workers = make([]*Worker, wp.maxWorkers)
	for i := 0; i < wp.maxWorkers; i++ {
		worker := &Worker{
			id:        i,
			pool:      wp,
			taskQueue: wp.taskQueue,
			ctx:       wp.ctx,
		}
		wp.workers[i] = worker
		
		// Start worker goroutine
		wp.wg.Add(1)
		go worker.run()
	}

	return nil
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() error {
	if !atomic.CompareAndSwapInt32(&wp.stopped, 0, 1) {
		return fmt.Errorf("worker pool already stopped")
	}

	// Cancel context to signal workers to stop
	wp.cancel()
	
	// Close task queue
	close(wp.taskQueue)
	
	// Wait for all workers to finish
	wp.wg.Wait()
	
	// Close result queue
	close(wp.resultQueue)
	
	return nil
}

// SubmitTask submits a task to the worker pool
func (wp *WorkerPool) SubmitTask(task Task) error {
	if atomic.LoadInt32(&wp.stopped) == 1 {
		return fmt.Errorf("worker pool is stopped")
	}

	select {
	case wp.taskQueue <- task:
		atomic.AddInt64(&wp.tasksSubmitted, 1)
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool context cancelled")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// GetResult returns a channel to receive results
func (wp *WorkerPool) GetResult() <-chan Result {
	return wp.resultQueue
}

// GetMetrics returns current pool metrics
func (wp *WorkerPool) GetMetrics() *PoolMetrics {
	uptime := time.Since(wp.startTime)
	throughput := float64(atomic.LoadInt64(&wp.tasksCompleted)) / uptime.Seconds()
	
	return &PoolMetrics{
		TasksSubmitted:   atomic.LoadInt64(&wp.tasksSubmitted),
		TasksCompleted:   atomic.LoadInt64(&wp.tasksCompleted),
		TasksFailed:      atomic.LoadInt64(&wp.tasksFailed),
		ActiveWorkers:    len(wp.workers),
		QueueSize:        len(wp.taskQueue),
		Uptime:           uptime,
		ThroughputPerSec: throughput,
	}
}

// run runs the worker
func (w *Worker) run() {
	defer w.pool.wg.Done()
	
	for {
		select {
		case task, ok := <-w.taskQueue:
			if !ok {
				// Task queue closed, worker should stop
				return
			}
			
			// Execute task
			result := task.Execute()
			
			// Send result
			select {
			case w.pool.resultQueue <- result:
				if result.GetError() != nil {
					atomic.AddInt64(&w.pool.tasksFailed, 1)
				} else {
					atomic.AddInt64(&w.pool.tasksCompleted, 1)
				}
			case <-w.ctx.Done():
				return
			}
			
		case <-w.ctx.Done():
			return
		}
	}
}

// BaseTask provides a base implementation for common task functionality
type BaseTask struct {
	ID       string
	Priority int
}

func (bt *BaseTask) GetID() string {
	return bt.ID
}

func (bt *BaseTask) GetPriority() int {
	return bt.Priority
}

// BaseResult provides a base implementation for common result functionality
type BaseResult struct {
	TaskID string
	Error  error
	Data   interface{}
}

func (br *BaseResult) GetTaskID() string {
	return br.TaskID
}

func (br *BaseResult) GetError() error {
	return br.Error
}

func (br *BaseResult) GetData() interface{} {
	return br.Data
}