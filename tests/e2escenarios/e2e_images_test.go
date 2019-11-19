package e2escenarios

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo supported images e2e tests", func() {
	//new clean project and context for each test
	var project string
	var context string

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

	verifySupportedImage := func(image, srcType, cmpType, project string) {

		// create the component
		helper.CopyExample(filepath.Join("source", srcType), context)
		helper.CmdShouldPass("odo", "create", cmpType, srcType+"-app", "--project", project, "--context", context)

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
			helper.ReplaceString(filepath.Join(context+"/src/main/java/MessageProducer.java"), "Hello", "Hello Java UPDATED")
			helper.CmdShouldPass("odo", "push", "--context", context)
			helper.HttpWaitFor(routeURL, "Hello Java UPDATED", 90, 1)
		} else {
			helper.ReplaceString(filepath.Join(context+"/server.js"), "Hello", "Hello nodejs UPDATED")
			helper.CmdShouldPass("odo", "push", "--context", context)
			helper.HttpWaitFor(routeURL, "Hello nodejs UPDATED", 90, 1)
		}

		//delete the component and validate
		helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		cmpLst := helper.CmdShouldPass("odo", "list", "--context", context)
		Expect(cmpLst).To(ContainSubstring("Not Pushed"))
	}

	Context("odo supported images deployment", func() {
		It("Should be able to verify the openjdk18-openshift image", func() {
			oc.ImportSupportedImage("redhat-openjdk-18/openjdk18-openshift:latest", "java:8", project)
			verifySupportedImage("redhat-openjdk-18/openjdk18-openshift:latest", "openjdk", "java:8", project)
		})

		// It("Should be able to verify the openjdk-11-rhel8 image", func() {
		// 	oc.ImportSupportedImage("openjdk/openjdk-11-rhel8:latest", "java:8", project)
		// 	verifySupportedImage("openjdk/openjdk-11-rhel8:latest", "openjdk", "java:8", project)
		// })

		It("Should be able to verify the openjdk-11-rhel7 image", func() {
			oc.ImportSupportedImage("openjdk/openjdk-11-rhel7:latest", "java:8", project)
			verifySupportedImage("openjdk/openjdk-11-rhel7:latest", "openjdk", "java:8", project)
		})

		It("Should be able to verify the nodejs-8-rhel7 image", func() {
			oc.ImportSupportedImage("rhscl/nodejs-8-rhel7:latest", "nodejs:8", project)
			verifySupportedImage("rhscl/nodejs-8-rhel7:latest", "nodejs", "nodejs:8", project)
		})

		It("Should be able to verify the nodejs-8 image", func() {
			oc.ImportSupportedImage("rhoar-nodejs/nodejs-8:latest", "nodejs:8", project)
			verifySupportedImage("rhoar-nodejs/nodejs-8:latest", "nodejs", "nodejs:8", project)
		})

		It("Should be able to verify the nodejs-10 image", func() {
			oc.ImportSupportedImage("rhoar-nodejs/nodejs-10:latest", "nodejs:8", project)
			verifySupportedImage("rhoar-nodejs/nodejs-10:latest", "nodejs", "nodejs:8", project)
		})

		// It("Should be able to verify the centos7-s2i-nodejs image", func() {
		// 	oc.ImportSupportedImage("bucharestgold/centos7-s2i-nodejs", "nodejs:8", project)
		// 	verifySupportedImage("bucharestgold/centos7-s2i-nodejs", "nodejs", "nodejs:8", project)
		// })

		// It("Should be able to verify the centos7-s2i-nodejs:10.x image", func() {
		// 	oc.ImportSupportedImage("bucharestgold/centos7-s2i-nodejs:10.x", "nodejs:8", project)
		// 	verifySupportedImage("bucharestgold/centos7-s2i-nodejs:10.x", "nodejs", "nodejs:8", project)
		// })

		// It("Should be able to verify the nodejs-8-centos7 image", func() {
		// 	oc.ImportSupportedImage("centos/nodejs-8-centos7:latest", "nodejs:8", project)
		// 	verifySupportedImage("centos/nodejs-8-centos7:latest", "nodejs", "nodejs:8", project)
		// })
	})
})
