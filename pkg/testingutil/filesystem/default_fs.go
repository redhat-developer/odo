/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
	This package is a FORK of https://github.com/kubernetes/kubernetes/blob/master/pkg/util/filesystem/defaultfs.go
	See above license
*/

package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// DefaultFs implements Filesystem using same-named functions from "os" and "io"
type DefaultFs struct{}

var _ Filesystem = DefaultFs{}

// Stat via os.Stat
func (DefaultFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Create via os.Create
func (DefaultFs) Create(name string) (File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &defaultFile{file}, nil
}

// Open via os.Open
func (DefaultFs) Open(name string) (File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	return &defaultFile{file}, nil
}

// OpenFile via os.OpenFile
func (DefaultFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return &defaultFile{file}, nil
}

// Rename via os.Rename
func (DefaultFs) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// MkdirAll via os.MkdirAll
func (DefaultFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Chtimes via os.Chtimes
func (DefaultFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}

// RemoveAll via os.RemoveAll
func (DefaultFs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Remove via os.RemoveAll
func (DefaultFs) Remove(name string) error {
	return os.Remove(name)
}

// Getwd via os.Getwd
func (DefaultFs) Getwd() (dir string, err error) {
	return os.Getwd()
}

// ReadFile via ioutil.ReadFile
func (DefaultFs) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile via ioutil.WriteFile
func (DefaultFs) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// TempDir via ioutil.TempDir
func (DefaultFs) TempDir(dir, prefix string) (string, error) {
	return os.MkdirTemp(dir, prefix)
}

// TempFile via ioutil.TempFile
func (DefaultFs) TempFile(dir, prefix string) (File, error) {
	file, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return nil, err
	}
	return &defaultFile{file}, nil
}

// ReadDir via ioutil.ReadDir
func (DefaultFs) ReadDir(dirname string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	infos := make([]fs.FileInfo, 0, len(entries))
	var info os.FileInfo
	for _, entry := range entries {
		info, err = entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Walk via filepath.Walk
func (DefaultFs) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// Chmod via os.Chmod
func (f DefaultFs) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

// defaultFile implements File using same-named functions from "os"
type defaultFile struct {
	file *os.File
}

var _ File = (*defaultFile)(nil)

// Name via os.File.Name
func (file *defaultFile) Name() string {
	return file.file.Name()
}

// Write via os.File.Write
func (file *defaultFile) Write(b []byte) (n int, err error) {
	return file.file.Write(b)
}

// WriteString via File.WriteString
func (file *defaultFile) WriteString(s string) (int, error) {
	return file.file.WriteString(s)
}

// Sync via os.File.Sync
func (file *defaultFile) Sync() error {
	return file.file.Sync()
}

// Close via os.File.Close
func (file *defaultFile) Close() error {
	return file.file.Close()
}

func (file *defaultFile) Readdir(n int) ([]os.FileInfo, error) {
	return file.file.Readdir(n)
}

func (file *defaultFile) Read(b []byte) (n int, err error) {
	return file.file.Read(b)
}

func (file *defaultFile) Chmod(name string, mode os.FileMode) error {
	return file.file.Chmod(mode)
}
