package genericclioptions

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

const devfileYAML = `
commands:
- exec:
    commandLine: npm install
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: /project
  id: install
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: /project
  id: run
- exec:
    commandLine: npm run debug
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: /project
  id: debug
- exec:
    commandLine: npm test
    component: runtime
    group:
      isDefault: true
      kind: test
    workingDir: /project
  id: test
components:
- container:
    endpoints:
    - name: http-3000
      targetPort: 3000
    image: registry.access.redhat.com/ubi8/nodejs-14:latest
    memoryLimit: 1024Mi
    mountSources: true
    sourceMapping: /project
  name: runtime
metadata:
  description: Stack with Node.js 14
  displayName: Node.js Runtime
  icon: https://nodejs.org/static/images/logos/nodejs-new-pantone-black.svg
  language: javascript
  name: nodejs-prj1-api-abhz
  projectType: nodejs
  tags:
  - NodeJS
  - Express
  - ubi8
  version: 1.0.1
schemaVersion: 2.0.0
starterProjects:
- git:
    remotes:
      origin: https://github.com/odo-devfiles/nodejs-ex.git
  name: nodejs-starter

`

func TestNew(t *testing.T) {

	prefixDir, err := os.MkdirTemp(os.TempDir(), "unittests-")
	if err != nil {
		t.Errorf("Error creating tem directory for tests")
		return
	}
	type input struct {
		// New params
		needDevfile bool

		// working dir
		workingDir         string
		populateWorkingDir func(fs filesystem.Filesystem)

		// flags
		componentFlag string
		outputFlag    string
		allFlagSet    bool

		// command
		parentCommandName string
		commandName       string
	}

	tests := []struct {
		name        string
		input       input
		expected    func() *Context
		expectedErr string
	}{
		{
			name: "flags set",
			input: input{
				needDevfile:   false,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName: getTestBaseName(),
						// empty when no devfile
						componentContext: "",
						devfilePath:      "",
					},
				}
			},
		},
		{
			name: "flags not set",
			input: input{
				needDevfile: false,
				workingDir:  filepath.Join(prefixDir, "myapp"),
				outputFlag:  "",
				allFlagSet:  false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName: getTestBaseName(),
						// empty when no devfile
						componentContext: "",
						devfilePath:      "",
					},
				}
			},
		},
		{
			name: "missing project for url create",
			input: input{
				needDevfile:       false,
				workingDir:        filepath.Join(prefixDir, "myapp"),
				outputFlag:        "",
				allFlagSet:        false,
				parentCommandName: "url",
				commandName:       "create",
				populateWorkingDir: func(fs filesystem.Filesystem) {
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName:    getTestBaseName(),
						componentContext: "",
						devfilePath:      "",
					},
				}
			},
		},
		{
			name: "flags set, needDevfile, devfile not found",
			input: input{
				needDevfile:   true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
				},
			},
			expectedErr: "The current directory does not represent an odo component",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName:    "",
						componentContext: filepath.Join(prefixDir, "myapp"),
						devfilePath:      "",
					},
				}
			},
		},
		{
			name: "flags set, needDevfile, .devfile.yaml is present",
			input: input{
				needDevfile:   true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".devfile.yaml"), []byte(devfileYAML), 0644)
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName:    "nodejs-prj1-api-abhz",
						componentContext: filepath.Join(prefixDir, "myapp"),
						devfilePath:      filepath.Join(prefixDir, "myapp", ".devfile.yaml"),
					},
				}
			},
		},
		{
			name: "flags set, needDevfile, devfile.yaml is present",
			input: input{
				needDevfile:   true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", "devfile.yaml"), []byte(devfileYAML), 0644)
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName:    "nodejs-prj1-api-abhz",
						componentContext: filepath.Join(prefixDir, "myapp"),
						devfilePath:      filepath.Join(prefixDir, "myapp", "devfile.yaml"),
					},
				}
			},
		},
		{
			name: "component flag not set, needDevfile, .devfile.yaml is present",
			input: input{
				needDevfile: true,
				workingDir:  filepath.Join(prefixDir, "myapp"),
				outputFlag:  "",
				allFlagSet:  false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".odo", "env", "env.yaml"), []byte{}, 0644)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".devfile.yaml"), []byte(devfileYAML), 0644)
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName:    "nodejs-prj1-api-abhz",
						componentContext: filepath.Join(prefixDir, "myapp"),
						devfilePath:      filepath.Join(prefixDir, "myapp", ".devfile.yaml"),
					},
				}
			},
		},
		{
			name: "component flag not set, needDevfile, devfile.yaml is present",
			input: input{
				needDevfile: true,
				workingDir:  filepath.Join(prefixDir, "myapp"),
				outputFlag:  "",
				allFlagSet:  false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".odo", "env", "env.yaml"), []byte{}, 0644)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", "devfile.yaml"), []byte(devfileYAML), 0644)
				},
			},
			expectedErr: "",
			expected: func() *Context {
				return &Context{
					internalCxt: internalCxt{
						componentName:    "nodejs-prj1-api-abhz",
						componentContext: filepath.Join(prefixDir, "myapp"),
						devfilePath:      filepath.Join(prefixDir, "myapp", "devfile.yaml"),
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// the first one is to cleanup the directory before execution (is ncase therea re remaining files from a previous execution)
			os.RemoveAll(prefixDir)
			// the second one to cleanup after execution
			defer os.RemoveAll(prefixDir)
			t.Setenv("KUBECONFIG", filepath.Join(prefixDir, ".kube", "config"))
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Fake Cobra
			cmdline := cmdline.NewMockCmdline(ctrl)
			cmdline.EXPECT().GetWorkingDirectory().Return(tt.input.workingDir, nil).AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("component").Return(tt.input.componentFlag).AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("o").Return(tt.input.outputFlag).AnyTimes()
			cmdline.EXPECT().IsFlagSet("all").Return(tt.input.allFlagSet).AnyTimes()
			cmdline.EXPECT().GetParentName().Return(tt.input.parentCommandName).AnyTimes()
			cmdline.EXPECT().GetName().Return(tt.input.commandName).AnyTimes()
			cmdline.EXPECT().GetRootName().Return(tt.input.parentCommandName).AnyTimes()

			// Fake fs
			// TODO(feloy) Unable to use memory FS because of devfile.ParseDevfileAndValidate not accepting FS parameter
			// mockFs := filesystem.NewFakeFs()
			// filesystem.Set(mockFs)
			tt.input.populateWorkingDir(filesystem.DefaultFs{})

			// Fake odo Kube client
			kclient := kclient.NewMockClientInterface(ctrl)

			kclient.EXPECT().SetNamespace(gomock.Any()).AnyTimes()

			// Call the tested function
			params := NewCreateParameters(cmdline)
			if tt.input.needDevfile {
				params = params.NeedDevfile(tt.input.workingDir)
			}
			result, err := New(params)

			// Checks
			if err == nil && tt.expectedErr != "" {
				t.Errorf("Expected %v, got no error", tt.expectedErr)
				return
			}
			if err != nil && tt.expectedErr == "" {
				t.Errorf("Expected no error, got %v", err.Error())
				return
			}
			if err != nil && tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err.Error())
				return
			}

			expected := tt.expected()
			if expected != nil && result == nil {
				t.Errorf("Expected non nil value, got nil result")
			}

			if expected == nil && result != nil {
				t.Errorf("Expected nil value, got non nil result")
			}

			if expected != nil && result != nil {
				if result.componentName != expected.componentName {
					t.Errorf("Expected componentName %s, got %s", expected.componentName, result.componentName)
				}
				if result.componentContext != expected.componentContext {
					t.Errorf("Expected component context %s, got %s", expected.componentContext, result.componentContext)
				}
				if result.devfilePath != expected.devfilePath {
					t.Errorf("Expected devfilePath %s, got %s", expected.devfilePath, result.devfilePath)
				}
			}
		})
	}
}

func getTestBaseName() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Base(filepath.Dir(b))
}
