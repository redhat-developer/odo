package envinfo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/util"

	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestSetEnvInfo(t *testing.T) {
	fs := filesystem.NewFakeFs()
	tempEnvFile, err := fs.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testURL := localConfigProvider.LocalURL{Name: "testURL", Host: "1.2.3.4.nip.io", TLSSecret: "testTLSSecret"}
	invalidParam := "invalidParameter"

	tests := []struct {
		name               string
		parameter          string
		value              interface{}
		existingEnvInfo    EnvInfo
		checkConfigSetting []string
		expectError        bool
	}{
		{
			name:      fmt.Sprintf("Case 1: %s to test", URL),
			parameter: URL,
			value:     testURL,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{},
			},
			checkConfigSetting: []string{"URL"},
			expectError:        false,
		},
		{
			name:      fmt.Sprintf("Case 2: %s to test", invalidParam),
			parameter: invalidParam,
			value:     testURL,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{},
			},
			checkConfigSetting: []string{"URL"},
			expectError:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi, err := NewEnvSpecificInfo("")
			if err != nil {
				t.Error(err)
			}
			esi.EnvInfo = tt.existingEnvInfo
			err = esi.SetConfiguration(tt.parameter, tt.value)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for SetConfiguration with %s: %v", tt.parameter, err)
			} else if !tt.expectError && err == nil {
				isSet := false
				for _, configSetting := range tt.checkConfigSetting {
					isSet = esi.IsSet(configSetting)
					if !isSet {
						t.Errorf("the setting '%s' is not set", configSetting)
					}
				}

			}

		})
	}
}

func TestUnsetEnvInfo(t *testing.T) {
	fs := filesystem.NewFakeFs()
	tempEnvFile, err := fs.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testURL := localConfigProvider.LocalURL{Name: "testURL", Host: "1.2.3.4.nip.io", TLSSecret: "testTLSSecret"}
	invalidParam := "invalidParameter"

	tests := []struct {
		name            string
		parameter       string
		existingEnvInfo EnvInfo
		expectError     bool
	}{
		{
			name:      fmt.Sprintf("Case 1: unset %s", URL),
			parameter: URL,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{testURL},
				},
			},
			expectError: false,
		},
		{
			name:      fmt.Sprintf("Case 2: unset %s", invalidParam),
			parameter: invalidParam,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{testURL},
				},
			},
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi, err := NewEnvSpecificInfo("")
			if err != nil {
				t.Error(err)
			}
			esi.EnvInfo = tt.existingEnvInfo
			err = esi.DeleteConfiguration(tt.parameter)
			if err == nil && tt.expectError {
				t.Errorf("expected error for DeleteConfiguration with %s", tt.parameter)
			} else if !tt.expectError {
				if err != nil {
					t.Error(err)
				}
				isSet := esi.IsSet(tt.parameter)
				if isSet {
					t.Errorf("the '%v' is not set to nil", tt.parameter)
				}
			}

		})
	}
}

func TestDeleteURLFromMultipleURLs(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()
	tempEnvFile, err := ioutil.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testURL1 := localConfigProvider.LocalURL{Name: "testURL1", Host: "1.2.3.4.nip.io", TLSSecret: "testTLSSecret"}
	testURL2 := localConfigProvider.LocalURL{Name: "testURL2", Host: "1.2.3.4.nip.io", TLSSecret: "testTLSSecret"}

	tests := []struct {
		name            string
		existingEnvInfo EnvInfo
		existingDevfile parser.DevfileObj
		deleteParam     string
		remainingParam  string
		singleURL       bool
		wantErr         bool
	}{
		{
			name: fmt.Sprintf("Case 1: delete %s from multiple URLs", testURL1.Name),
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{testURL1, testURL2},
				},
			},
			existingDevfile: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "runtime",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       testURL1.Name,
											TargetPort: 3000,
										},
										{
											Name:       testURL2.Name,
											TargetPort: 8080,
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			deleteParam:    testURL1.Name,
			remainingParam: testURL2.Name,
			singleURL:      false,
		},
		{
			name: fmt.Sprintf("Case 2: delete %s fro URL array with single element", testURL1.Name),
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					URL: &[]localConfigProvider.LocalURL{testURL1},
				},
			},
			existingDevfile: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "runtime",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       testURL1.Name,
											TargetPort: 3000,
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			deleteParam: testURL1.Name,
			singleURL:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi, err := NewEnvSpecificInfo("")
			if err != nil {
				t.Error(err)
			}
			esi.EnvInfo = tt.existingEnvInfo
			esi.SetDevfileObj(tt.existingDevfile)

			urls, err := esi.ListURLs()
			if err != nil {
				t.Error(err)
			}
			oldURLLength := len(urls)
			err = esi.DeleteURL(tt.deleteParam)
			if err != nil {
				t.Error(err)
			}

			urls, err = esi.ListURLs()
			if err != nil {
				t.Error(err)
			}
			newURLLength := len(urls)
			if newURLLength+1 != oldURLLength {
				t.Errorf("DeleteURL is expected to delete element %s from the URL array.", tt.deleteParam)
			}
			if tt.singleURL {
				if newURLLength != 0 {
					t.Errorf("Expect to have empty URL array if delete URL from URL array with only 1 element")
				}
			} else {
				if urls[0].Name != tt.remainingParam {
					t.Errorf("Expect to have element %s in the URL array", tt.remainingParam)
				}
			}

		})
	}

}

func TestEnvSpecificInfonitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	filename, _, err := getEnvInfoFile("")
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(filename)

	conf, err := NewEnvSpecificInfo("")
	if err != nil {
		t.Errorf("error while creating envinfo %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("local env.yaml shouldn't exist yet")
	}
}

func TestDeleteEnvDirIfEmpty(t *testing.T) {
	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a odo config directory on fake fs
	configDir, err := fs.TempDir(os.TempDir(), "odo")
	if err != nil {
		t.Error(err)
	}
	// create a mock env info from above fake fs & dir
	esi, err := mockEnvSpecificInfo(configDir, fs)
	if err != nil {
		t.Error(err)
	}

	envDir := filepath.Join(configDir, ".odo", "env")
	if _, err = fs.Stat(envDir); os.IsNotExist(err) {
		t.Error("config directory doesn't exist")
	}

	tests := []struct {
		name string
		// create indicates if a file is supposed to be created in the odo config dir
		create     bool
		setupEnv   func(create bool, fs filesystem.Filesystem, envDir string) error
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

		err := tt.setupEnv(tt.create, fs, envDir)
		if err != nil {
			t.Error(err)
		}

		err = esi.DeleteEnvDirIfEmpty()
		if err != nil {
			t.Error(err)
		}

		file, err := fs.Stat(envDir)
		if !tt.wantOdoDir && !os.IsNotExist(err) {
			// we don't want odo dir but odo dir exists
			fmt.Println(file.Size())
			t.Error("odo env directory exists even after deleting it")
			t.Errorf("Error in test %q", tt.name)
		} else if tt.wantOdoDir && os.IsNotExist(err) {
			// we want odo dir to exist after odo delete --all but it does not exist
			t.Error("wanted odo directory to exist after odo delete --all")
			t.Errorf("Error in test %q", tt.name)
		}
	}
}

func TestAddEndpointInDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()
	urlName := "testURL"
	urlName2 := "testURL2"
	tests := []struct {
		name           string
		devObj         parser.DevfileObj
		endpoint       devfilev1.Endpoint
		container      string
		wantComponents []devfilev1.Component
	}{
		{
			name: "Case 1: devfile has single container with existing endpoint",
			endpoint: devfilev1.Endpoint{
				Name:       urlName,
				TargetPort: 8080,
				Secure:     util.GetBoolPtr(false),
			},
			container: "testcontainer1",
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
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
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
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
									Name:       urlName,
									TargetPort: 8080,
									Secure:     util.GetBoolPtr(false),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Case 2: devfile has single container with no endpoint",
			endpoint: devfilev1.Endpoint{
				Name:       urlName,
				TargetPort: 8080,
				Secure:     util.GetBoolPtr(false),
			},
			container: "testcontainer1",
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "quay.io/nodejs-12",
							},
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Secure:     util.GetBoolPtr(false),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Case 3: devfile has multiple containers",
			endpoint: devfilev1.Endpoint{
				Name:       urlName,
				TargetPort: 8080,
				Secure:     util.GetBoolPtr(false),
			},
			container: "testcontainer1",
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
								},
							},
						},
						{
							Name: "testcontainer2",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       urlName2,
											TargetPort: 9090,
											Secure:     util.GetBoolPtr(true),
											Path:       "/testpath",
											Exposure:   devfilev1.InternalEndpointExposure,
											Protocol:   devfilev1.HTTPSEndpointProtocol,
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "quay.io/nodejs-12",
							},
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Secure:     util.GetBoolPtr(false),
								},
							},
						},
					},
				},
				{
					Name: "testcontainer2",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       urlName2,
									TargetPort: 9090,
									Secure:     util.GetBoolPtr(true),
									Path:       "/testpath",
									Exposure:   devfilev1.InternalEndpointExposure,
									Protocol:   devfilev1.HTTPSEndpointProtocol,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addEndpointInDevfile(tt.devObj, tt.endpoint, tt.container)
			if err != nil {
				t.Errorf("Unexpected err from UpdateEndpointsInDevfile: %v", err)
			}
			components, err := tt.devObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Errorf("Unexpected err from addEndpointInDevfile: %v", err)
			}
			if !reflect.DeepEqual(components, tt.wantComponents) {
				t.Errorf("Expected: %v, got %v", tt.wantComponents, components)
			}

		})
	}
}

func TestRemoveEndpointInDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()
	urlName := "testURL"
	urlName2 := "testURL2"
	tests := []struct {
		name           string
		devObj         parser.DevfileObj
		endpoint       devfilev1.Endpoint
		urlName        string
		wantComponents []devfilev1.Component
		wantErr        bool
	}{
		{
			name:    "Case 1: devfile has single container with multiple existing endpoint",
			urlName: urlName,
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
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
											Name:       urlName,
											TargetPort: 8080,
											Secure:     util.GetBoolPtr(false),
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
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
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Case 2: devfile has single container with a single endpoint",
			urlName: urlName,
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       urlName,
											TargetPort: 8080,
											Secure:     util.GetBoolPtr(false),
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "quay.io/nodejs-12",
							},
							Endpoints: []devfilev1.Endpoint{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Case 3: devfile has multiple containers",
			urlName: urlName,
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       urlName,
											TargetPort: 8080,
											Secure:     util.GetBoolPtr(false),
										},
									},
								},
							},
						},
						{
							Name: "testcontainer2",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       urlName2,
											TargetPort: 9090,
											Secure:     util.GetBoolPtr(true),
											Path:       "/testpath",
											Exposure:   devfilev1.InternalEndpointExposure,
											Protocol:   devfilev1.HTTPSEndpointProtocol,
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "quay.io/nodejs-12",
							},
							Endpoints: []devfilev1.Endpoint{},
						},
					},
				},
				{
					Name: "testcontainer2",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       urlName2,
									TargetPort: 9090,
									Secure:     util.GetBoolPtr(true),
									Path:       "/testpath",
									Exposure:   devfilev1.InternalEndpointExposure,
									Protocol:   devfilev1.HTTPSEndpointProtocol,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Case 4: delete an invalid endpoint",
			urlName: "invalidurl",
			devObj: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "testcontainer1",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       urlName,
											TargetPort: 8080,
											Secure:     util.GetBoolPtr(false),
										},
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
			wantComponents: []devfilev1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "quay.io/nodejs-12",
							},
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Secure:     util.GetBoolPtr(false),
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := removeEndpointInDevfile(tt.devObj, tt.urlName)
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected err from removeEndpointInDevfile: %v", err)
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			}
			components, err := tt.devObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Errorf("Unexpected err from removeEndpointInDevfile: %v", err)
			}
			if !reflect.DeepEqual(components, tt.wantComponents) {
				t.Errorf("Expected: %v, got %v", tt.wantComponents, components)
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

func mockEnvSpecificInfo(configDir string, fs filesystem.Filesystem) (*EnvSpecificInfo, error) {

	esi := &EnvSpecificInfo{
		Filename: filepath.Join(configDir, ".odo", "env", "env.yaml"),
		fs:       fs,
	}
	err := fs.MkdirAll(filepath.Join(configDir, ".odo", "env"), os.ModePerm)
	if err != nil {
		return nil, err
	}

	return esi, nil

}
