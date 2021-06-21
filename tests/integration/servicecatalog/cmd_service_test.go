package integration

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo service command tests", func() {
	var app, serviceName string
	/*
		Uncomment when we uncomment the test specs
		var context1, context2 string
		var oc helper.OcRunner
	*/
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		//oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
		// context1 = helper.CreateNewContext()
		// context2 = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
		// helper.DeleteDir(context1)
		// helper.DeleteDir(context2)
	})

	Context("when running help for service command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "service", "-h").ShouldPass().Out()
			Expect(appHelp).To(ContainSubstring("Perform service catalog operations"))
		})
	})

	Context("check catalog service search functionality", func() {
		It("check that a service does not exist", func() {
			serviceRandomName := helper.RandString(7)
			output := helper.Cmd("odo", "catalog", "search", "service", serviceRandomName).ShouldFail().Err()
			Expect(output).To(ContainSubstring("no service matched the query: " + serviceRandomName))
		})
	})

	Context("checking machine readable output for service catalog", func() {
		It("should succeed listing catalog components", func() {
			// Since service catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			output := helper.Cmd("odo", "catalog", "list", "services", "-o", "json").ShouldPass().Out()
			Expect(output).To(ContainSubstring("List"))
		})
	})

	Context("checking machine readable output for service catalog", func() {
		It("should succeed listing catalog components", func() {
			// Since service catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			helper.Cmd("odo", "catalog", "list", "services", "-o", "json").ShouldPass()
		})
	})

	Context("check search functionality", func() {

		It("should pass with searching for part of a service name", func() {

			// We just use "sql" as some catalogs only have postgresql-persistent and
			// others dh-postgresql-db. So let's just see if there's "any" postgresql to begin
			// with
			output := helper.Cmd("odo", "catalog", "search", "service", "sql").ShouldPass().Out()
			Expect(output).To(ContainSubstring("postgresql"))
		})

	})

	Context("create service with Env non-interactively", func() {
		JustBeforeEach(func() {
			app = helper.RandString(7)
		})

		It("should be able to create postgresql with env", func() {
			helper.Cmd("odo", "service", "create", "dh-postgresql-apb", "--project", commonVar.Project, "--app", app,
				"--plan", "dev", "-p", "postgresql_user=lukecage", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w").ShouldPass()
			// there is only a single pod in the project
			ocArgs := []string{"describe", "pod", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "lukecage")
			})

			// Delete the service
			helper.Cmd("odo", "service", "delete", "dh-postgresql-apb", "-f", "--app", app, "--project", commonVar.Project).ShouldPass()
		})

		It("should be able to create postgresql with env multiple times", func() {
			helper.Cmd("odo", "service", "create", "dh-postgresql-apb", "--project", commonVar.Project, "--app", app,
				"--plan", "dev", "-p", "postgresql_user=lukecage", "-p", "postgresql_user=testworker", "-p", "postgresql_password=secret",
				"-p", "postgresql_password=universe", "-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w").ShouldPass()
			// there is only a single pod in the project
			ocArgs := []string{"describe", "pod", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "testworker")
			})

			// Delete the service
			helper.Cmd("odo", "service", "delete", "dh-postgresql-apb", "-f", "--app", app, "--project", commonVar.Project).ShouldPass()
		})
	})

	/*
		Context("When creating with a spring boot application", func() {
			JustBeforeEach(func() {
				context = helper.CreateNewContext()
				os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
				project = helper.CreateRandProject()
				originalDir = helper.Getwd()
				helper.Chdir(context)
			})
			JustAfterEach(func() {
				helper.DeleteProject(project)
				helper.Chdir(originalDir)
				helper.DeleteDir(context)
				os.Unsetenv("GLOBALODOCONFIG")
			})
				It("should be able to create postgresql and link it with springboot", func() {
					oc.ImportJavaIS(project)
					helper.CopyExample(filepath.Join("source", "openjdk-sb-postgresql"), context)

					// Local config needs to be present in order to create service https://github.com/openshift/odo/issues/1602
					helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "sb-app", "--project", project)

					// Create a URL
					helper.CmdShouldPass("odo", "url", "create", "--port", "8080")

					// push
					helper.CmdShouldPass("odo", "push")

					// create the postgres service
					helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", project, "--plan", "dev",
						"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
						"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w")

					// link the service
					helper.CmdShouldPass("odo", "link", "dh-postgresql-apb", "--project", project, "-w", "--wait-for-target")
					odoArgs := []string{"service", "list"}
					helper.WaitForCmdOut("odo", odoArgs, 1, true, func(output string) bool {
						return strings.Contains(output, "dh-postgresql-apb") &&
							strings.Contains(output, "ProvisionedAndLinked")
					})

					routeURL := helper.DetermineRouteURL("")

					// Ping said URL
					helper.HttpWaitFor(routeURL, "Spring Boot", 90, 1)

					// Delete the service
					helper.CmdShouldPass("odo", "service", "delete", "dh-postgresql-apb", "-f")

					// Delete the component and the config
					helper.CmdShouldPass("odo", "delete", "sb-app", "-f", "--all")
				})
		})
	*/

	// TODO: auth issue, we need to find a proper way how to test it without requiring cluster admin privileges

	// Context("odo hides a hidden service in service catalog", func() {
	// 	It("not show a hidden service in the catalog", func() {
	// 		runCmdShouldPass("oc apply -f https://github.com/openshift/library/raw/master/official/sso/templates/sso72-https.json -n openshift")
	// 		outputErr := runCmdShouldFail("odo catalog search service sso72-https")
	// 		Expect(outputErr).To(ContainSubstring("No service matched the query: sso72-https"))
	// 	})
	// })

	Context("When working from outside a component dir", func() {
		JustBeforeEach(func() {
			app = helper.RandString(7)
			serviceName = "odo-postgres-service"
			helper.Chdir(commonVar.Context)
		})

		It("should be able to create, list and delete a service using a given value for --context", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "python", "--app", app, "--project", commonVar.Project).ShouldPass()

			// cd to the originalDir to create service using --context
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			helper.Cmd("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", serviceName,
				"--context", commonVar.Context,
			).ShouldPass()

			// now check if listing the service using --context works
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, serviceName)
			})
			stdOut := helper.Cmd("odo", "service", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(stdOut).To(ContainSubstring(serviceName))

			// now check if deleting the service using --context works
			stdOut = helper.Cmd("odo", "service", "delete", "-f", serviceName, "--context", commonVar.Context).ShouldPass().Out()
			Expect(stdOut).To(ContainSubstring(serviceName))
		})

		It("should be able to list services, as well as json list in a given app and project combination", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--app", app, "--project", commonVar.Project).ShouldPass()

			// create a service from within a component directory
			helper.Cmd("odo", "service", "create", "dh-prometheus-apb", "--plan", "ephemeral",
				"--app", app, "--project", commonVar.Project,
			).ShouldPass()
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-prometheus-apb")
			})

			// Listing the services should work as expected from within the component directory.
			// This means, it should not require --app or --project flags
			stdOut := helper.Cmd("odo", "service", "list").ShouldPass().Out()
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// Check json output
			stdOut = helper.Cmd("odo", "service", "list", "-o", "json").ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"dh-prometheus-apb", "List"})

			// cd to a non-component directory and list services
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			stdOut = helper.Cmd("odo", "service", "list", "--app", app, "--project", commonVar.Project).ShouldPass().Out()
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// Check json output
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			stdOut = helper.Cmd("odo", "service", "list", "--app", app, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"dh-prometheus-apb", "List"})
		})
	})

	Context("When working from outside a component dir", func() {
		It("should be able to create, list and delete services without a context and using --app and --project flags instaed", func() {
			app = helper.RandString(7)
			// create the service
			helper.Cmd("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6",
				"--app", app, "--project", commonVar.Project).ShouldPass()

			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			// list the service using app and project flags
			stdOut := helper.Cmd("odo", "service", "list", "--app", app, "--project", commonVar.Project).ShouldPass().Out()
			Expect(stdOut).To(ContainSubstring("dh-postgresql-apb"))

			// delete the service using app and project flags
			helper.Cmd("odo", "service", "delete", "-f", "dh-postgresql-apb", "--app", app, "--project", commonVar.Project).ShouldPass()
		})
	})

	/*
		Context("When link backend between component and service", func() {
			JustBeforeEach(func() {
				preSetup()
			})
			JustAfterEach(func() {
				cleanPreSetup()
			})
			It("should link backend to service successfully", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context1)
				helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--context", context1, "--project", project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "python"), context2)
				helper.CmdShouldPass("odo", "create", "--s2i", "python", "backend", "--context", context2, "--project", project)
				helper.CmdShouldPass("odo", "push", "--context", context2)
				helper.CmdShouldPass("odo", "link", "backend", "--context", context1)
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
	*/

	/*
		Context("When deleting service and unlink the backend from the frontend", func() {
			JustBeforeEach(func() {
				preSetup()
			})
			JustAfterEach(func() {
				cleanPreSetup()
			})
			It("should pass", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context1)
				helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--context", context1, "--project", project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "python"), context2)
				helper.CmdShouldPass("odo", "create", "--s2i", "python", "backend", "--context", context2, "--project", project)
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
	*/

	/*
		Context("When linking or unlinking a service or component", func() {
			JustBeforeEach(func() {
				preSetup()
			})
			JustAfterEach(func() {
				cleanPreSetup()
			})

			It("should print the environment variables being linked/unlinked", func() {
				helper.CopyExample(filepath.Join("source", "python"), context1)
				helper.CmdShouldPass("odo", "create", "--s2i", "python", "component1", "--context", context1, "--project", project)
				helper.CmdShouldPass("odo", "push", "--context", context1)
				helper.CopyExample(filepath.Join("source", "nodejs"), context2)
				helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "component2", "--context", context2, "--project", project)
				helper.CmdShouldPass("odo", "push", "--context", context2)

				// tests for linking a component to a component
				stdOut := helper.CmdShouldPass("odo", "link", "component2", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were added", "COMPONENT_COMPONENT2_HOST", "COMPONENT_COMPONENT2_PORT"})

				// tests for unlinking a component from a component
				stdOut = helper.CmdShouldPass("odo", "unlink", "component2", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were removed", "COMPONENT_COMPONENT2_HOST", "COMPONENT_COMPONENT2_PORT"})

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
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were added", "DB_PORT", "DB_HOST"})

				// tests for unlinking a service to a component
				stdOut = helper.CmdShouldPass("odo", "unlink", "dh-postgresql-apb", "--context", context1)
				helper.MatchAllInOutput(stdOut, []string{"The below secret environment variables were removed", "DB_PORT", "DB_HOST"})
			})
		})
	*/

	Context("When describing services", func() {
		It("should succeed when we're describing service that could have integer value for default field", func() {
			// https://github.com/openshift/odo/issues/2488
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "python", "component1", "--context", commonVar.Context, "--project", commonVar.Project).ShouldPass()
			helper.Chdir(commonVar.Context)

			helper.Cmd("odo", "catalog", "describe", "service", "dh-es-apb").ShouldPass()
			helper.Cmd("odo", "catalog", "describe", "service", "dh-import-vm-apb").ShouldPass()
		})
	})

	Context("When the application is deleted", func() {
		JustBeforeEach(func() {
			app = helper.RandString(6)
		})
		It("should delete the service(s) in the application as well", func() {
			helper.Cmd("odo", "service", "create", "--app", app, "-w", "dh-postgresql-apb", "--project", commonVar.Project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6").ShouldPass()
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			helper.Cmd("odo", "app", "delete", app, "--project", commonVar.Project, "-f").ShouldPass()

			ocArgs = []string{"get", "serviceinstances", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "No resources found")
			}, true)
		})
	})
})
