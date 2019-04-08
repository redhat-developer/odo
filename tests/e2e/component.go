package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/e2e/helper"
)

// SourceTest checks the component-source-type and the source url in the annotation of the bc and dc
// appTestName is the name of the app
// sourceType is the type of the source of the component i.e git/binary/local
// source is the source of the component i.e gitURL or path to the directory or binary file
func SourceTest(appTestName string, sourceType string, source string) {
	// checking for source-type in dc
	getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	Expect(getDc).To(ContainSubstring(sourceType))

	// checking for source in dc
	getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='{{index .metadata.annotations \"app.kubernetes.io/url\"}}'")
	Expect(getDc).To(ContainSubstring(source))
}

func componentTests(componentCmdPrefix string) {
	const initContainerName = "copy-files-to-volume"
	const wildflyURI1 = "https://github.com/marekjelen/katacoda-odo-backend"
	const wildflyURI2 = "https://github.com/mik-dass/katacoda-odo-backend"
	const appRootVolumeName = "-testing-s2idata"

	t := strconv.FormatInt(time.Now().Unix(), 10)
	projName := generateTimeBasedName("odocmp")
	const appTestName = "testing"
	/*
		tmpDir, err := ioutil.TempDir("", "odoCmp")
		if err != nil {
			Fail(err.Error())
		}
	*/
	Context("odo component creation without application", func() {
		It("creating a component without an application should create one", func() {
			// new project == no app
			projectName := generateTimeBasedName("project")
			odoCreateProject(projectName)
			//Expect(runCmdShouldPass("odo app list")).To(ContainSubstring("no applications"))

			componentName := generateTimeBasedName("frontend")

			// create a frontend component, an app should have been created
			runCmdShouldPass(componentCmdPrefix + " create nodejs --project " + projectName + " " + componentName + " --ref master --git https://github.com/openshift/nodejs-ex")
			runCmdShouldPass("odo push")
			appName := runCmdShouldPass("odo app list")
			Expect(appName).ToNot(BeEmpty())

			// clean up
			runCmdShouldPass("odo component delete " + componentName + " -f")
			runCmdShouldPass("odo app delete -f")

		})
	})

	Context("Regression : listing component outside of component directory should fail", func() {
		It("creates a component from local context, tries to list components from outside and fails", func() {
			dirName := generateTimeBasedName("context_dir")
			// simulate .odo not being present
			runCmdShouldPass("mv .odo .odo_tmp")
			session := runCmdShouldFail("odo component list")
			Expect(session).To(ContainSubstring("the current directory does not represent an odo component"))
			session = runCmdShouldFail("odo app list")
			Expect(session).To(ContainSubstring("the current directory does not represent an odo component"))
			session = runCmdShouldFail("odo config view")
			Expect(session).To(ContainSubstring("the current directory does not represent an odo component"))
			// clean up
			runCmdShouldPass("mv .odo_tmp .odo")
			os.RemoveAll(dirName)
		})
	})

	Context("odo component creation", func() {

		It("should create the project", func() {
			odoCreateProject(projName)
		})

		//appTestName

		It("should show an error when ref flag is provided with sources except git", func() {
			outputErr := runCmdShouldFail(fmt.Sprintf(componentCmdPrefix+" create nodejs cmp-git-%s --ref test", t))
			Expect(outputErr).To(ContainSubstring("The --ref flag is only valid for --git flag"))
		})

		It("should create the component from the branch ref when provided", func() {
			runCmdShouldPass(fmt.Sprintf(componentCmdPrefix+" create ruby ref-test-%s --git https://github.com/girishramnani/ruby-ex.git --ref develop", t))
			runCmdShouldPass("odo push")
			runCmdShouldPass(fmt.Sprintf("odo url create ref-test-%s", t))

			routeURL := determineRouteURL() + "/health"
			responseStringMatchStatus := matchResponseSubString(routeURL, "develop", 180, 1)
			Expect(responseStringMatchStatus).Should(BeTrue())
		})

		It("should be able to create a component with git source", func() {
			runCmdShouldPass("mkdir -p cmp-git")
			runCmdShouldPass(componentCmdPrefix + " create nodejs cmp-git --git https://github.com/openshift/nodejs-ex --min-memory 100Mi --max-memory 300Mi --min-cpu 0.1 --max-cpu 2 --context cmp-git/ --app " + appTestName)
			runCmdShouldPass("odo push --context cmp-git/")
			getMemoryLimit := runCmdShouldPass("oc get dc cmp-git-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
			)
			Expect(getMemoryLimit).To(ContainSubstring("300Mi"))
			getMemoryRequest := runCmdShouldPass("oc get dc cmp-git-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
			)
			Expect(getMemoryRequest).To(ContainSubstring("100Mi"))

			getCPULimit := runCmdShouldPass("oc get dc cmp-git-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.cpu}}{{end}}'",
			)
			Expect(getCPULimit).To(ContainSubstring("2"))
			getCPURequest := runCmdShouldPass("oc get dc cmp-git-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.cpu}}{{end}}'",
			)
			Expect(getCPURequest).To(ContainSubstring("100m"))
			runCmdShouldPass("cd cmp-git && odo component delete -f")
			runCmdShouldPass("rm -rf cmp-git")
		})

		It("should list the component", func() {
			runCmdShouldPass("mkdir -p cmp-git")
			runCmdShouldPass(componentCmdPrefix + " create nodejs cmp-git --git https://github.com/openshift/nodejs-ex --min-memory 100Mi --max-memory 300Mi --min-cpu 0.1 --max-cpu 2 --context cmp-git/ --app " + appTestName)
			runCmdShouldPass("odo push --context cmp-git/")
			cmpList := runCmdShouldPass("cd cmp-git && " + componentCmdPrefix + " list")
			Expect(cmpList).To(ContainSubstring("cmp-git"))
			runCmdShouldPass("cd cmp-git && odo component delete -f")
			runCmdShouldPass("rm -rf cmp-git")
		})
		/*
			It("should be in component description", func() {
				cmpDesc := runCmdShouldPass(componentCmdPrefix + " describe cmp-git")
				Expect(cmpDesc).To(ContainSubstring("Source: https://github.com/openshift/nodejs-ex"))
			})

			It("should be in application description", func() {
				appDesc := runCmdShouldPass(componentCmdPrefix + " describe")
				Expect(appDesc).To(ContainSubstring("Source: https://github.com/openshift/nodejs-ex"))
			})
		*/
		It("should list the components in the catalog", func() {
			getProj := runCmdShouldPass("odo catalog list components")
			Expect(getProj).To(ContainSubstring("wildfly"))
			Expect(getProj).To(ContainSubstring("ruby"))
			Expect(getProj).To(ContainSubstring("nodejs"))

			// check that the nodejs string does not contain the hidden versions
			lines := strings.Split(strings.Replace(getProj, "\r\n", "\n", -1), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "nodejs") {
					Expect(getProj).To(Not(ContainSubstring("0.10")))
				}
			}
		})
	})
	/*
		Context("updating the component", func() {
			It("should be able to create binary component", func() {
				runCmdShouldPass("curl -L -o " + tmpDir + "/sample-binary-testing-1.war " +
					"https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/e5bc575ac8b14ba2b23d66b5cb4873657e1a1489/sample.war")
				runCmdShouldPass(componentCmdPrefix + " create wildfly wildfly --binary " + tmpDir + "/sample-binary-testing-1.war --memory 500Mi")

				// TODO: remove this once https://github.com/openshift/odo/issues/943 is implemented
				time.Sleep(90 * time.Second)

				// Run push
				runCmdShouldPass(componentCmdPrefix + " push -v 4")

				// Verify memory limits to be same as configured
				getMemoryLimit := runCmdShouldPass("oc get dc wildfly-" +
					appTestName +
					" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
				)
				Expect(getMemoryLimit).To(ContainSubstring("500Mi"))
				getMemoryRequest := runCmdShouldPass("oc get dc wildfly-" +
					appTestName +
					" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
				)
				Expect(getMemoryRequest).To(ContainSubstring("500Mi"))

				cmpList := runCmdShouldPass(componentCmdPrefix + " list")
				Expect(cmpList).To(ContainSubstring("wildfly"))

			})

			It("should update component from binary to binary", func() {
				runCmdShouldPass("curl -L -o " + tmpDir + "/sample-binary-testing-2.war " +
					"'https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/85354d9ee8583a9c1e64a331425eede235b07a9e/sample%2520(1).war'")

				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --binary " + tmpDir + "/sample-binary-testing-2.war")

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-2.war")
			})

			It("should update component from binary to local", func() {
				runCmdShouldPass("git clone " + wildflyURI1 + " " +
					tmpDir + "/katacoda-odo-backend-1")

				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// Verify memory limits to be same as configured
				getMemoryLimit := runCmdShouldPass("oc get dc wildfly-" +
					appTestName +
					" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
				)
				Expect(getMemoryLimit).To(ContainSubstring("500Mi"))
				getMemoryRequest := runCmdShouldPass("oc get dc wildfly-" +
					appTestName +
					" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
				)
				Expect(getMemoryRequest).To(ContainSubstring("500Mi"))

				SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
			})

			It("should update component from local to local", func() {
				runCmdShouldPass("git clone " + wildflyURI2 + " " +
					tmpDir + "/katacoda-odo-backend-2")

				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --local " + tmpDir + "/katacoda-odo-backend-2")

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-2")
			})

			It("should update component from local to git", func() {
				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --git " + wildflyURI1)

				// checking bc for updates
				getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
				Expect(getBc).To(Equal(wildflyURI1))

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "git", wildflyURI1)
			})
			It("should update component from git to git", func() {
				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --git " + wildflyURI2)

				// checking bc for updates
				getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
				Expect(getBc).To(Equal(wildflyURI2))

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "git", wildflyURI2)
			})

			// This is expected to be removed at the time of fixing https://github.com/openshift/odo/issues/1008
			It("should create a wildfly git component", func() {
				runCmdShouldPass(componentCmdPrefix + " delete wildfly -f")
				runCmdShouldPass(componentCmdPrefix + " create wildfly wildfly --git " + wildflyURI1)
			})

			It("should update component from git to local", func() {
				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
			})

			It("should update component from local to binary", func() {
				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
			})

			It("should create a wildfly git component", func() {
				runCmdShouldPass(componentCmdPrefix + " delete wildfly -f")
				runCmdShouldPass(componentCmdPrefix + " create wildfly wildfly --git " + wildflyURI1)
			})

			It("should update component from git to binary", func() {
				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
			})

			It("should update component from binary to git", func() {
				waitForDCOfComponentToRolloutCompletely("wildfly")
				runCmdShouldPass(componentCmdPrefix + " update wildfly --git " + wildflyURI1)

				// checking bc for updates
				getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
				Expect(getBc).To(Equal(wildflyURI1))

				// checking for init containers
				getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.initContainers}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring(initContainerName))

				// checking for volumes
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.volumes}}" +
					"{{.name}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

				// checking for volumes mounts
				getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
					"{{.name}}{{end}}{{end}}'")
				Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

				SourceTest(appTestName, "git", wildflyURI1)
			})
		})
	*/
	Context("cleaning up", func() {
		It("should delete the application", func() {

			runCmdShouldPass("odo app delete " + appTestName + " -f")

			runCmdShouldPass("odo project delete " + projName + " -f")
			waitForDeleteCmd("odo project list", projName)
		})
	})

	Context("Test odo push with --source and --config flags", func() {
		//new clean project and context for each test
		var project string
		var context string

		//  current directory and project (before eny test is run) so it can restored  after all testing is done
		var originalDir string

		// Setup up state for each test spec
		// create new project (not set as active) and new context directory for each test spec
		// This is before every spec (It)
		BeforeEach(func() {
			project = helper.OcCreateRandProject()
			context = helper.CreateNewContext()
		})

		// Clean up after the test
		// This is run after every Spec (It)
		AfterEach(func() {
			helper.OcDeleteProject(project)
			helper.DeleteDir(context)
		})

		Context("when project flag(--project) is used", func() {

			Context("when using current directory", func() {
				// we will be testing components that are created from the current directory
				// switch to the clean context dir before each test
				JustBeforeEach(func() {
					originalDir = helper.Getwd()
					helper.Chdir(context)
				})

				// go back to original directory after each test
				JustAfterEach(func() {
					helper.Chdir(originalDir)
				})

				It("create local nodejs component and push source and code separately", func() {
					appName := "nodejs-push-test"
					cmpName := "nodejs"
					helper.CopyExample(filepath.Join("source", "nodejs"), context)

					helper.CmdShouldPass("odo component create nodejs " + cmpName + " --app " + appName + " --project " + project)

					// component doesn't exist yet so attempt to only push source should fail
					helper.CmdShouldFail("odo push --source")

					// Push only config and see that the component is created but wothout any source copied
					helper.CmdShouldPass("odo push --config")
					helper.VerifyCmpExists(cmpName, appName, project)

					// Push only source and see that the component is updated with source code
					helper.CmdShouldPass("odo push --source")
					helper.VerifyCmpExists(cmpName, appName, project)
					remoteCmdExecPass := helper.CheckCmdOpInRemoteCmpPod(
						cmpName,
						appName,
						project,
						"ls -lai /tmp/src/package.json",
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

					helper.CmdShouldPass("odo component create nodejs " + cmpName + " --app " + appName + " --project " + project)

					// Push only config and see that the component is created but wothout any source copied
					helper.CmdShouldPass("odo push")
					helper.VerifyCmpExists(cmpName, appName, project)
					remoteCmdExecPass := helper.CheckCmdOpInRemoteCmpPod(
						cmpName,
						appName,
						project,
						"ls -lai /tmp/src/package.json",
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

					helper.CmdShouldPass("odo component create nodejs " + cmpName + " --context " + context + " --app " + appName + " --project " + project)
					//TODO: verify that config was properly created

					// component doesn't exist yet so attempt to only push source should fail
					helper.CmdShouldFail("odo push --source --context " + context)

					// Push only config and see that the component is created but wothout any source copied
					helper.CmdShouldPass("odo push --config --context " + context)
					helper.VerifyCmpExists(cmpName, appName, project)

					// Push only source and see that the component is updated with source code
					helper.CmdShouldPass("odo push --source --context " + context)
					helper.VerifyCmpExists(cmpName, appName, project)
					remoteCmdExecPass := helper.CheckCmdOpInRemoteCmpPod(
						cmpName,
						appName,
						project,
						"ls -lai /tmp/src/package.json",
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

					helper.CmdShouldPass("odo component create nodejs " + cmpName + " --app " + appName + " --context " + context + " --project " + project)

					// Push both config and source
					helper.CmdShouldPass("odo push --context " + context)
					helper.VerifyCmpExists(cmpName, appName, project)
					remoteCmdExecPass := helper.CheckCmdOpInRemoteCmpPod(
						cmpName,
						appName,
						project,
						"ls -lai /tmp/src/package.json",
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
	})

}
