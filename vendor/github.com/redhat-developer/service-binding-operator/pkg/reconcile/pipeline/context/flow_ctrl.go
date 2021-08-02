package context

import "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"

type flowCtrl struct {
	retry bool
	stop  bool
	err   error
}

func (i *flowCtrl) FlowStatus() pipeline.FlowStatus {
	return pipeline.FlowStatus{
		Retry: i.retry,
		Stop:  i.stop,
		Err:   i.err,
	}
}

func (i *flowCtrl) RetryProcessing(reason error) {
	i.retry = true
	i.stop = true
	i.err = reason
}

func (i *flowCtrl) Error(err error) {
	i.err = err
}

func (i *flowCtrl) StopProcessing() {
	i.stop = true
}
