package integration

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests", func() {

	//new clean project and context for each test
	var project string
	var context, context1, context2 string
	var originalDir string
	var oc helper.OcRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		oc = helper.NewOcRunner("oc")
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for link command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "link", "-h")
			Expect(appHelp).To(ContainSubstring("Link component to a service or component"))
		})
	})

	Context("when running help for unlink command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "unlink", "-h")
			Expect(appHelp).To(ContainSubstring("Unlink component or service from a component"))
		})
	})

	Context("When link between components using wrong port", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			stdErr := helper.CmdShouldFail("odo", "link", "backend", "--context", context1, "--port", "1234")
			Expect(stdErr).To(ContainSubstring("Unable to properly link to component backend using port 1234"))
		})
	})

	Context("When handling link/unlink between components", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("should link the frontend application to the backend and then unlink successfully", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context1)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			frontendURL := helper.DetermineRouteURL(context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "url", "create", "--context", context2)
			helper.CmdShouldPass("odo", "push", "--context", context2)

			helper.CmdShouldPass("odo", "link", "backend", "--context", context1)

			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app", project)
			Expect(envFromOutput).To(ContainSubstring("backend"))

			dcName := oc.GetDcName("frontend", project)
			// wait for DeploymentConfig rollout to finish, so we can check if application is successfully running
			oc.WaitForDCRollout(dcName, project, 20*time.Second)
			helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

			outputErr := helper.CmdShouldFail("odo", "link", "backend", "--context", context1)
			Expect(outputErr).To(ContainSubstring("been linked"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--context", context1)
		})
	})

	Context("When link backend between component and service", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
			originalDir = helper.Getwd()
		})
		JustAfterEach(func() {
			helper.Chdir(originalDir)
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("should link backend to service successfully", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			helper.CmdShouldPass("odo", "link", "backend", "--context", context1) // context1 is the frontend
			// Switching to context2 dir because --context flag is not supported with service command
			helper.Chdir(context2)
			helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

			ocArgs := []string{"get", "serviceinstance", "-n", project, "-o", "name"}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			helper.CmdShouldPass("odo", "link", "mysql-persistent", "--wait-for-target", "--component", "backend", "--project", project)
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("backend", "app", project)
			Expect(envFromOutput).To(ContainSubstring("mysql-persistent"))
			outputErr := helper.CmdShouldFail("odo", "link", "mysql-persistent", "--context", context2)
			Expect(outputErr).To(ContainSubstring("been linked"))
		})
	})

	Context("When deleting service and unlink the backend from the frontend", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
			originalDir = helper.Getwd()
		})
		JustAfterEach(func() {
			helper.Chdir(originalDir)
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("should pass", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			helper.CmdShouldPass("odo", "link", "backend", "--context", context1)
			helper.Chdir(context2)
			helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

			ocArgs := []string{"get", "serviceinstance", "-n", project, "-o", "name"}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			helper.CmdShouldPass("odo", "service", "delete", "mysql-persistent", "-f")
			// ensure that the backend no longer has an envFrom value
			backendEnvFromOutput := oc.GetEnvFromEntry("backend", "app", project)
			Expect(backendEnvFromOutput).To(Equal("''"))
			// ensure that the frontend envFrom was not changed
			frontEndEnvFromOutput := oc.GetEnvFromEntry("frontend", "app", project)
			Expect(frontEndEnvFromOutput).To(ContainSubstring("backend"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--component", "frontend", "--project", project)
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app", project)
			Expect(envFromOutput).To(Equal("''"))
		})
	})

	Context("When linking or unlinking a service or component", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
			originalDir = helper.Getwd()
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})

		It("should print the environment variables being linked/unlinked", func() {
			helper.CopyExample(filepath.Join("source", "python"), context1)
			helper.CmdShouldPass("odo", "create", "python", "component1", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "nodejs"), context2)
			helper.CmdShouldPass("odo", "create", "nodejs", "component2", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context2)

			// tests for linking a component to a component
			stdOut := helper.CmdShouldPass("odo", "link", "component2", "--context", context1)
			Expect(stdOut).To(ContainSubstring("The below secret environment variables were added"))
			Expect(stdOut).To(ContainSubstring("COMPONENT_COMPONENT2_HOST"))
			Expect(stdOut).To(ContainSubstring("COMPONENT_COMPONENT2_PORT"))

			// tests for unlinking a component from a component
			stdOut = helper.CmdShouldPass("odo", "unlink", "component2", "--context", context1)
			Expect(stdOut).To(ContainSubstring("The below secret environment variables were removed"))
			Expect(stdOut).To(ContainSubstring("COMPONENT_COMPONENT2_HOST"))
			Expect(stdOut).To(ContainSubstring("COMPONENT_COMPONENT2_PORT"))

			// first create a service
			helper.CmdShouldPass("odo", "service", "create", "-w", "dh-postgresql-apb", "--project", project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6")
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			// tests for linking a service to a component
			stdOut = helper.CmdShouldPass("odo", "link", "dh-postgresql-apb", "--context", context1)
			Expect(stdOut).To(ContainSubstring("The below secret environment variables were added"))
			Expect(stdOut).To(ContainSubstring("DB_PORT"))
			Expect(stdOut).To(ContainSubstring("DB_HOST"))

			// tests for unlinking a service to a component
			stdOut = helper.CmdShouldPass("odo", "unlink", "dh-postgresql-apb", "--context", context1)
			Expect(stdOut).To(ContainSubstring("The below secret environment variables were removed"))
			Expect(stdOut).To(ContainSubstring("DB_PORT"))
			Expect(stdOut).To(ContainSubstring("DB_HOST"))
		})
	})
})
