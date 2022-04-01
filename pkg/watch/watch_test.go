package watch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func processEventsHandler(events []fsnotify.Event, _ WatchParameters, _ *fsnotify.Watcher, out io.Writer) {
	fmt.Fprintf(out, "processing %d events\n", len(events))
}

func cleanupHandler(_ WatchParameters, out io.Writer) error {
	fmt.Fprintf(out, "cleanup done\n")
	return nil
}

func Test_forLoopFunc(t *testing.T) {
	watcher, _ := fsnotify.NewWatcher()
	type args struct {
		ctx                  context.Context
		watcher              *fsnotify.Watcher
		parameters           WatchParameters
		processEventsHandler processEventsFunc
		cleanupHandler       cleanupFunc
	}
	tests := []struct {
		name          string
		args          args
		wantOut       string
		wantErr       bool
		watcherEvents []fsnotify.Event
		watcherErrors []error
	}{
		{
			name: "Case 1: Multiple events, no errors",
			args: args{
				ctx:                  nil,
				watcher:              watcher,
				parameters:           WatchParameters{},
				processEventsHandler: processEventsHandler,
				cleanupHandler:       cleanupHandler,
			},
			wantOut:       "processing 2 events\n",
			wantErr:       false,
			watcherEvents: []fsnotify.Event{{Name: "event1", Op: 1}, {Name: "event2", Op: 2}},
			watcherErrors: nil,
		},
		{
			name: "Case 2: Multiple events, one error",
			args: args{
				ctx:                  nil,
				watcher:              watcher,
				parameters:           WatchParameters{},
				processEventsHandler: processEventsHandler,
				cleanupHandler:       cleanupHandler,
			},
			wantOut:       "",
			wantErr:       true,
			watcherEvents: []fsnotify.Event{{Name: "event1", Op: 1}, {Name: "event2", Op: 2}},
			watcherErrors: []error{fmt.Errorf("error")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cancel context.CancelFunc
			tt.args.ctx, cancel = context.WithCancel(context.Background())
			out := &bytes.Buffer{}

			go func() {
				err := forLoopFunc(tt.args.ctx, tt.args.watcher, tt.args.parameters, out, tt.args.processEventsHandler, tt.args.cleanupHandler)
				if (err != nil) != tt.wantErr {
					t.Errorf("forLoopFunc() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}()

			for i := range tt.watcherEvents {
				watcher.Events <- tt.watcherEvents[i]
			}

			for i := range tt.watcherErrors {
				watcher.Errors <- tt.watcherErrors[i]
			}

			<-time.After(300 * time.Millisecond)
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("forLoopFunc() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			out.Reset()
			cancel()
		})
	}
}
