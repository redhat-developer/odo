package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestOdoAlizer(t *testing.T) {

	for _, tt := range []struct {
		name            string
		clientset       func() clientset.Clientset
		fsPopulator     func(fs filesystem.Filesystem) error
		args            []string
		wantErr         string
		wantStdout      string
		wantStderr      string
		checkJsonOutput func(t *testing.T, b []byte)
	}{
		{
			name: "analyze without json output",
			clientset: func() clientset.Clientset {
				return clientset.Clientset{}
			},
			args:    []string{"analyze"},
			wantErr: "this command can be run with json output only, please use the flag: -o json",
		},
		{
			name: "analyze with json output in an empty directory",
			clientset: func() clientset.Clientset {
				return clientset.Clientset{}
			},
			args:    []string{"analyze", "-o", "json"},
			wantErr: "No valid devfile found for project in",
		},
		{
			name: "analyze with json output",
			clientset: func() clientset.Clientset {
				ctrl := gomock.NewController(t)
				fs := filesystem.NewFakeFs()
				alizerClient := alizer.NewMockClient(ctrl)
				path := "/"
				alizerClient.EXPECT().DetectFramework(gomock.Any(), path).
					Return(
						model.DevFileType{
							Name: "framework-name",
						},
						"1.1.1",
						api.Registry{
							Name: "TheRegistryName",
						},
						nil,
					)
				alizerClient.EXPECT().DetectPorts(path).Return([]int{8080, 3000}, nil)
				alizerClient.EXPECT().DetectName(path).Return("aName", nil)
				return clientset.Clientset{
					FS:           fs,
					AlizerClient: alizerClient,
				}
			},
			args: []string{"analyze", "-o", "json"},
			checkJsonOutput: func(t *testing.T, b []byte) {
				var output []api.DetectionResult
				err := json.Unmarshal(b, &output)
				if err != nil {
					t.Fatal(err)
				}
				checkEqual(t, output[0].Devfile, "framework-name")
				checkEqual(t, output[0].DevfileRegistry, "TheRegistryName")
				checkEqual(t, output[0].Name, "aName")
				checkEqual(t, output[0].DevfileVersion, "1.1.1")
				checkEqual(t, output[0].ApplicationPorts[0], 8080)
				checkEqual(t, output[0].ApplicationPorts[1], 3000)
			},
		},
		{
			name: "analyze should not error out even if there is an invalid Devfile in the current directory",
			clientset: func() clientset.Clientset {
				ctrl := gomock.NewController(t)
				fs := filesystem.NewFakeFs()
				alizerClient := alizer.NewMockClient(ctrl)
				path := "/"
				alizerClient.EXPECT().DetectFramework(gomock.Any(), path).
					Return(
						model.DevFileType{
							Name: "framework-name",
						},
						"1.1.1",
						api.Registry{
							Name: "TheRegistryName",
						},
						nil,
					)
				alizerClient.EXPECT().DetectPorts(path).Return([]int{8080, 3000}, nil)
				alizerClient.EXPECT().DetectName(path).Return("aName", nil)
				return clientset.Clientset{
					FS:           fs,
					AlizerClient: alizerClient,
				}
			},
			fsPopulator: func(fs filesystem.Filesystem) error {
				cwd, err := fs.Getwd()
				if err != nil {
					return err
				}
				err = fs.WriteFile(
					filepath.Join(cwd, "main.go"),
					[]byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}
`),
					0644)
				if err != nil {
					return err
				}

				return fs.WriteFile(filepath.Join(cwd, "devfile.yaml"), []byte("some-invalid-content"), 0644)
			},
			args: []string{"analyze", "-o", "json"},
			checkJsonOutput: func(t *testing.T, b []byte) {
				var output []api.DetectionResult
				err := json.Unmarshal(b, &output)
				if err != nil {
					t.Fatal(err)
				}
				checkEqual(t, output[0].Devfile, "framework-name")
				checkEqual(t, output[0].DevfileRegistry, "TheRegistryName")
				checkEqual(t, output[0].Name, "aName")
				checkEqual(t, output[0].DevfileVersion, "1.1.1")
				checkEqual(t, output[0].ApplicationPorts[0], 8080)
				checkEqual(t, output[0].ApplicationPorts[1], 3000)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cs := clientset.Clientset{}
			if tt.clientset != nil {
				cs = tt.clientset()
			}
			runCommand(t, tt.args, runOptions{}, cs, tt.fsPopulator, func(err error, stdout, stderr string) {
				if (err != nil) != (tt.wantErr != "") {
					t.Fatalf("errWanted: %v\nGot: %v", tt.wantErr != "", err)
				}

				if tt.wantErr != "" {
					if !strings.Contains(err.Error(), tt.wantErr) {
						t.Errorf("%q\nerror does not contain:\n%q", err.Error(), tt.wantErr)
					}
				}

				if tt.wantStdout != "" {
					if !strings.Contains(stdout, tt.wantStdout) {
						t.Errorf("%q\nstdout does not contain:\n%q", stdout, tt.wantStdout)
					}
				}

				if tt.wantStderr != "" {
					if !strings.Contains(stderr, tt.wantStderr) {
						t.Errorf("%q\nstderr does not contain:\n%q", stderr, tt.wantStderr)
					}
				}

				if tt.checkJsonOutput != nil {
					tt.checkJsonOutput(t, []byte(stdout))
				}
			})
		})
	}
}
