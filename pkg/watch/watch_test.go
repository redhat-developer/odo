package watch

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/watch"

	"github.com/fsnotify/fsnotify"

	"github.com/redhat-developer/odo/pkg/dev"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
)

func evaluateChangesHandler(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) ([]string, []string) {
	var changedFiles []string
	var deletedPaths []string

	for _, event := range events {
		for _, file := range fileIgnores {
			// if file is in fileIgnores, don't do anything
			if event.Name == file {
				continue
			}
		}
		// this if condition is copied from original implementation in watch.go. Code within the if block is simplified for tests
		if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
			// On remove/rename, stop watching the resource
			_ = watcher.Remove(event.Name)
			// Append the file to list of deleted files
			deletedPaths = append(deletedPaths, event.Name)
			continue
		}
		// this if condition is copied from original implementation in watch.go. Code within the if block is simplified for tests
		if event.Op&fsnotify.Remove != fsnotify.Remove {
			changedFiles = append(changedFiles, event.Name)
		}
	}
	return changedFiles, deletedPaths
}

func processEventsHandler(ctx context.Context, params WatchParameters, changedFiles, deletedPaths []string, componentStatus *ComponentStatus) error {
	fmt.Fprintf(params.StartOptions.Out, "changedFiles %s deletedPaths %s\n", changedFiles, deletedPaths)
	return nil
}

type fakeWatcher struct{}

func (o fakeWatcher) Stop() {
}

func (o fakeWatcher) ResultChan() <-chan watch.Event {
	return make(chan watch.Event, 1)
}

func Test_eventWatcher(t *testing.T) {
	type args struct {
		parameters WatchParameters
	}
	tests := []struct {
		name          string
		args          args
		wantOut       string
		wantErr       bool
		watcherEvents []fsnotify.Event
		watcherError  error
	}{
		{
			name: "Case 1: Multiple events, no errors",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "Pushing files...\n\nchangedFiles [file1 file2] deletedPaths []\n",
			wantErr:       false,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Create}, {Name: "file2", Op: fsnotify.Write}},
			watcherError:  nil,
		},
		{
			name: "Case 2: Multiple events, one error",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "",
			wantErr:       true,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Create}, {Name: "file2", Op: fsnotify.Write}},
			watcherError:  fmt.Errorf("error"),
		},
		{
			name: "Case 3: Delete file, no error",
			args: args{
				parameters: WatchParameters{
					StartOptions: dev.StartOptions{
						IgnorePaths: []string{"file1"},
					},
				},
			},
			wantOut:       "Pushing files...\n\nchangedFiles [] deletedPaths [file1 file2]\n",
			wantErr:       false,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Remove}, {Name: "file2", Op: fsnotify.Rename}},
			watcherError:  nil,
		},
		{
			name: "Case 4: Only errors",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "",
			wantErr:       true,
			watcherEvents: nil,
			watcherError:  fmt.Errorf("error1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, _ := fsnotify.NewWatcher()
			fileWatcher, _ := fsnotify.NewWatcher()
			var cancel context.CancelFunc
			ctx, cancel := context.WithCancel(context.Background())
			ctx = odocontext.WithDevfilePath(ctx, "/path/to/devfile")
			ctx = odocontext.WithApplication(ctx, "odo")
			ctx = odocontext.WithComponentName(ctx, "my-component")
			out := &bytes.Buffer{}

			go func() {
				for _, event := range tt.watcherEvents {
					watcher.Events <- event
				}

				if tt.watcherError != nil {
					watcher.Errors <- tt.watcherError
				}
				<-time.After(500 * time.Millisecond)
				cancel()
			}()

			componentStatus := ComponentStatus{}
			componentStatus.SetState(StateReady)

			o := WatchClient{
				sourcesWatcher:    watcher,
				deploymentWatcher: fakeWatcher{},
				podWatcher:        fakeWatcher{},
				warningsWatcher:   fakeWatcher{},
				devfileWatcher:    fileWatcher,
				keyWatcher:        make(chan byte),
			}
			tt.args.parameters.StartOptions.Out = out

			err := o.eventWatcher(ctx, tt.args.parameters, evaluateChangesHandler, processEventsHandler, componentStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventWatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("eventWatcher() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
