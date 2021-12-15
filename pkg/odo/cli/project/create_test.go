package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreate(t *testing.T) {

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
		//		existingApps       []string
		wantProjectName string
		//wantErrValidate string
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
			args:            []string{"project-name-to-create"},
			wantProjectName: "project-name-to-create",
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
			opts := NewProjectCreateOptions(prjClient)

			/* COMPLETE */
			err = opts.Complete("create", cmdline, tt.args)

			if err != nil {
				t.Errorf("Expected nil error, got %s", err)
				return
			}
			if opts.projectName != tt.wantProjectName {
				t.Errorf("Got appName %q, expected %q", opts.projectName, tt.wantProjectName)
			}

			/* VALIDATE */
			err = opts.Validate()
			if err != nil {
				return
			}

			/* Mocks for Run */
			prjClient.EXPECT().Create(tt.wantProjectName, false).Times(1)
			prjClient.EXPECT().SetCurrent(tt.wantProjectName).Times(1)

			/* RUN */
			err = opts.Run()
		})
	}
}
