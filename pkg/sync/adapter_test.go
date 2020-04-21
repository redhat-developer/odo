package sync

import (
	"path/filepath"
	"reflect"
	"testing"

	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient"
)

func TestGetSyncFolder(t *testing.T) {
	projectNames := []string{"some-name", "another-name"}
	projectRepos := []string{"https://github.com/some/repo.git", "https://github.com/another/repo.git"}
	projectClonePath := "src/github.com/golang/example/"
	invalidClonePaths := []string{"/var", "../var", "pkg/../../var"}

	tests := []struct {
		name     string
		projects []versionsCommon.DevfileProject
		want     string
		wantErr  bool
	}{
		{
			name:     "Case 1: No projects",
			projects: []versionsCommon.DevfileProject{},
			want:     kclient.OdoSourceVolumeMount,
			wantErr:  false,
		},
		{
			name: "Case 2: One project",
			projects: []versionsCommon.DevfileProject{
				{
					Name: projectNames[0],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[0],
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(kclient.OdoSourceVolumeMount, projectNames[0])),
			wantErr: false,
		},
		{
			name: "Case 3: Multiple projects",
			projects: []versionsCommon.DevfileProject{
				{
					Name: projectNames[0],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[0],
					},
				},
				{
					Name: projectNames[1],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[1],
					},
				},
			},
			want:    kclient.OdoSourceVolumeMount,
			wantErr: false,
		},
		{
			name: "Case 4: Clone path set",
			projects: []versionsCommon.DevfileProject{
				{
					ClonePath: &projectClonePath,
					Name:      projectNames[0],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[0],
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(kclient.OdoSourceVolumeMount, projectClonePath)),
			wantErr: false,
		},
		{
			name: "Case 5: Invalid clone path, set with absolute path",
			projects: []versionsCommon.DevfileProject{
				{
					ClonePath: &invalidClonePaths[0],
					Name:      projectNames[0],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[0],
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case 6: Invalid clone path, starts with ..",
			projects: []versionsCommon.DevfileProject{
				{
					ClonePath: &invalidClonePaths[1],
					Name:      projectNames[0],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[0],
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case 7: Invalid clone path, contains ..",
			projects: []versionsCommon.DevfileProject{
				{
					ClonePath: &invalidClonePaths[2],
					Name:      projectNames[0],
					Source: versionsCommon.DevfileProjectSource{
						Type:     versionsCommon.DevfileProjectTypeGit,
						Location: projectRepos[0],
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		syncFolder, err := getSyncFolder(tt.projects)

		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, actual %v", tt.wantErr, err)
		}

		if syncFolder != tt.want {
			t.Errorf("expected %s, actual %s", tt.want, syncFolder)
		}
	}
}

func TestGetCmdToCreateSyncFolder(t *testing.T) {
	tests := []struct {
		name       string
		syncFolder string
		want       []string
	}{
		{
			name:       "Case 1: Sync to /projects",
			syncFolder: kclient.OdoSourceVolumeMount,
			want:       []string{"mkdir", "-p", kclient.OdoSourceVolumeMount},
		},
		{
			name:       "Case 2: Sync subdir of /projects",
			syncFolder: kclient.OdoSourceVolumeMount + "/someproject",
			want:       []string{"mkdir", "-p", kclient.OdoSourceVolumeMount + "/someproject"},
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
			syncFolder: kclient.OdoSourceVolumeMount,
			want:       []string{"rm", "-rf", kclient.OdoSourceVolumeMount + "/test.txt"},
		},
		{
			name:       "Case 2: Multiple deleted files, default sync folder",
			delFiles:   []string{"test.txt", "hello.c"},
			syncFolder: kclient.OdoSourceVolumeMount,
			want:       []string{"rm", "-rf", kclient.OdoSourceVolumeMount + "/test.txt", kclient.OdoSourceVolumeMount + "/hello.c"},
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
