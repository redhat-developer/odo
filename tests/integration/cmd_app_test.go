package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo app command tests", func() {
	var commonVar helper.CommonVar
	appName := "app"
	cmpName := "nodejs"
	mountPath := "/data"
	size := "1Gi"

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("when running help for app command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "app", "-h").ShouldPass().Out()
			// Trimmed the end of the message string to make it compatible across clusters
			Expect(appHelp).To(ContainSubstring("Performs application operations related to"))
		})
	})

	Context("when running app delete, describe and list command on fresh cluster", func() {
		It("should error out display the help", func() {
			appList := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appList).To(ContainSubstring("There are no applications deployed"))
			actual := helper.Cmd("odo", "app", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			values := gjson.GetMany(actual, "kind", "metadata", "items")
			expected := []string{"List", "{}", "[]"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))

			appDelete := helper.Cmd("odo", "app", "delete", "test", "--project", commonVar.Project, "-f").ShouldFail().Err()
			Expect(appDelete).To(ContainSubstring("test app does not exists"))
			appDescribe := helper.Cmd("odo", "app", "describe", "test", "--project", commonVar.Project).ShouldFail().Err()
			Expect(appDescribe).To(ContainSubstring("test app does not exists"))
		})
	})

	Context("when running app command without app parameter in directory that contains .odo config directory", func() {
		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// changing directory to the context directory
			helper.Chdir(commonVar.Context)

			appListOutput := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.Cmd("odo", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			valuesL := gjson.GetMany(actualCompListJSON, "kind", "devfileComponents.0.metadata.name", "devfileComponents.0.metadata.namespace")
			expectedL := []string{"List", "nodejs", commonVar.Project}
			Expect(helper.GjsonMatcher(valuesL, expectedL)).To(Equal(true))
			helper.Cmd("odo", "app", "describe", "--project", commonVar.Project).ShouldPass()
			actualDesAppJSON := helper.Cmd("odo", "app", "describe", "app", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			valuesDes := gjson.GetMany(actualDesAppJSON, "kind", "metadata.name", "metadata.namespace")
			expectedDes := []string{"Application", "app", commonVar.Project}
			Expect(helper.GjsonMatcher(valuesDes, expectedDes)).To(Equal(true))

			helper.Cmd("odo", "app", "delete", "-f", "--project", commonVar.Project).ShouldPass()
		})
	})

	Context("when running app command without app parameter in directory that doesn't contain .odo config directory", func() {
		It("should fail without app parameter (except the list command)", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// list should pass as the project exists
			appListOutput := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.Cmd("odo", "app", "describe", "--project", commonVar.Project).ShouldFail()
			helper.Cmd("odo", "app", "delete", "-f", "--project", commonVar.Project).ShouldFail()
		})
	})

	Context("when running app command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			appListOutput := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.Cmd("odo", "app", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			valuesList := gjson.GetMany(actualCompListJSON, "kind", "items.#.metadata.name", "items.#.metadata.namespace")
			expectedList := []string{"List", "app", commonVar.Project}
			Expect(helper.GjsonMatcher(valuesList, expectedList)).To(Equal(true))

			helper.Cmd("odo", "app", "describe", appName, "--project", commonVar.Project).ShouldPass()
			actualDesAppJSON := helper.Cmd("odo", "app", "describe", appName, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
			valuesDes := gjson.GetMany(actualDesAppJSON, "kind", "metadata.name", "metadata.namespace")
			expectedDes := []string{"Application", appName, commonVar.Project}
			Expect(helper.GjsonMatcher(valuesDes, expectedDes)).To(Equal(true))

			helper.Cmd("odo", "app", "delete", appName, "--project", commonVar.Project, "-f").ShouldPass()
		})

	})

	Context("When running app describe with storage added in component in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute describe", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", "storage-one", "--context", commonVar.Context, "--path", mountPath, "--size", size).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "app", "describe", appName, "--project", commonVar.Project).ShouldPass()

		})

	})
})
