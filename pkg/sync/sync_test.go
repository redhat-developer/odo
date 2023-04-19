package sync

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/tests/helper"
)

func TestGetCmdToCreateSyncFolder(t *testing.T) {
	tests := []struct {
		name       string
		syncFolder string
		want       []string
	}{
		{
			name:       "Case 1: Sync to /projects",
			syncFolder: generator.DevfileSourceVolumeMount,
			want:       []string{"mkdir", "-p", generator.DevfileSourceVolumeMount},
		},
		{
			name:       "Case 2: Sync subdir of /projects",
			syncFolder: generator.DevfileSourceVolumeMount + "/someproject",
			want:       []string{"mkdir", "-p", generator.DevfileSourceVolumeMount + "/someproject"},
		},
	}
	for _, tt := range tests {
		cmdArr := getCmdToCreateSyncFolder(tt.syncFolder)
		if diff := cmp.Diff(tt.want, cmdArr); diff != "" {
			t.Errorf("getCmdToCreateSyncFolder() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestGetCmdToDeleteFiles(t *testing.T) {
	syncFolder := "/projects/hello-world"

	tests := []struct {
		name       string
		delFiles   []string
		syncFolder string
		want       []string
	}{
		{
			name:       "Case 1: One deleted file",
			delFiles:   []string{"test.txt"},
			syncFolder: generator.DevfileSourceVolumeMount,
			want:       []string{"rm", "-rf", generator.DevfileSourceVolumeMount + "/test.txt"},
		},
		{
			name:       "Case 2: Multiple deleted files, default sync folder",
			delFiles:   []string{"test.txt", "hello.c"},
			syncFolder: generator.DevfileSourceVolumeMount,
			want:       []string{"rm", "-rf", generator.DevfileSourceVolumeMount + "/test.txt", generator.DevfileSourceVolumeMount + "/hello.c"},
		},
		{
			name:       "Case 2: Multiple deleted files, different sync folder",
			delFiles:   []string{"test.txt", "hello.c"},
			syncFolder: syncFolder,
			want:       []string{"rm", "-rf", syncFolder + "/test.txt", syncFolder + "/hello.c"},
		},
	}
	for _, tt := range tests {
		cmdArr := getCmdToDeleteFiles(tt.delFiles, tt.syncFolder)
		if diff := cmp.Diff(tt.want, cmdArr); diff != "" {
			t.Errorf("getCmdToDeleteFiles() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestSyncFiles(t *testing.T) {

	testComponentName := "test"

	// create a temp dir for the file indexer
	directory := t.TempDir()

	jsFile, e := os.Create(filepath.Join(directory, "red.js"))
	if e != nil {
		t.Errorf("TestSyncFiles error: error creating temporary file for the indexer: %v", e)
	}

	ctrl := gomock.NewController(t)
	kc := kclient.NewMockClientInterface(ctrl)
	kc.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Assert that Bar() is invoked.
	defer ctrl.Finish()

	tests := []struct {
		name               string
		syncParameters     SyncParameters
		wantErr            bool
		wantIsPushRequired bool
	}{
		{
			name: "Case 1: Component does not exist",
			syncParameters: SyncParameters{
				Path:              directory,
				WatchFiles:        []string{},
				WatchDeletedFiles: []string{},
				IgnoredFiles:      []string{},
				CompInfo: ComponentInfo{
					ContainerName: "abcd",
				},
				ForcePush: true,
			},
			wantErr:            false,
			wantIsPushRequired: true,
		},
		{
			name: "Case 2: Component does exist",
			syncParameters: SyncParameters{
				Path:              directory,
				WatchFiles:        []string{},
				WatchDeletedFiles: []string{},
				IgnoredFiles:      []string{},
				CompInfo: ComponentInfo{
					ContainerName: "abcd",
				},
				ForcePush: false,
			},
			wantErr:            false,
			wantIsPushRequired: false, // always false after case 1
		},
		{
			name: "Case 3: FakeErrorClient error",
			syncParameters: SyncParameters{
				Path:              directory,
				WatchFiles:        []string{},
				WatchDeletedFiles: []string{},
				IgnoredFiles:      []string{},
				CompInfo: ComponentInfo{
					ContainerName: "abcd",
				},
				ForcePush: false,
			},
			wantErr:            true,
			wantIsPushRequired: false,
		},
		{
			name: "Case 4: File change",
			syncParameters: SyncParameters{
				Path:              directory,
				WatchFiles:        []string{path.Join(directory, "test.log")},
				WatchDeletedFiles: []string{},
				IgnoredFiles:      []string{},
				CompInfo: ComponentInfo{
					ComponentName: testComponentName,
					ContainerName: "abcd",
				},
				ForcePush: false,
			},
			wantErr:            false,
			wantIsPushRequired: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execClient := exec.NewExecClient(kc)
			syncAdapter := NewSyncClient(kc, execClient)
			isPushRequired, err := syncAdapter.SyncFiles(context.Background(), tt.syncParameters)
			if !tt.wantErr && err != nil {
				t.Errorf("TestSyncFiles error: unexpected error when syncing files %v", err)
			} else if !tt.wantErr && isPushRequired != tt.wantIsPushRequired {
				t.Errorf("TestSyncFiles error: isPushRequired mismatch, wanted: %v, got: %v", tt.wantIsPushRequired, isPushRequired)
			}
		})
	}

	err := jsFile.Close()
	if err != nil {
		t.Errorf("TestSyncFiles error: error deleting the temp dir %s, err: %v", directory, err)
	}
}

func TestPushLocal(t *testing.T) {

	testComponentName := "test"

	// create a temp dir for the file indexer
	directory := t.TempDir()

	newFilePath := filepath.Join(directory, "foobar.txt")
	if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
		t.Errorf("TestPushLocal error: the foobar.txt file was not created: %v", err)
	}

	ctrl := gomock.NewController(t)
	kc := kclient.NewMockClientInterface(ctrl)
	kc.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Assert that Bar() is invoked.
	defer ctrl.Finish()

	syncClient := func(ComponentInfo, string, io.Reader) error {
		return nil
	}

	errorSyncClient := func(ComponentInfo, string, io.Reader) error {
		return errors.New("err")
	}

	tests := []struct {
		name        string
		client      SyncExtracter
		path        string
		files       []string
		delFiles    []string
		isForcePush bool
		compInfo    ComponentInfo
		wantErr     bool
	}{
		{
			name:        "Case 1: File change",
			client:      syncClient,
			path:        directory,
			files:       []string{path.Join(directory, "test.log")},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 2: File change with fake error client",
			client:      errorSyncClient,
			path:        directory,
			files:       []string{path.Join(directory, "test.log")},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: true,
		},
		{
			name:        "Case 3: No file change",
			client:      syncClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 4: Deleted file",
			client:      syncClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{path.Join(directory, "test.log")},
			isForcePush: false,
			compInfo: ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 5: Force push",
			client:      syncClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{},
			isForcePush: true,
			compInfo: ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 6: Source mapping folder set",
			client:      syncClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: ComponentInfo{
				ComponentName: testComponentName,
				ContainerName: "abcd",
				SyncFolder:    "/some/path",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execClient := exec.NewExecClient(kc)
			syncAdapter := NewSyncClient(kc, execClient)
			err := syncAdapter.pushLocal(context.Background(), tt.path, tt.files, tt.delFiles, tt.isForcePush, []string{}, tt.compInfo, util.IndexerRet{})
			if !tt.wantErr && err != nil {
				t.Errorf("TestPushLocal error: error pushing files: %v", err)
			}

		})
	}
}

func TestUpdateIndexWithWatchChanges(t *testing.T) {

	tests := []struct {
		name                 string
		initialFilesToCreate []string
		watchDeletedFiles    []string
		watchAddedFiles      []string
		expectedFilesInIndex []string
	}{
		{
			name:                 "Case 1 - Watch file deleted should remove file from index",
			initialFilesToCreate: []string{"file1", "file2"},
			watchDeletedFiles:    []string{"file1"},
			expectedFilesInIndex: []string{"file2"},
		},
		{
			name:                 "Case 2 - Watch file added should add file to index",
			initialFilesToCreate: []string{"file1"},
			watchAddedFiles:      []string{"file2"},
			expectedFilesInIndex: []string{"file1", "file2"},
		},
		{
			name:                 "Case 3 - No watch changes should mean no index changes",
			initialFilesToCreate: []string{"file1"},
			expectedFilesInIndex: []string{"file1"},
		},
	}
	for _, tt := range tests {

		// create a temp dir for the fake component
		directory := t.TempDir()

		fileIndexPath, err := util.ResolveIndexFilePath(directory)
		if err != nil {
			t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to resolve index file path: %v", err)
		}

		if err := os.MkdirAll(filepath.Dir(fileIndexPath), 0750); err != nil {
			t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to create directories for %s: %v", fileIndexPath, err)
		}

		t.Run(tt.name, func(t *testing.T) {

			indexData := map[string]util.FileData{}

			// Create initial files
			for _, fileToCreate := range tt.initialFilesToCreate {
				filePath := filepath.Join(directory, fileToCreate)

				if err := os.WriteFile(filePath, []byte("non-empty-string"), 0644); err != nil {
					t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to write to index file path: %v", err)
				}

				key, fileDatum, err := util.GenerateNewFileDataEntry(filePath, directory)
				if err != nil {
					t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to generate new file: %v", err)
				}
				indexData[key] = *fileDatum
			}

			// Write the index based on those files
			if err := util.WriteFile(indexData, fileIndexPath); err != nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to write index file: %v", err)
			}

			syncParams := SyncParameters{
				Path: directory,
			}

			// Add deleted files to pushParams (also delete the files)
			for _, deletedFile := range tt.watchDeletedFiles {
				deletedFilePath := filepath.Join(directory, deletedFile)
				syncParams.WatchDeletedFiles = append(syncParams.WatchDeletedFiles, deletedFilePath)

				if err := os.Remove(deletedFilePath); err != nil {
					t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to delete file %s %v", deletedFilePath, err)
				}
			}

			// Add added files to pushParams (also create the files)
			for _, addedFile := range tt.watchAddedFiles {
				addedFilePath := filepath.Join(directory, addedFile)
				syncParams.WatchFiles = append(syncParams.WatchFiles, addedFilePath)

				if err := os.WriteFile(addedFilePath, []byte("non-empty-string"), 0644); err != nil {
					t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to write to index file path: %v", err)
				}
			}

			if err := updateIndexWithWatchChanges(syncParams); err != nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal: unexpected error: %v", err)
			}

			postFileIndex, err := util.ReadFileIndex(fileIndexPath)
			if err != nil || postFileIndex == nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: read new file index: %v", err)
			}

			// Locate expected files
			if len(postFileIndex.Files) != len(tt.expectedFilesInIndex) {
				t.Fatalf("Mismatch between number expected files and actual files in index, post-index: %v   expected: %v", postFileIndex.Files, tt.expectedFilesInIndex)
			}
			for _, expectedFile := range tt.expectedFilesInIndex {
				if _, exists := postFileIndex.Files[expectedFile]; !exists {
					t.Fatalf("Unable to find '%s' in post index file, %v", expectedFile, postFileIndex.Files)
				}
			}
		})
	}
}
