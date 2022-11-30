package watch

import "k8s.io/apimachinery/pkg/watch"

type NoOpWatcher struct{}

var _ watch.Interface = (*NoOpWatcher)(nil)

func NewNoOpWatcher() NoOpWatcher {
	return NoOpWatcher{}
}

func (o NoOpWatcher) Stop() {}

func (o NoOpWatcher) ResultChan() <-chan watch.Event {
	return make(chan watch.Event)
}
