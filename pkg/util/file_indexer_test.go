package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestCheckGitIgnoreFile(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a context directory on fake fs
	contextDir, err := fs.TempDir(os.TempDir(), "context")
	if err != nil {
		t.Error(err)
	}

	gitignorePath := filepath.Join(contextDir, DotGitIgnoreFile)

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

			gitIgnoreFilePath, isNew, err := touchGitIgnoreFile(tt.directory, fs)

			if tt.want != gitIgnoreFilePath {
				t.Errorf("touchGitIgnoreFile unexpected error %v, while creating .gitignore file", err)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("touchGitIgnoreFile unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if tt.create != isNew {
				t.Errorf("touchGitIgnoreFile: expected tt.create=%v, got %v", tt.create, isNew)
			}
		})
	}

}

func TestAddOdoDirectory(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a context directory on fake fs
	contextDir, err := fs.TempDir(os.TempDir(), "context")
	if err != nil {
		t.Error(err)
	}

	gitignorePath := filepath.Join(contextDir, DotGitIgnoreFile)

	tests := []struct {
		testName        string
		create          bool
		gitIgnoreCreate func(create bool, contextDir string, fs filesystem.Filesystem) error
		directory       string
		wantErr         bool
	}{
		{
			testName:        "Test when .odo added to .gitignore",
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

			err := addOdoDirectory(tt.directory, fs)

			if !tt.wantErr == (err != nil) {
				t.Errorf("addOdoFileIndex unexpected error %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func mockDirectoryInfo(create bool, contextDir string, fs filesystem.Filesystem) error {

	if !create {
		err := fs.MkdirAll(filepath.Join(contextDir, DotGitIgnoreFile), os.ModePerm)
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

	err = createGitFolderAndFiles(tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	jsFileName := "red.js"
	jsFile, jsFileStat, err := createAndStat(jsFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsFileAbsPath := jsFile.Name()

	readmeFileName := "README.txt"
	readmeFile, readmeFileStat, err := createAndStat(readmeFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	readmeFileAbsPath := readmeFile.Name()

	specialCharFolderName := "[devfile-registry]"
	specialCharFolderPath := filepath.Join(tempDirectoryName, specialCharFolderName)
	err = fs.MkdirAll(specialCharFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fileInsideSpecialCharFolderRelPath := filepath.Join(specialCharFolderName, "index.tsx")
	fileInsideSpecialCharFolderFile, fileInsideSpecialCharFolderStat, err := createAndStat(fileInsideSpecialCharFolderRelPath, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	fileInsideSpecialCharFolderAbsPath := fileInsideSpecialCharFolderFile.Name()
	specialCharFolderStat, err := fs.Stat(specialCharFolderPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	viewsFolderName := "views"
	viewsFolderPath := filepath.Join(tempDirectoryName, viewsFolderName)
	err = fs.MkdirAll(viewsFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	htmlRelFilePath := filepath.Join(viewsFolderName, "view.html")
	htmlFile, htmlFileStat, err := createAndStat(filepath.Join("views", "view.html"), tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	htmlFileAbsPath := htmlFile.Name()

	targetFolderName := "target"
	targetFolderRelPath := filepath.Join(viewsFolderName, targetFolderName)
	targetFolderPath := filepath.Join(tempDirectoryName, targetFolderRelPath)
	err = fs.MkdirAll(targetFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	targetFileName := "someFile.txt"
	targetFileRelPath := filepath.Join(viewsFolderName, targetFolderName, targetFileName)
	targetFilePath := filepath.Join(tempDirectoryName, targetFileRelPath)
	_, targetFileStat, err := createAndStat(targetFileRelPath, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	targetFolderStat, err := fs.Stat(filepath.Join(tempDirectoryName, viewsFolderName, targetFolderName))
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
		targetFolderRelPath: {
			Size:             targetFolderStat.Size(),
			LastModifiedDate: targetFolderStat.ModTime(),
		},
		targetFileRelPath: {
			Size:             targetFileStat.Size(),
			LastModifiedDate: targetFileStat.ModTime(),
		},
		specialCharFolderName: {
			Size:             specialCharFolderStat.Size(),
			LastModifiedDate: specialCharFolderStat.ModTime(),
		},
		fileInsideSpecialCharFolderRelPath: {
			Size:             fileInsideSpecialCharFolderStat.Size(),
			LastModifiedDate: fileInsideSpecialCharFolderStat.ModTime(),
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
				existingFileIndex: FileIndex{},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFileAbsPath, jsFileAbsPath, viewsFolderPath, targetFolderPath, targetFilePath, htmlFileAbsPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
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
						readmeFileName: {
							Size:             readmeFileStat.Size() + 100,
							LastModifiedDate: readmeFileStat.ModTime(),
						},
						jsFileName:                         normalFileMap[jsFileName],
						viewsFolderName:                    normalFileMap[viewsFolderName],
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						targetFolderRelPath:                normalFileMap[targetFolderRelPath],
						targetFileRelPath:                  normalFileMap[targetFileRelPath],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFileAbsPath},
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
						readmeFileName: normalFileMap[readmeFileName],
						jsFileName:     normalFileMap[jsFileName],
						viewsFolderName: {
							Size:             viewsFolderStat.Size() + 100,
							LastModifiedDate: viewsFolderStat.ModTime(),
						},
						htmlRelFilePath:     normalFileMap[htmlRelFilePath],
						targetFolderRelPath: normalFileMap[targetFolderRelPath],
						targetFileRelPath:   normalFileMap[targetFileRelPath],
						specialCharFolderName: {
							Size:             specialCharFolderStat.Size() + 100,
							LastModifiedDate: specialCharFolderStat.ModTime(),
						},
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{viewsFolderPath, specialCharFolderPath},
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
						readmeFileName: {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime().Add(100),
						},
						jsFileName:                         normalFileMap[jsFileName],
						viewsFolderName:                    normalFileMap[viewsFolderName],
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						targetFolderRelPath:                normalFileMap[targetFolderRelPath],
						targetFileRelPath:                  normalFileMap[targetFileRelPath],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFileAbsPath},
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
						readmeFileName: normalFileMap[readmeFileName],
						jsFileName:     normalFileMap[jsFileName],
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime().Add(100),
						},
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						targetFolderRelPath:                normalFileMap[targetFolderRelPath],
						targetFileRelPath:                  normalFileMap[targetFileRelPath],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
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
						readmeFileName: {
							Size:             readmeFileStat.Size() + 100,
							LastModifiedDate: readmeFileStat.ModTime(),
						},
						jsFileName: normalFileMap[jsFileName],
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime().Add(100),
						},
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						targetFolderRelPath:                normalFileMap[targetFolderRelPath],
						targetFileRelPath:                  normalFileMap[targetFileRelPath],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFileAbsPath, viewsFolderPath},
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
					htmlRelFilePath: filepath.Join("new", "Folder", "views.html"),
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
					jsFileName: {
						Size:             jsFileStat.Size(),
						LastModifiedDate: jsFileStat.ModTime(),
						RemoteAttribute:  "red.js",
					},
					viewsFolderName: {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  "views",
					},
					targetFolderRelPath: {
						Size:             targetFolderStat.Size(),
						LastModifiedDate: targetFolderStat.ModTime(),
						RemoteAttribute:  targetFolderRelPath,
					},
					targetFileRelPath: {
						Size:             targetFileStat.Size(),
						LastModifiedDate: targetFileStat.ModTime(),
						RemoteAttribute:  targetFileRelPath,
					},
					specialCharFolderName: {
						Size:             specialCharFolderStat.Size(),
						LastModifiedDate: specialCharFolderStat.ModTime(),
						RemoteAttribute:  specialCharFolderName,
					},
					fileInsideSpecialCharFolderRelPath: {
						Size:             fileInsideSpecialCharFolderStat.Size(),
						LastModifiedDate: fileInsideSpecialCharFolderStat.ModTime(),
						RemoteAttribute:  fileInsideSpecialCharFolderRelPath,
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
						readmeFileName:  normalFileMap[readmeFileName],
						jsFileName:      normalFileMap[jsFileName],
						viewsFolderName: normalFileMap[viewsFolderName],
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder", "views.html"),
						},
						targetFolderRelPath:                normalFileMap[targetFolderRelPath],
						targetFileRelPath:                  normalFileMap[targetFileRelPath],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{htmlFileAbsPath},
				RemoteDeleted: []string{"new", filepath.Join("new", "Folder"), filepath.Join("new", "Folder", "views.html")},
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
						readmeFileName: normalFileMap[readmeFileName],
						jsFileName:     normalFileMap[jsFileName],
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  "new/Folder/views",
						},
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder", "views.html"),
						},
						targetFolderRelPath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/target",
						},
						targetFileRelPath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/target/someFile.txt",
						},
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{viewsFolderPath, targetFolderPath, targetFilePath, htmlFileAbsPath},
				RemoteDeleted: []string{"new", filepath.Join("new", "Folder"), "new/Folder/target", "new/Folder/target/someFile.txt", filepath.Join("new", "Folder", "views.html"), "new/Folder/views"},
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
					viewsFolderName: viewsFolderName,
				},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						viewsFolderName: {
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
				FilesChanged:  []string{viewsFolderPath, targetFolderPath, targetFilePath, htmlFileAbsPath},
				RemoteDeleted: []string{"new", filepath.Join("new", "Folder"), "new/Folder/views", "new/Folder/views/view.html"},
				NewFileMap: map[string]FileData{
					viewsFolderName: {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  filepath.ToSlash(viewsFolderName),
					},
					targetFolderRelPath: {
						Size:             targetFolderStat.Size(),
						LastModifiedDate: targetFolderStat.ModTime(),
						RemoteAttribute:  filepath.ToSlash(targetFolderRelPath),
					},
					targetFileRelPath: {
						Size:             targetFileStat.Size(),
						LastModifiedDate: targetFileStat.ModTime(),
						RemoteAttribute:  filepath.ToSlash(targetFileRelPath),
					},
					htmlRelFilePath: {
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
						jsFileName:      normalFileMap["red.js"],
						viewsFolderName: normalFileMap["views"],
					},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{htmlFileAbsPath},
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
				FilesChanged:  []string{readmeFileAbsPath},
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
						jsFileName:      normalFileMap[jsFileName],
						viewsFolderName: normalFileMap[viewsFolderName],
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{readmeFileAbsPath},
				RemoteDeleted: []string{"new", filepath.Join("new", "Folder"), filepath.Join("new", "Folder", "text"), filepath.Join("new", "Folder", "text", readmeFileName)},
				NewFileMap: map[string]FileData{
					readmeFileName: normalFileMap[readmeFileName],
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
						jsFileName:      normalFileMap["red.js"],
						viewsFolderName: normalFileMap["views"],
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
					readmeFileName: readmeFileName,
				},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{
						readmeFileName: {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime(),
							RemoteAttribute:  "new/Folder/README.txt",
						},
					},
				},
			},
			want: IndexerRet{
				FilesChanged:  []string{readmeFileAbsPath},
				RemoteDeleted: []string{"new", filepath.Join("new", "Folder"), filepath.Join("new", "Folder", "README.txt")},
				NewFileMap: map[string]FileData{
					readmeFileName: {
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
				FilesChanged: []string{readmeFileAbsPath, jsFileAbsPath, viewsFolderPath, targetFolderPath, targetFilePath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
				NewFileMap: map[string]FileData{
					jsFileName:                         normalFileMap["red.js"],
					viewsFolderName:                    normalFileMap["views"],
					readmeFileStat.Name():              normalFileMap["README.txt"],
					targetFolderRelPath:                normalFileMap[targetFolderRelPath],
					targetFileRelPath:                  normalFileMap[targetFileRelPath],
					specialCharFolderName:              normalFileMap[specialCharFolderName],
					fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
				},
			},
			wantErr: false,
		},
		{
			name: "case 21: ignore given folder",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{viewsFolderName},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFileAbsPath, jsFileAbsPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
				NewFileMap: map[string]FileData{
					jsFileName:                         normalFileMap["red.js"],
					readmeFileStat.Name():              normalFileMap["README.txt"],
					specialCharFolderName:              normalFileMap[specialCharFolderName],
					fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
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
				FilesChanged: []string{readmeFileAbsPath, filepath.Join(tempDirectoryName, "emptyDir"), jsFileAbsPath, viewsFolderPath, targetFolderPath, targetFilePath, htmlFileAbsPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
				NewFileMap:   normalFileMap,
			},
			wantErr: false,
		},
		{
			name: "Case 24: subfolder is ignored",
			args: args{
				directory:         tempDirectoryName,
				srcBase:           tempDirectoryName,
				ignoreRules:       []string{"target/"},
				remoteDirectories: map[string]string{},
				existingFileIndex: FileIndex{
					Files: map[string]FileData{},
				},
			},
			want: IndexerRet{
				FilesChanged: []string{readmeFileAbsPath, jsFileAbsPath, viewsFolderPath, targetFolderPath, htmlFileAbsPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
				NewFileMap: map[string]FileData{
					readmeFileName:                     normalFileMap[readmeFileName],
					jsFileName:                         normalFileMap[jsFileName],
					viewsFolderName:                    normalFileMap[viewsFolderName],
					htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
					targetFolderRelPath:                normalFileMap[targetFolderRelPath],
					specialCharFolderName:              normalFileMap[specialCharFolderName],
					fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
				},
			},
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

			sortOpt := cmpopts.SortSlices(func(x, y string) bool {
				return x < y
			})
			if diff := cmp.Diff(tt.want.FilesChanged, got.FilesChanged, sortOpt); diff != "" {
				t.Errorf("recursiveChecker() FilesChanged mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want.FilesDeleted, got.FilesDeleted, sortOpt); diff != "" {
				t.Errorf("recursiveChecker() FilesDeleted mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want.RemoteDeleted, got.RemoteDeleted, sortOpt); diff != "" {
				t.Errorf("recursiveChecker() RemoteDeleted mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want.NewFileMap, got.NewFileMap); diff != "" {
				t.Errorf("recursiveChecker() NewFileMap mismatch (-want +got):\n%s", diff)
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

	err = createGitFolderAndFiles(tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	jsFileName := "red.js"
	jsFile, jsFileStat, err := createAndStat(jsFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsFileAbsPath := jsFile.Name()

	readmeFileName := "README.txt"
	readmeFile, readmeFileStat, err := createAndStat(readmeFileName, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	readmeFileAbsPath := readmeFile.Name()

	specialCharFolderName := "[devfile-registry]"
	specialCharFolderPath := filepath.Join(tempDirectoryName, specialCharFolderName)
	err = fs.MkdirAll(specialCharFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fileInsideSpecialCharFolderRelPath := filepath.Join(specialCharFolderName, "index.tsx")
	fileInsideSpecialCharFolderFile, fileInsideSpecialCharFolderFileStat, err := createAndStat(fileInsideSpecialCharFolderRelPath, tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	fileInsideSpecialCharFolderAbsPath := fileInsideSpecialCharFolderFile.Name()

	specialCharFolderStat, err := fs.Stat(specialCharFolderPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	viewsFolderName := "views"
	viewsFolderPath := filepath.Join(tempDirectoryName, viewsFolderName)
	err = fs.MkdirAll(viewsFolderPath, 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	htmlRelFilePath := filepath.Join(viewsFolderName, "view.html")
	htmlFile, htmlFileStat, err := createAndStat(filepath.Join("views", "view.html"), tempDirectoryName, fs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	htmlFileAbsPath := htmlFile.Name()

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
		specialCharFolderName: {
			Size:             specialCharFolderStat.Size(),
			LastModifiedDate: specialCharFolderStat.ModTime(),
		},
		fileInsideSpecialCharFolderRelPath: {
			Size:             fileInsideSpecialCharFolderFileStat.Size(),
			LastModifiedDate: fileInsideSpecialCharFolderFileStat.ModTime(),
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
				FilesChanged: []string{readmeFileAbsPath, jsFileAbsPath, viewsFolderPath, htmlFileAbsPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
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
				FilesChanged: []string{readmeFileAbsPath, jsFileAbsPath, viewsFolderPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
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
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						jsFileName:                         normalFileMap[jsFileName],
						viewsFolderName:                    normalFileMap[viewsFolderName],
						readmeFileName:                     normalFileMap[readmeFileName],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
						"blah":                             {},
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
				remoteDirectories: map[string]string{viewsFolderName: filepath.Join("new", "Folder"), htmlRelFilePath: filepath.Join("new", "Folder0", "views.html")},
				existingFileIndex: &FileIndex{},
			},
			wantRet: IndexerRet{
				FilesChanged: []string{viewsFolderPath, htmlFileAbsPath},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
					},
					viewsFolderName: {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder"),
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
				remoteDirectories: map[string]string{htmlRelFilePath: filepath.Join("new", "Folder0", "views.html"), viewsFolderName: filepath.Join("new", "Folder")},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
						},
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder"),
						},
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
					},
					viewsFolderName: {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 7: with remote directories and files deleted",
			args: args{
				directory:   tempDirectoryName,
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					htmlRelFilePath: filepath.Join("new", "Folder0", "views.html"),
					viewsFolderName: filepath.Join("new", "Folder"),
				},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
						},
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder"),
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
						RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
					},
					viewsFolderName: {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 8: remote changed",
			args: args{
				directory:   tempDirectoryName,
				ignoreRules: []string{},
				remoteDirectories: map[string]string{
					htmlRelFilePath: filepath.Join("new", "Folder0", "views.html"),
					viewsFolderName: filepath.Join("new", "blah", "Folder"),
				},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
						},
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder"),
						},
					},
				},
			},
			wantRet: IndexerRet{
				FilesChanged:  []string{viewsFolderPath},
				RemoteDeleted: []string{filepath.Join("new", "Folder")},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
					},
					viewsFolderName: {
						Size:             viewsFolderStat.Size(),
						LastModifiedDate: viewsFolderStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "blah", "Folder"),
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
				remoteDirectories: map[string]string{htmlRelFilePath: filepath.Join("new", "Folder0", "views.html")},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
						},
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder"),
						},
					},
				},
			},
			wantRet: IndexerRet{
				RemoteDeleted: []string{filepath.Join("new", "Folder")},
				NewFileMap: map[string]FileData{
					htmlRelFilePath: {
						Size:             htmlFileStat.Size(),
						LastModifiedDate: htmlFileStat.ModTime(),
						RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
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
						readmeFileName: {
							Size:             readmeFileStat.Size(),
							LastModifiedDate: readmeFileStat.ModTime(),
							RemoteAttribute:  readmeFileStat.Name(),
						},
						htmlRelFilePath: {
							Size:             htmlFileStat.Size(),
							LastModifiedDate: htmlFileStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder0", "views.html"),
						},
						viewsFolderName: {
							Size:             viewsFolderStat.Size(),
							LastModifiedDate: viewsFolderStat.ModTime(),
							RemoteAttribute:  filepath.Join("new", "Folder"),
						},
					},
				},
			},
			wantRet: IndexerRet{
				FilesChanged:  []string{jsFileAbsPath, viewsFolderPath, htmlFileAbsPath, specialCharFolderPath, fileInsideSpecialCharFolderAbsPath},
				RemoteDeleted: []string{"new", filepath.Join("new", "Folder"), filepath.Join("new", "Folder0"), filepath.Join("new", "Folder0", "views.html")},
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
				ignoreRules:       []string{readmeFileName},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						viewsFolderName:                    normalFileMap[viewsFolderName],
						jsFileName:                         normalFileMap[jsFileName],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
					viewsFolderName:                    normalFileMap[viewsFolderName],
					jsFileName:                         normalFileMap[jsFileName],
					specialCharFolderName:              normalFileMap[specialCharFolderName],
					fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
				},
			},
			wantErr: false,
		},
		{
			name: "case 13: ignore a deleted file due to ignore rules",
			args: args{
				directory:         tempDirectoryName,
				ignoreRules:       []string{"blah"},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						readmeFileName:                     normalFileMap[readmeFileName],
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						viewsFolderName:                    normalFileMap[viewsFolderName],
						jsFileName:                         normalFileMap[jsFileName],
						"blah":                             {},
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
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
				ignoreRules:       []string{readmeFileName},
				remoteDirectories: map[string]string{},
				existingFileIndex: &FileIndex{
					Files: map[string]FileData{
						htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
						viewsFolderName:                    normalFileMap[viewsFolderName],
						jsFileName:                         normalFileMap[jsFileName],
						specialCharFolderName:              normalFileMap[specialCharFolderName],
						fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
					},
				},
			},
			wantRet: IndexerRet{
				NewFileMap: map[string]FileData{
					htmlRelFilePath:                    normalFileMap[htmlRelFilePath],
					viewsFolderName:                    normalFileMap[viewsFolderName],
					jsFileName:                         normalFileMap[jsFileName],
					specialCharFolderName:              normalFileMap[specialCharFolderName],
					fileInsideSpecialCharFolderRelPath: normalFileMap[fileInsideSpecialCharFolderRelPath],
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

			sortOpt := cmpopts.SortSlices(func(x, y string) bool {
				return x < y
			})

			if diff := cmp.Diff(tt.wantRet.FilesChanged, gotRet.FilesChanged, sortOpt); diff != "" {
				t.Errorf("runIndexerWithExistingFileIndex() FilesChanged mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantRet.NewFileMap, gotRet.NewFileMap); diff != "" {
				t.Errorf("runIndexerWithExistingFileIndex() NewFileMap mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantRet.FilesDeleted, gotRet.FilesDeleted, sortOpt); diff != "" {
				t.Errorf("runIndexerWithExistingFileIndex() FilesDeleted mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantRet.RemoteDeleted, gotRet.RemoteDeleted, sortOpt); diff != "" {
				t.Errorf("runIndexerWithExistingFileIndex() RemoteDeleted mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
