package push

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// indexPath is the relative path to the index file
const indexPath = ".odo/odo-file-index.json"

// fileInfo holds file size and mod time to test for changes.
// A subset of os.FileInfo
type fileInfo struct {
	Size    int64
	ModTime time.Time
}

func makeFileInfo(fi os.FileInfo) fileInfo {
	return fileInfo{Size: fi.Size(), ModTime: fi.ModTime()}
}

// Index is a map of paths to size/modtime structs.
//
// Internally paths are made relative to the location
// of the index file, but they are adjusted by Get
// and Put so the caller doesn't need to consider that.
type Index struct {
	index     map[string]fileInfo
	path, dir string // Location of index file
}

// NewIndex a new empty index associated with path, but do not load or save.
func NewIndex(path string) *Index {
	i := Index{index: map[string]fileInfo{}}
	i.path = path
	i.dir = filepath.Dir(path)
	return &i
}

func (i *Index) put(path string, info os.FileInfo) (err error) {
	defer func() { err = errors.Wrapf(err, "Index.Put %v", path) }()
	path, err = filepath.Rel(i.dir, path)
	i.index[path] = makeFileInfo(info)
	return err
}

func (i *Index) get(path string) (info fileInfo, ok bool) {
	path, err := filepath.Rel(i.dir, path)
	if err != nil {
		return info, false
	}
	info, ok = i.index[path]
	return info, ok
}

func (i *Index) clear()         { i.index = map[string]fileInfo{} }
func (i *Index) remove() error  { return os.RemoveAll(i.path) }
func (i *Index) String() string { return i.path }

// Update from a file set.
func (i *Index) Update(fs FileSet) (err error) {
	defer func() { err = errors.Wrapf(err, "Updating index from %v", fs.Root) }()
	i.clear()
	return fs.Walk(i.put)
}

// Load from file
func (i *Index) Load() (err error) {
	defer func() { err = errors.Wrapf(err, "Loading index %v", i.path) }()
	i.clear()
	f, err := os.Open(i.path)
	if err == nil {
		defer func() { _ = f.Close() }()
		err = json.NewDecoder(f).Decode(&i.index)
	}
	if err != nil {
		_ = i.remove()
	}
	return err
}

// Save to file
func (i *Index) Save() (err error) {
	defer func() { err = errors.Wrapf(err, "Saving index %v", i.path) }()
	if err := os.MkdirAll(i.dir, 0777); err != nil {
		return err
	}
	f, err := os.Create(i.path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if err = json.NewEncoder(f).Encode(&i.index); err != nil {
		_ = i.remove() // Don't leave possibly corrupt index file
	}
	return err
}

// Modified returns true if info does not match the index entry for path.
func (i *Index) Modified(path string, info os.FileInfo) bool {
	info2, ok := i.get(path)
	return !ok || info2 != makeFileInfo(info)
}
