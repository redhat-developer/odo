package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/tests/helper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Context

type testContext struct {
	platform        string
	fscontent       string
	runningInOption string
	filesOption     string
	nameOption      string
}

// Platform

type platformFunc func(t *testing.T, env map[string]string, config map[string]string, clientset *clientset.Clientset)

var noPlatformPlatform platformFunc = func(t *testing.T, env map[string]string, config map[string]string, clientset *clientset.Clientset) {
	env["KUBECONFIG"] = "/dev/null"
	config["PODMAN_CMD"] = "not-found"
}

var podmanOnlyPlatform = func() platformFunc {
	return func(t *testing.T, env map[string]string, config map[string]string, clientset *clientset.Clientset) {
		env["KUBECONFIG"] = "/dev/null"
		ctrl := gomock.NewController(t)
		// Podman is accessible
		podmanClient := podman.NewMockClient(ctrl)
		clientset.PodmanClient = podmanClient
	}
}

var kubernetesOnlyPlatform = func() platformFunc {
	return func(t *testing.T, env map[string]string, config map[string]string, clientset *clientset.Clientset) {
		config["PODMAN_CMD"] = "not-found"
		ctrl := gomock.NewController(t)
		// kubernetes is accessible
		kubeClient := kclient.NewMockClientInterface(ctrl)
		kubeClient.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
		clientset.KubernetesClient = kubeClient
	}
}

var kubernetesAndPodmanPlatform = func() platformFunc {
	return func(t *testing.T, env map[string]string, config map[string]string, clientset *clientset.Clientset) {
		ctrl := gomock.NewController(t)
		// kubernetes is accessible
		kubeClient := kclient.NewMockClientInterface(ctrl)
		clientset.KubernetesClient = kubeClient
		kubeClient.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
		// Podman is accessible
		podmanClient := podman.NewMockClient(ctrl)
		clientset.PodmanClient = podmanClient
	}
}

var allPlatforms = map[string]platformFunc{
	"no platform":           noPlatformPlatform,
	"podman only":           podmanOnlyPlatform(),
	"kubernetes only":       kubernetesOnlyPlatform(),
	"kubernetes and podman": kubernetesAndPodmanPlatform(),
}

// FS content

type fscontentFunc func(fs filesystem.Filesystem)

var noContentFscontent fscontentFunc = func(fs filesystem.Filesystem) {}

var nodeJsSourcesFsContent fscontentFunc = func(fs filesystem.Filesystem) {
	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), ".")
}

type fsOptions struct {
	dotOdoExists bool
	generated    []string
}

var nodeJsSourcesAndDevfileFsContent = func(devfilePath string, options fsOptions) fscontentFunc {
	return func(fs filesystem.Filesystem) {
		helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), ".")
		helper.CopyExampleDevFile(
			devfilePath,
			"devfile.yaml",
			"my-component")
		if options.dotOdoExists || options.generated != nil {
			helper.MakeDir(util.DotOdoDirectory)
		}
		if options.generated != nil {
			err := helper.CreateFileWithContent(filepath.Join(util.DotOdoDirectory, "generated"), strings.Join(options.generated, "\n"))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

		}
	}
}

var allFscontents = map[string]fscontentFunc{
	"no content":                 noContentFscontent,
	"nodeJS sources":             nodeJsSourcesFsContent,
	"nodeJS sources and Devfile": nodeJsSourcesAndDevfileFsContent(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), fsOptions{}),
	"nodeJS sources, Devfile and .odo": nodeJsSourcesAndDevfileFsContent(
		filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
		fsOptions{
			dotOdoExists: true,
		}),
	"nodeJS sources and generated Devfile": nodeJsSourcesAndDevfileFsContent(
		filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
		fsOptions{
			generated: []string{"devfile.yaml"},
		}),
}

// runningIn option

type runningInOption []string

var noRunningInOption = []string{}
var devRunninInOption = []string{"--running-in", "dev"}
var deployRunninInOption = []string{"--running-in", "deploy"}

var allRunningInOptions = map[string]runningInOption{
	"no":     noRunningInOption,
	"dev":    devRunninInOption,
	"deploy": deployRunninInOption,
}

// files option

type filesOption []string

var noFilesOptions = []string{}
var yesFilesOptions = []string{"--files"}

var allFilesOptions = map[string]filesOption{
	"no":  noFilesOptions,
	"yes": yesFilesOptions,
}

// name option

type nameOption []string

var noNameOptions = []string{}
var yesNameOptions = []string{"--name", "my-component"}

var allNameOptions = map[string]nameOption{
	"no":  noNameOptions,
	"yes": yesNameOptions,
}

// calls checks

var checkCallsNonDeployedComponent = func(t *testing.T, clientset clientset.Clientset, testContext testContext) {
	if strings.Contains(testContext.platform, "podman") &&
		testContext.runningInOption != "deploy" {
		podmanMock := clientset.PodmanClient.(*podman.MockClient)
		podmanMock.EXPECT().PodLs()
	}
	if strings.Contains(testContext.platform, "kubernetes") {
		kubeMock := clientset.KubernetesClient.(*kclient.MockClientInterface)
		dep := appsv1.Deployment{}
		if testContext.runningInOption != "deploy" {
			kubeMock.EXPECT().GetDeploymentByName("my-component-app").Return(&dep, nil)
		}
		selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/managed-by=odo,app.kubernetes.io/part-of=app"
		if testContext.runningInOption == "dev" {
			selector = selector + ",odo.dev/mode=Dev"
		} else if testContext.runningInOption == "deploy" {
			selector = selector + ",odo.dev/mode=Deploy"
		}
		kubeMock.EXPECT().GetAllResourcesFromSelector(selector, "a-namespace").Return(nil, nil).AnyTimes()
	}
}

var checkCallsDeployedComponent = func(t *testing.T, clientset clientset.Clientset, testContext testContext) {
	if strings.Contains(testContext.platform, "podman") &&
		testContext.runningInOption != "deploy" {
		podmanMock := clientset.PodmanClient.(*podman.MockClient)
		podmanMock.EXPECT().PodLs().Return(map[string]bool{"other-pod": true, "my-component-app": true}, nil)
		pod := corev1.Pod{}
		pod.SetName("my-component-app")
		podmanMock.EXPECT().KubeGenerate("my-component-app").Return(&pod, nil)
		// The pod and its volumes should be deleted
		podmanMock.EXPECT().CleanupPodResources(&pod, true)
	}
	if strings.Contains(testContext.platform, "kubernetes") {
		kubeMock := clientset.KubernetesClient.(*kclient.MockClientInterface)
		dep := appsv1.Deployment{}
		dep.Kind = "Deployment"
		dep.SetName("my-component-app")
		if testContext.runningInOption != "deploy" {
			kubeMock.EXPECT().GetDeploymentByName("my-component-app").Return(&dep, nil)
		}
		selector := "app.kubernetes.io/instance=my-component,app.kubernetes.io/managed-by=odo,app.kubernetes.io/part-of=app"
		if testContext.runningInOption == "dev" {
			selector = selector + ",odo.dev/mode=Dev"
		} else if testContext.runningInOption == "deploy" {
			selector = selector + ",odo.dev/mode=Deploy"
		}
		kubeMock.EXPECT().GetAllResourcesFromSelector(selector, "a-namespace").Return(nil, nil).AnyTimes()
		kubeMock.EXPECT().GetRestMappingFromUnstructured(gomock.Any()).Return(&meta.RESTMapping{
			Resource: schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
		}, nil)
		kubeMock.EXPECT().DeleteDynamicResource("my-component-app", schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}, false)

	}
}

func TestOdoDeleteMatrix(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string

		platforms        map[string]platformFunc
		fscontents       map[string]fscontentFunc
		runningInOptions map[string]runningInOption
		filesOptions     map[string]filesOption
		nameOptions      map[string]nameOption

		wantErr     string
		checkOutput func(t *testing.T, s string)
		checkFS     func(t *testing.T, fs filesystem.Filesystem)
		checkCalls  func(t *testing.T, clientset clientset.Clientset, tetsContext testContext)
	}{
		{
			name: "delete component when Devfile is not present in the directory",
			args: []string{"delete", "component", "-f"},

			platforms: allPlatforms,
			fscontents: map[string]fscontentFunc{
				"no content":     noContentFscontent,
				"nodeJS sources": nodeJsSourcesFsContent,
			},
			runningInOptions: allRunningInOptions,
			filesOptions:     allFilesOptions,
			nameOptions: map[string]nameOption{
				"no": noNameOptions,
			},

			wantErr: "The current directory does not represent an odo component",
		},
		{
			name: "delete component using both --files and --name",
			args: []string{"delete", "component", "-f"},

			platforms:        allPlatforms,
			fscontents:       allFscontents,
			runningInOptions: allRunningInOptions,
			filesOptions: map[string]filesOption{
				"yes": yesFilesOptions,
			},
			nameOptions: map[string]nameOption{
				"yes": yesNameOptions,
			},

			wantErr: "'--files' cannot be used with '--name'; '--files' must be used from a directory containing a Devfile",
		},
		{
			name: "delete component passing an invalid running-in",
			args: []string{"delete", "component", "-f", "--running-in", "invalid-value"},

			platforms:  allPlatforms,
			fscontents: allFscontents,
			runningInOptions: map[string]runningInOption{
				"no": noRunningInOption,
			},
			filesOptions: allFilesOptions,
			nameOptions:  allNameOptions,

			wantErr: "invalid value for --running-in: \"invalid-value\". Acceptable values are: dev, deploy",
		},
		{
			name: "using --files in a directory where Devfile was not generated by odo",
			args: []string{"delete", "component", "-f"},

			platforms: allPlatforms,
			fscontents: map[string]fscontentFunc{
				"nodeJS sources and Devfile": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{}),
				"nodeJS sources, Devfile and .odo": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						dotOdoExists: true,
					}),
			},
			runningInOptions: allRunningInOptions,
			filesOptions: map[string]filesOption{
				"yes": yesFilesOptions,
			},
			nameOptions: map[string]nameOption{
				"no": noNameOptions,
			},

			checkOutput: func(t *testing.T, s string) {
				gomega.Expect(s).ToNot(gomega.ContainSubstring("devfile.yaml"), "should not list the devfile.yaml")
			},
			checkFS: func(t *testing.T, fs filesystem.Filesystem) {
				fileList := helper.ListFilesInDir(".")
				gomega.Expect(fileList).Should(gomega.ContainElement("devfile.yaml"), "should not delete the devfile.yaml")
			},
			checkCalls: checkCallsNonDeployedComponent,
		},
		{
			name: "using --files in a directory where Devfile was generated by odo",
			args: []string{"delete", "component", "-f"},

			platforms: allPlatforms,
			fscontents: map[string]fscontentFunc{
				"nodeJS sources and generated Devfile": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						generated: []string{"devfile.yaml"},
					}),
			},
			runningInOptions: allRunningInOptions,
			filesOptions: map[string]filesOption{
				"yes": yesFilesOptions,
			},
			nameOptions: map[string]nameOption{
				"no": noNameOptions,
			},

			checkOutput: func(t *testing.T, s string) {
				gomega.Expect(s).To(gomega.ContainSubstring("devfile.yaml"), "should list the devfile.yaml")
			},
			checkFS: func(t *testing.T, fs filesystem.Filesystem) {
				fileList := helper.ListFilesInDir(".")
				gomega.Expect(fileList).ShouldNot(gomega.ContainElement("devfile.yaml"), "should delete the devfile.yaml")
			},
			checkCalls: checkCallsNonDeployedComponent,
		},
		{
			name: "delete a non deployed component",
			args: []string{"delete", "component", "-f"},

			platforms: allPlatforms,
			fscontents: map[string]fscontentFunc{
				"nodeJS sources and Devfile": nodeJsSourcesAndDevfileFsContent(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), fsOptions{}),
				"nodeJS sources, Devfile and .odo": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						dotOdoExists: true,
					}),
				"nodeJS sources and generated Devfile": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						generated: []string{"devfile.yaml"},
					}),
			},
			runningInOptions: allRunningInOptions,
			filesOptions:     allFilesOptions,
			nameOptions: map[string]nameOption{
				"no": noNameOptions,
			},

			checkOutput: func(t *testing.T, s string) {
				gomega.Expect(s).To(gomega.ContainSubstring("No resource found for component %q", "my-component"))
			},
			checkCalls: checkCallsNonDeployedComponent,
		},
		{
			name: "delete a component deployed on podman",
			args: []string{"delete", "component", "-f"},

			platforms: map[string]platformFunc{
				"podman only":           podmanOnlyPlatform(),
				"kubernetes and podman": kubernetesAndPodmanPlatform(),
			},
			fscontents: map[string]fscontentFunc{
				"nodeJS sources and Devfile": nodeJsSourcesAndDevfileFsContent(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), fsOptions{}),
				"nodeJS sources, Devfile and .odo": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						dotOdoExists: true,
					}),
				"nodeJS sources and generated Devfile": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						generated: []string{"devfile.yaml"},
					}),
			},
			runningInOptions: map[string]runningInOption{
				"no":  noRunningInOption,
				"dev": devRunninInOption,
			},
			filesOptions: allFilesOptions,
			nameOptions: map[string]nameOption{
				"no": noNameOptions,
			},

			checkOutput: func(t *testing.T, s string) {
				gomega.Expect(s).To(gomega.ContainSubstring("The following pods and associated volumes will get deleted from podman"))
				gomega.Expect(s).To(gomega.ContainSubstring("- my-component-app"))
			},
			checkCalls: checkCallsDeployedComponent,
		},
		{
			name: "delete a component deployed on kubernetes",
			args: []string{"delete", "component", "-f"},

			platforms: map[string]platformFunc{
				"kubernetes only":       kubernetesOnlyPlatform(),
				"kubernetes and podman": kubernetesAndPodmanPlatform(),
			},
			fscontents: map[string]fscontentFunc{
				"nodeJS sources and Devfile": nodeJsSourcesAndDevfileFsContent(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), fsOptions{}),
				"nodeJS sources, Devfile and .odo": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						dotOdoExists: true,
					}),
				"nodeJS sources and generated Devfile": nodeJsSourcesAndDevfileFsContent(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					fsOptions{
						generated: []string{"devfile.yaml"},
					}),
			},
			runningInOptions: map[string]runningInOption{
				"no":  noRunningInOption,
				"dev": devRunninInOption,
			},
			filesOptions: allFilesOptions,
			nameOptions: map[string]nameOption{
				"no": noNameOptions,
			},

			checkOutput: func(t *testing.T, s string) {
				gomega.Expect(s).To(gomega.ContainSubstring("The following resources will get deleted from cluster"))
				gomega.Expect(s).To(gomega.ContainSubstring("- Deployment: my-component-app"))
			},
			checkCalls: checkCallsDeployedComponent,
		},
	} {
		if tt.platforms == nil {
			t.Fatal("platforms cannot be nil")
		}
		for platform, platformFunc := range tt.platforms {
			platform := platform
			platformFunc := platformFunc
			if tt.fscontents == nil {
				t.Fatal("fscontents cannot be nil")
			}
			for fscontent, fscontentFunc := range tt.fscontents {
				fscontent := fscontent
				fscontentFunc := fscontentFunc
				if tt.runningInOptions == nil {
					t.Fatal("runningInOptions cannot be nil")
				}
				for runningInOption, runningInOptionValue := range tt.runningInOptions {
					runningInOption := runningInOption
					runningInOptionValue := runningInOptionValue
					if tt.filesOptions == nil {
						t.Fatal("filesOptions cannot be nil")
					}
					for filesOption, filesOptionValue := range tt.filesOptions {
						filesOption := filesOption
						filesOptionValue := filesOptionValue
						if tt.nameOptions == nil {
							t.Fatal("nameOptions cannot be nil")
						}
						for nameOption, nameOptionValue := range tt.nameOptions {
							nameOption := nameOption
							nameOptionValue := nameOptionValue

							testCtx := testContext{
								platform:        platform,
								fscontent:       fscontent,
								runningInOption: runningInOption,
								filesOption:     filesOption,
								nameOption:      nameOption,
							}
							t.Run(
								tt.name+
									fmt.Sprintf(" [platform=%s]", platform)+
									fmt.Sprintf(" [fscontent=%s]", fscontent)+
									fmt.Sprintf(" [runningInOptions=%s]", runningInOption)+
									fmt.Sprintf(" [filesOption=%s]", filesOption)+
									fmt.Sprintf(" [nameOption=%s]", nameOption),
								func(t *testing.T) {
									gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
										t.Fatalf(message)
									})
									clientset := clientset.Clientset{}
									env := map[string]string{}
									config := map[string]string{}
									platformFunc(t, env, config, &clientset)
									if tt.checkCalls != nil {
										tt.checkCalls(t, clientset, testCtx)
									}

									args := append(tt.args, runningInOptionValue...)
									args = append(args, filesOptionValue...)
									args = append(args, nameOptionValue...)
									runCommand(t, args, runOptions{env: env, config: config}, clientset, fscontentFunc, func(err error, stdout, stderr string) {
										if (err != nil) != (tt.wantErr != "") {
											t.Fatalf("errWanted: %v\nGot: %v (%s)", tt.wantErr != "", err != nil, err)
										}

										if tt.wantErr != "" {
											if !strings.Contains(err.Error(), tt.wantErr) {
												t.Errorf("%q\nerror does not contain:\n%q", err.Error(), tt.wantErr)
											}
										}
										if tt.checkOutput != nil {
											tt.checkOutput(t, stdout)
										}

										if tt.checkFS != nil {
											tt.checkFS(t, clientset.FS)
										}
									})
								})
						}
					}
				}
			}
		}
	}
}
