package devfile

import (
	"github.com/tidwall/gjson"
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

	It("should display the help for app command", func() {
		appHelp := helper.Cmd("odo", "app", "-h").ShouldPass().Out()
		// Trimmed the end of the message string to make it compatible across clusters
		Expect(appHelp).To(ContainSubstring("Performs application operations related to"))
	})

	Context("on a fresh new project", func() {

		BeforeEach(func() {
			appList := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appList).To(ContainSubstring("There are no applications deployed"))
			actual := helper.Cmd("odo", "app", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			values := gjson.GetMany(actual, "kind", "metadata", "items")
			expected := []string{"List", "{}", "[]"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})

		It("should fail deleting non existing app", func() {
			appDelete := helper.Cmd("odo", "app", "delete", "test", "--project", commonVar.Project, "-f").ShouldFail().Err()
			Expect(appDelete).To(ContainSubstring("test app does not exists"))
		})

		It("should fail describing non existing app", func() {
			appDescribe := helper.Cmd("odo", "app", "describe", "test", "--project", commonVar.Project).ShouldFail().Err()
			Expect(appDescribe).To(ContainSubstring("test app does not exists"))
		})

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
			createComponent(component0, app0, namespace, context0)

			// create second component in the second app
			context1 = helper.CreateNewContext()
			component1 = helper.RandString(4)
			createComponent(component1, app1, namespace, context1)
		})

		AfterEach(func() {
			helper.Cmd("odo", "delete", "-f", "--context", context0).ShouldPass()
			helper.Cmd("odo", "delete", "-f", "--context", context1).ShouldPass()

			helper.DeleteDir(context0)
			helper.DeleteDir(context1)
		})

		When("the user creates and pushes a third devfile component", func() {

			var context00 string
			var info testInfo

			BeforeEach(func() {
				context00 = helper.CreateNewContext()
				component00 := helper.RandString(4)
				storage00 := helper.RandString(4)
				url00 := helper.RandString(4)

				helper.Cmd("odo", "create", "nodejs", "--project", namespace, component00, "--context", context00, "--app", app0).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context00)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"), filepath.Join(context00, "devfile.yaml"))
				helper.Cmd("odo", "storage", "create", storage00, "--path", "/data", "--size", "1Gi", "--context", context00).ShouldPass()
				helper.Cmd("odo", "url", "create", url00, "--port", "3000", "--context", context00, "--host", "com", "--ingress").ShouldPass()
				helper.Cmd("odo", "push", "--context", context00).ShouldPass()

				info = testInfo{
					app0:      app0,
					app1:      app1,
					comp0:     component0,
					comp00:    component00,
					url00:     url00,
					storage00: storage00,
					namespace: namespace,
				}
			})

			AfterEach(func() {
				helper.Cmd("odo", "delete", "-f", "--context", context00).ShouldPass()
				helper.DeleteDir(context00)
			})

			It("should list, describe and delete the app properly with json output", func() {
				runner(info)
			})
		})
	})

	When("the user creates two components with the same name in different apps", func() {
		var context0 string
		var context1 string
		var component string

		var app0 string
		var app1 string

		BeforeEach(func() {
			app0 = helper.RandString(4)
			app1 = helper.RandString(4)

			component = helper.RandString(4)

			// create first component in the first app
			context0 = helper.CreateNewContext()
			createComponent(component, app0, namespace, context0)

			// create second component in the second app
			context1 = helper.CreateNewContext()
			createComponent(component, app1, namespace, context1)
		})

		AfterEach(func() {
			helper.Cmd("odo", "delete", "-f", "--context", context0).ShouldPass()
			helper.Cmd("odo", "delete", "-f", "--context", context1).ShouldPass()

			helper.DeleteDir(context0)
			helper.DeleteDir(context1)
		})

		It("should list the components", func() {
			output := helper.Cmd("odo", "list", "--app", app0).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{app0, component})
			output = helper.Cmd("odo", "list", "--app", app1).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{app1, component})
		})
	})
})

// createComponent creates with the given parameters and pushes it
func createComponent(componentName, appName, project, context string) {
	helper.Cmd("odo", "create", "nodejs", "--project", project, componentName, "--context", context, "--app", appName).ShouldPass()
	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
	helper.Cmd("odo", "push", "--context", context).ShouldPass()
}

// testInfo holds the information to run the matchers in the runner function
type testInfo struct {
	// the name of the two apps
	app0 string
	app1 string

	// the name of the components belonging to the first app
	comp0, comp00 string

	// the url and storage belonging to one of the components in the first app
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
