package task

import (
	"fmt"
	"time"

	"k8s.io/klog"
)

// Retryable represents a task that can be retried.
type Retryable struct {
	// description of the task
	description string

	// errorIfTimeout indicates whether an error should be returned if the task does not complete successfully
	// within the given schedule.
	errorIfTimeout bool

	// runner is the actual function that is expected to be retried.
	runner Runner
}

// Runner is a function that will get invoked via RetryWithSchedule. If exitCondition is false, the function will get invoked again, until
// the given timeout schedule expires. It then returns a result of any type along with a potential error.
type Runner func() (exitCondition bool, result interface{}, err error)

// NewRetryable creates and returns a new Retryable task.
func NewRetryable(description string, runner Runner, errorIfTimeout bool) Retryable {
	return Retryable{
		description:    description,
		runner:         runner,
		errorIfTimeout: errorIfTimeout,
	}
}

// RetryWithSchedule invokes the retryable runner function, and keeps retrying until this runner returns an exitCondition that evaluates to false,
// or the given timeout expires. The timeout schedule can be a seen as a backoff schedule, in the sense that before recalling the runner function,
// RetryWithSchedule waits for each duration defined in the given schedule.
// If the exitCondition is not true after all retries, the behavior is governed by the errorIfTimeout flag.
// If errorIfTimeout is true, then an error is returned.
func (r Retryable) RetryWithSchedule(schedule ...time.Duration) (interface{}, error) {
	var err error
	var result interface{}
	if len(schedule) == 0 {
		_, result, err = r.runner()
		return result, err
	}

	var exitCondition bool
	var totalWaitTime float64
	for _, s := range schedule {
		seconds := s.Seconds()
		klog.V(3).Infof("waiting for %0.f second(s) before trying task %q", seconds, r.description)
		time.Sleep(s)
		totalWaitTime += seconds
		exitCondition, result, err = r.runner()
		if exitCondition {
			break
		}
	}

	if !exitCondition {
		msg := "aborted retrying task %q which is still not ok after %0.f second(s)"
		if r.errorIfTimeout {
			if err != nil {
				return result, fmt.Errorf(msg+": %w", r.description, totalWaitTime, err)
			}
			return result, fmt.Errorf(msg, r.description, totalWaitTime)
		}
		klog.V(3).Infof(msg, r.description, totalWaitTime)
	}

	return result, err
}
