package genericclioptions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		isOffline   bool

		// working dir
		workingDir         string
		populateWorkingDir func(fs filesystem.Filesystem)

		// current namespace
		currentNamespace string

		// flags
		projectFlag   string
		appFlag       string
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
		expected    Context
		expectedErr string
	}{
		{
			name: "flags set",
			input: input{
				needDevfile:   false,
				isOffline:     true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				projectFlag:   "myproject",
				appFlag:       "myapp",
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
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
			},
			expectedErr: "",
			expected: Context{
				internalCxt: internalCxt{
					project:     "myproject",
					application: "myapp",
					component:   "mycomponent",
					// empty when no devfile
					componentContext: "",
					outputFlag:       "",
					devfilePath:      "devfile.yaml",
				},
			},
		},
		{
			name: "flags not set",
			input: input{
				needDevfile: false,
				isOffline:   true,
				workingDir:  filepath.Join(prefixDir, "myapp"),
				outputFlag:  "",
				allFlagSet:  false,
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
			},
			expectedErr: "",
			expected: Context{
				internalCxt: internalCxt{
					project:     "a-project",
					application: "an-app-name",
					component:   "a-name",
					// empty when no devfile
					componentContext: "",
					outputFlag:       "",
					devfilePath:      "devfile.yaml",
				},
			},
		},
		{
			name: "missing project for url create",
			input: input{
				needDevfile:       false,
				isOffline:         true,
				workingDir:        filepath.Join(prefixDir, "myapp"),
				outputFlag:        "",
				allFlagSet:        false,
				parentCommandName: "url",
				commandName:       "create",
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
					env, err := envinfo.NewEnvSpecificInfo(filepath.Join(prefixDir, "myapp"))
					if err != nil {
						return
					}
					_ = env.SetComponentSettings(envinfo.ComponentSettings{
						Name:    "a-name",
						AppName: "an-app-name",
					})
				},
			},
			expectedErr: "",
			expected: Context{
				internalCxt: internalCxt{
					project:          "",
					application:      "an-app-name",
					component:        "a-name",
					componentContext: "",
					outputFlag:       "",
					devfilePath:      "devfile.yaml",
				},
			},
		},
		{
			name: "flags set, needDevfile",
			input: input{
				needDevfile:   true,
				isOffline:     true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				projectFlag:   "myproject",
				appFlag:       "myapp",
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".odo", "env", "env.yaml"), []byte{}, 0644)
				},
			},
			expectedErr: "no devfile found",
			expected: Context{
				internalCxt: internalCxt{
					project:          "myproject",
					application:      "myapp",
					component:        "mycomponent",
					componentContext: filepath.Join(prefixDir, "myapp"),
					outputFlag:       "",
					devfilePath:      filepath.Join(prefixDir, "myapp", "devfile.yaml"),
				},
			},
		},
		{
			name: "flags set, needDevfile, .devfile.yaml is present",
			input: input{
				needDevfile:   true,
				isOffline:     true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				projectFlag:   "myproject",
				appFlag:       "myapp",
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".odo", "env", "env.yaml"), []byte{}, 0644)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".devfile.yaml"), []byte(devfileYAML), 0644)
				},
			},
			expectedErr: "",
			expected: Context{
				internalCxt: internalCxt{
					project:          "myproject",
					application:      "myapp",
					component:        "mycomponent",
					componentContext: filepath.Join(prefixDir, "myapp"),
					outputFlag:       "",
					devfilePath:      filepath.Join(prefixDir, "myapp", ".devfile.yaml"),
				},
			},
		},
		{
			name: "flags set, needDevfile, devfile.yaml is present",
			input: input{
				needDevfile:   true,
				isOffline:     true,
				workingDir:    filepath.Join(prefixDir, "myapp"),
				projectFlag:   "myproject",
				appFlag:       "myapp",
				componentFlag: "mycomponent",
				outputFlag:    "",
				allFlagSet:    false,
				populateWorkingDir: func(fs filesystem.Filesystem) {
					_ = fs.MkdirAll(filepath.Join(prefixDir, "myapp", ".odo", "env"), 0755)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", ".odo", "env", "env.yaml"), []byte{}, 0644)
					_ = fs.WriteFile(filepath.Join(prefixDir, "myapp", "devfile.yaml"), []byte(devfileYAML), 0644)
				},
			},
			expectedErr: "",
			expected: Context{
				internalCxt: internalCxt{
					project:          "myproject",
					application:      "myapp",
					component:        "mycomponent",
					componentContext: filepath.Join(prefixDir, "myapp"),
					outputFlag:       "",
					devfilePath:      filepath.Join(prefixDir, "myapp", "devfile.yaml"),
				},
			},
		},
		{
			name: "no env file",
			input: input{
				needDevfile: false,
				isOffline:   true,
				workingDir:  filepath.Join(prefixDir, "myapp"),
				populateWorkingDir: func(fs filesystem.Filesystem) {
				},
			},
			expectedErr: "The current directory does not represent an odo component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// the first one is to cleanup the directory before execution (is ncase therea re remaining files from a previous execution)
			os.RemoveAll(prefixDir)
			// the second one to cleanup after execution
			defer os.RemoveAll(prefixDir)
			os.Setenv("KUBECONFIG", filepath.Join(prefixDir, ".kube", "config"))
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Fake Cobra
			cmdline := cmdline.NewMockCmdline(ctrl)
			cmdline.EXPECT().GetWorkingDirectory().Return(tt.input.workingDir, nil).AnyTimes()
			cmdline.EXPECT().CheckIfConfigurationNeeded().Return(false, nil).AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("project").Return(tt.input.projectFlag).AnyTimes()
			cmdline.EXPECT().FlagValueIfSet("app").Return(tt.input.appFlag).AnyTimes()
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

			kclient.EXPECT().SetNamespace(tt.expected.project).AnyTimes()

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.input.projectFlag,
				},
			}
			kclient.EXPECT().GetNamespaceNormal(tt.expected.project).Return(ns, nil).AnyTimes()

			depName := tt.expected.component + "-" + tt.expected.application
			dep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: depName,
				},
			}
			kclient.EXPECT().GetDeploymentByName(depName).Return(dep, nil).AnyTimes()
			kclient.EXPECT().GetCurrentNamespace().Return(tt.input.currentNamespace).AnyTimes()
			cmdline.EXPECT().GetKubeClient().Return(kclient, nil).AnyTimes()

			// Call the tested function
			params := NewCreateParameters(cmdline)
			if tt.input.needDevfile {
				params = params.NeedDevfile(tt.input.workingDir)
			}
			if tt.input.isOffline {
				params = params.IsOffline()
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
			if err != nil {
				return
			}

			if result.project != tt.expected.project {
				t.Errorf("Expected project %s, got %s", tt.expected.project, result.project)
			}
			if result.application != tt.expected.application {
				t.Errorf("Expected application %s, got %s", tt.expected.application, result.application)
			}
			if result.component != tt.expected.component {
				t.Errorf("Expected component %s, got %s", tt.expected.component, result.component)
			}
			if result.componentContext != tt.expected.componentContext {
				t.Errorf("Expected component context %s, got %s", tt.expected.componentContext, result.componentContext)
			}
			if result.outputFlag != tt.expected.outputFlag {
				t.Errorf("Expected output flag %s, got %s", tt.expected.outputFlag, result.outputFlag)
			}
			if result.devfilePath != tt.expected.devfilePath {
				t.Errorf("Expected devfilePath %s, got %s", tt.expected.devfilePath, result.devfilePath)
			}
		})
	}
}
