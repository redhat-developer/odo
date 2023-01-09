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
	This package is a FORK of https://github.com/kubernetes/kubernetes/blob/master/pkg/util/filesystem/fakefs.go
	See above license
*/

package filesystem

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// fakeFs is implemented in terms of afero
type fakeFs struct {
	a afero.Afero
}

// NewFakeFs returns a fake Filesystem that exists in-memory, useful for unit tests
func NewFakeFs() Filesystem {
	return &fakeFs{a: afero.Afero{Fs: afero.NewMemMapFs()}}
}

// Stat via afero.Fs.Stat
func (fs *fakeFs) Stat(name string) (os.FileInfo, error) {
	return fs.a.Fs.Stat(name)
}

// Create via afero.Fs.Create
func (fs *fakeFs) Create(name string) (File, error) {
	file, err := fs.a.Fs.Create(name)
	if err != nil {
		return nil, err
	}
	return &fakeFile{file}, nil
}

// Open via afero.Fs.Open
func (fs *fakeFs) Open(name string) (File, error) {
	file, err := fs.a.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &fakeFile{file}, nil
}

// OpenFile via afero.Fs.OpenFile
func (fs *fakeFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	file, err := fs.a.Fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &fakeFile{file}, nil
}

// Rename via afero.Fs.Rename
func (fs *fakeFs) Rename(oldpath, newpath string) error {
	return fs.a.Fs.Rename(oldpath, newpath)
}

// MkdirAll via afero.Fs.MkdirAll
func (fs *fakeFs) MkdirAll(path string, perm os.FileMode) error {
	return fs.a.Fs.MkdirAll(path, perm)
}

// Chtimes via afero.Fs.Chtimes
func (fs *fakeFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return fs.a.Fs.Chtimes(name, atime, mtime)
}

// ReadFile via afero.ReadFile
func (fs *fakeFs) ReadFile(filename string) ([]byte, error) {
	return fs.a.ReadFile(filename)
}

// WriteFile via afero.WriteFile
func (fs *fakeFs) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return fs.a.WriteFile(filename, data, perm)
}

// TempDir via afero.TempDir
func (fs *fakeFs) TempDir(dir, prefix string) (string, error) {
	return fs.a.TempDir(dir, prefix)
}

// TempFile via afero.TempFile
func (fs *fakeFs) TempFile(dir, prefix string) (File, error) {
	file, err := fs.a.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}
	return &fakeFile{file}, nil
}

// ReadDir via afero.ReadDir
func (fs *fakeFs) ReadDir(dirname string) ([]os.FileInfo, error) {
	return fs.a.ReadDir(dirname)
}

// Walk via afero.Walk
func (fs *fakeFs) Walk(root string, walkFn filepath.WalkFunc) error {
	return fs.a.Walk(root, walkFn)
}

// RemoveAll via afero.RemoveAll
func (fs *fakeFs) RemoveAll(path string) error {
	return fs.a.RemoveAll(path)
}

func (fs *fakeFs) Getwd() (dir string, err error) {
	return ".", nil
}

// Remove via afero.RemoveAll
func (fs *fakeFs) Remove(name string) error {
	return fs.a.Remove(name)
}

// Chmod via afero.Chmod
func (fs *fakeFs) Chmod(name string, mode os.FileMode) error {
	return fs.a.Chmod(name, mode)
}

// fakeFile implements File; for use with fakeFs
type fakeFile struct {
	file afero.File
}

// Name via afero.File.Name
func (file *fakeFile) Name() string {
	return file.file.Name()
}

// Write via afero.File.Write
func (file *fakeFile) Write(b []byte) (n int, err error) {
	return file.file.Write(b)
}

// WriteString via afero.File.WriteString
func (file *fakeFile) WriteString(s string) (n int, err error) {
	return file.file.WriteString(s)
}

// Sync via afero.File.Sync
func (file *fakeFile) Sync() error {
	return file.file.Sync()
}

// Close via afero.File.Close
func (file *fakeFile) Close() error {
	return file.file.Close()
}

func (file *fakeFile) Readdir(n int) ([]os.FileInfo, error) {
	return file.file.Readdir(n)
}

func (file *fakeFile) Read(b []byte) (n int, err error) {
	return file.file.Read(b)
}
