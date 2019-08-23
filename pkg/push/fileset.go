package push

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

// normalize converts any path to a clean filepath path.
// We always use filepath paths when referring to local files.
func normalize(notNormalPath string) string {
	return filepath.FromSlash(filepath.Clean(notNormalPath))
}

// FileSet is a set of local files.
type FileSet struct {
	// Root directory of file set
	Root      string
	rootSlash string
	ignore    glob.Glob // Glob matching any of patterns
}

// NewFileSet returns a FileSet selecting all files under root,
// except those matching any of the glob.Glob patterns in ignore.
// Platform-native or "/" separators are allowed in glob patterns.
//
// NOTE: ignore patterns are interpreted relative to fs.Root,
// they should not include fs.Root.
func NewFileSet(root string, ignore []string) (fs FileSet, err error) {
	defer func() { err = errors.WithStack(err) }()
	fs = FileSet{Root: normalize(root)}
	fs.rootSlash = fs.Root
	if !strings.HasSuffix(fs.rootSlash, string(filepath.Separator)) {
		fs.rootSlash += string(filepath.Separator)
	}
	joined := make([]string, len(ignore))
	for i, p := range ignore {
		joined[i] = fs.join(p)
	}
	// A pattern like {pattern1,pattern2,...} will match any of the listed patterns.
	// See https://godoc.org/github.com/gobwas/glob#Compile
	matchAll := fmt.Sprintf("{%s}", strings.Join(joined, ","))
	fs.ignore, err = glob.Compile(matchAll, '/', '\\')
	return fs, err
}

func (fs *FileSet) remote(path string) string {
	return filepath.ToSlash(strings.TrimPrefix(path, fs.rootSlash))
}

func (fs *FileSet) join(path string) string {
	return filepath.Join(fs.Root, normalize(path))
}

// WalkFunc is the function for Walk.
type WalkFunc = func(path string, info os.FileInfo) error

// Walk files in FileSet, skip any that match Ignore() patterns.
// Ignores os.IsNotExist() errors.
func (fs FileSet) Walk(f WalkFunc) (err error) {
	defer func() { err = errors.WithStack(err) }()
	return filepath.Walk(fs.Root,
		func(path string, info os.FileInfo, err error) error {
			if fs.Root == path {
				return nil // Only walk children of root
			}
			if fs.ignore != nil && fs.ignore.Match(path) {
				if info != nil && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if err == nil {
				err = f(path, info)
			}
			return nilNotExist(err)
		})
}

func nilNotExist(err error) error {
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}
