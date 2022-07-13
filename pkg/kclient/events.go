package kclient

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/log"
)

// We use a mutex here in order to make 100% sure that functions such as CollectEvents
// so that there are no race conditions
var mu sync.Mutex

const (
	failedEventCount = 5
)

// CollectEvents collects events in a Goroutine by manipulating a spinner.
// We don't care about the error (it's usually ran in a go routine), so erroring out is not needed.
func (c *Client) CollectEvents(selector string, events map[string]corev1.Event, quit <-chan int) {

	// Secondly, we will start a go routine for watching for events related to the pod and update our pod status accordingly.
	eventWatcher, err := c.KubeClient.CoreV1().Events(c.Namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Warningf("Unable to watch for events: %s", err)
		return
	}
	defer eventWatcher.Stop()

	// Create an endless loop for collecting
	for {
		select {
		case <-quit:
			klog.V(3).Info("Quitting collect events")
			return
		case val, ok := <-eventWatcher.ResultChan():
			mu.Lock()
			if !ok {
				log.Warning("Watch channel was closed")
				return
			}
			if e, ok := val.Object.(*corev1.Event); ok {

				// If there are many warning events happening during deployment, let's log them.
				if e.Type == "Warning" {

					if e.Count >= failedEventCount {
						newEvent := e
						(events)[e.Name] = *newEvent
						klog.V(3).Infof("Warning Event: Count: %d, Reason: %s, Message: %s", e.Count, e.Reason, e.Message)
					}

				}

			} else {
				log.Warning("Unable to convert object to event")
				return
			}
			mu.Unlock()
		}
	}
}

func (c *Client) PodWarningEventWatcher(ctx context.Context) (watch.Interface, error) {
	selector := "involvedObject.kind=Pod,involvedObject.apiVersion=v1,type=Warning"
	ns := c.GetCurrentNamespace()
	return c.GetClient().CoreV1().Events(ns).
		Watch(ctx, metav1.ListOptions{
			FieldSelector: selector,
		})
}
