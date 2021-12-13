package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDelete(t *testing.T) {

	prefixDir, err := os.MkdirTemp(os.TempDir(), "unittests-")
	if err != nil {
		t.Errorf("Error creating temp directory for tests")
		return
	}
	workingDir := filepath.Join(prefixDir, "myapp")

	tests := []struct {
		name               string
		populateWorkingDir func(fs filesystem.Filesystem)
		args               []string
		existingApps       []string
		wantAppName        string
		wantErrValidate    string
	}{
		{
			name: "default app",
			populateWorkingDir: func(fs filesystem.Filesystem) {
				_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
				env, err := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
				if err != nil {
					return
				}
				_ = env.SetComponentSettings(envinfo.ComponentSettings{
					Name:    "a-name",
					Project: "a-project",
					AppName: "an-app-name",
				})
			},
			existingApps: []string{"an-app-name", "another-app-name"},
			wantAppName:  "an-app-name",
		},
		{
			name: "app from args",
			populateWorkingDir: func(fs filesystem.Filesystem) {
				_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
				env, err := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
				if err != nil {
					return
				}
				_ = env.SetComponentSettings(envinfo.ComponentSettings{
					Name:    "a-name",
					Project: "a-project",
					AppName: "an-app-name",
				})
			},
			args:         []string{"another-app-name"},
			existingApps: []string{"an-app-name", "another-app-name"},
			wantAppName:  "another-app-name",
		},
		{
			name: "empty app name",
			populateWorkingDir: func(fs filesystem.Filesystem) {
				_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
				env, err := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
				if err != nil {
					return
				}
				_ = env.SetComponentSettings(envinfo.ComponentSettings{
					Name:    "a-name",
					Project: "a-project",
					AppName: "",
				})
			},
			existingApps:    []string{"an-app-name", "another-app-name"},
			wantAppName:     "",
			wantErrValidate: "Please specify the application name and project name",
		},
		{
			name: "non existing app name",
			populateWorkingDir: func(fs filesystem.Filesystem) {
				_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
				env, err := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
				if err != nil {
					return
				}
				_ = env.SetComponentSettings(envinfo.ComponentSettings{
					Name:    "a-name",
					Project: "a-project",
					AppName: "an-app-name",
				})
			},
			args:            []string{"an-unknown-app-name"},
			existingApps:    []string{"an-app-name", "another-app-name"},
			wantAppName:     "an-unknown-app-name",
			wantErrValidate: " app does not exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// the first one is to cleanup the directory before execution (in case there are remaining files from a previous execution)
			os.RemoveAll(prefixDir)
			// the second one to cleanup after execution
			defer os.RemoveAll(prefixDir)

			// Fake Cobra
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cmdline := cmdline.NewMockCmdline(ctrl)

			// Fake odo Kube client
			kclient := kclient.NewMockClientInterface(ctrl)

			/* Mocks for Complete */
			cmdline.EXPECT().GetWorkingDirectory().Return(workingDir, nil).AnyTimes()
			cmdline.EXPECT().CheckIfConfigurationNeeded().Return(false, nil).AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("project").Return("").AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("app").Return("").AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("component").Return("").AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("o").Return("").AnyTimes()
			cmdline.EXPECT().GetKubeClient().Return(kclient, nil).AnyTimes()

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "a-project",
				},
			}
			kclient.EXPECT().GetNamespaceNormal("a-project").Return(ns, nil).AnyTimes()
			kclient.EXPECT().SetNamespace("a-project").AnyTimes()

			tt.populateWorkingDir(filesystem.DefaultFs{})

			/* Mocks for Complete */
			appClient := application.NewMockClient(ctrl)
			appClient.EXPECT().Exists(tt.wantAppName).Return(func() bool {
				for _, app := range tt.existingApps {
					if tt.wantAppName == app {
						return true
					}
				}
				return false
			}(), nil).AnyTimes()
			opts := NewDeleteOptions(appClient)

			/* COMPLETE */
			err := opts.Complete("delete", cmdline, tt.args)

			if err != nil {
				return
			}
			if opts.appName != tt.wantAppName {
				t.Errorf("Got appName %q, expected %q", opts.appName, tt.wantAppName)
			}

			/* VALIDATE */
			err = opts.Validate()

			if err == nil && tt.wantErrValidate != "" {
				t.Errorf("Expected %v, got no error", tt.wantErrValidate)
				return
			}
			if err != nil && tt.wantErrValidate == "" {
				t.Errorf("Expected no error, got %v", err.Error())
				return
			}
			if err != nil && tt.wantErrValidate != "" && !strings.Contains(err.Error(), tt.wantErrValidate) {
				t.Errorf("Expected error %v, got %v", tt.wantErrValidate, err.Error())
				return
			}
			if err != nil {
				return
			}
		})
	}
}
