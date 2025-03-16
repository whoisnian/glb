package tasklane_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/whoisnian/glb/tasklane"
)

const panicReason = "[expected panic]"

type TestTask struct {
	wg    *sync.WaitGroup
	out   chan struct{}
	panic bool
}

func (task *TestTask) Start(context.Context) {
	defer task.wg.Done()
	if task.panic {
		panic(panicReason)
	}
	task.out <- struct{}{}
}

func TestTaskLane(t *testing.T) {
	wg := new(sync.WaitGroup)
	out := make(chan struct{})

	laneSize, queueSize := 2, 3
	tl := tasklane.New(t.Context(), laneSize, queueSize)
	if st := tl.Status(); st.LaneSize != laneSize || st.QueueSize != queueSize || st.PendingSize != 0 || st.LastPanic != nil {
		t.Fatalf("tasklane.New() = %#v, want %#v", st, tasklane.LaneStatus{laneSize, queueSize, 0, nil})
	}

	taskCnt := 10 // taskCnt should less than or equal to ((queueSize + 2) * laneSize)
	wg.Add(taskCnt)
	for i := range taskCnt {
		time.Sleep(time.Millisecond) // wait TaskQueue to sync internal channel
		if err := tl.PushTask(&TestTask{wg, out, false}, tl.ShortestQueueIndex()); err != nil {
			t.Fatalf("tasklane.PushTask()#%d error %v, want nil", i, err)
		}
	}
	if tl.Status().PendingSize != taskCnt-laneSize {
		t.Fatalf("LaneStatus.PendingSize = %d, want %d", tl.Status().PendingSize, taskCnt-laneSize)
	}

	resCnt := 0
	for range laneSize {
		<-out
		resCnt++
	}
	time.Sleep(time.Millisecond) // wait TaskQueue to sync internal channel
	if tl.Status().PendingSize != taskCnt-laneSize-resCnt {
		t.Fatalf("LaneStatus.PendingSize = %d, want %d", tl.Status().PendingSize, taskCnt-laneSize*2)
	}

	for range taskCnt - resCnt {
		<-out
	}
	time.Sleep(time.Millisecond) // wait TaskQueue to sync internal channel
	if tl.Status().PendingSize != 0 {
		t.Fatalf("LaneStatus.PendingSize = %d, want %d", tl.Status().PendingSize, 0)
	}
}

func TestTaskPanic(t *testing.T) {
	wg := new(sync.WaitGroup)
	out := make(chan struct{}, 32)

	laneSize, queueSize := 2, 3
	tl := tasklane.New(t.Context(), laneSize, queueSize)

	taskCnt, panicCnt := 10, 3 // taskCnt should less than or equal to ((queueSize + 2) * laneSize)
	wg.Add(taskCnt)
	for i := range taskCnt {
		time.Sleep(time.Millisecond) // wait TaskQueue to sync internal channel
		if err := tl.PushTask(&TestTask{wg, out, i < panicCnt}, tl.ShortestQueueIndex()); err != nil {
			t.Fatalf("tasklane.PushTask()#%d error %v, want nil", i, err)
		}
	}

	wg.Wait()
	st := tl.Status()
	if st.PendingSize != 0 {
		t.Fatalf("LaneStatus.PendingSize = %d, want %d", st.PendingSize, 0)
	}
	if st.LastPanic.(string) != panicReason {
		t.Fatalf("LaneStatus.LastPanic = %v, want %s", st.LastPanic, panicReason)
	}
	if len(out) != taskCnt-panicCnt {
		t.Fatalf("testTask panic %d, want %d", taskCnt-len(out), panicCnt)
	}
}

func TestPushTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	out := make(chan struct{})

	laneSize, queueSize := 2, 3
	tl := tasklane.New(ctx, laneSize, queueSize)
	tl.SetTimeout(time.Millisecond)

	taskCnt := 10 // taskCnt should less than or equal to ((queueSize + 2) * laneSize)
	wg.Add(taskCnt)
	for i := range taskCnt {
		time.Sleep(time.Millisecond) // wait TaskQueue to sync internal channel
		if err := tl.PushTask(&TestTask{wg, out, false}, tl.ShortestQueueIndex()); err != nil {
			t.Fatalf("tasklane.PushTask()#%d error %v, want nil", i, err)
		}
	}

	if err := tl.PushTask(&TestTask{wg, out, false}, tl.ShortestQueueIndex()); err != tasklane.ErrTimeout {
		t.Fatalf("tasklane.PushTask() error %v, want %v", err, tasklane.ErrTimeout)
	}

	cancel()
	go func() {
		for range out {
		}
	}()
	tl.Wait()
	close(out)
	if err := tl.PushTask(&TestTask{wg, out, false}, tl.ShortestQueueIndex()); err != context.Canceled {
		t.Fatalf("tasklane.PushTask() error %v, want %v", err, context.Canceled)
	}
}
