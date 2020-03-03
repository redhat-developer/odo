package e2escenarios

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo supported images e2e tests", func() {
	//new clean project and context for each test
	var project string
	var context string
	appName := "app"

	var oc helper.OcRunner

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	OdoWatch := func(srcType, routeURL, project, appName, context string) {

		startSimulationCh := make(chan bool)
		go func() {
			startMsg := <-startSimulationCh
			if startMsg {
				err := os.MkdirAll(filepath.Join(context, ".abc"), 0755)
				if err != nil {
					panic(err)
				}
				err = os.MkdirAll(filepath.Join(context, "abcd"), 0755)
				if err != nil {
					panic(err)
				}
				_, err = os.Create(filepath.Join(context, "a.txt"))
				if err != nil {
					panic(err)
				}

				helper.DeleteDir(filepath.Join(context, "abcd"))

				if srcType == "openjdk" {
					helper.ReplaceString(filepath.Join(context, "src", "main", "java", "MessageProducer.java"), "Hello", "Hello odo")
				} else {
					helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello odo")
				}
			}
		}()

		success, err := helper.WatchNonRetCmdStdOut(
			("odo watch " + srcType + "-app" + " -v 4 " + "--context " + context),
			time.Duration(5)*time.Minute,
			func(output string) bool {
				curlURL := helper.CmdShouldPass("curl", routeURL)
				if strings.Contains(curlURL, "Hello odo") {
					// Verify delete from component pod
					podName := oc.GetRunningPodNameOfComp(srcType+"-app", project)
					envs := oc.GetEnvs(srcType+"-app", appName, project)
					dir := envs["ODO_S2I_SRC_BIN_PATH"]
					stdOut := oc.ExecListDir(podName, project, filepath.Join(dir, "src"))
					Expect(stdOut).To(ContainSubstring(("a.txt")))
					Expect(stdOut).To(Not(ContainSubstring("abcd")))
				}
				return strings.Contains(curlURL, "Hello odo")
			},
			startSimulationCh,
			func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})

		Expect(success).To(Equal(true))
		Expect(err).To(BeNil())

		// Verify memory limits to be same as configured
		getMemoryLimit := oc.MaxMemory(srcType+"-app", appName, project)
		Expect(getMemoryLimit).To(ContainSubstring("700Mi"))
		getMemoryRequest := oc.MinMemory(srcType+"-app", appName, project)
		Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
	}

	// verifySupportedImage takes arguments supported images, source type, image type, namespace and application name.
	// Also verify the flow of odo commands with respect to supported images only.
	verifySupportedImage := func(image, srcType, cmpType, project, appName string) {

		// create the component
		helper.CopyExample(filepath.Join("source", srcType), context)
		helper.CmdShouldPass("odo", "create", cmpType, srcType+"-app", "--project", project, "--context", context, "--app", appName, "--min-memory", "400Mi", "--max-memory", "700Mi")

		// push component and validate
		helper.CmdShouldPass("odo", "push", "--context", context)
		cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
		Expect(cmpList).To(ContainSubstring(srcType + "-app"))

		// create a url
		helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context)
		helper.CmdShouldPass("odo", "push", "--context", context)
		routeURL := helper.DetermineRouteURL(context)

		// Ping said URL
		helper.HttpWaitFor(routeURL, "Hello", 90, 1)

		// edit source and validate
		if srcType == "openjdk" {
			helper.ReplaceString(filepath.Join(context, "src", "main", "java", "MessageProducer.java"), "Hello", "Hello Java UPDATED")
			helper.CmdShouldPass("odo", "push", "--context", context)
			helper.HttpWaitFor(routeURL, "Hello Java UPDATED", 90, 1)
		} else {
			helper.ReplaceString(filepath.Join(context, "server.js"), "Hello", "Hello nodejs UPDATED")
			helper.CmdShouldPass("odo", "push", "--context", context)
			helper.HttpWaitFor(routeURL, "Hello nodejs UPDATED", 90, 1)
		}

		// odo watch and validate
		OdoWatch(srcType, routeURL, project, appName, context)

		// delete the component and validate
		helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		cmpLst := helper.CmdShouldPass("odo", "list", "--context", context)
		Expect(cmpLst).To(ContainSubstring("Not Pushed"))
	}

	Context("odo supported images deployment", func() {
		It("Should be able to verify the openjdk18-openshift image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("redhat-openjdk-18", "openjdk18-openshift:latest"), "java:8", project)
			verifySupportedImage(filepath.Join("redhat-openjdk-18", "openjdk18-openshift:latest"), "openjdk", "java:8", project, appName)
		})

		It("Should be able to verify the openjdk-11-rhel7 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("openjdk", "openjdk-11-rhel7:latest"), "java:8", project)
			verifySupportedImage(filepath.Join("openjdk", "openjdk-11-rhel7:latest"), "openjdk", "java:8", project, appName)
		})

		It("Should be able to verify the nodejs-8-rhel7 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("rhscl", "nodejs-8-rhel7:latest"), "nodejs:latest", project)
			verifySupportedImage(filepath.Join("rhscl", "nodejs-8-rhel7:latest"), "nodejs", "nodejs:latest", project, appName)
		})

		It("Should be able to verify the nodejs-8 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("rhoar-nodejs", "nodejs-8:latest"), "nodejs:latest", project)
			verifySupportedImage(filepath.Join("rhoar-nodejs", "nodejs-8:latest"), "nodejs", "nodejs:latest", project, appName)
		})

		It("Should be able to verify the nodejs-10 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("rhoar-nodejs", "nodejs-10:latest"), "nodejs:latest", project)
			verifySupportedImage(filepath.Join("rhoar-nodejs", "nodejs-10:latest"), "nodejs", "nodejs:latest", project, appName)
		})

		It("Should be able to verify the centos7-s2i-nodejs image", func() {
			oc.ImportImageFromRegistry("docker.io", filepath.Join("bucharestgold", "centos7-s2i-nodejs"), "nodejs:latest", project)
			verifySupportedImage(filepath.Join("bucharestgold", "centos7-s2i-nodejs"), "nodejs", "nodejs:latest", project, appName)
		})

		It("Should be able to verify the centos7-s2i-nodejs:10.x image", func() {
			oc.ImportImageFromRegistry("docker.io", filepath.Join("bucharestgold", "centos7-s2i-nodejs:10.x"), "nodejs:latest", project)
			verifySupportedImage(filepath.Join("bucharestgold", "centos7-s2i-nodejs:10.x"), "nodejs", "nodejs:latest", project, appName)
		})

		It("Should be able to verify the nodejs-8-centos7 image", func() {
			oc.ImportImageFromRegistry("docker.io", filepath.Join("centos", "nodejs-8-centos7:latest"), "nodejs:latest", project)
			verifySupportedImage(filepath.Join("centos", "nodejs-8-centos7:latest"), "nodejs", "nodejs:latest", project, appName)
		})
	})
})
