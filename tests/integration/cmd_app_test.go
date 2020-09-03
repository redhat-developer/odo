package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo app command tests", func() {
	var project string
	var context string
	var originalDir string

	appName := "app"
	cmpName := "nodejs"
	mountPath := "/data"
	size := "1Gi"

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
		originalDir = helper.Getwd()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for app command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "app", "-h")
			Expect(appHelp).To(ContainSubstring("Performs application operations related to your OpenShift project."))
		})
	})

	Context("when running app delete, describe and list command on fresh cluster", func() {
		It("should error out display the help", func() {
			appList := helper.CmdShouldPass("odo", "app", "list", "--project", project)
			Expect(appList).To(ContainSubstring("There are no applications deployed"))
			actual := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", project)
			desired := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[]}`
			Expect(desired).Should(MatchJSON(actual))

			appDelete := helper.CmdShouldFail("odo", "app", "delete", "test", "--project", project, "-f")
			Expect(appDelete).To(ContainSubstring("test app does not exists"))
			appDescribe := helper.CmdShouldFail("odo", "app", "describe", "test", "--project", project)
			Expect(appDescribe).To(ContainSubstring("test app does not exists"))
		})
	})

	Context("when running app command without app parameter in directory that contains .odo config directory", func() {

		JustAfterEach(func() {
			helper.Chdir(originalDir)
		})

		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// changing directory to the context directory
			helper.Chdir(context)
			appListOutput := helper.CmdShouldPass("odo", "app", "list")
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.CmdShouldPass("odo", "list", "-o", "json")

			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null, "namespace":"%s"},"spec":{"type":"nodejs","app":"app","sourceType": "local","env":[{"name":"DEBUG_PORT","value":"5858"}]},"status":{"state":"Pushed"}}]}`, project)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			helper.CmdShouldPass("odo", "app", "describe")
			desiredDesAppJSON := fmt.Sprintf(`{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"app","namespace":"%s","creationTimestamp":null},"spec":{"components": ["nodejs"]}}`, project)
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", "app", "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			helper.CmdShouldPass("odo", "app", "delete", "-f")
		})
	})

	Context("when running app command without app parameter in directory that doesn't contain .odo config directory", func() {
		It("should fail without app parameter (except the list command)", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// list should pass as the project exists
			appListOutput := helper.CmdShouldPass("odo", "app", "list", "--project", project)
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.CmdShouldFail("odo", "app", "describe")
			helper.CmdShouldFail("odo", "app", "delete", "-f")
		})
	})

	Context("when running app command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list, describe and delete along with machine readable output", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			appListOutput := helper.CmdShouldPass("odo", "app", "list", "--project", project)
			Expect(appListOutput).To(ContainSubstring(appName))
			actualCompListJSON := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", project)
			//desiredCompListJSON := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[]}`
			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"app","namespace":"%s","creationTimestamp":null},"spec":{"components":["%s"]}}]}`, project, cmpName)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			helper.CmdShouldPass("odo", "app", "describe", appName, "--project", project)
			desiredDesAppJSON := fmt.Sprintf(`{"kind":"Application","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"spec":{"components":["%s"]}}`, appName, project, cmpName)
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", appName, "--project", project, "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			helper.CmdShouldPass("odo", "app", "delete", appName, "--project", project, "-f")
		})

	})

	Context("When running app describe with storage added in component in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute describe", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-one", "--context", context, "--path", mountPath, "--size", size)
			helper.CmdShouldPass("odo", "push", "--context", context)
			helper.CmdShouldPass("odo", "app", "describe", appName, "--project", project)

		})

	})
})
