// This test file verifies all the supported container images listed in the
// file https://github.com/openshift/odo-init-image/blob/master/language-scripts/image-mappings.json
package e2escenarios

import (
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo supported images e2e tests", func() {
	var oc helper.OcRunner
	var commonVar helper.CommonVar
	appName := "app"

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		// initialize oc runner
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// verifySupportedImage takes arguments supported images, source type, image type, namespace and application name.
	// Also verify the flow of odo commands with respect to supported images only.
	verifySupportedImage := func(image, srcType, cmpType, project, appName, context string) {

		cmpName := srcType + "-app"
		// create the component
		helper.CopyExample(filepath.Join("source", srcType), commonVar.Context)
		helper.CmdShouldPass("odo", "create", "--s2i", cmpType, cmpName, "--project", project, "--context", context, "--app", appName)

		// push component and validate
		helper.CmdShouldPass("odo", "push", "--context", context)
		cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
		Expect(cmpList).To(ContainSubstring(srcType + "-app"))
		// push again just to confirm it works
		helper.CmdShouldPass("odo", "push", "--context", context)
		// get the url
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

		// odo watch and validate
		utils.OdoWatch(utils.OdoV1Watch{},
			utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing s2i-assemble command", "Executing s2i-run command"},
				FolderToCheck:      "/tmp/projects",
				SrcType:            srcType,
			}, project, context, watchFlag, oc, "kube")

		// delete the component and validate
		helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		cmpLst := helper.CmdShouldPass("odo", "list", "--context", context)
		Expect(cmpLst).To(ContainSubstring("Not Pushed"))
	}

	Context("odo supported images deployment on amd64", func() {
		JustBeforeEach(func() {
			if runtime.GOARCH != "amd64" {
				Skip("Skipping test because these images are not supported.")
			}
		})

		It("Should be able to verify the nodejs-10 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", "rhoar-nodejs/nodejs-10:latest", "nodejs:latest", commonVar.Project)
			verifySupportedImage("rhoar-nodejs/nodejs-10:latest", "nodejs", "nodejs:latest", commonVar.Project, appName, commonVar.Context)
		})
	})

	Context("odo supported images deployment", func() {
		It("Should be able to verify the openjdk18-openshift image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", "redhat-openjdk-18/openjdk18-openshift:latest", "java:8", commonVar.Project)
			verifySupportedImage("redhat-openjdk-18/openjdk18-openshift:latest", "openjdk", "java:8", commonVar.Project, appName, commonVar.Context)
		})

		It("Should be able to verify the nodejs-10-rhel7 image", func() {
			oc.ImportImageFromRegistry("registry.access.redhat.com", "rhscl/nodejs-10-rhel7:latest", "nodejs:latest", commonVar.Project)
			verifySupportedImage("rhscl/nodejs-10-rhel7:latest", "nodejs", "nodejs:latest", commonVar.Project, appName, commonVar.Context)
		})

		It("Should be able to verify the nodejs-10-centos7 image", func() {
			oc.ImportImageFromRegistry("quay.io", "centos7/nodejs-10-centos7:latest", "nodejs:latest", commonVar.Project)
			verifySupportedImage("centos7/nodejs-10-centos7:latest", "nodejs", "nodejs:latest", commonVar.Project, appName, commonVar.Context)
		})

		It("Should be able to verify the nodejs-12-centos7 image", func() {
			oc.ImportImageFromRegistry("quay.io", "centos7/nodejs-12-centos7:latest", "nodejs:latest", commonVar.Project)
			verifySupportedImage("centos7/nodejs-12-centos7:latest", "nodejs", "nodejs:latest", commonVar.Project, appName, commonVar.Context)
		})
	})

	Context("odo supported private registry images deployment", func() {
		JustBeforeEach(func() {
			// Issue for configuring login secret for travis CI https://github.com/openshift/odo/issues/3640
			if os.Getenv("CI") != "openshift" {
				Skip("Skipping it on travis CI, skipping")
			}
		})

		It("Should be able to verify the nodejs-12 image", func() {
			redhatNodejs12UBI8Project := util.GetEnvWithDefault("REDHAT_NODEJS12_UBI8_PROJECT", "nodejs-12")
			oc.ImportImageFromRegistry("registry.redhat.io", "ubi8/nodejs-12:latest", "nodejs:latest", redhatNodejs12UBI8Project)
			verifySupportedImage("ubi8/nodejs-12:latest", "nodejs", "nodejs:latest", redhatNodejs12UBI8Project, appName, commonVar.Context)
		})

		It("Should be able to verify the nodejs-12-rhel7 image", func() {
			redhatNodejs12RHEL7Project := util.GetEnvWithDefault("REDHAT_NODEJS12_RHEL7_PROJECT", "nodejs-12-rhel7")
			oc.ImportImageFromRegistry("registry.redhat.io", "rhscl/nodejs-12-rhel7:latest", "nodejs:latest", redhatNodejs12RHEL7Project)
			verifySupportedImage("rhscl/nodejs-12-rhel7:latest", "nodejs", "nodejs:latest", redhatNodejs12RHEL7Project, appName, commonVar.Context)
		})

		It("Should be able to verify the openjdk-11 image", func() {
			redhatOpenjdk11UBI8Project := util.GetEnvWithDefault("REDHAT_OPENJDK11_UBI8_PROJECT", "openjdk-11")
			oc.ImportImageFromRegistry("registry.redhat.io", "ubi8/openjdk-11:latest", "java:8", redhatOpenjdk11UBI8Project)
			verifySupportedImage("ubi8/openjdk-11:latest", "openjdk", "java:8", redhatOpenjdk11UBI8Project, appName, commonVar.Context)
		})

		It("Should be able to verify the openjdk-11-rhel8 image", func() {
			redhatOpenjdk12RHEL8Project := util.GetEnvWithDefault("REDHAT_OPENJDK11_RHEL8_PROJECT", "openjdk-11-rhel8")
			oc.ImportImageFromRegistry("registry.redhat.io", "openjdk/openjdk-11-rhel8:latest", "java:8", redhatOpenjdk12RHEL8Project)
			verifySupportedImage("openjdk/openjdk-11-rhel8:latest", "openjdk", "java:8", redhatOpenjdk12RHEL8Project, appName, commonVar.Context)
		})

		It("Should be able to verify the nodejs-14 image", func() {
			redhatNodeJS14UBI8Project := util.GetEnvWithDefault("REDHAT_NODEJS14_UBI8_PROJECT", "nodejs-14")
			oc.ImportImageFromRegistry("registry.redhat.io", "ubi8/nodejs-14:latest", "nodejs:latest", redhatNodeJS14UBI8Project)
			verifySupportedImage("ubi8/nodejs-14:latest", "nodejs", "nodejs:latest", redhatNodeJS14UBI8Project, appName, commonVar.Context)
		})
	})

})
