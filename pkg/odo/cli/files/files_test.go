package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

func TestGetFilesGeneratedByOdo(t *testing.T) {
	type args struct {
		fsProvider func() filesystem.Filesystem
		rootDir    string
	}
	type test struct {
		name    string
		args    args
		setup   func() error
		wantErr bool
		want    []string
	}
	fakeFs := filesystem.NewFakeFs()
	for _, tt := range []test{
		{
			name: "error reading .odo/generated file",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						openFile: func(name string, flag int, perm os.FileMode) (filesystem.File, error) {
							return nil, errors.New("not implemented yet")
						},
					}
				},
				rootDir: "/path/to/root/directory/11",
			},
			wantErr: true,
			want:    []string{util.DotOdoDirectory},
		},
		{
			name: "non-existing .odo/generated file",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						openFile: func(name string, flag int, perm os.FileMode) (filesystem.File, error) {
							return nil, fmt.Errorf("no such file or directory: %w", fs.ErrNotExist)
						},
					}
				},
				rootDir: "/path/to/root/directory/12",
			},
			wantErr: false,
			want:    []string{util.DotOdoDirectory},
		},
		{
			name: "empty .odo/generated file",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir: "/path/to/root/directory/2",
			},
			setup: func() error {
				_, err := fakeFs.Create(filepath.Join("/path/to/root/directory/2", util.DotOdoDirectory, _dotOdoGenerated))
				return err
			},
			wantErr: false,
			want:    []string{util.DotOdoDirectory},
		},
		{
			name: ".odo/generated file with 1 line",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir: "/path/to/root/directory/3",
			},
			setup: func() error {
				return fakeFs.WriteFile(
					filepath.Join("/path/to/root/directory/3", util.DotOdoDirectory, _dotOdoGenerated),
					[]byte("devfile.yaml"),
					0644,
				)
			},
			wantErr: false,
			want:    []string{util.DotOdoDirectory, "devfile.yaml"},
		},
		{
			name: ".odo/generated file with multiple lines",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir: "/path/to/root/directory/4",
			},
			setup: func() error {
				return fakeFs.WriteFile(
					filepath.Join("/path/to/root/directory/4", util.DotOdoDirectory, _dotOdoGenerated),
					[]byte("devfile.yaml\n.gitignore\n"),
					0644,
				)
			},
			wantErr: false,
			want:    []string{util.DotOdoDirectory, "devfile.yaml", ".gitignore"},
		},
		{
			name: ".odo/generated file with blank lines that should be ignored",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir: "/path/to/root/directory/5",
			},
			setup: func() error {
				return fakeFs.WriteFile(
					filepath.Join("/path/to/root/directory/5", util.DotOdoDirectory, _dotOdoGenerated),
					[]byte("\n\ndevfile.yaml\n\n\t\n.gitignore\n\n/path/to/a/file\na-path with spaces\n"),
					0644,
				)
			},
			wantErr: false,
			want:    []string{util.DotOdoDirectory, "devfile.yaml", ".gitignore", "/path/to/a/file", "a-path with spaces"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				if err != nil {
					t.Errorf("error when setting up test: %v", err)
					return
				}
			}

			got, err := GetFilesGeneratedByOdo(tt.args.fsProvider(), tt.args.rootDir)

			if tt.wantErr != (err != nil) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetFilesGeneratedByOdo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReportLocalFileGeneratedByOdo(t *testing.T) {
	type args struct {
		fsProvider func() filesystem.Filesystem
		rootDir    string
		filename   string
	}
	type test struct {
		name        string
		args        args
		wantErr     bool
		wantContent string
	}
	fakeFs := filesystem.NewFakeFs()
	for _, tt := range []test{
		{
			name: "error when creating .odo directory",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						mkdirAll: func(path string, perm os.FileMode) error {
							return errors.New("not implemented yet")
						},
					}
				},
				rootDir:  "/path/to/root/directory/11",
				filename: "a-file",
			},
			wantErr: true,
		},
		{
			name: "error when opening .odo/generated file",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						mkdirAll: fakeFs.MkdirAll,
						openFile: func(name string, flag int, perm os.FileMode) (filesystem.File, error) {
							return nil, errors.New("not implemented yet")
						},
					}
				},
				rootDir:  "/path/to/root/directory/12",
				filename: "a-file",
			},
			wantErr: true,
		},
		{
			name: "related file path not related to root directory",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						mkdirAll: fakeFs.MkdirAll,
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir:  "/path/to/root/directory/13",
				filename: "../b/c",
			},
			wantErr:     false,
			wantContent: "../b/c\n",
		},
		{
			name: "relative file path",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						mkdirAll: fakeFs.MkdirAll,
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir:  "/path/to/root/directory/2",
				filename: "a-file",
			},
			wantContent: "a-file\n",
		},
		{
			name: "absolute file path not related to root directory",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						mkdirAll: fakeFs.MkdirAll,
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir:  "/path/to/root/directory/31",
				filename: "/b/c",
			},
			wantErr:     false,
			wantContent: "/b/c\n",
		},
		{
			name: "absolute file path",
			args: args{
				fsProvider: func() filesystem.Filesystem {
					return partialFs{
						mkdirAll: fakeFs.MkdirAll,
						openFile: fakeFs.OpenFile,
					}
				},
				rootDir:  "/path/to/root/directory/32",
				filename: filepath.Join("/path/to/root/directory/32", "a-path", "to", "some-file"),
			},
			wantContent: filepath.Join("a-path", "to", "some-file") + "\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := ReportLocalFileGeneratedByOdo(tt.args.fsProvider(), tt.args.rootDir, tt.args.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			got, err := fakeFs.ReadFile(filepath.Join(tt.args.rootDir, util.DotOdoDirectory, _dotOdoGenerated))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.wantContent != string(got) {
				t.Errorf("expected content=%q, got %q", tt.wantContent, got)
			}
		})
	}
}

type partialFs struct {
	mkdirAll func(path string, perm os.FileMode) error
	openFile func(name string, flag int, perm os.FileMode) (filesystem.File, error)
}

var _ filesystem.Filesystem = partialFs{}

func (p partialFs) Stat(name string) (os.FileInfo, error) {
	return nil, errors.New("not implemented yet")
}
func (p partialFs) Create(name string) (filesystem.File, error) {
	return nil, errors.New("not implemented yet")
}
func (p partialFs) Open(name string) (filesystem.File, error) {
	return nil, errors.New("not implemented yet")
}
func (p partialFs) OpenFile(name string, flag int, perm os.FileMode) (filesystem.File, error) {
	return p.openFile(name, flag, perm)
}
func (p partialFs) Rename(oldpath, newpath string) error {
	return errors.New("not implemented yet")
}
func (p partialFs) MkdirAll(path string, perm os.FileMode) error {
	return p.mkdirAll(path, perm)
}
func (p partialFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return errors.New("not implemented yet")
}
func (p partialFs) RemoveAll(path string) error {
	return errors.New("not implemented yet")
}
func (p partialFs) Remove(name string) error {
	return errors.New("not implemented yet")
}
func (p partialFs) Chmod(name string, mode os.FileMode) error {
	return errors.New("not implemented yet")
}
func (p partialFs) Getwd() (dir string, err error) {
	return "", errors.New("not implemented yet")
}
func (p partialFs) ReadFile(filename string) ([]byte, error) {
	return nil, errors.New("not implemented yet")
}
func (p partialFs) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return errors.New("not implemented yet")
}
func (p partialFs) TempDir(dir, prefix string) (string, error) {
	return "", errors.New("not implemented yet")
}
func (p partialFs) TempFile(dir, prefix string) (filesystem.File, error) {
	return nil, errors.New("not implemented yet")
}
func (p partialFs) ReadDir(dirname string) ([]os.FileInfo, error) {
	return nil, errors.New("not implemented yet")
}
func (p partialFs) Walk(root string, walkFn filepath.WalkFunc) error {
	return errors.New("not implemented yet")
}
