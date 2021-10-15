package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestCheckGitIgnoreFile(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a context directory on fake fs
	contextDir, err := fs.TempDir(os.TempDir(), "context")
	if err != nil {
		t.Error(err)
	}

	gitignorePath := filepath.Join(contextDir, ".gitignore")

	tests := []struct {
		testName        string
		create          bool
		gitIgnoreCreate func(create bool, contextDir string, fs filesystem.Filesystem) error
		directory       string
		want            string
		wantErr         bool
	}{
		{
			testName:        "Test when .gitignore does not exist",
			create:          true,
			gitIgnoreCreate: mockDirectoryInfo,
			directory:       contextDir,
			want:            gitignorePath,
			wantErr:         false,
		},
		{
			testName:        "Test when .gitignore exists",
			create:          false,
			gitIgnoreCreate: mockDirectoryInfo,
			directory:       contextDir,
			want:            gitignorePath,
			wantErr:         false,
		},
	}

	for _, tt := range tests {

		err := tt.gitIgnoreCreate(tt.create, tt.directory, fs)
		if err != nil {
			t.Error(err)
		}

		t.Run(tt.testName, func(t *testing.T) {

			gitIgnoreFilePath, err := checkGitIgnoreFile(tt.directory, fs)

			if tt.want != gitIgnoreFilePath {
				t.Errorf("checkGitIgnoreFile unexpected error %v, while creating .gitignore file", err)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("checkGitIgnoreFile unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestAddOdoFileIndex(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a context directory on fake fs
	contextDir, err := fs.TempDir(os.TempDir(), "context")
	if err != nil {
		t.Error(err)
	}

	gitignorePath := filepath.Join(contextDir, ".gitignore")

	tests := []struct {
		testName        string
		create          bool
		gitIgnoreCreate func(create bool, contextDir string, fs filesystem.Filesystem) error
		directory       string
		wantErr         bool
	}{
		{
			testName:        "Test when odo-file-index.json added to .gitignore",
			create:          false,
			gitIgnoreCreate: mockDirectoryInfo,
			directory:       gitignorePath,
			wantErr:         false,
		},
	}

	for _, tt := range tests {

		err := tt.gitIgnoreCreate(tt.create, tt.directory, fs)
		if err != nil {
			t.Error(err)
		}

		t.Run(tt.testName, func(t *testing.T) {

			err := addOdoFileIndex(tt.directory, fs)

			if !tt.wantErr == (err != nil) {
				t.Errorf("addOdoFileIndex unexpected error %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func mockDirectoryInfo(create bool, contextDir string, fs filesystem.Filesystem) error {

	if !create {
		err := fs.MkdirAll(filepath.Join(contextDir, ".gitignore"), os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestCalculateFileDataKeyFromPath(t *testing.T) {

	// create a temp dir for the fake component
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: error creating temporary directory for the indexer: %v", err)
	}

	tests := []struct {
		absolutePath   string
		rootDirectory  string
		expectedResult string
	}{
		{
			absolutePath:   filepath.Join(directory, "/path/file1"),
			rootDirectory:  filepath.Join(directory, "/path"),
			expectedResult: "file1",
		},
		{
			absolutePath:   filepath.Join(directory, "/path/path2/file1"),
			rootDirectory:  filepath.Join(directory, "/path/"),
			expectedResult: "path2/file1",
		},
		{
			absolutePath:   filepath.Join(directory, "/path"),
			rootDirectory:  filepath.Join(directory, "/"),
			expectedResult: "path",
		},
	}

	for _, tt := range tests {

		t.Run("Expect result: "+tt.expectedResult, func(t *testing.T) {

			result, err := CalculateFileDataKeyFromPath(tt.absolutePath, tt.rootDirectory)
			if err != nil {
				t.Fatalf("unexpecter error occurred %v", err)
			}

			if result != filepath.FromSlash(tt.expectedResult) {
				t.Fatalf("unexpected result: %v %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGenerateNewFileDataEntry(t *testing.T) {

	// create a temp dir for the fake component
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: error creating temporary directory for the indexer: %v", err)
	}

	tests := []struct {
		testName      string
		absolutePath  string
		rootDirectory string
		expectedKey   string
	}{
		{
			absolutePath:  filepath.Join(directory, "/path1/file1"),
			rootDirectory: filepath.Join(directory, "/path1"),
			expectedKey:   "file1",
		},
		{
			absolutePath:  filepath.Join(directory, "/path2/path2/file1"),
			rootDirectory: filepath.Join(directory, "/path2"),
			expectedKey:   "path2/file1",
		},
		{
			absolutePath:  filepath.Join(directory, "/path3"),
			rootDirectory: filepath.Join(directory, "/"),
			expectedKey:   "path3",
		},
	}

	for _, tt := range tests {

		t.Run("Expected key '"+tt.expectedKey+"'", func(t *testing.T) {

			if err := os.MkdirAll(filepath.Dir(tt.absolutePath), 0750); err != nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to create directories for %s: %v", tt.absolutePath, err)
			}

			if err := ioutil.WriteFile(tt.absolutePath, []byte("non-empty-string"), 0644); err != nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to write to index file path: %v", err)
			}

			key, filedata, err := GenerateNewFileDataEntry(tt.absolutePath, tt.rootDirectory)

			if err != nil {
				t.Fatalf("Unexpected error occurred %v", err)
			}

			// Keys are platform specific, so swap to forward slash for Windows before comparison
			key = strings.ReplaceAll(key, "\\", "/")

			if key != tt.expectedKey {
				t.Fatalf("Key %s did not match expected key %s", key, tt.expectedKey)
			}

			if filedata == nil {
				t.Fatalf("Filedata should not be null")
			}

			if filedata.Size == 0 || filedata.LastModifiedDate.IsZero() {
				t.Fatalf("Invalid filedata values %v %v", filedata.Size, filedata.LastModifiedDate)
			}

		})
	}
}

func createAndStat(fileName, tempDirectoryName string, fs filesystem.Filesystem) (filesystem.File, os.FileInfo, error) {
	file, err := fs.Create(filepath.Join(tempDirectoryName, fileName))
	if err != nil {
		return nil, nil, err
	}
	stat, err := fs.Stat(file.Name())
	if err != nil {
		return nil, nil, err
	}
	return file, stat, nil
}

func createGitFolderAndFiles(tempDirectoryName string, fs filesystem.Filesystem) error {
	err := fs.MkdirAll(filepath.Join(tempDirectoryName, ".git"), 0755)
	if err != nil {
		return err
	}

	err = fs.MkdirAll(filepath.Join(tempDirectoryName, DotOdoDirectory), 0755)
	if err != nil {
		return err
	}

	_, err = fs.Create(filepath.Join(tempDirectoryName, ".git", "someFile.txt"))
	if err != nil {
		return err
	}
	return nil
}

func Test_recursiveChecker(t *testing.T) {
	fs := filesystem.DefaultFs{}

	tempDirectoryName, err := fs.TempDir(os.TempDir(), "dir0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	jsFileName := "red.js"
	jsFile, jsFileStat, err := createAndStat(jsFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	readmeFileName := "README.txt"
	readmeFile, readmeFileStat, err := createAndStat(readmeFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	viewsFolderName := "views"
	viewsFolderPath := filepath.Join(tempDirectoryName, viewsFolderName)
	err = fs.MkdirAll(viewsFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = createGitFolderAndFiles(tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	htmlRelFilePath := filepath.Join(viewsFolderName, "view.html")
	htmlFile, htmlFileStat, err := createAndStat(filepath.Join("views", "view.html"), tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	viewsFolderStat, err := fs.Stat(filepath.Join(tempDirectoryName, viewsFolderName))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defer os.RemoveAll(tempDirectoryName)

	normalFileMap := map[string]FileData{
		readmeFileName: {
			Size:             readmeFileStat.Size(),
			LastModifiedDate: readmeFileStat.ModTime(),
		},
		jsFileName: {
			Size:             jsFileStat.Size(),
			LastModifiedDate: jsFileStat.ModTime(),
		},
		viewsFolderName: {
			Size:             viewsFolderStat.Size(),
			LastModifiedDate: viewsFolderStat.ModTime(),
		},
		htmlRelFilePath: {
			Size:             htmlFileStat.Size(),
			LastModifiedDate: htmlFileStat.ModTime(),
		},
	}

	type args struct {
		directory         string
		srcBase           string
		srcFile           string
		destBase          string
		destFile          string
		ignoreRules       []string
		remoteDirectories map[string]string
		existingFileIndex FileIndex
	}
	tests := []struct {
		name     string
		args     args
		want     IndexerRet
		emptyDir bool
		wantErr  bool
	}{
		{
			name: "case 1: existing index is empty",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), jsFile.Name(), viewsFolderPath, htmlFile.Name()},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 2: existing index exists and no file or folder changes occurs",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: normalFileMap,
				},
			},
			want: IndexerRet{
				NewFileMap: normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 3: file size changed",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: normalFileMap[htmlRelFilePath],
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size() + 100,
							LastModifiedDate: readmeFileStat.ModTime(),
						},
						jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
						viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name()},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 4: folder size changed",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:       normalFileMap[htmlRelFilePath],
						readmeFileStat.Name(): normalFileMap[readmeFileStat.Name()],
						jsFileStat.Name():     normalFileMap[jsFileStat.Name()],
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size() + 100,
							LastModifiedDate: viewsFolderStat.ModTime(),
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{viewsFolderPath},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 5: file modified",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: normalFileMap[htmlRelFilePath],
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime().Add(100),
						},
						jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
						viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name()},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 6: folder modified",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:       normalFileMap[htmlRelFilePath],
						readmeFileStat.Name(): normalFileMap[readmeFileStat.Name()],
						jsFileStat.Name():     normalFileMap[jsFileStat.Name()],
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime().Add(100),
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{viewsFolderPath},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 7: both file and folder modified",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: normalFileMap[htmlRelFilePath],
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size() + 100,
							LastModifiedDate: readmeFileStat.ModTime(),
						},
						jsFileStat.Name(): normalFileMap[jsFileStat.Name()],
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime().Add(100),
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), viewsFolderPath},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},

		{
			name: "case 8: ignore file with changes if remote exists",
			args: args{
				directory:   tempDirectoryName,
				srcBase:     tempDirectoryName,
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					htmlRelFilePath: "new/Folder/view.html",
				},
				existingFileIndex: FileIndex{
					Files: normalFileMap,
				},
			},
			want: IndexerRet{
				NewFileMap: map[string]FileData{
					readmeFileStat.Name(): {
						Size:             readmeFileStat.Size(),
						LastModifiedDate: readmeFileStat.ModTime(),
						RemoteAttribute:  "README.txt",
					},
					jsFileStat.Name(): {
						Size:             jsFileStat.Size(),
						LastModifiedDate: jsFileStat.ModTime(),
						RemoteAttribute:  "red.js",
					},
					viewsFolderStat.Name(): {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  "views",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 9: remote removed for a file containing different remote destination",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/view.html",
						},
						readmeFileStat.Name():  normalFileMap["README.txt"],
						jsFileStat.Name():      normalFileMap["red.js"],
						viewsFolderStat.Name(): normalFileMap["views"],
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{htmlFile.Name()},
				RemoteDeleted: []string{"new", "new/Folder", "new/Folder/view.html"},
				NewFileMap:    normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 10: remote removed for a folder containing different remote destination",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/view.html",
						},
						readmeFileStat.Name(): normalFileMap["README.txt"],
						jsFileStat.Name():     normalFileMap["red.js"],
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder/views",
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{viewsFolderPath, htmlFile.Name()},
				RemoteDeleted: []string{"new", "new/Folder", "new/Folder/view.html", "new/Folder/views"},
				NewFileMap:    normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 11: folder remote changed to local path",
			args: args{
				directory:   tempDirectoryName,
				srcBase:     tempDirectoryName,
				srcFile:     viewsFolderName,
				destFile:    viewsFolderName,
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					viewsFolderStat.Name(): viewsFolderStat.Name(),
				},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder/views",
						},
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/views/view.html",
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{viewsFolderPath, htmlFile.Name()},
				RemoteDeleted: []string{"new", "new/Folder", "new/Folder/views", "new/Folder/views/view.html"},
				NewFileMap: map[string]FileData{
					viewsFolderStat.Name(): {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  filepath.ToSlash(viewsFolderStat.Name()),
					}, htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  filepath.ToSlash(htmlRelFilePath),
					}},
			},
			wantErr: false,
		},

		{
			name: "case 12: only a single file is checked and others are ignored",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           filepath.Join(tempDirectoryName, "views"),
				srcFile:           "view.html",
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size() + 100,
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "",
						},
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size() + 100,
							LastModifiedDate: readmeFileStat.ModTime(),
						},
						jsFileStat.Name():      normalFileMap["red.js"],
						viewsFolderStat.Name(): normalFileMap["views"],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{htmlFile.Name()},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 13: only a single file with a different remote is checked",
			args: args{
				directory:   tempDirectoryName,
				srcBase:     tempDirectoryName,
				srcFile:     "README.txt",
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					readmeFileStat.Name(): "new/Folder/text/README.txt",
				},
				existingFileIndex: FileIndex{
					Files: normalFileMap,
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{readmeFile.Name()},
				RemoteDeleted: []string{filepath.ToSlash(readmeFileStat.Name())},
				NewFileMap: map[string]FileData{
					readmeFileStat.Name(): {
						Size:             readmeFileStat.Size(),
						LastModifiedDate: readmeFileStat.ModTime(),
						RemoteAttribute:  "new/Folder/text/README.txt",
					}},
			},
			wantErr: false,
		},
		{
			name: "case 14: only a single file is checked with a remote removed",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				srcFile:           "README.txt",
				destBase:          tempDirectoryName,
				destFile:          "README.txt",
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: normalFileMap["views/view.html"],
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/text/README.txt",
						},
						jsFileStat.Name():      normalFileMap["red.js"],
						viewsFolderStat.Name(): normalFileMap["views"],
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{readmeFile.Name()},
				RemoteDeleted: []string{"new", "new/Folder", "new/Folder/text", "new/Folder/text/README.txt"},
				NewFileMap: map[string]FileData{
					readmeFileStat.Name(): normalFileMap["README.txt"],
				},
			},
			wantErr: false,
		},
		{
			name: "case 15: only a single file is checked with the same remote path earlier",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				srcFile:           "README.txt",
				destBase:          tempDirectoryName,
				destFile:          "README.txt",
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: normalFileMap["views/view.html"],
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime(),
							RemoteAttribute:  "README.txt",
						},
						jsFileStat.Name():      normalFileMap["red.js"],
						viewsFolderStat.Name(): normalFileMap["views"],
					},
				},
			},
			want: IndexerRet{
				NewFileMap: map[string]FileData{
					readmeFileStat.Name(): normalFileMap["README.txt"],
				},
			},
			wantErr: false,
		},
		{
			name: "case 16: only a single file is checked and there is no modification",
			args: args{
				directory:   tempDirectoryName,
				srcBase:     viewsFolderPath,
				srcFile:     "view.html",
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					htmlRelFilePath: "new/views/view.html",
				},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/views/view.html",
						},
					},
				},
			},
			want: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  "new/views/view.html",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 17: file remote changed to local path",
			args: args{
				directory:   tempDirectoryName,
				srcBase:     tempDirectoryName,
				srcFile:     "README.txt",
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					readmeFileStat.Name(): "README.txt",
				},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/README.txt",
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{readmeFile.Name()},
				RemoteDeleted: []string{"new", "new/Folder", "new/Folder/README.txt"},
				NewFileMap: map[string]FileData{
					readmeFileStat.Name(): {
						Size:             readmeFileStat.Size(),
						LastModifiedDate: readmeFileStat.ModTime(),
						RemoteAttribute:  readmeFileStat.Name(),
					}},
			},
			wantErr: false,
		},

		{
			name: "case 18: file doesn't exist",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           viewsFolderPath,
				srcFile:           "views.html",
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{},
			},
			want: IndexerRet{
				NewFileMap: map[string]FileData{},
			},
			wantErr: true,
		},
		{
			name: "case 19: folder doesn't exist",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName + "blah",
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			want: IndexerRet{
				NewFileMap: map[string]FileData{},
			},
			wantErr: true,
		},

		{
			name: "case 20: ignore given file",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{"*.html"},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), jsFile.Name(), viewsFolderPath},
				NewFileMap: map[string]FileData{
					jsFileStat.Name():      normalFileMap["red.js"],
					viewsFolderStat.Name(): normalFileMap["views"],
					readmeFileStat.Name():  normalFileMap["README.txt"],
				},
			},
			wantErr: false,
		},
		{
			name: "case 21: ignore given folder",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{viewsFolderPath},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), jsFile.Name()},
				NewFileMap: map[string]FileData{
					jsFileStat.Name():     normalFileMap["red.js"],
					readmeFileStat.Name(): normalFileMap["README.txt"],
				},
			},
			wantErr: false,
		},

		{
			name: "case 22: only empty Dir with different remote location is checked",
			args: args{
				directory:   tempDirectoryName,
				srcBase:     filepath.Join(tempDirectoryName, "emptyDir"),
				srcFile:     "",
				destBase:    filepath.Join(tempDirectoryName, "emptyDir"),
				destFile:    "",
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					"emptyDir": "new/Folder/",
				},
				existingFileIndex: FileIndex{
					Files: normalFileMap,
				},
			},
			emptyDir: true,
			want: IndexerRet{
				FilesChanged: []string{filepath.Join(tempDirectoryName, "emptyDir")},
				NewFileMap:   map[string]FileData{},
			},
			wantErr: false,
		},
		{
			name: "case 23: folder containing a empty directory",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			emptyDir: true,
			want: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), filepath.Join(tempDirectoryName, "emptyDir"), jsFile.Name(), viewsFolderPath, htmlFile.Name()},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.emptyDir {
				emptyDirPath := filepath.Join(tempDirectoryName, "emptyDir")
				err = fs.MkdirAll(emptyDirPath, 0755)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				defer func(name string) {
					err := os.Remove(name)
					if err != nil {
						t.Errorf("enexpected error: %v", err)
					}
				}(emptyDirPath)

				emptyDirStat, err := fs.Stat(emptyDirPath)
				if err != nil {
					t.Errorf("enexpected error: %v", err)
				}

				tt.want.NewFileMap[emptyDirStat.Name()] = FileData{
					Size:             emptyDirStat.Size(),
					LastModifiedDate: emptyDirStat.ModTime(),
					RemoteAttribute:  tt.args.remoteDirectories[emptyDirStat.Name()],
				}
			}
			pathsOptions := recursiveCheckerPathOptions{tt.args.directory, tt.args.srcBase, tt.args.srcFile, tt.args.destBase, tt.args.destFile}
			got, err := recursiveChecker(pathsOptions, tt.args.ignoreRules, tt.args.remoteDirectories, tt.args.existingFileIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("recursiveChecker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr {
				return
			}

			sort.Strings(got.FilesDeleted)
			sort.Strings(got.FilesChanged)
			sort.Strings(got.RemoteDeleted)
			if !reflect.DeepEqual(got.FilesChanged, tt.want.FilesChanged) {
				t.Errorf("recursiveChecker() FilesChanged got = %v, want %v", got.FilesChanged, tt.want.FilesChanged)
			}

			if !reflect.DeepEqual(got.FilesDeleted, tt.want.FilesDeleted) {
				t.Errorf("recursiveChecker() FilesDeleted got = %v, want %v", got.FilesDeleted, tt.want.FilesDeleted)
			}

			if !reflect.DeepEqual(got.RemoteDeleted, tt.want.RemoteDeleted) {
				t.Errorf("recursiveChecker() RemoteDeleted got = %v, want %v", got.RemoteDeleted, tt.want.RemoteDeleted)
			}

			if !reflect.DeepEqual(tt.want.NewFileMap, got.NewFileMap) {
				t.Errorf("recursiveChecker() new file map is different, difference = %v", pretty.Compare(got.NewFileMap, tt.want.NewFileMap))
			}
		})
	}
}

func Test_runIndexerWithExistingFileIndex(t *testing.T) {
	fs := filesystem.DefaultFs{}

	tempDirectoryName, err := fs.TempDir(os.TempDir(), "dir0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	jsFileName := "red.js"
	jsFile, jsFileStat, err := createAndStat(jsFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	readmeFileName := "README.txt"
	readmeFile, readmeFileStat, err := createAndStat(readmeFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	viewsFolderName := "views"
	viewsFolderPath := filepath.Join(tempDirectoryName, viewsFolderName)
	err = fs.MkdirAll(viewsFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = createGitFolderAndFiles(tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	htmlRelFilePath := filepath.Join(viewsFolderName, "view.html")
	htmlFile, htmlFileStat, err := createAndStat(filepath.Join("views", "view.html"), tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	viewsFolderStat, err := fs.Stat(filepath.Join(tempDirectoryName, viewsFolderName))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defer os.RemoveAll(tempDirectoryName)

	normalFileMap := map[string]FileData{
		readmeFileName: {
			Size:             readmeFileStat.Size(),
			LastModifiedDate: readmeFileStat.ModTime(),
		},
		jsFileName: {
			Size:             jsFileStat.Size(),
			LastModifiedDate: jsFileStat.ModTime(),
		},
		viewsFolderName: {
			Size:             viewsFolderStat.Size(),
			LastModifiedDate: viewsFolderStat.ModTime(),
		},
		htmlRelFilePath: {
			Size:             htmlFileStat.Size(),
			LastModifiedDate: htmlFileStat.ModTime(),
		},
	}

	type args struct {
		directory         string
		ignoreRules       []string
		remoteDirectories map[string]string
		existingFileIndex *FileIndex
	}
	tests := []struct {
		name    string
		args    args
		wantRet IndexerRet
		wantErr bool
	}{
		{
			name: "case 1: normal directory with no existing file index data",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{},
			},
			wantRet: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), jsFile.Name(), viewsFolderPath, htmlFile.Name()},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 2: normal directory with existing file index data",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: normalFileMap,
				},
			},
			wantRet: IndexerRet{
				NewFileMap: normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 3: normal directory with existing file index data and new files are added",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: normalFileMap[htmlRelFilePath],
					},
				},
			},
			wantRet: IndexerRet{
				FilesChanged: []string{readmeFile.Name(), jsFile.Name(), viewsFolderPath},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 4: normal directory with existing file index data and files are deleted",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:        normalFileMap[htmlRelFilePath],
						jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
						viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
						readmeFileStat.Name():  normalFileMap[readmeFileStat.Name()],
						"blah":                 {},
					},
				},
			},
			wantRet: IndexerRet{
				FilesDeleted: []string{"blah"},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},

		{
			name: "case 5: with remote directories and no existing file index",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{viewsFolderStat.Name(): "new/Folder", htmlRelFilePath: "new/Folder0/view.html"},
				existingFileIndex: &FileIndex{},
			},
			wantRet: IndexerRet{
				FilesChanged: []string{viewsFolderPath, htmlFile.Name()},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  "new/Folder0/view.html",
					},
					viewsFolderStat.Name(): {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  "new/Folder",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 6: with remote directories and no modification",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{htmlRelFilePath: "new/Folder0/view.html", viewsFolderStat.Name(): "new/Folder"},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder0/view.html",
						},
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder",
						},
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  "new/Folder0/view.html",
					},
					viewsFolderStat.Name(): {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  "new/Folder",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 7: with remote directories and files deleted",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{htmlRelFilePath: "new/Folder0/view.html", viewsFolderStat.Name(): "new/Folder"},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder0/view.html",
						},
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder",
						},
						"blah": {},
					},
				},
			},
			wantRet: IndexerRet{
				FilesDeleted: []string{"blah"},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  "new/Folder0/view.html",
					},
					viewsFolderStat.Name(): {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  "new/Folder",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 8: remote changed",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{htmlRelFilePath: "new/Folder0/view.html", viewsFolderStat.Name(): "new/blah/Folder"},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder0/view.html",
						},
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder",
						},
					},
				},
			},
			wantRet: IndexerRet{
				FilesChanged:  []string{viewsFolderPath},
				RemoteDeleted: []string{"new/Folder"},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  "new/Folder0/view.html",
					},
					viewsFolderStat.Name(): {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  "new/blah/Folder",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 9: remote of a file removed",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{htmlRelFilePath: "new/Folder0/view.html"},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder0/view.html",
						},
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder",
						},
					},
				},
			},
			wantRet: IndexerRet{
				RemoteDeleted: []string{"new/Folder"},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  "new/Folder0/view.html",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 10: all remotes removed",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						readmeFileStat.Name(): {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime(),
							RemoteAttribute:  readmeFileStat.Name(),
						},
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder0/view.html",
						},
						viewsFolderStat.Name(): {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder",
						},
					},
				},
			},
			wantRet: IndexerRet{
				FilesChanged:  []string{jsFile.Name(), viewsFolderPath, htmlFile.Name()},
				RemoteDeleted: []string{"new", "new/Folder", "new/Folder0", "new/Folder0/view.html"},
				NewFileMap:    normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 11: remote added for a file but local path and remote destination are same",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{htmlRelFilePath: htmlRelFilePath},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
						},
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  filepath.ToSlash(htmlRelFilePath),
					},
				},
			},
			wantErr: false,
		},

		{
			name: "case 12: ignore a modified file due to ignore rules",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{filepath.Join(tempDirectoryName, readmeFileStat.Name())},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:        normalFileMap[htmlRelFilePath],
						viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
						jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath:        normalFileMap[htmlRelFilePath],
					viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
					jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
				},
			},
			wantErr: false,
		},
		{
			name: "case 13: ignore a deleted file due to ignore rules",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{filepath.Join(tempDirectoryName, "blah")},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						readmeFileStat.Name():  normalFileMap[readmeFileStat.Name()],
						htmlRelFilePath:        normalFileMap[htmlRelFilePath],
						viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
						jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
						"blah":                 {},
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "case 14: ignore a added file due to ignore rules",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{filepath.Join(tempDirectoryName, readmeFileStat.Name())},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:        normalFileMap[htmlRelFilePath],
						viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
						jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath:        normalFileMap[htmlRelFilePath],
					viewsFolderStat.Name(): normalFileMap[viewsFolderStat.Name()],
					jsFileStat.Name():      normalFileMap[jsFileStat.Name()],
				},
			},
			wantErr: false,
		},
		{
			name: "case 15: local file doesn't exist",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{htmlRelFilePath + "blah": htmlRelFilePath},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{},
				},
			},
			wantRet: IndexerRet{},
			wantErr: true,
		},
		{
			name: "case 16: local folder doesn't exist",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{},
				remoteDirectories: map[string]string{viewsFolderPath + "blah": viewsFolderPath},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{},
				},
			},
			wantRet: IndexerRet{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := runIndexerWithExistingFileIndex(tt.args.directory, tt.args.ignoreRules, tt.args.remoteDirectories, tt.args.existingFileIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("runIndexerWithExistingFileIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr {
				return
			}

			sort.Strings(gotRet.FilesDeleted)
			sort.Strings(gotRet.FilesChanged)
			sort.Strings(gotRet.RemoteDeleted)
			if !reflect.DeepEqual(gotRet.FilesChanged, tt.wantRet.FilesChanged) {
				t.Errorf("runIndexerWithExistingFileIndex() fileChanged gotRet = %v, want %v", gotRet.FilesChanged, tt.wantRet.FilesChanged)
			}

			if !reflect.DeepEqual(gotRet.NewFileMap, tt.wantRet.NewFileMap) {
				t.Errorf("runIndexerWithExistingFileIndex() new file map is different = %v", pretty.Compare(gotRet.NewFileMap, tt.wantRet.NewFileMap))
			}

			if !reflect.DeepEqual(gotRet.FilesDeleted, tt.wantRet.FilesDeleted) {
				t.Errorf("runIndexerWithExistingFileIndex() files deleted gotRet = %v, want %v", gotRet.FilesDeleted, tt.wantRet.FilesDeleted)
			}

			if !reflect.DeepEqual(gotRet.RemoteDeleted, tt.wantRet.RemoteDeleted) {
				t.Errorf("runIndexerWithExistingFileIndex() files remote changed gotRet = %v, want %v", gotRet.RemoteDeleted, tt.wantRet.RemoteDeleted)
			}
		})
	}
}
