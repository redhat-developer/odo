package devfile

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile app command tests", func() {

	var namespace string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		namespace = commonVar.Project
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("the user creates and pushes two new devfile components in different apps", func() {
		var context0 string
		var context1 string
		var component0 string
		var component1 string

		var app0 string
		var app1 string

		BeforeEach(func() {
			app0 = helper.RandString(4)
			app1 = helper.RandString(4)

			// create first component in the first app
			context0 = helper.CreateNewContext()
			component0 = helper.RandString(4)
			helper.Cmd("odo", "create", "nodejs", "--project", namespace, component0, "--context", context0, "--app", app0).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context0)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context0, "devfile.yaml"))
			helper.Cmd("odo", "push", "--context", context0).ShouldPass()

			// create second component in the second app
			context1 = helper.CreateNewContext()
			component1 = helper.RandString(4)
			helper.Cmd("odo", "create", "nodejs", "--project", namespace, component1, "--context", context1, "--app", app1).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context1)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context1, "devfile.yaml"))
			helper.Cmd("odo", "push", "--context", context1).ShouldPass()
		})

		When("the user creates and pushes a third s2i component on a openshift cluster", func() {

			var context00 string
			var component00 string

			var storage00 string
			var url00 string

			BeforeEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					Skip("This is a OpenShift specific scenario, skipping")
				}

				context00 = helper.CreateNewContext()
				component00 = helper.RandString(4)
				storage00 = helper.RandString(4)
				url00 = helper.RandString(4)
				helper.CopyExample(filepath.Join("source", "nodejs"), context00)
				helper.Cmd("odo", "component", "create", "--s2i", "nodejs", component00, "--app", app0, "--project", namespace, "--context", context00).ShouldPass()
				helper.Cmd("odo", "storage", "create", storage00, "--path", "/data", "--size", "1Gi", "--context", context00).ShouldPass()
				helper.Cmd("odo", "url", "create", url00, "--port", "8080", "--context", context00).ShouldPass()
				helper.Cmd("odo", "push", "--context", context00).ShouldPass()
			})

			It("should list, describe and delete the app properly with json output", func() {
				runner(testInfo{
					app0:      app0,
					app1:      app1,
					comp0:     component0,
					comp00:    component00,
					url00:     url00,
					storage00: storage00,
					namespace: namespace,
				})
			})

			AfterEach(func() {
				helper.Cmd("odo", "delete", "-f", "--context", context00).ShouldPass()
				helper.DeleteDir(context00)
			})
		})

		When("the user creates and pushes a third devfile component", func() {

			var context00 string
			var component00 string

			var storage00 string
			var url00 string

			BeforeEach(func() {
				context00 = helper.CreateNewContext()
				component00 = helper.RandString(4)
				storage00 = helper.RandString(4)
				url00 = helper.RandString(4)

				helper.Cmd("odo", "create", "nodejs", "--project", namespace, component00, "--context", context00, "--app", app0).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context00)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context00, "devfile.yaml"))
				helper.Cmd("odo", "storage", "create", storage00, "--path", "/data", "--size", "1Gi", "--context", context00).ShouldPass()
				helper.Cmd("odo", "url", "create", url00, "--port", "3000", "--context", context00, "--host", "com", "--ingress").ShouldPass()
				helper.Cmd("odo", "push", "--context", context00).ShouldPass()
			})

			It("should list, describe and delete the app properly with json output", func() {
				runner(testInfo{
					app0:      app0,
					app1:      app1,
					comp0:     component0,
					comp00:    component00,
					url00:     url00,
					storage00: storage00,
					namespace: namespace,
				})
			})

			AfterEach(func() {
				helper.Cmd("odo", "delete", "-f", "--context", context00).ShouldPass()
				helper.DeleteDir(context00)
			})
		})

		AfterEach(func() {
			helper.Cmd("odo", "delete", "-f", "--context", context0).ShouldPass()
			helper.Cmd("odo", "delete", "-f", "--context", context1).ShouldPass()

			helper.DeleteDir(context0)
			helper.DeleteDir(context1)
		})
	})
})

// testInfo holds the information to run the matchers in the runner function
type testInfo struct {
	app0 string
	app1 string

	comp0, comp00    string
	url00, storage00 string

	namespace string
}

func runner(info testInfo) {

	stdOut := helper.Cmd("odo", "app", "list", "--project", info.namespace).ShouldPass().Out()
	helper.MatchAllInOutput(stdOut, []string{info.app0, info.app1})

	// test the json output
	stdOut = helper.Cmd("odo", "app", "list", "--project", info.namespace, "-o", "json").ShouldPass().Out()
	helper.MatchAllInOutput(stdOut, []string{info.app0, info.app1})
	Expect(helper.IsJSON(stdOut)).To(BeTrue())

	stdOut = helper.Cmd("odo", "app", "describe", info.app0, "--project", info.namespace).ShouldPass().Out()
	helper.MatchAllInOutput(stdOut, []string{info.app0, info.comp0, info.comp00, info.storage00, info.url00, "http", "3000"})

	// test the json output
	stdOut = helper.Cmd("odo", "app", "describe", info.app0, "--project", info.namespace, "-o", "json").ShouldPass().Out()
	helper.MatchAllInOutput(stdOut, []string{info.app0, info.comp0, info.comp00})
	Expect(helper.IsJSON(stdOut)).To(BeTrue())

	// delete the app
	stdOut = helper.Cmd("odo", "app", "delete", info.app0, "--project", info.namespace, "-f").ShouldPass().Out()
	helper.MatchAllInOutput(stdOut, []string{info.app0, info.comp0, info.comp00, info.url00, info.storage00})

	// test the list output again
	stdOut = helper.Cmd("odo", "app", "list", "--project", info.namespace).ShouldPass().Out()
	helper.MatchAllInOutput(stdOut, []string{info.app1})
}
