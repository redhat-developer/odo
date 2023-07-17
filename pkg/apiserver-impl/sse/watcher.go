package sse

import (
	"context"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog"
)

func (n *Notifier) watchDevfileChanges(ctx context.Context, devfileFiles []string) error {
	devfileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				klog.V(2).Infof("context done, reason: %v", ctx.Err())
				cErr := devfileWatcher.Close()
				if cErr != nil {
					klog.V(2).Infof("error closing devfileWater: %v", cErr)
				}
				return
			case ev, ok := <-devfileWatcher.Events:
				if !ok {
					return
				}
				klog.V(7).Infof("event: %v", ev)
				devfileContent, rErr := n.fsys.ReadFile(n.devfilePath)
				if rErr != nil {
					klog.V(1).Infof("unable to read Devfile at path %q: %v", n.devfilePath, rErr)
					continue
				}
				n.eventsChan <- Event{
					eventType: DevfileUpdated,
					data: map[string]string{
						"path":      ev.Name,
						"operation": ev.Op.String(),
						"content":   string(devfileContent),
					},
				}
				if ev.Has(fsnotify.Remove) {
					// For some reason, depending on the editor used to edit the file, changes would be detected only once.
					// Workaround recommended is to re-add the path to the watcher.
					// See https://github.com/fsnotify/fsnotify/issues/363
					wErr := devfileWatcher.Remove(ev.Name)
					if wErr != nil {
						klog.V(7).Infof("error removing file watch: %v", wErr)
					}
					wErr = devfileWatcher.Add(ev.Name)
					if wErr != nil {
						klog.V(0).Infof("error re-adding file watch: %v", wErr)
					}
				}
			case wErr, ok := <-devfileWatcher.Errors:
				if !ok {
					return
				}
				klog.V(0).Infof("error on file watch: %v", wErr)
			}
		}
	}()

	for _, f := range devfileFiles {
		err = devfileWatcher.Add(f)
		if err != nil {
			klog.V(0).Infof("error adding watcher for path %q: %v", f, err)
		} else {
			klog.V(7).Infof("added watcher for path %q", f)
		}
	}

	return nil
}
