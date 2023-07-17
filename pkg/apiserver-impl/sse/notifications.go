package sse

import (
	"context"
	"net/http"
	"sync"
	"time"

	"k8s.io/klog"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type Notifier struct {
	fsys filesystem.Filesystem

	devfilePath string

	// eventsChan is a channel for all events that will be broadcast to all subscribers.
	// Because it is not natively possible to read a same value twice from a Go channel,
	// we are storing the list of channels to broadcast the event to into the subscribers list.
	eventsChan chan Event

	// subscribers is a list of all channels where any event from eventsChan will be broadcast to.
	subscribers []chan Event

	// newSubscriptionChan is a channel where new subscribers can register the channel on which they wish to be notified.
	// Such channels are stored into the subscribers list.
	newSubscriptionChan chan chan Event

	// cancelSubscriptionChan is a write-only channel where subscribers can cancel their registration and stop being broadcast new events coming from eventsChan.
	cancelSubscriptionChan chan (<-chan Event)
}

func NewNotifier(ctx context.Context, fsys filesystem.Filesystem, devfilePath string, devfileFiles []string) (*Notifier, error) {
	notifier := Notifier{
		fsys:                   fsys,
		devfilePath:            devfilePath,
		eventsChan:             make(chan Event),
		subscribers:            make([]chan Event, 0),
		newSubscriptionChan:    make(chan chan Event),
		cancelSubscriptionChan: make(chan (<-chan Event)),
	}

	err := notifier.watchDevfileChanges(ctx, devfileFiles)
	if err != nil {
		return nil, err
	}

	go notifier.manageSubscriptions(ctx)

	// Heartbeat as a keep-alive mechanism to prevent some clients from closing inactive connections (notifications might not be sent regularly).
	go func() {
		ticker := time.NewTicker(7 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				notifier.eventsChan <- Event{
					eventType: Heartbeat,
				}
			}
		}
	}()

	return &notifier, nil
}

func (n *Notifier) manageSubscriptions(ctx context.Context) {
	defer func() {
		for _, listener := range n.subscribers {
			if listener != nil {
				close(listener)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case newSubscriber := <-n.newSubscriptionChan:
				n.subscribers = append(n.subscribers, newSubscriber)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case subscriberToRemove := <-n.cancelSubscriptionChan:
				for i, ch := range n.subscribers {
					if ch == subscriberToRemove {
						n.subscribers[i] = n.subscribers[len(n.subscribers)-1]
						n.subscribers = n.subscribers[:len(n.subscribers)-1]
						close(ch)
						break
					}
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case val, ok := <-n.eventsChan:
			if !ok {
				return
			}
			var wg sync.WaitGroup
			for _, subscriber := range n.subscribers {
				subscriber := subscriber
				if subscriber == nil {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case subscriber <- val:
					case <-ctx.Done():
						return
					}
				}()
			}
			wg.Wait()
		}
	}
}

func (n *Notifier) Routes() openapi.Routes {
	return openapi.Routes{
		{
			Name:        "ServerSentEvents",
			Method:      http.MethodGet,
			Pattern:     "/api/v1/notifications",
			HandlerFunc: n.handler,
		},
	}
}

func (n *Notifier) handler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	newListener := make(chan Event)
	n.newSubscriptionChan <- newListener
	defer func() {
		n.cancelSubscriptionChan <- newListener
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Headers sent back as early as possible to clients
	flusher.Flush()

	for {
		select {
		case ev := <-newListener:
			func() {
				defer flusher.Flush()
				dataToWrite, err := ev.toSseString()
				if err != nil {
					klog.V(2).Infof("error writing notification data: %v", err)
					return
				}
				_, err = w.Write([]byte(dataToWrite))
				if err != nil {
					klog.V(2).Infof("error writing notification data: %v", err)
					return
				}
			}()

		case <-r.Context().Done():
			klog.V(8).Infof("Connection closed!")
			return
		}
	}
}
