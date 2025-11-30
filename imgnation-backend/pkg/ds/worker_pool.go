package ds

import (
	"context"
	"sync"
)

type Task func(ctx context.Context)

type WorkersCount struct {
	Lim  int32
	Busy int32
}

type IWorkerPool interface {
	Exec(ctx context.Context, task Task)
	QueueLen() int
	WorkersCount() WorkersCount
	SetWorkersCountLim(newSize int32)
	Run(ctx context.Context)
}

type WorkerPool struct {
	taskQueue   []Task
	mu          sync.RWMutex
	countLim    int32
	countBusy   int32
	taskChan    chan Task
	wg          sync.WaitGroup
	poolCtx     context.Context
	cancelPool  context.CancelFunc
	shutdownWg  sync.WaitGroup
	shutdownSig chan struct{}
}

func NewWorkerPool(workersCountLim int32, taskChanBufferSize int) *WorkerPool {
	return &WorkerPool{
		taskQueue:   []Task{},
		countLim:    workersCountLim,
		taskChan:    make(chan Task, taskChanBufferSize),
		shutdownSig: make(chan struct{}),
	}
}

func (wp *WorkerPool) Run(ctx context.Context) {
	wp.poolCtx, wp.cancelPool = context.WithCancel(context.Background())
	wp.shutdownWg.Add(1)

	go func() {
		defer wp.shutdownWg.Done()
		wp.dispatcher()
	}()

	// Handle external context cancellation
	go func() {
		select {
		case <-ctx.Done():
			wp.cancelPool()
		case <-wp.poolCtx.Done():
		}
	}()
}

func (wp *WorkerPool) dispatcher() {
	defer wp.wg.Wait()

	for {
		select {
		case task := <-wp.taskChan:
			wp.mu.Lock()
			if wp.countBusy >= wp.countLim {
				wp.taskQueue = append(wp.taskQueue, task)
				wp.mu.Unlock()
				continue
			}
			wp.countBusy++
			wp.mu.Unlock()

			wp.wg.Add(1)
			go wp.worker(task)

		case <-wp.poolCtx.Done():
			close(wp.shutdownSig)
			return
		}
	}
}

func (wp *WorkerPool) worker(initialTask Task) {
	defer wp.wg.Done()

	// Create worker-local context for task execution
	taskCtx, cancel := context.WithCancel(wp.poolCtx)
	defer cancel()

	// Process the initial task
	initialTask(taskCtx)

	// Process tasks from the queue
	for {
		// Check for shutdown signal
		select {
		case <-wp.shutdownSig:
			return
		default:
		}

		// Get next task from queue
		wp.mu.Lock()
		if len(wp.taskQueue) == 0 {
			wp.countBusy--
			wp.mu.Unlock()
			return
		}
		task := wp.taskQueue[0]
		wp.taskQueue = wp.taskQueue[1:]
		wp.mu.Unlock()

		// Execute the task
		task(taskCtx)
	}
}

func (wp *WorkerPool) Exec(ctx context.Context, task Task) {
	// Check if task context is already canceled
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Check if pool is shutting down
	select {
	case <-wp.poolCtx.Done():
		return
	default:
	}

	// Try to send without blocking first
	select {
	case wp.taskChan <- task:
		return
	default:
	}

	// If channel is full, use queue with lock
	wp.mu.Lock()
	wp.taskQueue = append(wp.taskQueue, task)
	wp.mu.Unlock()
}

func (wp *WorkerPool) QueueLen() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return len(wp.taskQueue)
}

func (wp *WorkerPool) WorkersCount() WorkersCount {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return WorkersCount{
		Lim:  wp.countLim,
		Busy: wp.countBusy,
	}
}

func (wp *WorkerPool) SetWorkersCountLim(newSize int32) {
	wp.mu.Lock()
	wp.countLim = newSize
	wp.mu.Unlock()
}

func (wp *WorkerPool) Stop() {
	wp.cancelPool()
	wp.shutdownWg.Wait()
}
