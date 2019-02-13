package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name    string
		output  *GlobalConfigInfo
		success bool
	}{
		{
			name: "Test filename is being set",
			output: &GlobalConfigInfo{
				Filename:     tempConfigFile.Name(),
				GlobalConfig: GlobalConfig{},
			},
			success: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfi, err := NewGlobalConfig()
			switch test.success {
			case true:
				if err != nil {
					t.Errorf("expected test to pass, but it failed with error: %v", err)
				}
			case false:
				if err == nil {
					t.Errorf("expected test to fail, but it passed!")
				}
			}
			if !reflect.DeepEqual(test.output, cfi) {
				t.Errorf("expected output: %#v", test.output)
				t.Errorf("actual output: %#v", cfi)
			}
		})
	}
}

func TestSetActiveComponent(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		existingConfig GlobalConfig
		component      string
		project        string
		application    string
		wantErr        bool
		result         []ApplicationInfo
	}{
		{
			name:           "activeComponents nil",
			existingConfig: GlobalConfig{},
			component:      "foo",
			project:        "bar",
			application:    "app",
			wantErr:        true,
			result:         nil,
		},
		{
			name: "activeComponents empty",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "a",
						Active:  true,
						Project: "test",
					},
				},
			},
			component:   "foo",
			project:     "test",
			application: "a",
			wantErr:     false,
			result: []ApplicationInfo{
				{
					Name:            "a",
					Active:          true,
					Project:         "test",
					ActiveComponent: "foo",
				},
			},
		},
		{
			name: "project doesn't exists",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          false,
						Project:         "test",
						ActiveComponent: "b",
					},
				},
			},
			component:   "foo",
			project:     "nonexisting",
			application: "a",
			wantErr:     true,
			result:      nil,
		},
		{
			name: "application doesn't exists",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          false,
						Project:         "test",
						ActiveComponent: "b",
					},
				},
			},
			component:   "foo",
			project:     "test",
			application: "nonexisting",
			wantErr:     true,
			result:      nil,
		},
		{
			name: "overwrite existing active component (apps with same name in different projects)",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          true,
						Project:         "test",
						ActiveComponent: "old",
					},
					{
						Name:            "a",
						Active:          false,
						Project:         "test2",
						ActiveComponent: "old2",
					},
				},
			},
			component:   "new",
			project:     "test",
			application: "a",
			wantErr:     false,
			result: []ApplicationInfo{
				{
					Name:            "a",
					Active:          true,
					Project:         "test",
					ActiveComponent: "new",
				},
				{
					Name:            "a",
					Active:          false,
					Project:         "test2",
					ActiveComponent: "old2",
				},
			},
		},
		{
			name: "overwrite existing active component (different apps in the same project)",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          true,
						Project:         "test",
						ActiveComponent: "old",
					},
					{
						Name:            "b",
						Active:          false,
						Project:         "test",
						ActiveComponent: "old2",
					},
				},
			},
			component:   "new",
			project:     "test",
			application: "a",
			wantErr:     false,
			result: []ApplicationInfo{
				{
					Name:            "a",
					Active:          true,
					Project:         "test",
					ActiveComponent: "new",
				},
				{
					Name:            "b",
					Active:          false,
					Project:         "test",
					ActiveComponent: "old2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig
			err = cfg.SetActiveComponent(tt.component, tt.application, tt.project)
			if tt.wantErr {
				if (err != nil) != tt.wantErr {
					t.Errorf("SetActiveComponent() unexpected error %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err == nil {
				if !reflect.DeepEqual(cfg.ActiveApplications, tt.result) {
					t.Errorf("expected output doesn't match what was returned: \n expected:\n%#v\n, returned:\n%#v\n", tt.result, cfg.ActiveApplications)
				}
			}

		})
	}
}

func TestGetActiveComponent(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name              string
		existingConfig    GlobalConfig
		activeApplication string
		activeProject     string
		activeComponent   string
	}{
		{
			name:              "empty config",
			existingConfig:    GlobalConfig{},
			activeApplication: "test",
			activeProject:     "test",
			activeComponent:   "",
		},
		{
			name: "ActiveApplications empty",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
			},
			activeApplication: "test",
			activeProject:     "test",
			activeComponent:   "",
		},
		{
			name: "no active component record for given application",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "a",
						Active:  false,
						Project: "test",
					},
				},
			},
			activeApplication: "test",
			activeProject:     "test",
			activeComponent:   "",
		},
		{
			name: "activeComponents for one project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          true,
						Project:         "test",
						ActiveComponent: "b",
					},
				},
			},
			activeApplication: "a",
			activeProject:     "test",
			activeComponent:   "b",
		},
		{
			name: "inactive project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          false,
						Project:         "test",
						ActiveComponent: "b",
					},
				},
			},
			activeApplication: "a",
			activeProject:     "test",
			activeComponent:   "",
		},
		{
			name: "multiple projects",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          false,
						Project:         "test",
						ActiveComponent: "b",
					},
					{
						Name:            "a",
						Active:          true,
						Project:         "test2",
						ActiveComponent: "b2",
					},
				},
			},
			activeApplication: "a",
			activeProject:     "test2",
			activeComponent:   "b2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GlobalConfigInfo{
				GlobalConfig: tt.existingConfig,
			}
			output := cfg.GetActiveComponent(tt.activeApplication, tt.activeProject)

			if output != tt.activeComponent {
				t.Errorf("active component doesn't match expected \ngot: %s \nexpected: %s\n", output, tt.activeComponent)
			}

		})
	}
}

func TestSetActiveApplication(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig GlobalConfig
		setApplication string
		project        string
		wantErr        bool
	}{
		{
			name:           "activeApplication nil",
			existingConfig: GlobalConfig{},
			setApplication: "app",
			project:        "proj",
			wantErr:        true,
		},
		{
			name: "activeApplication empty",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
			},
			setApplication: "app",
			project:        "proj",
			wantErr:        true,
		},
		{
			name: "no Active value",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Project:         "proj",
						ActiveComponent: "b",
					},
				},
			},
			setApplication: "app",
			project:        "proj",
		},
		{
			name: "multiple apps in the same project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app2",
						Active:          false,
						Project:         "proj",
						ActiveComponent: "b2",
					},
				},
			},
			setApplication: "app2",
			project:        "proj",
		},
		{
			name: "same app name in different projects",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app",
						Active:          false,
						Project:         "proj2",
						ActiveComponent: "b2",
					},
				},
			},
			setApplication: "app",
			project:        "proj2",
		},
		{
			name: "nonexisting application",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
				},
			},
			setApplication: "app-non-existing",
			project:        "proj",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig

			err = cfg.SetActiveApplication(tt.setApplication, tt.project)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, but there was no error returned")
				} else {
					return
				}
			} else {
				if err != nil {
					t.Error(err)
				}
			}

			found := false
			for _, aa := range cfg.ActiveApplications {
				fmt.Printf("%#v\n", aa)
				if aa.Project == tt.project && aa.Name == tt.setApplication {
					found = true
				}
			}
			if !found {
				t.Errorf("application %s/%s was not set as current", tt.project, tt.setApplication)
			}

		})
	}
}

func TestAddApplication(t *testing.T) {
	tests := []struct {
		name           string
		existingConfig GlobalConfig
		resultConfig   GlobalConfig
		addApplication string
		project        string
		wantErr        bool
	}{
		{
			name:           "activeApplication nil",
			existingConfig: GlobalConfig{},
			resultConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "app",
						Project: "proj",
						Active:  false,
					},
				},
			},
			addApplication: "app",
			project:        "proj",
			wantErr:        false,
		},
		{
			name: "activeApplication empty",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
			},
			resultConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "app",
						Project: "proj",
						Active:  false,
					},
				},
			},
			addApplication: "app",
			project:        "proj",
			wantErr:        false,
		},
		{
			name: "multiple apps in the same project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app2",
						Active:          false,
						Project:         "proj",
						ActiveComponent: "b2",
					},
				},
			},
			resultConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app2",
						Active:          false,
						Project:         "proj",
						ActiveComponent: "b2",
					},
					{
						Name:    "app3",
						Project: "proj",
						Active:  false,
					},
				},
			},
			addApplication: "app3",
			project:        "proj",
		},
		{
			name: "same app name in different projects",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app",
						Active:          false,
						Project:         "proj2",
						ActiveComponent: "b2",
					},
				},
			},
			resultConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app",
						Active:          false,
						Project:         "proj2",
						ActiveComponent: "b2",
					},
					{
						Name:    "app2",
						Project: "proj2",
						Active:  false,
					},
				},
			},
			addApplication: "app2",
			project:        "proj2",
		},
		{
			name: "application already exist",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
				},
			},
			resultConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
				},
			},
			addApplication: "app",
			project:        "proj",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig

			err = cfg.AddApplication(tt.addApplication, tt.project)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, but there was no error returned")
				}
				// if there is an error and we expected it, check if existingConfig matched resultedConfig anyway
			} else {
				if err != nil {
					t.Error(err)
				}
			}

			if !reflect.DeepEqual(cfg.GlobalConfig, tt.resultConfig) {
				t.Errorf("expected output doesn't match what was returned: \n expected:\n%#v\n returned:\n%#v\n", tt.resultConfig, cfg.GlobalConfig)
			}

		})
	}
}

func TestGetActiveApplication(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name              string
		existingConfig    GlobalConfig
		activeProject     string
		activeApplication string
	}{
		{
			name:              "activeApplication nil",
			existingConfig:    GlobalConfig{},
			activeApplication: "",
			activeProject:     "proj",
		},
		{
			name: "activeApplication empty",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
			},
			activeApplication: "",
			activeProject:     "proj",
		},
		{
			name: "no Active value",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Project:         "proj",
						ActiveComponent: "b",
					},
				},
			},
			activeApplication: "",
			activeProject:     "proj",
		},
		{
			name: "multiple apps in the same project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app2",
						Active:          false,
						Project:         "proj",
						ActiveComponent: "b2",
					},
				},
			},
			activeApplication: "app",
			activeProject:     "proj",
		},
		{
			name: "same app name in different projects",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "app",
						Active:          true,
						Project:         "proj",
						ActiveComponent: "b",
					},
					{
						Name:            "app",
						Active:          false,
						Project:         "proj2",
						ActiveComponent: "b2",
					},
				},
			},
			activeApplication: "app",
			activeProject:     "proj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GlobalConfigInfo{
				GlobalConfig: tt.existingConfig,
			}
			output := cfg.GetActiveApplication(tt.activeProject)

			if output != tt.activeApplication {
				t.Errorf("active application doesn't match expected \ngot: %s \nexpected: %s\n", output, tt.activeApplication)
			}

		})
	}
}

func TestDeleteApplication(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		existingConfig GlobalConfig
		application    string
		project        string
		wantErr        bool
		result         []ApplicationInfo
	}{
		{
			name:           "empty config",
			existingConfig: GlobalConfig{},
			application:    "foo",
			project:        "bar",
			wantErr:        true,
			result:         nil,
		},
		{
			name: "delete not existing application",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "a",
						Active:  true,
						Project: "test",
					},
				},
			},
			application: "b",
			project:     "test",
			wantErr:     false,
			result: []ApplicationInfo{
				{},
			},
		},
		{
			name: "delete existing application",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          false,
						Project:         "test",
						ActiveComponent: "b",
					},
				},
			},
			application: "a",
			project:     "test",
			wantErr:     false,
			result:      []ApplicationInfo{},
		},
		{
			name: "delete application (apps with same name in different projects)",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          true,
						Project:         "test",
						ActiveComponent: "old",
					},
					{
						Name:            "a",
						Active:          false,
						Project:         "test2",
						ActiveComponent: "old2",
					},
				},
			},
			application: "a",
			project:     "test",
			wantErr:     false,
			result: []ApplicationInfo{
				{
					Name:            "a",
					Active:          false,
					Project:         "test2",
					ActiveComponent: "old2",
				},
			},
		},
		{
			name: "delete application (different apps in the same project)",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:            "a",
						Active:          true,
						Project:         "test",
						ActiveComponent: "comp",
					},
					{
						Name:            "b",
						Active:          false,
						Project:         "test",
						ActiveComponent: "comp2",
					},
				},
			},
			application: "b",
			project:     "test",
			wantErr:     false,
			result: []ApplicationInfo{
				{
					Name:            "a",
					Active:          true,
					Project:         "test",
					ActiveComponent: "comp",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig

			err = cfg.DeleteApplication(tt.application, tt.project)
			if tt.wantErr {
				if (err != nil) != tt.wantErr {
					t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err == nil {
				if !reflect.DeepEqual(cfg.ActiveApplications, tt.result) {
					t.Errorf("expected output doesn't match what was returned: \n expected:\n%#v\n returned:\n%#v\n", tt.result, cfg.ActiveApplications)
				}
			}

		})
	}
}

func TestGetTimeout(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	zeroValue := 0
	nonzeroValue := 5
	tests := []struct {
		name           string
		existingConfig GlobalConfig
		want           int
	}{
		{
			name:           "Case 1: validating value 1 from config in default case",
			existingConfig: GlobalConfig{},
			want:           1,
		},

		{
			name: "Case 2: validating value 0 from config",
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					Timeout: &zeroValue,
				},
			},
			want: 0,
		},

		{
			name: "Case 3: validating value 5 from config",
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					Timeout: &nonzeroValue,
				},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig

			output := cfg.GetTimeout()
			if output != tt.want {
				t.Errorf("GetTimeout returned unexpeced value expected \ngot: %d \nexpected: %d\n", output, tt.want)
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false
	fakePrefix := "name"

	tests := []struct {
		name           string
		existingConfig GlobalConfig
		project        string
		wantErr        bool
		result         GlobalConfig
	}{
		{
			name: "test case 1: no applications to the project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &trueValue,
				},
			},
			project: "project-1",
			wantErr: false,
			result: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &trueValue,
				},
			},
		},
		{
			name: "test case 2: one application to the project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "blah",
						Project: "project-1",
					},
				},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &trueValue,
				},
			},
			project: "project-1",
			wantErr: false,
			result: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &trueValue,
				},
			},
		},
		{
			name: "test case 3: two applications to the project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "blah",
						Project: "project-1",
					},
					{
						Name:    "blah-1",
						Project: "project-1",
					},
				},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &trueValue,
				},
			},
			project: "project-1",
			wantErr: false,
			result: GlobalConfig{
				ActiveApplications: []ApplicationInfo{},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &trueValue,
				},
			},
		},
		{
			name: "test case 4: two applications to the project and one in another project",
			existingConfig: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "blah",
						Project: "project-1",
					},
					{
						Name:    "blah-1",
						Project: "project-1",
					},
					{
						Name:    "blah",
						Project: "project-3",
					},
				},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &falseValue,
				},
			},
			project: "project-1",
			wantErr: false,
			result: GlobalConfig{
				ActiveApplications: []ApplicationInfo{
					{
						Name:    "blah",
						Project: "project-3",
					},
				},
				OdoSettings: OdoSettings{
					NamePrefix:         &fakePrefix,
					UpdateNotification: &falseValue,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig
			err = cfg.DeleteProject(tt.project)

			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(cfg.ActiveApplications, tt.result.ActiveApplications) {
					t.Errorf("test failed, config file active applications values are different, wanted: %v, got: %v", tt.result.ActiveApplications, cfg.ActiveApplications)
				}
				if !reflect.DeepEqual(*cfg.OdoSettings.UpdateNotification, *tt.result.OdoSettings.UpdateNotification) {
					t.Errorf("test failed, config file updateNotification values are different, wanted: %v, got: %v", *tt.result.OdoSettings.UpdateNotification, *cfg.OdoSettings.UpdateNotification)
				}
				if !reflect.DeepEqual(*cfg.OdoSettings.NamePrefix, *tt.result.OdoSettings.NamePrefix) {
					t.Errorf("test failed, config file NamePrefix values are different, wanted: %v, got: %v", *tt.result.OdoSettings.NamePrefix, *cfg.OdoSettings.NamePrefix)
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}

func TestSetConfiguration(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false
	zeroValue := 0

	tests := []struct {
		name           string
		parameter      string
		value          string
		existingConfig GlobalConfig
		wantErr        bool
		want           interface{}
	}{
		// update notification
		{
			name:           fmt.Sprintf("Case 1: %s set nil to true", UpdateNotificationSetting),
			parameter:      UpdateNotificationSetting,
			value:          "true",
			existingConfig: GlobalConfig{},
			want:           true,
			wantErr:        false,
		},
		{
			name:      fmt.Sprintf("Case 2: %s set true to false", UpdateNotificationSetting),
			parameter: UpdateNotificationSetting,
			value:     "false",
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					UpdateNotification: &trueValue,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:      fmt.Sprintf("Case 3: %s set false to true", UpdateNotificationSetting),
			parameter: UpdateNotificationSetting,
			value:     "true",
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					UpdateNotification: &falseValue,
				},
			},
			want:    true,
			wantErr: false,
		},

		{
			name:           fmt.Sprintf("Case 4: %s invalid value", UpdateNotificationSetting),
			parameter:      UpdateNotificationSetting,
			value:          "invalid_value",
			existingConfig: GlobalConfig{},
			wantErr:        true,
		},
		// time out
		{
			name:      fmt.Sprintf("Case 5: %s set to 5 from 0", TimeoutSetting),
			parameter: TimeoutSetting,
			value:     "5",
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					Timeout: &zeroValue,
				},
			},
			want:    5,
			wantErr: false,
		},
		{
			name:           fmt.Sprintf("Case 6: %s set to 300", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "300",
			existingConfig: GlobalConfig{},
			want:           300,
			wantErr:        false,
		},
		{
			name:           fmt.Sprintf("Case 7: %s set to 0", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "0",
			existingConfig: GlobalConfig{},
			want:           0,
			wantErr:        false,
		},
		{
			name:           fmt.Sprintf("Case 8: %s set to -1", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "-1",
			existingConfig: GlobalConfig{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 9: %s invalid value", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "this",
			existingConfig: GlobalConfig{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 10: %s set to 300 with mixed case in parameter name", TimeoutSetting),
			parameter:      "TimeOut",
			value:          "300",
			existingConfig: GlobalConfig{},
			want:           300,
			wantErr:        false,
		},
		// invalid parameter
		{
			name:           "Case 11: invalid parameter",
			parameter:      "invalid_parameter",
			existingConfig: GlobalConfig{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewGlobalConfig()
			if err != nil {
				t.Error(err)
			}
			cfg.GlobalConfig = tt.existingConfig

			err = cfg.SetConfiguration(tt.parameter, tt.value)

			if !tt.wantErr && err == nil {
				// validating the value after executing Serconfiguration
				// according to component in positive cases
				switch tt.parameter {
				case "updatenotification":
					if *cfg.OdoSettings.UpdateNotification != tt.want {
						t.Errorf("unexpeced value after execution of SetConfiguration \ngot: %t \nexpected: %t\n", *cfg.OdoSettings.UpdateNotification, tt.want)
					}
				case "timeout":
					if *cfg.OdoSettings.Timeout != tt.want {
						t.Errorf("unexpeced value after execution of SetConfiguration \ngot: %v \nexpected: %d\n", cfg.OdoSettings.Timeout, tt.want)
					}
				}
			} else if tt.wantErr && err != nil {
				// negative cases
				switch tt.parameter {
				case "updatenotification":
				case "timeout":
					typedval, err := strconv.Atoi(tt.value)
					// if err is found in cases other than value <0 or !ok
					if !(typedval < 0 || err != nil) {
						t.Error(err)
					}
				}
			} else {
				t.Error(err)
			}

		})
	}
}

func TestGetupdateNotification(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false

	tests := []struct {
		name           string
		existingConfig GlobalConfig
		want           bool
	}{
		{
			name:           fmt.Sprintf("Case 1: %s nil", UpdateNotificationSetting),
			existingConfig: GlobalConfig{},
			want:           true,
		},
		{
			name: fmt.Sprintf("Case 2: %s true", UpdateNotificationSetting),
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					UpdateNotification: &trueValue,
				},
			},
			want: true,
		},
		{
			name: fmt.Sprintf("Case 3: %s false", UpdateNotificationSetting),
			existingConfig: GlobalConfig{
				OdoSettings: OdoSettings{
					UpdateNotification: &falseValue,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GlobalConfigInfo{
				GlobalConfig: tt.existingConfig,
			}
			output := cfg.GetUpdateNotification()

			if output != tt.want {
				t.Errorf("GetUpdateNotification returned unexpeced value expected \ngot: %t \nexpected: %t\n", output, tt.want)
			}

		})
	}
}

func TestFormatSupportedParameters(t *testing.T) {
	expected := `
Available Parameters:
%s - %s
%s - %s
%s - %s
`
	expected = fmt.Sprintf(expected,
		NamePrefixSetting, NamePrefixSettingDescription,
		TimeoutSetting, TimeoutSettingDescription,
		UpdateNotificationSetting, UpdateNotificationSettingDescription)
	actual := FormatSupportedParameters()
	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func TestLowerCaseParameters(t *testing.T) {
	expected := map[string]bool{"nameprefix": true, "timeout": true, "updatenotification": true}
	actual := getLowerCaseParameters(GetSupportedParameters())
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected '%v', got '%v'", expected, actual)
	}
}

func TestLowerCaseParameterForLocalParameters(t *testing.T) {
	expected := map[string]bool{"componenttype": true, "componentname": true, "minmemory": true, "maxmemory": true,
		"ignore": true, "mincpu": true, "maxcpu": true, "cpu": true}
	actual := getLowerCaseParameters(GetLocallySupportedParameters())
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected '%v', got '%v'", expected, actual)
	}
}

func TestLocalConfigInitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	os.RemoveAll(path.Join(".odo", configFileName))

	conf, err := NewLocalConfig()
	if err != nil {
		t.Errorf("error while creating local config %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("local odo-config.yaml shouldn't exist yet")
	}
}

func TestIsSupportedParameter(t *testing.T) {
	tests := []struct {
		testName      string
		param         string
		expectedLower string
		expected      bool
	}{
		{
			testName:      "existing, lower case",
			param:         "timeout",
			expectedLower: "timeout",
			expected:      true,
		},
		{
			testName:      "existing, from description",
			param:         "Timeout",
			expectedLower: "timeout",
			expected:      true,
		},
		{
			testName:      "existing, mixed case",
			param:         "TimeOut",
			expectedLower: "timeout",
			expected:      true,
		},
		{
			testName: "empty",
			param:    "",
			expected: false,
		},
		{
			testName: "unexisting",
			param:    "foo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual, ok := asSupportedParameter(tt.param)
			if tt.expected != ok && tt.expectedLower != actual {
				t.Fail()
			}
		})
	}
}
