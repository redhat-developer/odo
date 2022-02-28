package kclient

import (
	"fmt"
	"strings"
	"testing"
	time "time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func fakeEventStatus(podName string, eventWarningMessage string, count int32) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Type:    "Warning",
		Count:   count,
		Reason:  eventWarningMessage,
		Message: "Foobar",
	}
}

func TestCollectEvents(t *testing.T) {
	tests := []struct {
		name                string
		podName             string
		eventWarningMessage string
	}{
		{
			name:                "Case 1: Collect an arbitrary amount of events",
			podName:             "ruby",
			eventWarningMessage: "Fake event warning message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create a fake client
			fakeClient, fakeClientSet := FakeNew()
			fakeEventWatch := watch.NewRaceFreeFake()
			podSelector := fmt.Sprintf("deploymentconfig=%s", tt.podName)

			// Create a fake event status / watch reactor for faking the events we are collecting
			fakeEvent := fakeEventStatus(tt.podName, tt.eventWarningMessage, 10)
			go func(event *corev1.Event) {
				fakeEventWatch.Add(event)
			}(fakeEvent)

			fakeClientSet.Kubernetes.PrependWatchReactor("events", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fakeEventWatch, nil
			})

			events := make(map[string]corev1.Event)
			quit := make(chan int)
			go fakeClient.CollectEvents(podSelector, events, quit)

			// Sleep in order to make sure we actually collect some events
			time.Sleep(2 * time.Second)
			close(quit)

			// We make sure to lock in order to prevent race conditions when retrieving the events (since they are a pointer
			// by default since we pass in a map)
			mu.Lock()
			if len(events) == 0 {
				t.Errorf("Expected events, got none")
			}
			mu.Unlock()

			// Collect the first event in the map
			var firstEvent corev1.Event
			for _, val := range events {
				firstEvent = val
			}

			if !strings.Contains(firstEvent.Reason, tt.eventWarningMessage) {
				t.Errorf("expected warning message: '%s' in event message: '%+v'", tt.eventWarningMessage, firstEvent.Reason)
			}

		})
	}
}
