package integration

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo app command tests", func() {
	var globals helper.Globals

	appName := "app"
	cmpName := "nodejs"
	mountPath := "/data"
	size := "1Gi"

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEach()

	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	Context("when running help for app command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "app", "-h")
			Expect(appHelp).To(ContainSubstring("Performs application operations related to your OpenShift project."))
		})
	})

	Context("when running app delete, describe and list command on fresh cluster", func() {
		It("should error out display the help", func() {
			appList := helper.CmdShouldPass("odo", "app", "list", "--project", globals.Project)
			Expect(appList).To(ContainSubstring("There are no applications deployed"))
			actual := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", globals.Project)
			desired := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[]}`
			Expect(desired).Should(MatchJSON(actual))

			appDelete := helper.CmdShouldFail("odo", "app", "delete", "test", "--project", globals.Project, "-f")
			Expect(appDelete).To(ContainSubstring("test app does not exists"))
			appDescribe := helper.CmdShouldPass("odo", "app", "describe", "test", "--project", globals.Project)
			Expect(appDescribe).To(ContainSubstring("Application test has no components or services deployed."))
		})
	})

	Context("when running app command without app parameter in directory that contains .odo config directory", func() {

		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", globals.Project, "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			// changing directory to the context directory
			helper.Chdir(globals.Context)
			appListOutput := helper.CmdShouldPass("odo", "app", "list")
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.CmdShouldPass("odo", "list", "-o", "json")

			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null, "namespace":"%s"},"spec":{"type":"nodejs","app":"app","sourceType": "local","env":[{"name":"DEBUG_PORT","value":"5858"}]},"status":{"state":"Pushed"}}]}`, globals.Project)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			helper.CmdShouldPass("odo", "app", "describe")
			desiredDesAppJSON := fmt.Sprintf(`{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"myapp","namespace":"%s","creationTimestamp":null},"spec":{}}`, globals.Project)
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", "myapp", "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			helper.CmdShouldPass("odo", "app", "delete", "-f")
		})
	})

	Context("when running app command without app parameter in directory that doesn't contain .odo config directory", func() {
		It("should fail without app parameter (except the list command)", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", globals.Project, "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			// list should pass as the project exists
			appListOutput := helper.CmdShouldPass("odo", "app", "list", "--project", globals.Project)
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.CmdShouldFail("odo", "app", "describe")
			helper.CmdShouldFail("odo", "app", "delete", "-f")
		})
	})

	Context("when running app command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", globals.Project, "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			appListOutput := helper.CmdShouldPass("odo", "app", "list", "--project", globals.Project)
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", globals.Project)
			//desiredCompListJSON := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[]}`
			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"app","namespace":"%s","creationTimestamp":null},"spec":{"components":["%s"]}}]}`, globals.Project, cmpName)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			helper.CmdShouldPass("odo", "app", "describe", appName, "--project", globals.Project)
			desiredDesAppJSON := fmt.Sprintf(`{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"spec":{"components":["%s"]}}`, appName, globals.Project, cmpName)
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", appName, "--project", globals.Project, "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			helper.CmdShouldPass("odo", "app", "delete", appName, "--project", globals.Project, "-f")
		})

	})

	Context("When running app describe with storage added in component in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute describe", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", globals.Project, "--context", globals.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-one", "--context", globals.Context, "--path", mountPath, "--size", size)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			helper.CmdShouldPass("odo", "app", "describe", appName, "--project", globals.Project)

		})

	})
})
