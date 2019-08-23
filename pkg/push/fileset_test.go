package push

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var walkAll = []string{
	"testdata/project/.odo",
	"testdata/project/.odo/odostuff.txt",
	"testdata/project/file0.link",
	"testdata/project/ignore-dir",
	"testdata/project/ignore-dir/file.txt",
	"testdata/project/ignore-file.txt",
	"testdata/project/stuff",
	"testdata/project/stuff/file0.txt",
	"testdata/project/stuff/morestuff",
	"testdata/project/stuff/morestuff/file1.txt",
	"testdata/project/stuff/morestuff/file2.txt",
	"testdata/project/stuff2",
	"testdata/project/stuff2/file10.txt",
	"testdata/project/stuff2/file11.txt",
	"testdata/project/stuff2/ignore-me.txt",
}

var walkIgnore = []string{
	"testdata/project/file0.link",
	"testdata/project/stuff",
	"testdata/project/stuff/file0.txt",
	"testdata/project/stuff/morestuff",
	"testdata/project/stuff/morestuff/file1.txt",
	"testdata/project/stuff/morestuff/file2.txt",
	"testdata/project/stuff2",
	"testdata/project/stuff2/file10.txt",
	"testdata/project/stuff2/file11.txt",
}

type fileList []string

func (x *fileList) Add(path string, info os.FileInfo) error {
	*x = append(*x, path)
	return nil
}

func (x *fileList) Get(fs FileSet) error { return fs.Walk(x.Add) }

func TestWalkAll(t *testing.T) {
	assert := assert.New(t)
	files := fileList{}
	fs, err := NewFileSet("testdata/project", nil)
	assert.NoError(err)
	assert.NoError(fs.Walk(files.Add))
	assert.Equal(walkAll, []string(files))
}

func TestWalkIgnore(t *testing.T) {
	assert := assert.New(t)
	files := fileList{}
	fs, err := NewFileSet("testdata/project", testIgnore)
	assert.NoError(err)
	assert.NoError(fs.Walk(files.Add))
	assert.Equal(walkIgnore, []string(files))
}

// File removed during walk should not cause an error.
func TestWalkRemoved(t *testing.T) {
	assert := assert.New(t)
	tmp, err := ioutil.TempDir("", t.Name())
	assert.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()
	// Make a directory hierarchy so we can delete files out from under the walk.
	assert.NoError(os.MkdirAll(filepath.Join(tmp, "a"), 0777))
	assert.NoError(os.MkdirAll(filepath.Join(tmp, "a", "a1"), 0777))
	assert.NoError(os.MkdirAll(filepath.Join(tmp, "a", "a2"), 0777))
	assert.NoError(os.MkdirAll(filepath.Join(tmp, "a", "a3"), 0777))
	fs, err := NewFileSet(tmp, nil)
	assert.NoError(err)
	err = fs.Walk(
		func(path string, _ os.FileInfo) error {
			switch path {
			case filepath.Join(tmp, "a", "a1"):
				return os.ErrNotExist
			case filepath.Join(tmp, "a", "a2"):
				assert.NoError(os.RemoveAll(filepath.Join(tmp, "a", "a3")))
				return nil
			case filepath.Join(tmp, "a", "a3"):
				assert.Fail("a3 should be gone")
			}
			return nil
		})
	assert.NoError(err)
}

func copyAll(dst, src string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dpath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if info.IsDir() {
			return os.MkdirAll(dpath, os.ModeDir|os.ModePerm)
		}
		sf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = sf.Close() }()
		df, err := os.Create(dpath)
		if err != nil {
			return err
		}
		defer func() { _ = df.Close() }()
		if _, err := io.Copy(df, sf); err != nil {
			return err
		}
		return nil
	})
}

func BenchmarkWalk(b *testing.B) {
	withBenchTemp(b, func(tf testFiles) {
		for n := 0; n < b.N; n++ {
			count := 0
			tf.fs.Walk(func(string, os.FileInfo) error { count++; return nil })
		}
	})
}
