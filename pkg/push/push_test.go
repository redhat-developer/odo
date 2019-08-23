package push

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testIgnore = []string{"ignore*", "**/ignore*", ".odo", ".git"}

func TestForce(t *testing.T) {
	assert := assert.New(t)
	tf := makeTestFiles(t, "testdata/project")
	defer tf.TearDown()

	// invalid index should be ignored by force and updated at the end.
	assert.NoError(ioutil.WriteFile(tf.fs.join(indexPath), []byte("xxxinvalidxxx"), 0777))
	rmt := &testRemote{}
	p := ForcePusher(tf.fs)
	defer p.Cancel()
	changes, err := p.Push(rmt)
	assert.NoError(err)
	assert.True(changes)
	assert.Equal(trimDirs(walkIgnore, "testdata"), trimDirs(rmt.copied, tf.tmp))
	assert.Equal(rmAll, rmt.removed)

	// The index file should be up-to-date
	indexFile := tf.fs.join(indexPath)
	want := NewIndex(indexFile)
	got := NewIndex(indexFile)
	assert.NoError(got.Load())
	assert.NoError(want.Update(tf.fs))
	assert.Equal(want, got)
}

func TestExact(t *testing.T) {
	assert := assert.New(t)
	tf := makeTestFiles(t, "testdata/project")
	defer tf.TearDown()
	changed := []string{tf.fs.join("stuff/file0.txt"), tf.fs.join("stuff2")}
	removed := []string{tf.fs.join("stuff/morestuff")}
	rmt := &testRemote{}
	p := ExactPusher(tf.fs, changed, removed)
	defer p.Cancel()
	changes, err := p.Push(rmt)
	assert.NoError(err)
	assert.True(changes)
	assert.Equal(changed, rmt.copied)
	assert.Equal(removed, rmt.removed)
}

func TestIndex(t *testing.T) {
	assert := assert.New(t)
	tf := makeTestFiles(t, "testdata/project")
	defer tf.TearDown()

	// Force push to create initial files.
	p := ForcePusher(tf.fs)
	defer p.Cancel()
	changes, err := p.Push(&testRemote{})
	assert.NoError(err)
	assert.True(changes)

	// Modify some local files, verify only changes are pushed.

	// Note: Set file times to the past, we only check for equality not
	// order.  File system time resolution can be very coarse, and we
	// just copied the files, so using time.Now might not actually change
	// the time. Tar clutters test output with warnings if file
	// times are in the future, so lets go waaaay back.
	ts := time.Now().Add(-time.Hour)
	assert.NoError(os.Chtimes(tf.fs.join("stuff/file0.txt"), ts, ts))
	assert.NoError(os.Chtimes(tf.fs.join("stuff"), ts, ts))
	// This file is ignored so should not be copied even if changed.
	assert.NoError(os.Chtimes(tf.fs.join("ignore-file.txt"), ts, ts))

	// Change file size by appending data
	f, err := os.OpenFile(tf.fs.join("stuff2/file10.txt"), os.O_WRONLY|os.O_APPEND, 0644)
	assert.NoError(err)
	_, err = f.WriteString(" - more data")
	assert.NoError(err)
	assert.NoError(f.Close())

	// Remove directory. NOTE: implicitly changes the mtime of the parent directory.
	assert.NoError(os.RemoveAll(tf.fs.join("stuff/morestuff")))

	// IndexPusher.Push to dummy remote, verify we only sent changes.
	rmt := &testRemote{}
	p = IndexPusher(tf.fs)
	defer p.Cancel()
	changes, err = p.Push(rmt)
	assert.NoError(err)
	assert.True(changes)
	sort.Strings(rmt.copied)
	want := []string{"project/stuff", "project/stuff/file0.txt", "project/stuff2/file10.txt"}
	assert.Equal(want, trimDirs(rmt.copied, tf.tmp))

	want = []string{
		"project/stuff/morestuff",
		"project/stuff/morestuff/file1.txt",
		"project/stuff/morestuff/file2.txt",
	}
	sort.Strings(rmt.removed)
	assert.Equal(want, trimDirs(rmt.removed, tf.tmp))

	// A second push should do nothing, the index should be up-to-date.
	rmt = &testRemote{}
	p = IndexPusher(tf.fs)
	defer p.Cancel()
	changes, err = p.Push(rmt)
	assert.NoError(err)
	assert.False(changes)
	assert.Empty(rmt.copied)
	assert.Empty(rmt.removed)
}

func trimDir(path, dir string) string {
	path, dir = filepath.FromSlash(path), filepath.FromSlash(dir)
	if path == dir {
		return ""
	}
	if !strings.HasSuffix(dir, string(filepath.Separator)) {
		dir += string(filepath.Separator)
	}
	return strings.TrimPrefix(path, dir)
}

func trimDirs(paths []string, dir string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = trimDir(p, dir)
	}
	return out
}

type testFiles struct {
	fs          FileSet
	tmp, remote string
}

func makeTestFiles(t testing.TB, originDir string) testFiles {
	require := require.New(t)
	t.Helper()
	tmp, err := ioutil.TempDir("", t.Name())
	require.NoError(err)
	fs, err := NewFileSet(filepath.Join(tmp, filepath.Base(originDir)), testIgnore)
	require.NoError(err)
	require.NoError(copyAll(fs.Root, originDir))

	tf := testFiles{fs: fs, tmp: tmp}
	tf.remote = filepath.Join(tf.tmp, "remote")
	require.NoError(os.MkdirAll(tf.remote, 0777))
	return tf
}

func (tf *testFiles) TearDown() { _ = os.RemoveAll(tf.tmp) }

type testRemote struct {
	copied  []string
	removed []string
}

func (r *testRemote) Run(actions <-chan action) error {
	for a := range actions {
		switch a.Type {
		case doRemoveAll:
			r.removed = append(r.removed, rmAll...)
		case doRemove:
			r.removed = append(r.removed, a.Path)
		case doCopy:
			r.copied = append(r.copied, a.Path)
		default:
			panic("internal error")
		}
	}
	return nil
}

func (r *testRemote) Cancel() {}
