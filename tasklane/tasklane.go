// Package tasklane creates multiple goroutines to wait and run tasks.
//
//	TaskLane  := []TaskQueue
//	TaskQueue := {
//	  Buffered Channel
//	  Blocking Channel
//	  Goroutine Worker
//	}
package tasklane

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Task interface {
	Start()
}

var (
	ErrTimeout = errors.New("timeout")
)

type TaskLane struct {
	// Context
	ctx context.Context
	wg  *sync.WaitGroup

	// Config
	laneSize  int
	queueSize int
	timeout   time.Duration

	// Queue
	bufferedQueueList []chan Task
	blockingQueueList []chan Task
	universalQueue    chan Task

	// Status
	blockingTaskCnt *atomic.Uint32
	lastPanic       any
}

func (tl *TaskLane) startQueue(index int) {
	defer tl.wg.Done()

	var task Task
	for {
		select {
		case <-tl.ctx.Done():
			return
		case task = <-tl.bufferedQueueList[index]:
		}
		tl.blockingTaskCnt.Add(1)
		select {
		case <-tl.ctx.Done():
			return
		default:
			select {
			case tl.blockingQueueList[index] <- task:
			default:
				select {
				case <-tl.ctx.Done():
					return
				case tl.blockingQueueList[index] <- task:
				case tl.universalQueue <- task:
				}
			}
		}
		tl.blockingTaskCnt.Add(^uint32(0)) // decrement blockingTaskCnt
	}
}

func (tl *TaskLane) startWorker(index int) {
	defer tl.wg.Done()

	var task Task
	for {
		select {
		case <-tl.ctx.Done():
			return
		default:
			select {
			case task = <-tl.blockingQueueList[index]:
			default:
				select {
				case <-tl.ctx.Done():
					return
				case task = <-tl.blockingQueueList[index]:
				case task = <-tl.universalQueue:
				}
			}
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					tl.lastPanic = err
				}
			}()
			task.Start()
		}()
	}
}

func New(ctx context.Context, laneSize, queueSize int) *TaskLane {
	// create channels for TaskQueue
	bufferedQueueList := make([]chan Task, laneSize)
	blockingQueueList := make([]chan Task, laneSize)
	for i := 0; i < laneSize; i++ {
		bufferedQueueList[i] = make(chan Task, queueSize)
		blockingQueueList[i] = make(chan Task)
	}

	// create TaskLane
	tl := &TaskLane{
		ctx: ctx,
		wg:  new(sync.WaitGroup),

		laneSize:  laneSize,
		queueSize: queueSize,
		timeout:   time.Second,

		bufferedQueueList: bufferedQueueList,
		blockingQueueList: blockingQueueList,
		universalQueue:    make(chan Task),

		blockingTaskCnt: &atomic.Uint32{},
	}

	// start TaskQueue
	tl.wg.Add(laneSize * 2)
	for i := 0; i < laneSize; i++ {
		go tl.startQueue(i)
		go tl.startWorker(i)
	}

	return tl
}

// SetTimeout sets the timeout for PushTask(). The default timeout is 1 second.
func (tl *TaskLane) SetTimeout(t time.Duration) {
	tl.timeout = t
}

// Wait blocks until all TaskQueue ended.
//
// TaskQueue will start exiting if the context was Done. Pending tasks will be aborted.
func (tl *TaskLane) Wait() {
	tl.wg.Wait()
}

type LaneStatus struct {
	LaneSize    int
	QueueSize   int
	PendingTask int
	LastPanic   any
}

// Status returns the status of TaskLane.
func (tl *TaskLane) Status() *LaneStatus {
	pending := 0
	for i := 0; i < tl.laneSize; i++ {
		pending += len(tl.bufferedQueueList[i])
	}
	pending += int(tl.blockingTaskCnt.Load())
	return &LaneStatus{
		LaneSize:    tl.laneSize,
		QueueSize:   tl.queueSize,
		PendingTask: pending,
		LastPanic:   tl.lastPanic,
	}
}

// ShortestQueueIndex returns the index of shortest TaskQueue.
func (tl *TaskLane) ShortestQueueIndex() int {
	index := 0
	for i := 1; i < tl.laneSize; i++ {
		if len(tl.bufferedQueueList[i]) < len(tl.bufferedQueueList[index]) {
			index = i
		}
	}
	return index
}

// PushTask will push task to BufferedChannel of specified TaskQueue.
//
// PushTask returns a non-nil error if failed to push task:
// context.Canceled or context.DeadlineExceeded if the context was Done.
// tasklane.ErrTimeout if specified TaskQueue is full until timeout.
func (tl *TaskLane) PushTask(task Task, index int) error {
	select {
	case <-tl.ctx.Done():
		return tl.ctx.Err()
	default:
		select {
		case <-tl.ctx.Done():
			return tl.ctx.Err()
		case tl.bufferedQueueList[index] <- task:
			return nil
		case <-time.After(tl.timeout):
			return ErrTimeout
		}
	}
}
