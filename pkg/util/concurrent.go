package util

import (
	"sync"
)

// A task to execute in a go-routine
type ConcurrentTask struct {
	ToRun func(errChannel chan error)
}

// run encapsulates the work to be done by calling the ToRun function
func (ct ConcurrentTask) run(errChannel chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	ct.ToRun(errChannel)
}

// Records tasks to be run concurrently with go-routines
type ConcurrentTasks struct {
	tasks []ConcurrentTask
}

// NewConcurrentTasks creates a new ConcurrentTasks instance, dimensioned to accept at least the specified number of tasks
func NewConcurrentTasks(taskNumber int) *ConcurrentTasks {
	return &ConcurrentTasks{tasks: make([]ConcurrentTask, 0, taskNumber)}
}

// Add adds the specified ConcurrentTask to the list of tasks to be run concurrently
func (ct *ConcurrentTasks) Add(task ConcurrentTask) {
	if len(ct.tasks) == 0 {
		ct.tasks = make([]ConcurrentTask, 0, 7)
	}
	ct.tasks = append(ct.tasks, task)
}

// Run concurrently runs the added tasks failing on the first error
// Based on https://garrypolley.com/2016/02/10/golang-routines-errors/
func (ct *ConcurrentTasks) Run() error {
	var wg sync.WaitGroup
	finished := make(chan bool, 1) // this along with wg.Wait() is why the error handling works and doesn't deadlock
	errChannel := make(chan error)

	for _, task := range ct.tasks {
		wg.Add(1)
		go task.run(errChannel, &wg)
	}

	// Put the wait group in a go routine.
	// By putting the wait group in the go routine we ensure either all pass
	// and we close the "finished" channel or we wait forever for the wait group
	// to finish.
	//
	// Waiting forever is okay because of the blocking select below.
	go func() {
		wg.Wait()
		close(finished)
	}()

	// This select will block until one of the two channels returns a value.
	// This means on the first failure in the go routines above the errChannel will release a
	// value first. Because there is a "return" statement in the err check this function will
	// exit when an error occurs.
	//
	// Due to the blocking on wg.Wait() the finished channel will not get a value unless all
	// the go routines before were successful because not all the wg.Done() calls would have
	// happened.
	select {
	case <-finished:
	case err := <-errChannel:
		if err != nil {
			return err
		}
	}

	return nil
}
