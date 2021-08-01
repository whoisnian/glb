package tasklane

import (
	"errors"
	"sync/atomic"
	"time"
)

type Task interface {
	Start()
}

var (
	ErrTimeout = errors.New("timeout")
)

// Multiple TaskQueue => One TaskLane
// One Buffered Channel + One Blocking Channel + One Goroutine(Worker) => One TaskQueue
type TaskLane struct {
	// LaneConfig
	laneSize  int
	queueSize int
	timeout   time.Duration

	// TaskQueue
	bufferedQueueList []chan Task
	blockingQueueList []chan Task
	universalQueue    chan Task

	// Status
	blockingTaskCnt int32
}

func (taskLane *TaskLane) startQueue(index int) {
	for {
		task := <-taskLane.bufferedQueueList[index]
		atomic.AddInt32(&taskLane.blockingTaskCnt, 1)
		select {
		case taskLane.blockingQueueList[index] <- task:
		default:
			select {
			case taskLane.blockingQueueList[index] <- task:
			case taskLane.universalQueue <- task:
			}
		}
		atomic.AddInt32(&taskLane.blockingTaskCnt, -1)
	}
}

func (taskLane *TaskLane) startWorker(index int) {
	var task Task
	for {
		select {
		case task = <-taskLane.blockingQueueList[index]:
		default:
			select {
			case task = <-taskLane.blockingQueueList[index]:
			case task = <-taskLane.universalQueue:
			}
		}
		task.Start()
	}
}

func New(laneSize, queueSize int) *TaskLane {
	// create buffered channels for TaskQueue
	bufferedQueueList := make([]chan Task, laneSize)
	for i := 0; i < laneSize; i++ {
		bufferedQueueList[i] = make(chan Task, queueSize)
	}

	// create TaskLane
	taskLane := &TaskLane{
		laneSize:  laneSize,
		queueSize: queueSize,
		timeout:   time.Second * 5,

		bufferedQueueList: bufferedQueueList,
		blockingQueueList: make([]chan Task, queueSize),
		universalQueue:    make(chan Task),

		blockingTaskCnt: 0,
	}

	// start TaskQueue
	for i := 0; i < taskLane.laneSize; i++ {
		go taskLane.startQueue(i)
		go taskLane.startWorker(i)
	}

	return taskLane
}

type LaneStatus struct {
	laneSize    int
	queueSize   int
	pendingTask int
}

func (taskLane *TaskLane) Status() *LaneStatus {
	pending := 0
	for i := 0; i < taskLane.laneSize; i++ {
		pending += len(taskLane.bufferedQueueList[i])
	}
	pending += int(atomic.LoadInt32(&taskLane.blockingTaskCnt))
	return &LaneStatus{
		laneSize:    taskLane.laneSize,
		queueSize:   taskLane.queueSize,
		pendingTask: pending,
	}
}

func (taskLane *TaskLane) ShortestQueueIndex() int {
	index := 0
	for i := 1; i < taskLane.laneSize; i++ {
		if len(taskLane.bufferedQueueList[i]) < len(taskLane.bufferedQueueList[index]) {
			index = i
		}
	}
	return index
}

func (taskLane *TaskLane) PushTask(task Task, index int) error {
	select {
	case taskLane.bufferedQueueList[index] <- task:
		return nil
	case <-time.After(taskLane.timeout):
		return ErrTimeout
	}
}
