package e2escenarios

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo supported images e2e tests", func() {

	appName := "app"

	var oc helper.OcRunner
	var globals helper.Globals

	BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		globals = helper.CommonBeforeEach()
	})

	AfterEach(func() {
		helper.CommonAfterEeach(globals)
	})

	// verifySupportedImage takes arguments supported images, source type, image type, namespace and application name.
	// Also verify the flow of odo commands with respect to supported images only.
	verifySupportedImage := func(image, srcType, cmpType, project, appName, context string) {

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

		watchFlag := ""
		odoV1Watch := utils.OdoV1Watch{
			SrcType:  srcType,
			RouteURL: routeURL,
			AppName:  appName,
		}
		// odo watch and validate
		utils.OdoWatch(odoV1Watch, utils.OdoV2Watch{}, project, context, watchFlag, oc, "kube")

		// delete the component and validate
		helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		cmpLst := helper.CmdShouldPass("odo", "list", "--context", context)
		Expect(cmpLst).To(ContainSubstring("Not Pushed"))
	}

	Context("odo supported images deployment", func() {
		It("Should be able to verify the openjdk18-openshift image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("redhat-openjdk-18", "openjdk18-openshift:latest"), "java:8", globals.Project)
			verifySupportedImage(filepath.Join("redhat-openjdk-18", "openjdk18-openshift:latest"), "openjdk", "java:8", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the openjdk-11-rhel7 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("openjdk", "openjdk-11-rhel7:latest"), "java:8", globals.Project)
			verifySupportedImage(filepath.Join("openjdk", "openjdk-11-rhel7:latest"), "openjdk", "java:8", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the nodejs-8-rhel7 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("rhscl", "nodejs-8-rhel7:latest"), "nodejs:latest", globals.Project)
			verifySupportedImage(filepath.Join("rhscl", "nodejs-8-rhel7:latest"), "nodejs", "nodejs:latest", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the nodejs-8 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("rhoar-nodejs", "nodejs-8:latest"), "nodejs:latest", globals.Project)
			verifySupportedImage(filepath.Join("rhoar-nodejs", "nodejs-8:latest"), "nodejs", "nodejs:latest", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the nodejs-10 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", filepath.Join("rhoar-nodejs", "nodejs-10:latest"), "nodejs:latest", globals.Project)
			verifySupportedImage(filepath.Join("rhoar-nodejs", "nodejs-10:latest"), "nodejs", "nodejs:latest", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the centos7-s2i-nodejs image", func() {
			oc.ImportImageFromRegistry("docker.io", filepath.Join("bucharestgold", "centos7-s2i-nodejs"), "nodejs:latest", globals.Project)
			verifySupportedImage(filepath.Join("bucharestgold", "centos7-s2i-nodejs"), "nodejs", "nodejs:latest", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the centos7-s2i-nodejs:10.x image", func() {
			oc.ImportImageFromRegistry("docker.io", filepath.Join("bucharestgold", "centos7-s2i-nodejs:10.x"), "nodejs:latest", globals.Project)
			verifySupportedImage(filepath.Join("bucharestgold", "centos7-s2i-nodejs:10.x"), "nodejs", "nodejs:latest", globals.Project, appName, globals.Context)
		})

		It("Should be able to verify the nodejs-8-centos7 image", func() {
			oc.ImportImageFromRegistry("docker.io", filepath.Join("centos", "nodejs-8-centos7:latest"), "nodejs:latest", globals.Project)
			verifySupportedImage(filepath.Join("centos", "nodejs-8-centos7:latest"), "nodejs", "nodejs:latest", globals.Project, appName, globals.Context)
		})
	})
})
