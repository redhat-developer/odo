package push

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRun(cmd []string, stdin io.Reader, stdout, stderr io.Writer) error {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Env = []string{}
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}

// Ensure all files in fs exist under remote and have matching FileInfo.
func assertFilesMatch(t testing.TB, fs FileSet, remote string) {
	t.Helper()
	assert := assert.New(t)
	assert.NoError(fs.Walk(
		func(path string, info os.FileInfo) error {
			other := filepath.Join(remote, trimDir(path, fs.Root))
			info2, err := os.Lstat(other)
			assert.NoError(err)
			if err != nil {
				return err
			}
			assert.Equal(info.Name(), info2.Name())
			assert.Equal(info.Mode(), info2.Mode())
			assert.Equal(info.IsDir(), info2.IsDir())
			if !info.IsDir() {
				assert.Equal(info.Size(), info2.Size())
			}
			return nil
		}))
}

func TestRemote(t *testing.T) {
	assert := assert.New(t)
	tt := makeTestFiles(t, "testdata/project")
	defer tt.TearDown()
	// Need to test empty dir, git won't allow it, so make one now.
	assert.NoError(os.MkdirAll(filepath.Join(tt.fs.Root, "empty"), 0777))
	extra := []string{filepath.Join(tt.tmp, "x0"), filepath.Join(tt.tmp, "x1")}
	rmt, err := UnixRemote(testRun, tt.fs, tt.remote, extra...)
	assert.NoError(err)
	p := IndexPusher(tt.fs)
	defer p.Cancel()
	changes, err := p.Push(rmt)
	assert.NoError(err)
	assert.True(changes)
	assertFilesMatch(t, tt.fs, tt.remote)

	// Make sure deletes propagate to extra dirs. Populate the extra dirs.
	for _, d := range extra {
		assert.NoError(os.MkdirAll(d, 0777))
		assert.NoError(copyAll(d, tt.remote))
		assert.FileExists(path.Join(d, "stuff/file0.txt"))
	}
	// Do the delete and make sure it is removed from from all dirs
	assert.NoError(os.Remove(path.Join(tt.fs.Root, "stuff/file0.txt")))
	p = IndexPusher(tt.fs)
	defer p.Cancel()
	rmt, err = UnixRemote(testRun, tt.fs, tt.remote, extra...)
	assert.NoError(err)
	changes, err = p.Push(rmt)
	assert.NoError(err)
	assert.True(changes)
	want := trimDirs(walkIgnore, "testdata/project")
	for i, p := range want {
		if p == "stuff/file0.txt" {
			want = append(want[:i], want[i+i:]...)
			break
		}
	}
	sort.Strings(want)
	for _, d := range append(extra, tt.remote) {
		assertFilesMatch(t, tt.fs, d)
	}
}

func withBenchTemp(b *testing.B, f func(tf testFiles)) {
	b.StopTimer()
	// Use odo/pkg as a bigger file set test for copying.
	_, src, _, ok := runtime.Caller(0)
	require.True(b, ok)
	tf := makeTestFiles(b, filepath.Dir(filepath.Dir(src)))
	defer func() {
		b.StopTimer()
		tf.TearDown()
		b.StartTimer()
	}()
	b.StartTimer()
	f(tf)
}

func benchPush(b *testing.B, p Pusher, tf testFiles) {
	defer p.Cancel()
	rmt, err := UnixRemote(testRun, tf.fs, tf.remote)
	require.NoError(b, err)
	_, err = p.Push(rmt)
	require.NoError(b, err)
}

func BenchmarkPushForce(b *testing.B) {
	withBenchTemp(b, func(tf testFiles) {
		for n := 0; n < b.N; n++ {
			benchPush(b, ForcePusher(tf.fs), tf)
		}
	})
}

func BenchmarkPushNoChange(b *testing.B) {
	withBenchTemp(b, func(tf testFiles) {
		b.StopTimer()
		benchPush(b, ForcePusher(tf.fs), tf) // Initial full push
		b.StartTimer()
		for n := 0; n < b.N; n++ {
			benchPush(b, IndexPusher(tf.fs), tf)
		}
	})
}

// Will do one big push then many small, on average should be quick
func BenchmarkPushOneChange(b *testing.B) {
	withBenchTemp(b, func(tf testFiles) {
		b.StopTimer()
		benchPush(b, ForcePusher(tf.fs), tf) // Initial full push
		b.StartTimer()
		for n := 0; n < b.N; n++ {
			path := filepath.Join(tf.fs.Root, "changeme.txt")
			err := ioutil.WriteFile(path, make([]byte, 1000+n%100), 0600)
			require.NoError(b, err)
			benchPush(b, IndexPusher(tf.fs), tf)
		}
	})
}
