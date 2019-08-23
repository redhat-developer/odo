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

func TestUpdateSaveLoad(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(err)
	defer func() { _ = os.RemoveAll(dir) }()
	project := filepath.Join(dir, "project")
	err = copyAll(project, "testdata/project")
	assert.NoError(err)

	indexPath := filepath.Join(project, "index", "index.json")
	i := NewIndex(indexPath)
	fs, err := NewFileSet(project, []string{"**/index*.json"})
	assert.NoError(err)
	assert.NoError(i.Update(fs))

	verify := func() {
		for _, path := range walkAll {
			path = filepath.Join(dir, strings.TrimPrefix(path, "testdata/"))
			info, ok := i.get(path)
			assert.True(ok, path)
			info2, err := os.Lstat(path)
			assert.NoError(err)
			assert.Equal(makeFileInfo(info2), info)
		}
	}
	verify()

	assert.NoError(i.Save())
	i.clear()
	assert.NoError(i.Load())
	verify()
}

func TestBadLoad(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dir) }()

	i := NewIndex("nosuchfile")
	assert.Error(t, i.Load())

	badFile := filepath.Join(dir, "bad.json")
	require.NoError(t, ioutil.WriteFile(badFile, []byte("invalid_json"), 0777))
	i = NewIndex(badFile)
	assert.Error(t, i.Load())
	// Failure to load should delete the file
	_, err = os.Lstat(badFile)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestWalk(t *testing.T) {
	assert := assert.New(t)
	fs, err := NewFileSet("testdata/project", nil)
	assert.NoError(err)

	// Build an index, and check that empty index considers everything modified.
	var empty Index
	index := NewIndex("")
	assert.NoError(fs.Walk(
		func(f string, fi os.FileInfo) error {
			assert.NoError(index.put(f, fi))
			assert.True(empty.Modified(f, fi), "%v, %v", f, fi)
			return nil
		}))

	// Verify the complete index matches the file system.
	modified := make([]string, 0, len(index.index))
	for path, info := range index.index {
		modified = append(modified, path)
		if fi, err := os.Lstat(path); err != nil {
			assert.Equal(makeFileInfo(fi), info)
		} else {
			assert.NoError(err)
		}
	}
	sort.Strings(modified)
	assert.Equal(walkAll, modified)

	// Make some index modifications, verify walk picks them up.
	info := index.index["testdata/project/stuff2/file10.txt"]
	info.Size += 100
	index.index["testdata/project/stuff2/file10.txt"] = info

	info = index.index["testdata/project/stuff2/file10.txt"]
	info.ModTime = info.ModTime.Add(time.Hour)
	index.index["testdata/project/stuff2/file11.txt"] = info

	modified = nil
	assert.NoError(fs.Walk(
		func(f string, fi os.FileInfo) error {
			if index.Modified(f, fi) {
				modified = append(modified, f)
			}
			return nil
		}))
	assert.Equal(
		[]string{"testdata/project/stuff2/file10.txt", "testdata/project/stuff2/file11.txt"},
		modified)
}
