package tasklane_test

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/whoisnian/glb/tasklane"
)

type ExampleTask struct {
	wg      *sync.WaitGroup
	content string
}

func (task *ExampleTask) Start() {
	defer task.wg.Done()
	fmt.Println(task.content)
}

func Example() {
	wg := new(sync.WaitGroup)
	tl := tasklane.New(context.Background(), 2, 2)

	index := tl.ShortestQueueIndex()
	taskCnt := 10
	wg.Add(taskCnt)
	for i := 0; i < taskCnt; i++ {
		if err := tl.PushTask(&ExampleTask{wg, "start task " + strconv.Itoa(i)}, index); err != nil {
			panic(err)
		}
	}
	wg.Wait()

	// Unordered output:
	// start task 0
	// start task 1
	// start task 2
	// start task 3
	// start task 4
	// start task 5
	// start task 6
	// start task 7
	// start task 8
	// start task 9
}
