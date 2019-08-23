package push

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	slashpath "path"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// rmAll is a set of standard Unix glob patterns that match all
// files, including hidden files starting with ".", except for the
// special file names "." and "..".
var rmAll = []string{"*", ".[!.]*", "..?*"}

type actionType byte

const (
	doCopy = iota
	doRemove
	doRemoveAll
)

type action struct {
	Type actionType
	Path string
	Info os.FileInfo
}

// Remote manages remote processes to execute file copies/deletions.
// Paths in Remote methods are local, they are adjusted automatically.
type Remote interface {
	// Execute actions until channel closes.
	Run(actions <-chan action) error
	// Cancel the Remote, no-op if Run() already called.
	Cancel()
}

// CmdFunc runs a command with the given IO streams.
//
// In odo we use the occlient remote command, for unit tests we d use
// a local exec.Cmd.
type CmdFunc = func(cmd []string, stdin io.Reader, stdout, stderr io.Writer) error

type unixRemote struct {
	FileSet
	prefix          string
	run             CmdFunc
	remoteDir       string
	remoteExtraDirs []string
	tarx, xrm       io.WriteCloser
	tw              *tar.Writer
	done            chan error
	buf             [32 * 1024]byte // Same size as io.Copy
}

// UnixRemote uses Unix tar and rm commands on the remote.
// Remove commands are also run in optional extraDirs.
func UnixRemote(run CmdFunc, fs FileSet, remoteDir string, extraDirs ...string) (Remote, error) {
	rmt := &unixRemote{
		FileSet:         fs,
		run:             run,
		remoteDir:       remoteDir,
		remoteExtraDirs: extraDirs,
		done:            make(chan error),
	}
	rmt.prefix, _ = filepath.Split(rmt.Root)
	var err error
	defer func() {
		if err != nil {
			rmt.Cancel()
		}
	}()
	rmt.tarx, err = rmt.start([]string{"tar", "xf", "-", "-C", remoteDir})
	if err != nil {
		return nil, err
	}
	rmt.xrm, err = rmt.start([]string{"xargs", "-0", "rm", "-rf"})
	if err != nil {
		return nil, err
	}
	rmt.tw = tar.NewWriter(rmt.tarx)
	return rmt, nil
}

func (rmt *unixRemote) start(cmd []string) (io.WriteCloser, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	go func() {
		var stderr bytes.Buffer // Capture error message
		var err error
		if rmt.run(cmd, r, nil, &stderr) != nil {
			err = fmt.Errorf("%v: %v", cmd, strings.TrimSpace(stderr.String()))
		}
		rmt.done <- err
	}()
	return w, nil
}

func (rmt *unixRemote) stop() error {
	for _, c := range []io.Closer{rmt.tw, rmt.xrm, rmt.tarx} {
		_ = c.Close()
	}
	return <-rmt.done
}

func (rmt *unixRemote) Cancel() { _ = rmt.stop() }

func (rmt *unixRemote) Run(actions <-chan action) (err error) {
	defer func() {
		if err2 := rmt.stop(); err == nil {
			err = err2
		}
	}()
	for a := range actions {
		var err error
		switch a.Type {
		case doRemoveAll:
			glog.V(4).Infof("Push remove all")
			err = rmt.rm(rmAll...)
		case doRemove:
			err = rmt.rm(a.Path)
		case doCopy:
			err = rmt.add(a.Path, a.Info)
		default:
			panic("internal error")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (rmt *unixRemote) rm(paths ...string) (err error) {
	defer func() { err = errors.WithStack(err) }()
	var null [1]byte
	dirs := append(rmt.remoteExtraDirs, rmt.remoteDir)
	// Write null-terminated paths for 'xargs -0' to avoid problems
	// with spaces or other shell-special characters in file names.
	for _, p := range paths {
		p = rmt.remote(p)
		for _, dir := range dirs {
			if _, err := io.WriteString(rmt.xrm, slashpath.Join(dir, p)); err != nil {
				return err
			}
			if _, err := rmt.xrm.Write(null[:]); err != nil {
				return err
			}
		}
	}
	return nil
}

// Add a tar entry.
func (rmt *unixRemote) add(path string, info os.FileInfo) (err error) {
	defer func() { err = errors.WithStack(err) }()
	var f *os.File
	if info.IsDir() || info.Mode().IsRegular() {
		if f, err = os.Open(path); err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
	}
	if info.IsDir() {
		// Only write a header for empty directories.
		// Full directories are inferred from their children.
		_, err := f.Readdir(1)
		if err == io.EOF {
			return rmt.writeHeader(path, info)
		}
		return nil
	}
	// Write a header for any other type of file
	if err := rmt.writeHeader(path, info); err != nil {
		return err
	}
	if info.Mode().IsRegular() {
		if _, err := io.CopyBuffer(rmt.tw, f, rmt.buf[:]); err != nil {
			return err
		}
	}
	return nil
}

func (rmt *unixRemote) writeHeader(path string, info os.FileInfo) (err error) {
	defer func() { err = errors.WithStack(err) }()
	link, err := rmt.linkTarget(path, info)
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	hdr.Name = rmt.remote(path)
	if info.IsDir() && !strings.HasSuffix(hdr.Name, "/") {
		hdr.Name += "/" // Directory needs trailing "/"
	}
	return rmt.tw.WriteHeader(hdr)
}

func (rmt *unixRemote) linkTarget(path string, info os.FileInfo) (link string, err error) {
	if info.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(path); err == nil {
			link = filepath.ToSlash(link)
		}
	}
	if err != nil {
		err = fmt.Errorf("bad symlink %#v -> %#v: %v", path, link, err)
	}
	return link, err
}
