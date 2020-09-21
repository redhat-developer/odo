package integration

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
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
			appHelp := helper.CmdShouldPass("odo", "app", "-h")
			Expect(appHelp).To(ContainSubstring("Performs application operations related to your project."))
		})
	})

	Context("when running app delete, describe and list command on fresh cluster", func() {
		It("should error out display the help", func() {
			appList := helper.CmdShouldPass("odo", "app", "list", "--project", commonVar.Project)
			Expect(appList).To(ContainSubstring("There are no applications deployed"))
			actual := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", commonVar.Project)
			desired := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[]}`
			Expect(desired).Should(MatchJSON(actual))

			appDelete := helper.CmdShouldFail("odo", "app", "delete", "test", "--project", commonVar.Project, "-f")
			Expect(appDelete).To(ContainSubstring("test app does not exists"))
			appDescribe := helper.CmdShouldFail("odo", "app", "describe", "test", "--project", commonVar.Project)
			Expect(appDescribe).To(ContainSubstring("test app does not exists"))
		})
	})

	Context("when running app command without app parameter in directory that contains .odo config directory", func() {
		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// changing directory to the context directory
			helper.Chdir(commonVar.Context)

			appListOutput := helper.CmdShouldPass("odo", "app", "list")
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.CmdShouldPass("odo", "list", "-o", "json")

			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"s2iComponents":[{"kind":"Component","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"nodejs","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"nodejs","sourceType":"local","env":[{"name":"DEBUG_PORT","value":"5858"}]},"status":{"state":"Pushed"}}],"devfileComponents":[]}`, commonVar.Project)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			helper.CmdShouldPass("odo", "app", "describe")
			desiredDesAppJSON := fmt.Sprintf(`{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"app","namespace":"%s","creationTimestamp":null},"spec":{"components": ["nodejs"]}}`, commonVar.Project)
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", "app", "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			helper.CmdShouldPass("odo", "app", "delete", "-f")
		})
	})

	Context("when running app command without app parameter in directory that doesn't contain .odo config directory", func() {
		It("should fail without app parameter (except the list command)", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// list should pass as the project exists
			appListOutput := helper.CmdShouldPass("odo", "app", "list", "--project", commonVar.Project)
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.CmdShouldFail("odo", "app", "describe")
			helper.CmdShouldFail("odo", "app", "delete", "-f")
		})
	})

	Context("when running app command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			appListOutput := helper.CmdShouldPass("odo", "app", "list", "--project", commonVar.Project)
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", commonVar.Project)
			//desiredCompListJSON := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[]}`
			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"app","namespace":"%s","creationTimestamp":null},"spec":{"components":["%s"]}}]}`, commonVar.Project, cmpName)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			helper.CmdShouldPass("odo", "app", "describe", appName, "--project", commonVar.Project)
			desiredDesAppJSON := fmt.Sprintf(`{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"spec":{"components":["%s"]}}`, appName, commonVar.Project, cmpName)
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", appName, "--project", commonVar.Project, "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			helper.CmdShouldPass("odo", "app", "delete", appName, "--project", commonVar.Project, "-f")
		})

	})

	Context("When running app describe with storage added in component in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute describe", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-one", "--context", commonVar.Context, "--path", mountPath, "--size", size)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "app", "describe", appName, "--project", commonVar.Project)

		})

	})
})
