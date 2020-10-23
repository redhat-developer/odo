package sync

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
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
		if !reflect.DeepEqual(tt.want, cmdArr) {
			t.Errorf("Expected %s, got %s", tt.want, cmdArr)
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
		if !reflect.DeepEqual(tt.want, cmdArr) {
			t.Errorf("Expected %s, got %s", tt.want, cmdArr)
		}
	}
}

func TestSyncFiles(t *testing.T) {

	testComponentName := "test"

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()
	// fkclient, _ := kclient.FakeNew() TODO: test kube fake client, oc fake client and sync

	// create a temp dir for the file indexer
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("TestSyncFiles error: error creating temporary directory for the indexer: %v", err)
	}

	tests := []struct {
		name               string
		client             SyncClient
		syncParameters     common.SyncParameters
		wantErr            bool
		wantIsPushRequired bool
	}{
		{
			name:   "Case 1: Component does not exist",
			client: fakeClient,
			syncParameters: common.SyncParameters{
				PushParams: common.PushParameters{
					Path:              directory,
					WatchFiles:        []string{},
					WatchDeletedFiles: []string{},
					IgnoredFiles:      []string{},
					ForceBuild:        false,
				},
				CompInfo: common.ComponentInfo{
					ContainerName: "abcd",
				},
				ComponentExists: false,
			},
			wantErr:            false,
			wantIsPushRequired: true,
		},
		{
			name:   "Case 2: Component does exist",
			client: fakeClient,
			syncParameters: common.SyncParameters{
				PushParams: common.PushParameters{
					Path:              directory,
					WatchFiles:        []string{},
					WatchDeletedFiles: []string{},
					IgnoredFiles:      []string{},
					ForceBuild:        false,
				},
				CompInfo: common.ComponentInfo{
					ContainerName: "abcd",
				},
				ComponentExists: true,
			},
			wantErr:            false,
			wantIsPushRequired: false, // always false after case 1
		},
		{
			name:   "Case 3: FakeErrorClient error",
			client: fakeErrorClient,
			syncParameters: common.SyncParameters{
				PushParams: common.PushParameters{
					Path:              directory,
					WatchFiles:        []string{},
					WatchDeletedFiles: []string{},
					IgnoredFiles:      []string{},
					ForceBuild:        true,
				},
				CompInfo: common.ComponentInfo{
					ContainerName: "abcd",
				},
				ComponentExists: true,
			},
			wantErr:            true,
			wantIsPushRequired: false,
		},
		{
			name:   "Case 4: File change",
			client: fakeClient,
			syncParameters: common.SyncParameters{
				PushParams: common.PushParameters{
					Path:              directory,
					WatchFiles:        []string{path.Join(directory, "test.log")},
					WatchDeletedFiles: []string{},
					IgnoredFiles:      []string{},
					ForceBuild:        false,
				},
				CompInfo: common.ComponentInfo{
					ContainerName: "abcd",
				},
				ComponentExists: true,
			},
			wantErr:            false,
			wantIsPushRequired: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []versionsCommon.DevfileComponent{},
				},
			}

			adapterCtx := common.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			syncAdapter := New(adapterCtx, tt.client)
			isPushRequired, err := syncAdapter.SyncFiles(tt.syncParameters)
			if !tt.wantErr && err != nil {
				t.Errorf("TestSyncFiles error: unexpected error when syncing files %v", err)
			} else if !tt.wantErr && isPushRequired != tt.wantIsPushRequired {
				t.Errorf("TestSyncFiles error: isPushRequired mismatch, wanted: %v, got: %v", tt.wantIsPushRequired, isPushRequired)
			}
		})
	}

	// Remove the temp dir created for the file indexer
	err = os.RemoveAll(directory)
	if err != nil {
		t.Errorf("TestSyncFiles error: error deleting the temp dir %s", directory)
	}
}

func TestPushLocal(t *testing.T) {

	testComponentName := "test"

	// create a temp dir for the file indexer
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("TestPushLocal error: error creating temporary directory for the indexer: %v", err)
	}

	newFilePath := filepath.Join(directory, "foobar.txt")
	if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
		t.Errorf("TestPushLocal error: the foobar.txt file was not created: %v", err)
	}

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()
	// fkclient, _ := kclient.FakeNew() TODO: test kube fake client, oc fake client and sync

	tests := []struct {
		name        string
		client      SyncClient
		path        string
		files       []string
		delFiles    []string
		isForcePush bool
		compInfo    common.ComponentInfo
		wantErr     bool
	}{
		{
			name:        "Case 1: File change",
			client:      fakeClient,
			path:        directory,
			files:       []string{path.Join(directory, "test.log")},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: common.ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 2: File change with fake error client",
			client:      fakeErrorClient,
			path:        directory,
			files:       []string{path.Join(directory, "test.log")},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: common.ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: true,
		},
		{
			name:        "Case 3: No file change",
			client:      fakeClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: common.ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 4: Deleted file",
			client:      fakeClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{path.Join(directory, "test.log")},
			isForcePush: false,
			compInfo: common.ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 5: Force push",
			client:      fakeClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{},
			isForcePush: true,
			compInfo: common.ComponentInfo{
				ContainerName: "abcd",
			},
			wantErr: false,
		},
		{
			name:        "Case 6: Source mapping folder set",
			client:      fakeClient,
			path:        directory,
			files:       []string{},
			delFiles:    []string{},
			isForcePush: false,
			compInfo: common.ComponentInfo{
				ContainerName: "abcd",
				SourceMount:   "/some/path",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []versionsCommon.DevfileComponent{},
				},
			}

			adapterCtx := common.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			syncAdapter := New(adapterCtx, tt.client)
			err := syncAdapter.pushLocal(tt.path, tt.files, tt.delFiles, tt.isForcePush, []string{}, tt.compInfo)
			if !tt.wantErr && err != nil {
				t.Errorf("TestPushLocal error: error pushing files: %v", err)
			}

		})
	}

	// Remove the temp dir created for the file indexer
	err = os.RemoveAll(directory)
	if err != nil {
		t.Errorf("TestPushLocal error: error deleting the temp dir %s", directory)
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
		directory, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: error creating temporary directory for the indexer: %v", err)
		}

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

				if err := ioutil.WriteFile(filePath, []byte("non-empty-string"), 0644); err != nil {
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

			pushParams := common.PushParameters{
				Path: directory,
			}

			// Add deleted files to pushParams (also delete the files)
			for _, deletedFile := range tt.watchDeletedFiles {
				deletedFilePath := filepath.Join(directory, deletedFile)
				pushParams.WatchDeletedFiles = append(pushParams.WatchDeletedFiles, deletedFilePath)

				if err := os.Remove(deletedFilePath); err != nil {
					t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to delete file %s %v", deletedFilePath, err)
				}
			}

			// Add added files to pushParams (also create the files)
			for _, addedFile := range tt.watchAddedFiles {
				addedFilePath := filepath.Join(directory, addedFile)
				pushParams.WatchFiles = append(pushParams.WatchFiles, addedFilePath)

				if err := ioutil.WriteFile(addedFilePath, []byte("non-empty-string"), 0644); err != nil {
					t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to write to index file path: %v", err)
				}
			}

			if err := updateIndexWithWatchChanges(pushParams); err != nil {
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

// func TestGetSyncFolder(t *testing.T) {
// 	projectNames := []string{"some-name", "another-name"}
// 	projectRepos := []string{"https://github.com/some/repo.git", "https://github.com/another/repo.git"}
// 	projectClonePath := "src/github.com/golang/example/"
// 	invalidClonePaths := []string{"/var", "../var", "pkg/../../var"}
// 	sourceVolumePath := "/projects/app"

// 	tests := []struct {
// 		name     string
// 		projects []versionsCommon.DevfileProject
// 		want     string
// 		wantErr  bool
// 	}{
// 		{
// 			name:     "Case 1: No projects",
// 			projects: []versionsCommon.DevfileProject{},
// 			want:     sourceVolumePath,
// 			wantErr:  false,
// 		},
// 		{
// 			name: "Case 2: One project",
// 			projects: []versionsCommon.DevfileProject{
// 				{
// 					Name: projectNames[0],
// 					Git: &versionsCommon.Git{
// 						GitLikeProjectSource: versionsCommon.GitLikeProjectSource{
// 							Remotes: map[string]string{"origin": projectRepos[0]},
// 						},
// 					},
// 				},
// 			},
// 			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectNames[0])),
// 			wantErr: false,
// 		},
// 		{
// 			name: "Case 3: Multiple projects",
// 			projects: []versionsCommon.DevfileProject{
// 				{
// 					Name: projectNames[0],
// 					Git: &versionsCommon.Git{
// 						GitLikeProjectSource: versionsCommon.GitLikeProjectSource{
// 							Remotes: map[string]string{"origin": projectRepos[0]},
// 						},
// 					},
// 				},
// 				{
// 					Name: projectNames[1],
// 					Github: &versionsCommon.Github{
// 						GitLikeProjectSource: versionsCommon.GitLikeProjectSource{
// 							Remotes: map[string]string{"origin": projectRepos[1]},
// 						},
// 					},
// 				},
// 				{
// 					Name: projectNames[1],
// 					Zip: &versionsCommon.Zip{
// 						Location: projectRepos[1],
// 					},
// 				},
// 			},
// 			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectNames[0])),
// 			wantErr: false,
// 		},
// 		{
// 			name: "Case 4: Clone path set",
// 			projects: []versionsCommon.DevfileProject{
// 				{
// 					ClonePath: projectClonePath,
// 					Name:      projectNames[0],
// 					Zip: &versionsCommon.Zip{
// 						Location: projectRepos[0],
// 					},
// 				},
// 			},
// 			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectClonePath)),
// 			wantErr: false,
// 		},
// 		{
// 			name: "Case 5: Invalid clone path, set with absolute path",
// 			projects: []versionsCommon.DevfileProject{
// 				{
// 					ClonePath: invalidClonePaths[0],
// 					Name:      projectNames[0],
// 					Github: &versionsCommon.Github{
// 						GitLikeProjectSource: versionsCommon.GitLikeProjectSource{
// 							Remotes: map[string]string{"origin": projectRepos[0]},
// 						},
// 					},
// 				},
// 			},
// 			want:    "",
// 			wantErr: true,
// 		},
// 		{
// 			name: "Case 6: Invalid clone path, starts with ..",
// 			projects: []versionsCommon.DevfileProject{
// 				{
// 					ClonePath: invalidClonePaths[1],
// 					Name:      projectNames[0],
// 					Git: &versionsCommon.Git{
// 						GitLikeProjectSource: versionsCommon.GitLikeProjectSource{
// 							Remotes: map[string]string{"origin": projectRepos[0]},
// 						},
// 					},
// 				},
// 			},
// 			want:    "",
// 			wantErr: true,
// 		},
// 		{
// 			name: "Case 7: Invalid clone path, contains ..",
// 			projects: []versionsCommon.DevfileProject{
// 				{
// 					ClonePath: invalidClonePaths[2],
// 					Name:      projectNames[0],
// 					Zip: &versionsCommon.Zip{
// 						Location: projectRepos[0],
// 					},
// 				},
// 			},
// 			want:    "",
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			syncFolder, err := GetSyncFolder(sourceVolumePath, tt.projects)

// 			if !tt.wantErr == (err != nil) {
// 				t.Errorf("expected %v, actual %v", tt.wantErr, err)
// 			}

// 			if syncFolder != tt.want {
// 				t.Errorf("expected %s, actual %s", tt.want, syncFolder)
// 			}
// 		})
// 	}
// }
