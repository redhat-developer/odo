package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

func componentTests(args ...string) {
	const initContainerName = "copy-files-to-volume"
	const wildflyURI1 = "https://github.com/marekjelen/katacoda-odo-backend"
	const wildflyURI2 = "https://github.com/mik-dass/katacoda-odo-backend"
	const appRootVolumeName = "-testing-s2idata"
	var oc helper.OcRunner
	var project string
	var context string
	var originalDir string

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		oc = helper.NewOcRunner("oc")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Generic machine readable output tests", func() {

		It("Command should fail if json is non-existent for a command", func() {
			output := helper.CmdShouldFail("odo", "version", "-o", "json")
			Expect(output).To(ContainSubstring("Machine readable output is not yet implemented for this command"))
		})

		It("Help for odo version should not contain machine output", func() {
			output := helper.CmdShouldPass("odo", "version", "--help")
			Expect(output).NotTo(ContainSubstring("Specify output format, supported format: json"))
		})

		It("Help for odo create should contain machine output", func() {
			output := helper.CmdShouldPass("odo", "create", "--help")
			Expect(output).To(ContainSubstring("Specify output format, supported format: json"))
		})

	})

	Context("Creating component", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
		})

		It("should create component even in new project", func() {
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--git", "https://github.com/openshift/nodejs-ex", "--project", project, "--context", context, "--app", "testing")...)
			helper.CmdShouldPass("odo", "push", "--context", context, "-v4")
			oc.SwitchProject(project)
			projectList := helper.CmdShouldPass("odo", "project", "list")
			Expect(projectList).To(ContainSubstring(project))
		})

		It("Without an application should create one", func() {
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "--project", project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex")...)
			helper.CmdShouldPass("odo", "push")
			appName := helper.CmdShouldPass("odo", "app", "list")
			Expect(appName).ToNot(BeEmpty())

			// checking if application name is set to "app"
			applicationName := helper.GetConfigValue("Application")
			Expect(applicationName).To(Equal("app"))

			// clean up
			helper.CmdShouldPass("odo", "app", "delete", "app", "-f")
			helper.CmdShouldFail("odo", "app", "delete", "app", "-f")
			helper.CmdShouldFail("odo", "component", "delete", componentName, "-f")

		})

		It("should create default named component when passed same context differently", func() {
			dir := filepath.Base(context)
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "--project", project, "--context", ".", "--app", "testing")...)
			componentName := helper.GetConfigValueWithContext("Name", context)
			Expect(componentName).To(ContainSubstring("nodejs-" + dir))
			helper.DeleteDir(filepath.Join(context, ".odo"))
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "--project", project, "--context", context, "--app", "testing")...)
			newComponentName := helper.GetConfigValueWithContext("Name", context)
			Expect(newComponentName).To(ContainSubstring("nodejs-" + dir))
		})

		It("should show an error when ref flag is provided with sources except git", func() {
			outputErr := helper.CmdShouldFail("odo", append(args, "create", "nodejs", "--project", project, "cmp-git", "--ref", "test")...)
			Expect(outputErr).To(ContainSubstring("The --ref flag is only valid for --git flag"))
		})

		It("create component twice fails from same directory", func() {
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "nodejs", "--project", project)...)
			output := helper.CmdShouldFail("odo", append(args, "create", "nodejs", "nodejs", "--project", project)...)
			Expect(output).To(ContainSubstring("this directory already contains a component"))
		})

		It("should list out component in json format along with path flag", func() {
			var contextPath string
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "nodejs", "--project", project)...)
			if runtime.GOOS == "windows" {
				contextPath = strings.Replace(strings.TrimSpace(context), "\\", "\\\\", -1)
			} else {
				contextPath = strings.TrimSpace(context)
			}
			// this orders the json
			desired, err := helper.Unindented(fmt.Sprintf(`{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"nodejs","source":"./","ports":["8080/TCP"]},"status":{"context":"%s","state":"Not Pushed"}}`, project, contextPath))
			Expect(err).Should(BeNil())

			actual, err := helper.Unindented(helper.CmdShouldPass("odo", append(args, "list", "-o", "json", "--path", filepath.Dir(context))...))
			Expect(err).Should(BeNil())
			// since the tests are run parallel, there might be many odo component directories in the root folder
			// so we only check for the presence of the current one
			Expect(actual).Should(ContainSubstring(desired))
		})

		It("should list out pushed components of different projects in json format along with path flag", func() {
			var contextPath string
			var contextPath2 string
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "nodejs", "--project", project)...)
			helper.CmdShouldPass("odo", append(args, "push")...)

			project2 := helper.CreateRandProject()
			context2 := helper.CreateNewContext()
			helper.Chdir(context2)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", append(args, "create", "python", "python", "--project", project2)...)
			helper.CmdShouldPass("odo", append(args, "push")...)

			if runtime.GOOS == "windows" {
				contextPath = strings.Replace(strings.TrimSpace(context), "\\", "\\\\", -1)
				contextPath2 = strings.Replace(strings.TrimSpace(context2), "\\", "\\\\", -1)
			} else {
				contextPath = strings.TrimSpace(context)
				contextPath2 = strings.TrimSpace(context2)
			}

			actual, err := helper.Unindented(helper.CmdShouldPass("odo", append(args, "list", "-o", "json", "--path", filepath.Dir(context))...))
			Expect(err).Should(BeNil())
			helper.Chdir(context)
			helper.DeleteDir(context2)
			helper.DeleteProject(project2)
			// this orders the json
			expected, err := helper.Unindented(fmt.Sprintf(`{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"nodejs","source":"./","ports":["8080/TCP"]},"status":{"context":"%s","state":"Pushed"}}`, project, contextPath))
			Expect(err).Should(BeNil())
			Expect(actual).Should(ContainSubstring(expected))
			// this orders the json
			expected, err = helper.Unindented(fmt.Sprintf(`{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"python","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"python","source":"./","ports":["8080/TCP"]},"status":{"context":"%s","state":"Pushed"}}`, project2, contextPath2))
			Expect(err).Should(BeNil())
			Expect(actual).Should(ContainSubstring(expected))

		})

		It("should create the component from the branch ref when provided", func() {
			helper.CmdShouldPass("odo", append(args, "create", "ruby", "ref-test", "--project", project, "--git", "https://github.com/girishramnani/ruby-ex.git", "--ref", "develop")...)
			helper.CmdShouldPass("odo", "push")
		})

		It("should list the component", func() {
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--project", project, "--git", "https://github.com/openshift/nodejs-ex", "--min-memory", "100Mi", "--max-memory", "300Mi", "--min-cpu", "0.1", "--max-cpu", "2", "--context", context, "--app", "testing")...)
			helper.CmdShouldPass("odo", "push", "--context", context)
			cmpList := helper.CmdShouldPass("odo", append(args, "list", "--project", project)...)
			Expect(cmpList).To(ContainSubstring("cmp-git"))
			actualCompListJSON := helper.CmdShouldPass("odo", append(args, "list", "--project", project, "-o", "json")...)
			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"cmp-git","namespace":"%s","creationTimestamp":null},"spec":{"app":"testing","type":"nodejs","source":"https://github.com/openshift/nodejs-ex","env":[{"name":"DEBUG_PORT","value":"5858"}]},"status":{"state":"Pushed"}}]}`, project)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))
			cmpAllList := helper.CmdShouldPass("odo", append(args, "list", "--all")...)
			Expect(cmpAllList).To(ContainSubstring("cmp-git"))
			helper.CmdShouldPass("odo", append(args, "delete", "cmp-git", "-f")...)
		})

		It("should list the component when it is not pushed", func() {
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--project", project, "--git", "https://github.com/openshift/nodejs-ex", "--min-memory", "100Mi", "--max-memory", "300Mi", "--min-cpu", "0.1", "--max-cpu", "2", "--context", context, "--app", "testing")...)
			cmpList := helper.CmdShouldPass("odo", append(args, "list", "--context", context)...)
			Expect(cmpList).To(ContainSubstring("cmp-git"))
			Expect(cmpList).To(ContainSubstring("Not Pushed"))
			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", context)...)
		})

		It("should list the component in the same app when one is pushed and the other one is not pushed", func() {
			helper.Chdir(originalDir)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--project", project, "--git", "https://github.com/openshift/nodejs-ex", "--context", context, "--app", "testing")...)
			helper.CmdShouldPass("odo", append(args, "push", "--context", context)...)

			context2 := helper.CreateNewContext()
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git-2", "--project", project, "--git", "https://github.com/openshift/nodejs-ex", "--context", context2, "--app", "testing")...)

			cmpList := helper.CmdShouldPass("odo", append(args, "list", "--context", context2)...)

			Expect(cmpList).To(ContainSubstring("cmp-git"))
			Expect(cmpList).To(ContainSubstring("cmp-git-2"))
			Expect(cmpList).To(ContainSubstring("Not Pushed"))
			Expect(cmpList).To(ContainSubstring("Pushed"))

			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", context)...)
			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", context2)...)
			helper.DeleteDir(context2)
		})

		It("should succeed listing catalog components", func() {

			// Since components catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-o", "json")
			Expect(output).To(ContainSubstring("List"))
		})

		It("binary component should not fail when --context is not set", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), context)
			// Was failing due to https://github.com/openshift/odo/issues/1969
			helper.CmdShouldPass("odo", "create", "java:8", "sb-jar-test", "--project",
				project, "--binary", filepath.Join(context, "sb.jar"))
		})

		It("binary component should fail when --binary is not in --context folder", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), context)

			newContext := helper.CreateNewContext()
			defer helper.DeleteDir(newContext)

			output := helper.CmdShouldFail("odo", "create", "java:8", "sb-jar-test", "--project",
				project, "--binary", filepath.Join(context, "sb.jar"), "--context", newContext)
			Expect(output).To(ContainSubstring("inside of the context directory"))
		})

		It("binary component is valid if path is relative and includes ../", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), context)

			relativeContext := fmt.Sprintf("..%c%s", filepath.Separator, filepath.Base(context))
			fmt.Printf("relativeContext = %#v\n", relativeContext)

			helper.CmdShouldPass("odo", "create", "java:8", "sb-jar-test", "--project",
				project, "--binary", filepath.Join(context, "sb.jar"), "--context", relativeContext)
		})

	})

	Context("Test odo push with --source and --config flags", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
		})
		Context("Using project flag(--project) and current directory", func() {
			It("create local nodejs component and push source and code separately", func() {
				appName := "nodejs-push-test"
				cmpName := "nodejs"
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--project", project)...)

				// component doesn't exist yet so attempt to only push source should fail
				helper.CmdShouldFail("odo", "push", "--source")

				// Push only config and see that the component is created but wothout any source copied
				helper.CmdShouldPass("odo", "push", "--config")
				oc.VerifyCmpExists(cmpName, appName, project)

				// Push only source and see that the component is updated with source code
				helper.CmdShouldPass("odo", "push", "--source")
				oc.VerifyCmpExists(cmpName, appName, project)
				remoteCmdExecPass := oc.CheckCmdOpInRemoteCmpPod(
					cmpName,
					appName,
					project,
					[]string{"ls", "-la", "/tmp/src/package.json"},
					func(cmdOp string, err error) bool {
						if err != nil {
							return false
						}
						return true
					},
				)
				Expect(remoteCmdExecPass).To(Equal(true))
			})

			It("create local nodejs component and push source and code at once", func() {
				appName := "nodejs-push-test"
				cmpName := "nodejs-push-atonce"
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--project", project)...)

				// Push only config and see that the component is created but wothout any source copied
				helper.CmdShouldPass("odo", "push")
				oc.VerifyCmpExists(cmpName, appName, project)
				remoteCmdExecPass := oc.CheckCmdOpInRemoteCmpPod(
					cmpName,
					appName,
					project,
					[]string{"ls", "-la", "/tmp/src/package.json"},
					func(cmdOp string, err error) bool {
						if err != nil {
							return false
						}
						return true
					},
				)
				Expect(remoteCmdExecPass).To(Equal(true))
			})

		})

		Context("when --context is used", func() {
			// don't need to switch to any dir here, as this test should use --context flag
			It("create local nodejs component and push source and code separately", func() {
				appName := "nodejs-push-context-test"
				cmpName := "nodejs"
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--context", context, "--app", appName, "--project", project)...)
				//TODO: verify that config was properly created

				// component doesn't exist yet so attempt to only push source should fail
				helper.CmdShouldFail("odo", "push", "--source", "--context", context)

				// Push only config and see that the component is created but wothout any source copied
				helper.CmdShouldPass("odo", "push", "--config", "--context", context)
				oc.VerifyCmpExists(cmpName, appName, project)

				// Push only source and see that the component is updated with source code
				helper.CmdShouldPass("odo", "push", "--source", "--context", context)
				oc.VerifyCmpExists(cmpName, appName, project)
				remoteCmdExecPass := oc.CheckCmdOpInRemoteCmpPod(
					cmpName,
					appName,
					project,
					[]string{"ls", "-la", "/tmp/src/package.json"},
					func(cmdOp string, err error) bool {
						if err != nil {
							return false
						}
						return true
					},
				)
				Expect(remoteCmdExecPass).To(Equal(true))
			})

			It("create local nodejs component and push source and code at once", func() {
				appName := "nodejs-push-context-test"
				cmpName := "nodejs-push-atonce"
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--context", context, "--project", project)...)

				// Push both config and source
				helper.CmdShouldPass("odo", "push", "--context", context)
				oc.VerifyCmpExists(cmpName, appName, project)
				remoteCmdExecPass := oc.CheckCmdOpInRemoteCmpPod(
					cmpName,
					appName,
					project,
					[]string{"ls", "-la", "/tmp/src/package.json"},
					func(cmdOp string, err error) bool {
						if err != nil {
							return false
						}
						return true
					},
				)
				Expect(remoteCmdExecPass).To(Equal(true))
			})
		})
	})

	Context("Creating Component even in new project", func() {
		var project string
		JustBeforeEach(func() {
			project = helper.RandString(10)
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
		})
		It("should create component", func() {
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--git", "https://github.com/openshift/nodejs-ex", "--project", project, "--context", context, "--app", "testing")...)
			helper.CmdShouldPass("odo", "push", "--context", context, "-v4")
			oc.SwitchProject(project)
			projectList := helper.CmdShouldPass("odo", "project", "list")
			Expect(projectList).To(ContainSubstring(project))
		})
	})

	Context("Test odo push with --now flag during creation", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
		})
		It("should successfully create config and push code in one create command with --now", func() {
			appName := "nodejs-create-now-test"
			cmpName := "nodejs-push-atonce"
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--project", project, "--now")...)

			oc.VerifyCmpExists(cmpName, appName, project)
			remoteCmdExecPass := oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				project,
				[]string{"ls", "-la", "/tmp/src/package.json"},
				func(cmdOp string, err error) bool {
					if err != nil {
						return false
					}
					return true
				},
			)
			Expect(remoteCmdExecPass).To(Equal(true))
		})
	})

	Context("when component is in the current directory and --project flag is used", func() {

		appName := "app"
		componentName := "my-component"

		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
			helper.DeleteProject(project)
		})

		It("create local nodejs component twice and fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "--project", project, "--env", "key=value,key1=value1")...)
			output := helper.CmdShouldFail("odo", append(args, "create", "nodejs", "--project", project, "--env", "key=value,key1=value1")...)
			Expect(output).To(ContainSubstring("this directory already contains a component"))
		})

		It("creates and pushes local nodejs component and then deletes --all", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", componentName, "--app", appName, "--project", project, "--env", "key=value,key1=value1")...)
			helper.CmdShouldPass("odo", append(args, "push", "--context", context)...)
			helper.CmdShouldPass("odo", append(args, "delete", "--context", context, "-f", "--all", "--app", appName)...)
			componentList := helper.CmdShouldPass("odo", append(args, "list", "--context", context, "--app", appName, "--project", project)...)
			Expect(componentList).NotTo(ContainSubstring(componentName))
			files := helper.ListFilesInDir(context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("creates a local python component, pushes it and then deletes it using --all flag", func() {
			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.CmdShouldPass("odo", append(args, "create", "python", componentName, "--app", appName, "--project", project, "--context", context)...)
			helper.CmdShouldPass("odo", append(args, "push", "--context", context)...)
			helper.CmdShouldPass("odo", append(args, "delete", "--context", context, "-f")...)
			helper.CmdShouldPass("odo", append(args, "delete", "--all", "-f", "--context", context)...)
			componentList := helper.CmdShouldPass("odo", append(args, "list", "--context", context, "--app", appName, "--project", project)...)
			Expect(componentList).NotTo(ContainSubstring(componentName))
			files := helper.ListFilesInDir(context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("creates a local python component and check for unsupported warning", func() {
			helper.CopyExample(filepath.Join("source", "python"), context)
			output := helper.CmdShouldPass("odo", append(args, "create", "python", componentName, "--app", appName, "--project", project, "--context", context)...)
			Expect(output).To(ContainSubstring("Warning: python is not fully supported by odo, and it is not guaranteed to work."))
		})

		It("creates a local nodejs component and check unsupported warning hasn't occured", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			output := helper.CmdShouldPass("odo", append(args, "create", "nodejs", componentName, "--app", appName, "--project", project, "--context", context)...)
			Expect(output).NotTo(ContainSubstring("Warning"))
		})
	})

	/*
			Enable once #1782 and #1778 are fixed

				Context("odo component updating", func() {
					JustBeforeEach(func() {
						project = helper.CreateRandProject()
					})

					JustAfterEach(func() {
						helper.DeleteProject(project)
					})

					It("should be able to create a git component and update it from local to git", func() {
						helper.CopyExample(filepath.Join("source", "nodejs"), context)
						helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--project", project, "--min-cpu", "0.1", "--max-cpu", "2", "--context", context, "--app", "testing")...)
						helper.CmdShouldPass("odo", "push", "--context", context, "-v", "4")
						getCPULimit := oc.MaxCPU("cmp-git", "testing", project)
						Expect(getCPULimit).To(ContainSubstring("2"))
						getCPURequest := oc.MinCPU("cmp-git", "testing", project)
						Expect(getCPURequest).To(ContainSubstring("100m"))

						// update the component config according to the git component
						helper.CmdShouldPass("odo", "config", "set", "sourcelocation", "https://github.com/openshift/nodejs-ex", "--context", context, "-f")
						helper.CmdShouldPass("odo", "config", "set", "sourcetype", "git", "--context", context, "-f")

						// check if the earlier resource requests are still valid
						helper.CmdShouldPass("odo", "push", "--context", context, "-v", "4")
						getCPULimit = oc.MaxCPU("cmp-git", "testing", project)
						Expect(getCPULimit).To(ContainSubstring("2"))
						getCPURequest = oc.MinCPU("cmp-git", "testing", project)
						Expect(getCPURequest).To(ContainSubstring("100m"))

						// check the source location and type in the deployment config
						getSourceLocation := oc.SourceLocationDC("cmp-git", "testing", project)
						Expect(getSourceLocation).To(ContainSubstring("https://github.com/openshift/nodejs-ex"))
						getSourceType := oc.SourceTypeDC("cmp-git", "testing", project)
						Expect(getSourceType).To(ContainSubstring("git"))

						// since the current component type is git
						// check the source location and type in the build config
						getSourceLocation = oc.SourceLocationBC("cmp-git", "testing", project)
						Expect(getSourceLocation).To(ContainSubstring("https://github.com/openshift/nodejs-ex"))
						getSourceType = oc.SourceTypeBC("cmp-git", "testing", project)
						Expect(getSourceType).To(ContainSubstring("Git"))
					})

					It("should be able to update a component from git to local", func() {
						helper.CmdShouldPass("odo", append(args, "create", "nodejs", "cmp-git", "--project", project, "--git", "https://github.com/openshift/nodejs-ex", "--min-memory", "100Mi", "--max-memory", "300Mi", "--context", context, "--app", "testing")...)
						helper.CmdShouldPass("odo", "push", "--context", context, "-v", "4")
						getMemoryLimit := oc.MaxMemory("cmp-git", "testing", project)
						Expect(getMemoryLimit).To(ContainSubstring("300Mi"))
						getMemoryRequest := oc.MinMemory("cmp-git", "testing", project)
						Expect(getMemoryRequest).To(ContainSubstring("100Mi"))

						// update the component config according to the git component
						helper.CopyExample(filepath.Join("source", "nodejs"), context)
						helper.CmdShouldPass("odo", "config", "set", "sourcelocation", "./", "--context", context, "-f")
						helper.CmdShouldPass("odo", "config", "set", "sourcetype", "local", "--context", context, "-f")

						// check if the earlier resource requests are still valid
						helper.CmdShouldPass("odo", "push", "--context", context, "-v", "4")
						getMemoryLimit = oc.MaxMemory("cmp-git", "testing", project)
						Expect(getMemoryLimit).To(ContainSubstring("300Mi"))
						getMemoryRequest = oc.MinMemory("cmp-git", "testing", project)
						Expect(getMemoryRequest).To(ContainSubstring("100Mi"))

						// check the source location and type in the deployment config
						getSourceLocation := oc.SourceLocationDC("cmp-git", "testing", project)
						var sourcePath string
						if runtime.GOOS == "windows" {
							sourcePath = "file:///./"
						} else {
							sourcePath = "file://./"
						}
						Expect(getSourceLocation).To(ContainSubstring(sourcePath))
						getSourceType := oc.SourceTypeDC("cmp-git", "testing", project)
						Expect(getSourceType).To(ContainSubstring("local"))
					})
				})
		})
	*/

	Context("odo component delete, list and describe", func() {
		appName := "app"
		cmpName := "nodejs"

		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
		})

		It("should pass inside a odo directory without component name as parameter", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)...)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// changing directory to the context directory
			helper.Chdir(context)
			cmpListOutput := helper.CmdShouldPass("odo", "list")
			Expect(cmpListOutput).To(ContainSubstring(cmpName))
			helper.CmdShouldPass("odo", "describe")
			helper.CmdShouldPass("odo", "delete", "-f")
		})

		It("should fail outside a odo directory without component name as parameter", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)...)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// list command should fail as no app flag is given
			helper.CmdShouldFail("odo", "list", "--project", project)
			// commands should fail as the component name is missing
			helper.CmdShouldFail("odo", "describe", "--app", appName, "--project", project)
			helper.CmdShouldFail("odo", "delete", "-f", "--app", appName, "--project", project)
		})

		It("should pass outside a odo directory with component name as parameter", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)...)
			helper.CmdShouldPass("odo", append(args, "push", "--context", context)...)

			cmpListOutput := helper.CmdShouldPass("odo", append(args, "list", "--app", appName, "--project", project)...)
			Expect(cmpListOutput).To(ContainSubstring(cmpName))

			actualDesCompJSON := helper.CmdShouldPass("odo", append(args, "describe", cmpName, "--app", appName, "--project", project, "-o", "json")...)
			var sourcePath string
			if runtime.GOOS == "windows" {
				sourcePath = "file:///./"
			} else {
				sourcePath = "file://./"
			}
			desiredDesCompJSON := fmt.Sprintf(`{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"nodejs","source":"%s","env":[{"name":"DEBUG_PORT","value":"5858"}]},"status":{"state":"Pushed"}}`, project, sourcePath)
			Expect(desiredDesCompJSON).Should(MatchJSON(actualDesCompJSON))

			helper.CmdShouldPass("odo", append(args, "delete", cmpName, "--app", appName, "--project", project, "-f")...)
		})
	})

	Context("when running odo push multiple times, check for existence of environment variables", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
		})

		It("should should retain the same environment variable on multiple push", func() {
			componentName := helper.RandString(6)
			appName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", componentName, "--app", appName, "--project", project, "--context", context)...)
			helper.CmdShouldPass("odo", "push", "--context", context)

			helper.Chdir(context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "FOO=BAR")
			helper.CmdShouldPass("odo", "push")

			dcName := oc.GetDcName(componentName, project)
			stdOut := helper.CmdShouldPass("oc", "get", "dc/"+dcName, "-n", project, "-o", "go-template={{ .spec.template.spec }}{{.env}}")
			Expect(stdOut).To(ContainSubstring("FOO"))

			helper.CmdShouldPass("odo", "push")
			stdOut = oc.DescribeDc(dcName, project)
			Expect(stdOut).To(ContainSubstring("FOO"))
		})
	})

	Context("Creating component with numeric named context", func() {
		var contextNumeric string
		JustBeforeEach(func() {
			var err error
			ts := time.Now().UnixNano()
			contextNumeric, err = ioutil.TempDir("", fmt.Sprint(ts))
			Expect(err).ToNot(HaveOccurred())
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.DeleteDir(contextNumeric)
		})

		It("should create default named component in a directory with numeric name", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), contextNumeric)
			helper.CmdShouldPass("odo", append(args, "create", "nodejs", "--project", project, "--context", contextNumeric, "--app", "testing")...)
			helper.CmdShouldPass("odo", "push", "--context", contextNumeric, "-v4")
		})
	})

	Context("when creating component with improper memory quantities", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
		})
		It("should fail gracefully with proper error message", func() {
			stdError := helper.CmdShouldFail("odo", append(args, "create", "java:8", "backend", "--memory", "1GB", "--project", project, "--context", context)...)
			Expect(stdError).ToNot(ContainSubstring("panic: cannot parse"))
			Expect(stdError).To(ContainSubstring("quantities must match the regular expression"))
		})
	})
}
