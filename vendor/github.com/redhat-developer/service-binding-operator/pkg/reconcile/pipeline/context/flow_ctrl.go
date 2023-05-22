package context

import (
	"time"

	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
)

type flowCtrl struct {
	retry bool
	stop  bool
	err   error
	delay time.Duration
}

func (i *flowCtrl) FlowStatus() pipeline.FlowStatus {
	return pipeline.FlowStatus{
		Retry: i.retry,
		Stop:  i.stop,
		Err:   i.err,
		Delay: i.delay,
	}
}

func (i *flowCtrl) RetryProcessing(reason error) {
	i.retry = true
	i.stop = true
	i.err = reason
	i.delay = 0
}

func (i *flowCtrl) RetryProcessingWithDelay(reason error, delay time.Duration) {
	i.retry = true
	i.stop = reason != nil // for label selection
	i.err = reason
	i.delay = delay
}

func (i *flowCtrl) Error(err error) {
	i.err = err
}

func (i *flowCtrl) StopProcessing() {
	i.stop = true
}
