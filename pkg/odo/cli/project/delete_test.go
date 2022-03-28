package project

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/project"
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
		projectExists      bool
		wantProjectName    string
		wantErrValidate    string
	}{
		{
			name: "project from args",
			populateWorkingDir: func(fs filesystem.Filesystem) {
				_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
				env, er := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
				if er != nil {
					return
				}
				_ = env.SetComponentSettings(envinfo.ComponentSettings{
					Name:    "a-name",
					Project: "a-project",
					AppName: "an-app-name",
				})
			},
			args:            []string{"project-name-to-delete"},
			projectExists:   true,
			wantProjectName: "project-name-to-delete",
		}, {
			name: "project from args not existing",
			populateWorkingDir: func(fs filesystem.Filesystem) {
				_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
				env, er := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
				if er != nil {
					return
				}
				_ = env.SetComponentSettings(envinfo.ComponentSettings{
					Name:    "a-name",
					Project: "a-project",
					AppName: "an-app-name",
				})
			},
			args:            []string{"project-name-to-delete"},
			projectExists:   false,
			wantProjectName: "project-name-to-delete",
			wantErrValidate: `The project "project-name-to-delete" does not exist`,
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
			cmdline.EXPECT().GetWorkingDirectory().Return(workingDir, nil).AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("project").Return("").AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("app").Return("").AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("component").Return("").AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("o").Return("").AnyTimes()
			cmdline.EXPECT().CheckIfConfigurationNeeded().Return(false, nil).AnyTimes()
			cmdline.EXPECT().Context().Return(context.Background()).AnyTimes()

			// Fake odo Kube client
			kclient := kclient.NewMockClientInterface(ctrl)
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
			prjClient := project.NewMockClient(ctrl)
			opts := NewProjectDeleteOptions()
			opts.SetClientset(&clientset.Clientset{
				ProjectClient: prjClient,
			})
			opts.forceFlag = true

			/* COMPLETE */
			err = opts.Complete(cmdline, tt.args)

			if err != nil {
				t.Errorf("Expected nil error, got %s", err)
				return
			}
			if opts.projectName != tt.wantProjectName {
				t.Errorf("Got appName %q, expected %q", opts.projectName, tt.wantProjectName)
			}

			/* Mocks for Validate */
			prjClient.EXPECT().Exists(tt.wantProjectName).Return(tt.projectExists, nil).Times(1)

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
			/* Mocks for Run */
			prjClient.EXPECT().SetCurrent(tt.wantProjectName).Times(1)
			prjClient.EXPECT().Delete(tt.wantProjectName, false).Times(1)

			/* RUN */
			err = opts.Run(cmdline.Context())
		})
	}
}
