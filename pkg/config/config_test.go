package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	devfilev1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestLocalConfigInitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	filename, err := getLocalConfigFile("")
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(filename)

	conf, err := NewLocalConfigInfo("")
	if err != nil {
		t.Errorf("error while creating local config %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("local config.yaml shouldn't exist yet")
	}
}

func TestMetaTypePopulatedInLocalConfig(t *testing.T) {
	ci, err := NewLocalConfigInfo("")

	if err != nil {
		t.Error(err)
	}
	if ci.typeMeta.APIVersion != localConfigAPIVersion || ci.typeMeta.Kind != localConfigKind {
		t.Error("the api version and kind in local config are incorrect")
	}
}

func TestDeleteConfigDirIfEmpty(t *testing.T) {
	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a odo config directory on fake fs
	configDir, err := fs.TempDir(os.TempDir(), "odo")
	if err != nil {
		t.Error(err)
	}
	// create a mock local configuration from above fake fs & dir
	lci, err := mockLocalConfigInfo(configDir, fs)
	if err != nil {
		t.Error(err)
	}

	odoDir := filepath.Join(configDir, ".odo")
	if _, err = fs.Stat(odoDir); os.IsNotExist(err) {
		t.Error("config directory doesn't exist")
	}

	tests := []struct {
		name string
		// create indicates if a file is supposed to be created in the odo config dir
		create     bool
		setupEnv   func(create bool, fs filesystem.Filesystem, odoDir string) error
		wantOdoDir bool
		wantErr    bool
	}{
		{
			name:       "Case 1: Empty config dir",
			create:     false,
			setupEnv:   createDirectoryAndFile,
			wantOdoDir: false,
		},
		{
			name:       "Case 2: Config dir with test file",
			create:     true,
			setupEnv:   createDirectoryAndFile,
			wantOdoDir: true,
		},
	}

	for _, tt := range tests {

		err := tt.setupEnv(tt.create, fs, odoDir)
		if err != nil {
			t.Error(err)
		}

		err = lci.DeleteConfigDirIfEmpty()
		if err != nil {
			t.Error(err)
		}

		file, err := fs.Stat(odoDir)
		if !tt.wantOdoDir && !os.IsNotExist(err) {
			// we don't want odo dir but odo dir exists
			fmt.Println(file.Size())
			t.Error("odo config directory exists even after deleting it")
			t.Errorf("Error in test %q", tt.name)
		} else if tt.wantOdoDir && os.IsNotExist(err) {
			// we want odo dir to exist after odo delete --all but it does not exist
			t.Error("wanted odo directory to exist after odo delete --all")
			t.Errorf("Error in test %q", tt.name)
		}
	}
}

func TestSetDevfileConfiguration(t *testing.T) {

	// Use fakeFs
	fs := devfilefs.NewFakeFs()

	tests := []struct {
		name           string
		args           map[string]string
		currentDevfile parser.DevfileObj
		wantDevFile    parser.DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: set memory to 500Mi",
			args: map[string]string{
				"memory": "500Mi",
			},
			currentDevfile: testingutil.GetTestDevfileObj(fs),
			wantDevFile: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: &testingutil.TestDevfileData{
					Commands: []devfilev1.Command{
						{
							Id: "devbuild",
							CommandUnion: devfilev1.CommandUnion{
								Exec: &devfilev1.ExecCommand{
									WorkingDir: "/projects/nodejs-starter",
								},
							},
						},
					},
					Components: []devfilev1.Component{
						{
							Name: "runtime",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image:       "quay.io/nodejs-12",
										MemoryLimit: "500Mi",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       "port-3030",
											TargetPort: 3000,
										},
									},
								},
							},
						},
						{
							Name: "loadbalancer",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image:       "quay.io/nginx",
										MemoryLimit: "500Mi",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 2: set ports array",
			args: map[string]string{
				"ports": "8080,8081/UDP,8080/TCP",
			},
			currentDevfile: testingutil.GetTestDevfileObj(fs),
			wantDevFile: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: &testingutil.TestDevfileData{
					Commands: []devfilev1.Command{
						{
							Id: "devbuild",
							CommandUnion: devfilev1.CommandUnion{
								Exec: &devfilev1.ExecCommand{
									WorkingDir: "/projects/nodejs-starter",
								},
							},
						},
					},
					Components: []devfilev1.Component{
						{
							Name: "runtime",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       "port-3030",
											TargetPort: 3000,
										},
										{
											Name:       "port-8080-tcp",
											TargetPort: 8080,
											Protocol:   "tcp",
										}, {
											Name:       "port-8081-udp",
											TargetPort: 8081,
											Protocol:   "udp",
										},
									},
								},
							},
						},
						{
							Name: "loadbalancer",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nginx",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       "port-8080-tcp",
											TargetPort: 8080,
											Protocol:   "tcp",
										}, {
											Name:       "port-8081-udp",
											TargetPort: 8081,
											Protocol:   "udp",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 3: set ports array fails due to validation",
			args: map[string]string{
				"ports": "8080,8081/UDP,8083/",
			},
			currentDevfile: testingutil.GetTestDevfileObj(fs),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for key, value := range tt.args {
				err := SetDevfileConfiguration(tt.currentDevfile, key, value)
				if tt.wantErr {
					if err == nil {
						t.Errorf("expected error but got nil")
					}
					// we dont expect an error here
				} else {
					if err != nil {
						t.Errorf("error while setting configuration %+v", err.Error())
					}
				}
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
					t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
				}
			}

		})
	}

}

func createDirectoryAndFile(create bool, fs filesystem.Filesystem, odoDir string) error {
	if !create {
		return nil
	}

	file, err := fs.Create(filepath.Join(odoDir, "testfile"))
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("hello world"))
	if err != nil {
		return err
	}

	file.Close()
	if err != nil {
		return err
	}
	return nil
}
