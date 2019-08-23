package push

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// Pusher reads local file system to send files and deletions.
type Pusher interface {
	// Push sends changes to remote and closes the Pusher.
	// Returns true if any changes were sent.
	Push(Remote) (bool, error)
	// Cancel the Pusher, no-op if Push() already called.
	Cancel()
	FileSet() FileSet
}

// IndexPusher uses the index to push incremental changes.
// The index file is updated on success, removed on failure.
func IndexPusher(fs FileSet) Pusher {
	var p indexPusher
	p.init(fs)
	go func() {
		var err error
		var force bool
		if err = p.oldIndex.Load(); err != nil {
			glog.Warningf("Error loading index %v, using force: %v", p.oldIndex, err)
			force = true
		}
		p.done <- p.walk(force)
	}()
	return &p
}

// ForcePusher resets the remote project and pushes all local files.
// The index file is updated on success, removed on failure.
func ForcePusher(fs FileSet) Pusher {
	var p indexPusher
	p.init(fs)
	go func() { p.done <- p.walk(true) }()
	return &p
}

// ExactPusher returns a Pusher that pushes an exact set of file
// copies and deletes.  The index file is ignored.
func ExactPusher(fs FileSet, copyFiles, removeFiles []string) Pusher {
	var p pusher
	p.init(fs)
	go func() {
		err := p.copyFiles(copyFiles)
		if err == nil {
			err = p.removeFiles(removeFiles)
		}
		p.done <- err
	}()
	return &p
}

// Base implementation used by all pushers
type pusher struct {
	fs         FileSet
	actions    chan action
	sent       int
	done       chan error
	cancel     chan struct{}
	cancelOnce sync.Once
}

// How many actions to buffer in send channel.
const pushAhead = 1024

func (p *pusher) init(fs FileSet) {
	p.fs = fs
	p.actions = make(chan action, pushAhead)
	p.done = make(chan error)
	p.cancel = make(chan struct{})
}

func (p *pusher) send(act actionType, path string, info os.FileInfo) (err error) {
	defer func() { err = errors.WithStack(err) }()
	p.sent++
	a := action{act, path, info}
	select {
	case p.actions <- a:
		return nil
	case <-p.cancel:
		return errors.New("Cancelled")
	}
}

func (p *pusher) copyFiles(files []string) (err error) {
	defer func() { err = errors.WithStack(err) }()
	for _, f := range files {
		info, err := os.Lstat(f)
		if err == nil {
			glog.V(4).Infof("Push sending %v", f)
			err = p.send(doCopy, f, info)
		}
		if err = nilNotExist(err); err != nil {
			return err
		}
	}
	return nil
}

func (p *pusher) removeFiles(files []string) (err error) {
	defer func() { err = errors.WithStack(err) }()
	for _, f := range files {
		glog.V(4).Infof("Push removing %v", f)
		if err := nilNotExist(p.send(doRemove, f, nil)); err != nil {
			return err
		}
	}
	return nil
}

func (p *pusher) Cancel() { p.cancelOnce.Do(func() { close(p.cancel) }) }

func (p *pusher) FileSet() FileSet { return p.fs }

func (p *pusher) Push(rmt Remote) (modified bool, err error) {
	defer func() { err = errors.Wrapf(err, "Pushing %v", p.fs.Root) }()
	rmtDone := make(chan error)
	go func() {
		err := rmt.Run(p.actions)
		if err != nil {
			p.Cancel()
		}
		rmtDone <- err
	}()
	err = <-p.done
	close(p.actions)
	err2 := <-rmtDone
	if err2 != nil {
		err = err2
	}
	return p.sent > 0, err
}

// Used by IndexPusher and ForcePusher since both update the index.
type indexPusher struct {
	pusher
	oldIndex, newIndex *Index
}

func (p *indexPusher) init(fs FileSet) {
	p.pusher.init(fs)
	indexFile := p.fs.join(indexPath)
	p.oldIndex = NewIndex(indexFile)
	p.newIndex = NewIndex(indexFile)
}

func (p *indexPusher) Push(rmt Remote) (bool, error) {
	modified, err := p.pusher.Push(rmt)
	if err == nil {
		err = p.newIndex.Save() // Update the index on success.
	}
	return modified, err
}

// walk the file system pushing changes to actions.
func (p *indexPusher) walk(force bool) (err error) {
	glog.V(4).Infof("Push with index in %v force=%v", p.fs.Root, force)
	defer func() { err = errors.WithStack(err) }()
	// Remove out of date index, will update on successful Push
	_ = p.oldIndex.remove()
	if force { // Clear all remote files
		if err := p.send(doRemoveAll, "", nil); err != nil {
			return err
		}
		// TODO(alanconway) should also flush remote language-specific
		// packages caches (npm, maven etc.) etc. both to prevent unbounded
		// cache growth, and to completely reset to a from-scratch build to
		// avoid surprises caused by unexpected leftovers.
	}
	err = p.fs.Walk(func(path string, info os.FileInfo) error {
		if err := p.newIndex.put(path, info); err != nil {
			return err
		}
		if p.oldIndex.Modified(path, info) {
			glog.V(4).Infof("Push sending modified %v", path)
			if err := p.send(doCopy, path, info); err != nil {
				return err
			}
		} else {
			glog.V(4).Infof("Push skipping unmodified %v", path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for path := range p.oldIndex.index { // In old but not in new means deleted.
		if _, ok := p.newIndex.index[path]; !ok {
			path = filepath.Join(p.oldIndex.dir, path)
			glog.V(4).Infof("Push removing missing %v", path)
			if err := p.send(doRemove, path, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
