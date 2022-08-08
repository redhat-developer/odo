package kclient

import (
	"context"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type NoOpWatch struct{}

func (o NoOpWatch) Stop() {}

func (o NoOpWatch) ResultChan() <-chan watch.Event {
	return make(chan watch.Event)
}

// PodWarningEventWatcher watch for events in the current directory. If the watch is forbidden, a NoOp
// implementation of watch.Interface is returned
func (c *Client) PodWarningEventWatcher(ctx context.Context) (result watch.Interface, isForbidden bool, err error) {
	selector := "involvedObject.kind=Pod,involvedObject.apiVersion=v1,type=Warning"
	ns := c.GetCurrentNamespace()
	result, err = c.GetClient().CoreV1().Events(ns).
		Watch(ctx, metav1.ListOptions{
			FieldSelector: selector,
		})

	if err != nil {
		if kerrors.IsForbidden(err) {
			return NoOpWatch{}, true, nil
		}
		return nil, false, err
	}
	return result, false, nil
}
