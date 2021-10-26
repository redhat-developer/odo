package sync

import (
	taro "archive/tar"
	"bytes"
	"io"
	"path"
	"path/filepath"
	"testing"

	"github.com/openshift/odo/v2/pkg/testingutil/filesystem"
	"github.com/openshift/odo/v2/pkg/util"
)

func Test_linearTar(t *testing.T) {
	// FileType custom type to indicate type of file
	type FileType int

	const (
		// RegularFile enum to represent regular file
		RegularFile FileType = 0
		// Directory enum to represent directory
		Directory FileType = 1
	)

	fs := filesystem.NewFakeFs()

	type args struct {
		srcBase  string
		srcFile  string
		destBase string
		destFile string
		data     string
	}
	tests := []struct {
		name          string
		args          args
		fileType      FileType
		notExistError bool
		wantErr       bool
	}{
		{
			name: "case 1: write a regular file",
			args: args{
				srcBase:  filepath.Join("tmp", "dir1"),
				srcFile:  "red.js",
				destBase: filepath.Join("tmp1", "dir2"),
				destFile: "red.js",
				data:     "hi",
			},
			fileType: RegularFile,
			wantErr:  false,
		},
		{
			name: "case 2: write a folder",
			args: args{
				srcBase:  filepath.Join("tmp", "dir1"),
				srcFile:  "dir0",
				destBase: filepath.Join("tmp1", "dir2"),
				destFile: "dir2",
			},
			fileType: Directory,
			wantErr:  false,
		},
		{
			name: "case 3: file source doesn't exist",
			args: args{
				srcBase:  filepath.Join("tmp", "dir1"),
				srcFile:  "red.js",
				destBase: filepath.Join("tmp1", "dir2"),
				destFile: "red.js",
				data:     "hi",
			},
			fileType:      RegularFile,
			notExistError: true,
			wantErr:       true,
		},
		{
			name: "case 4: folder source doesn't exist",
			args: args{
				srcBase:  filepath.Join("tmp", "dir1"),
				srcFile:  "dir0",
				destBase: filepath.Join("tmp1", "dir2"),
				destFile: "dir2",
			},
			fileType:      Directory,
			notExistError: true,
			wantErr:       true,
		},
		{
			name: "case 5: dest is empty",
			args: args{
				srcBase:  filepath.Join("tmp", "dir1"),
				srcFile:  "dir0",
				destBase: "",
				destFile: "",
			},
			fileType: Directory,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := path.Join(tt.args.srcBase, tt.args.srcFile)

			if tt.fileType == RegularFile {
				f, err := fs.Create(filepath)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if _, err := io.Copy(f, bytes.NewBuffer([]byte(tt.args.data))); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				defer f.Close()
			} else {
				if err := fs.MkdirAll(filepath, 0755); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.notExistError == true {
				tt.args.srcBase += "blah"
			}

			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()

			tarWriter := taro.NewWriter(writer)

			go func() {
				defer tarWriter.Close()
				if err := linearTar(tt.args.srcBase, tt.args.srcFile, tt.args.destBase, tt.args.destFile, tarWriter, fs); (err != nil) != tt.wantErr {
					t.Errorf("linearTar() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()

			tarReader := taro.NewReader(reader)
			for {
				hdr, err := tarReader.Next()
				if err == io.EOF {
					break
				} else if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if hdr.Name != tt.args.destFile {
					t.Errorf("expected %q as destination filename, saw: %q", tt.args.destFile, hdr.Name)
				}
			}
		})
	}
}

func Test_makeTar(t *testing.T) {
	fs := filesystem.NewFakeFs()

	dir0, err := fs.TempDir("", "dir0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = fs.Create(filepath.Join(dir0, "red.js"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_, err = fs.Create(filepath.Join(dir0, "README.txt"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = fs.MkdirAll(filepath.Join(dir0, "views"), 0644)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_, err = fs.Create(filepath.Join(dir0, "views", "view.html"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	type args struct {
		srcPath  string
		destPath string
		files    []string
		globExps []string
		ret      util.IndexerRet
	}
	tests := []struct {
		name      string
		args      args
		wantFiles map[string]bool
		wantErr   bool
	}{
		{
			name: "case 1: normal tar making",
			args: args{
				srcPath:  dir0,
				destPath: filepath.Join("tmp", "dir1"),
				files: []string{
					filepath.Join(dir0, "red.js"),
					filepath.Join(dir0, "README.txt"),
					filepath.Join(dir0, "views"),
					filepath.Join(dir0, "views", "view.html")},
				globExps: []string{},
				ret: util.IndexerRet{
					NewFileMap: map[string]util.FileData{
						"red.js": {
							RemoteAttribute: "red.js",
						},
						"README.txt": {
							RemoteAttribute: "README.txt",
						},
						"views": {
							RemoteAttribute: "views",
						},
						filepath.Join("views", "view.html"): {
							RemoteAttribute: "views/view.html",
						},
					},
				},
			},
			wantFiles: map[string]bool{
				"red.js":          true,
				"views/view.html": true,
				"README.txt":      true,
			},
		},
		{
			name: "case 2: normal tar making with matching glob expression",
			args: args{
				srcPath:  dir0,
				destPath: filepath.Join("tmp", "dir1"),
				files: []string{
					filepath.Join(dir0, "red.js"),
					filepath.Join(dir0, "README.txt"),
					filepath.Join(dir0, "views"),
					filepath.Join(dir0, "views", "view.html")},
				globExps: []string{filepath.Join(dir0, "README.txt")},
				ret: util.IndexerRet{
					NewFileMap: map[string]util.FileData{
						"red.js": {
							RemoteAttribute: "red.js",
						},
						"README.txt": {
							RemoteAttribute: "README.txt",
						},
						"views": {
							RemoteAttribute: "views",
						},
						filepath.Join("views", "view.html"): {
							RemoteAttribute: "views/view.html",
						},
					},
				},
			},
			wantFiles: map[string]bool{
				"red.js":          true,
				"views/view.html": true,
			},
		},
		{
			name: "case 3: normal tar making different remote than local",
			args: args{
				srcPath:  dir0,
				destPath: filepath.Join("tmp", "dir1"),
				files: []string{
					filepath.Join(dir0, "red.js"),
					filepath.Join(dir0, "README.txt"),
					filepath.Join(dir0, "views"),
					filepath.Join(dir0, "views", "view.html")},
				globExps: []string{},
				ret: util.IndexerRet{
					NewFileMap: map[string]util.FileData{
						"red.js": {
							RemoteAttribute: "red.js",
						},
						"README.txt": {
							RemoteAttribute: "text/README.txt",
						},
						"views": {
							RemoteAttribute: "views",
						},
						filepath.Join("views", "view.html"): {
							RemoteAttribute: "views/view.html",
						},
					},
				},
			},
			wantFiles: map[string]bool{
				"red.js":          true,
				"views/view.html": true,
				"text/README.txt": true,
			},
		},
		{
			name: "case 4: ignore no existent file or folder",
			args: args{
				srcPath:  dir0,
				destPath: filepath.Join("tmp", "dir1"),
				files: []string{
					filepath.Join(dir0, "red.js"),
					filepath.Join(dir0, "README.txt"),
					filepath.Join("blah", "views"),
					filepath.Join(dir0, "views", "view.html")},
				globExps: []string{},
				ret: util.IndexerRet{
					NewFileMap: map[string]util.FileData{
						"red.js": {
							RemoteAttribute: "red.js",
						},
						"README.txt": {
							RemoteAttribute: "text/README.txt",
						},
						"views": {
							RemoteAttribute: "views",
						},
						filepath.Join("views", "view.html"): {
							RemoteAttribute: "views/view.html",
						},
					},
				},
			},
			wantFiles: map[string]bool{
				"red.js":          true,
				"views/view.html": true,
				"text/README.txt": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()

			tarWriter := taro.NewWriter(writer)
			go func() {
				defer tarWriter.Close()
				wantErr := tt.wantErr
				if err := makeTar(tt.args.srcPath, tt.args.destPath, writer, tt.args.files, tt.args.globExps, tt.args.ret, fs); (err != nil) != wantErr {
					t.Errorf("makeTar() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}()

			gotFiles := make(map[string]bool)
			tarReader := taro.NewReader(reader)
			for {
				hdr, err := tarReader.Next()
				if err == io.EOF {
					break
				} else if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if _, ok := tt.wantFiles[hdr.Name]; !ok {
					t.Errorf("unexpected file name in tar, : %q", hdr.Name)
				}

				gotFiles[hdr.Name] = true
			}

			for fileName := range tt.wantFiles {
				if _, ok := gotFiles[fileName]; !ok {
					t.Errorf("missed file, : %q", fileName)
				}
			}
		})
	}
}
