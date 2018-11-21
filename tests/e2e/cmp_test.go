// +build !race

package e2e

import (
	"strings"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/testingutil"

	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

// SourceTest checks the component-source-type and the source url in the annotation of the bc and dc
// appTestName is the name of the app
// sourceType is the type of the source of the component i.e git/binary/local
// source is the source of the component i.e gitURL or path to the directory or binary file
func SourceTest(appTestName string, sourceType string, source string) {
	// checking for source-type in dc
	getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	Expect(getDc).To(ContainSubstring(sourceType))

	// checking for source in dc
	getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='{{index .metadata.annotations \"app.kubernetes.io/url\"}}'")
	Expect(getDc).To(ContainSubstring(source))
}

var _ = Describe("odoCmpE2e", func() {
	const bootStrapSupervisorURI = "https://github.com/kadel/bootstrap-supervisored-s2i"
	const initContainerName = "copy-files-to-volume"
	const wildflyUri1 = "https://github.com/marekjelen/katacoda-odo-backend"
	const wildflyUri2 = "https://github.com/mik-dass/katacoda-odo-backend"
	const appRootVolumeName = "-testing-s2idata"

	var t = strconv.FormatInt(time.Now().Unix(), 10)
	var projName = fmt.Sprintf("odocmp-%s", t)
	const appTestName = "testing"

	tmpDir, err := ioutil.TempDir("", "odoCmp")
	if err != nil {
		Fail(err.Error())
	}

	Context("odo component creation", func() {

		It("should create the project and application", func() {
			runCmd("odo project create " + projName)
			runCmd("odo app create " + appTestName)
		})

		It("should be able to create a component with git source", func() {
			runCmd("odo create nodejs cmp-git --git https://github.com/openshift/nodejs-ex")
		})

		It("should list the component", func() {
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("cmp-git"))
		})

		It("should be in component description", func() {
			cmpDesc := runCmd("odo describe cmp-git")
			Expect(cmpDesc).To(ContainSubstring("source in https://github.com/openshift/nodejs-ex"))
		})

		It("should be in application description", func() {
			appDesc := runCmd("odo describe")
			Expect(appDesc).To(ContainSubstring("source in https://github.com/openshift/nodejs-ex"))
		})

		It("should list the components in the catalog", func() {
			getProj := runCmd("odo catalog list components")
			Expect(getProj).To(ContainSubstring("wildfly"))
			Expect(getProj).To(ContainSubstring("ruby"))
		})
	})

	Context("updating the component", func() {
		It("should be able to create binary component", func() {
			runCmd("curl -L -o " + tmpDir + "/sample-binary-testing-1.war " +
				"https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/e5bc575ac8b14ba2b23d66b5cb4873657e1a1489/sample.war")
			runCmd("odo create wildfly wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// TODO: remove this once https://github.com/redhat-developer/odo/issues/943 is implemented
			time.Sleep(90 * time.Second)

			// Run push
			runCmd("odo push -v 4")

			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("wildfly"))

			runCmd("oc get dc")
		})

		It("should update component from binary to binary", func() {
			runCmd("curl -L -o " + tmpDir + "/sample-binary-testing-2.war " +
				"'https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/85354d9ee8583a9c1e64a331425eede235b07a9e/sample%2520(1).war'")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-2.war")

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-2.war")
		})

		It("should update component from binary to local", func() {
			runCmd("git clone " + wildflyUri1 + " " +
				tmpDir + "/katacoda-odo-backend-1")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
		})

		It("should watch the local sources for any changes", func() {
			runCmd("odo create wildfly wildfly-watch --local " + tmpDir + "/katacoda-odo-backend-1")
			runCmd("odo push -v 4")
			startSimulationCh := make(chan bool)
			go func() {
				startMsg := <-startSimulationCh
				if startMsg {
					fmt.Println("Received signal, starting file modification simulation")
					fileModification := testingutil.FileProperties{
						FileParent:       "src/main/java/eu/mjelen/katacoda/odo/",
						FilePath:         "BackendServlet.java",
						FileType:         testingutil.RegularFile,
						ModificationType: testingutil.APPEND,
					}
					_, err := testingutil.SimulateFileModifications(filepath.Join(tmpDir, "katacoda-odo-backend-1"), fileModification)
					fmt.Printf("Triggered file modification %+v\n\n", fileModification)
					if err != nil {
						fmt.Printf("Failed performing file operation with error %v", err)
					}
				}
			}()
			success, err := pollNonRetCmdStdOutForString("odo watch wildfly-watch -v 4", time.Duration(5)*time.Minute, func(output string) bool {
				return strings.Contains(output, fmt.Sprintf("File %s changed", filepath.Join(filepath.Join(tmpDir, "katacoda-odo-backend-1"), "src/main/java/eu/mjelen/katacoda/odo/BackendServlet.java")))
			}, startSimulationCh, func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})
			Expect(success).To(Equal(true))
			Expect(err).To(BeNil())
		})

		It("should update component from local to local", func() {
			runCmd("git clone " + wildflyUri2 + " " +
				tmpDir + "/katacoda-odo-backend-2")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-2")

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-2")
		})

		It("should update component from local to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --git " + wildflyUri1)

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyUri1))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyUri1)
		})
		It("should update component from git to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --git " + wildflyUri2)

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyUri2))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyUri2)
		})

		// This is expected to be removed at the time of fixing https://github.com/redhat-developer/odo/issues/1008
		It("should create a wildfly git component", func() {
			runCmd("odo delete wildfly -f")
			runCmd("odo create wildfly wildfly --git " + wildflyUri1)
		})

		It("should update component from git to local", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
		})

		It("should update component from local to binary", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
		})

		It("should create a wildfly git component", func() {
			runCmd("odo delete wildfly -f")
			runCmd("odo create wildfly wildfly --git " + wildflyUri1)
		})

		It("should update component from git to binary", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
		})

		It("should update component from binary to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --git " + wildflyUri1)

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyUri1))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyUri1)
		})

	})

	Context("cleaning up", func() {
		It("should delete the application", func() {
			runCmd("odo app delete " + appTestName + " -f")

			runCmd("odo project delete " + projName + " -f")
			waitForDeleteCmd("odo project list", projName)
		})
	})
})

// ensures that the DeploymentConfig of the specified component
// has completely rolled out and that none of the old pods are running
// this is very useful to avoid race conditions that can occur when
// updating the component
func waitForDCOfComponentToRolloutCompletely(componentName string) {
	fullDCName := runCmd(fmt.Sprintf("oc get dc -l app.kubernetes.io/component-name=%s -o name | tr -d '\n'", componentName))
	// oc rollout status ensures that the existing DC is fully rolled out before it terminates
	// we need this because a rolling DC could cause odo update to fail due to its use
	// of the read/update-in-memory/write-changes pattern
	runCmd("oc rollout status " + fullDCName)

	simpleDCName := strings.Replace(fullDCName, "deploymentconfig.apps.openshift.io/", "", -1)
	// ensure that no more changes will occur to the name DC by waiting until there is only one pod running (the old one has terminated)
	waitForEqualCmd(fmt.Sprintf("oc get pod -o name -l deploymentconfig=%s | wc -l | tr -d '\n'", simpleDCName), "1", 2)

	// done in order to make sure that Openshift has updated the DC with the latest events
	time.Sleep(5 * time.Second)
}
