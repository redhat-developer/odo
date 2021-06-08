package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests", func() {
	//new clean context for each test
	/*
		Uncomment when we uncomment the test specs
		var context1, context2 string
		var oc helper.OcRunner
	*/
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		// oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
		//context1 = helper.CreateNewContext()
		//context2 = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		//helper.DeleteDir(context1)
		//helper.DeleteDir(context2)
		helper.CommonAfterEach(commonVar)
	})

	Context("when running help for link and unlink command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "link", "-h").ShouldPass().Out()
			Expect(appHelp).To(ContainSubstring("Link component to a service"))
			appHelp = helper.Cmd("odo", "unlink", "-h").ShouldPass().Out()
			Expect(appHelp).To(ContainSubstring("Unlink component or service from a component"))
		})
	})

	/*
		Context("When link between components using wrong port", func() {
			It("should fail", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context1)
				helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "python"), context2)
				helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context2)
				stdErr := helper.CmdShouldFail("odo", "link", "backend", "--context", context1, "--port", "1234")
				Expect(stdErr).To(ContainSubstring("Unable to properly link to component backend using port 1234"))
			})
		})

		Context("When handling link/unlink between components", func() {
			It("should link the frontend application to the backend and then unlink successfully", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context1)
				helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context1)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				frontendURL := helper.DetermineRouteURL(context1)
				helper.CopyExample(filepath.Join("source", "python"), context2)
				helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "url", "create", "--context", context2)
				helper.CmdShouldPass("odo", "push", "--context", context2)

				helper.CmdShouldPass("odo", "link", "backend", "--context", context1)

				// ensure that the proper envFrom entry was created
				envFromOutput := oc.GetEnvFromEntry("frontend", "app", commonVar.Project)
				Expect(envFromOutput).To(ContainSubstring("backend"))

				dcName := oc.GetDcName("frontend", commonVar.Project)
				// wait for DeploymentConfig rollout to finish, so we can check if application is successfully running
				oc.WaitForDCRollout(dcName, commonVar.Project, 20*time.Second)
				helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

				outputErr := helper.CmdShouldFail("odo", "link", "backend", "--context", context1)
				Expect(outputErr).To(ContainSubstring("been linked"))
				helper.CmdShouldPass("odo", "unlink", "backend", "--context", context1)
			})
		})

		Context("When link backend between component and service", func() {
			It("should link backend to service successfully", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context1)
				helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "python"), context2)
				helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context2)
				helper.CmdShouldPass("odo", "link", "backend", "--context", context1) // context1 is the frontend
				// Switching to context2 dir because --context flag is not supported with service command
				helper.Chdir(context2)
				helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

				ocArgs := []string{"get", "serviceinstance", "-n", commonVar.Project, "-o", "name"}
				helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
					return strings.Contains(output, "mysql-persistent")
				})
				helper.CmdShouldPass("odo", "link", "mysql-persistent", "--wait-for-target", "--component", "backend", "--project", commonVar.Project)
				// ensure that the proper envFrom entry was created
				envFromOutput := oc.GetEnvFromEntry("backend", "app", commonVar.Project)
				Expect(envFromOutput).To(ContainSubstring("mysql-persistent"))
				outputErr := helper.CmdShouldFail("odo", "link", "mysql-persistent", "--context", context2)
				Expect(outputErr).To(ContainSubstring("been linked"))
			})
		})

		Context("When deleting service and unlink the backend from the frontend", func() {
			It("should pass", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context1)
				helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "python"), context2)
				helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context2)
				helper.CmdShouldPass("odo", "link", "backend", "--context", context1)
				helper.Chdir(context2)
				helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

				ocArgs := []string{"get", "serviceinstance", "-n", commonVar.Project, "-o", "name"}
				helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
					return strings.Contains(output, "mysql-persistent")
				})
				helper.CmdShouldPass("odo", "service", "delete", "mysql-persistent", "-f")
				// ensure that the backend no longer has an envFrom value
				backendEnvFromOutput := oc.GetEnvFromEntry("backend", "app", commonVar.Project)
				Expect(backendEnvFromOutput).To(Equal("''"))
				// ensure that the frontend envFrom was not changed
				frontEndEnvFromOutput := oc.GetEnvFromEntry("frontend", "app", commonVar.Project)
				Expect(frontEndEnvFromOutput).To(ContainSubstring("backend"))
				helper.CmdShouldPass("odo", "unlink", "backend", "--component", "frontend", "--project", commonVar.Project)
				// ensure that the proper envFrom entry was created
				envFromOutput := oc.GetEnvFromEntry("frontend", "app", commonVar.Project)
				Expect(envFromOutput).To(Equal("''"))
			})
		})

		Context("When linking or unlinking a service or component", func() {
			It("should print the environment variables being linked/unlinked", func() {
				helper.CopyExample(filepath.Join("source", "python"), context1)
				helper.CmdShouldPass("odo", "create", "python", "component1", "--context", context1, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "nodejs"), context2)
				helper.CmdShouldPass("odo", "create", "nodejs", "component2", "--context", context2, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", context2)

				// tests for linking a component to a component
				stdOut := helper.CmdShouldPass("odo", "link", "component2", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were added", "COMPONENT_COMPONENT2_HOST", "COMPONENT_COMPONENT2_PORT"})

				// tests for unlinking a component from a component
				stdOut = helper.CmdShouldPass("odo", "unlink", "component2", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were removed", "COMPONENT_COMPONENT2_HOST", "COMPONENT_COMPONENT2_PORT"})

				// first create a service
				helper.CmdShouldPass("odo", "service", "create", "-w", "dh-postgresql-apb", "--project", commonVar.Project, "--plan", "dev",
					"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
					"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6")
				ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", commonVar.Project}
				helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
					return strings.Contains(output, "dh-postgresql-apb")
				})

				// tests for linking a service to a component
				stdOut = helper.CmdShouldPass("odo", "link", "dh-postgresql-apb", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were added", "DB_PORT", "DB_HOST"})

				// tests for unlinking a service to a component
				stdOut = helper.CmdShouldPass("odo", "unlink", "dh-postgresql-apb", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were removed", "DB_PORT", "DB_HOST"})
			})
		})

	*/
})
