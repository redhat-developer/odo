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

var _ = Describe("odo service command tests", func() {
	//new clean project and context for each test
	var project string
	var context, context1, context2 string
	var app string
	var serviceName string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string
	var oc helper.OcRunner
	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
	})

	preSetup := func() {
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
		context1 = helper.CreateNewContext()
		context2 = helper.CreateNewContext()
		originalDir = helper.Getwd()
	}

	cleanPreSetup := func() {
		helper.Chdir(originalDir)
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		helper.DeleteDir(context1)
		helper.DeleteDir(context2)
		os.Unsetenv("GLOBALODOCONFIG")
	}

	Context("when running help for service command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "service", "-h")
			Expect(appHelp).To(ContainSubstring("Perform service catalog operations"))
		})
	})

	Context("check catalog service search functionality", func() {
		It("check that a service does not exist", func() {
			serviceRandomName := helper.RandString(7)
			output := helper.CmdShouldFail("odo", "catalog", "search", "service", serviceRandomName)
			Expect(output).To(ContainSubstring("no service matched the query: " + serviceRandomName))
		})
	})

	Context("checking machine readable output for service catalog", func() {
		It("should succeed listing catalog components", func() {
			// Since service catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			output := helper.CmdShouldPass("odo", "catalog", "list", "services", "-o", "json")
			Expect(output).To(ContainSubstring("List"))
		})
	})

	Context("checking machine readable output for service catalog", func() {
		It("should succeed listing catalog components", func() {
			// Since service catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			helper.CmdShouldPass("odo", "catalog", "list", "services", "-o", "json")
		})
	})

	Context("check search functionality", func() {

		It("should pass with searching for part of a service name", func() {

			// We just use "sql" as some catalogs only have postgresql-persistent and
			// others dh-postgresql-db. So let's just see if there's "any" postgresql to begin
			// with
			output := helper.CmdShouldPass("odo", "catalog", "search", "service", "sql")
			Expect(output).To(ContainSubstring("postgresql"))
		})

	})

	Context("create service with Env non-interactively", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			app = helper.RandString(7)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
		})

		It("should be able to create postgresql with env", func() {
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", project, "--app", app,
				"--plan", "dev", "-p", "postgresql_user=lukecage", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w")
			// there is only a single pod in the project
			ocArgs := []string{"describe", "pod", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "lukecage")
			})

			// Delete the service
			helper.CmdShouldPass("odo", "service", "delete", "dh-postgresql-apb", "-f", "--app", app, "--project", project)
		})

		It("should be able to create postgresql with env multiple times", func() {
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", project, "--app", app,
				"--plan", "dev", "-p", "postgresql_user=lukecage", "-p", "postgresql_user=testworker", "-p", "postgresql_password=secret",
				"-p", "postgresql_password=universe", "-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w")
			// there is only a single pod in the project
			ocArgs := []string{"describe", "pod", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "testworker")
			})

			// Delete the service
			helper.CmdShouldPass("odo", "service", "delete", "dh-postgresql-apb", "-f", "--app", app, "--project", project)
		})
	})

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
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			app = helper.RandString(7)
			serviceName = "odo-postgres-service"
			originalDir = helper.Getwd()
			helper.Chdir(context)
			SetDefaultConsistentlyDuration(30 * time.Second)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})

		It("should be able to create, list and delete a service using a given value for --context", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "python", "--app", app, "--project", project)

			// cd to the originalDir to create service using --context
			helper.Chdir(originalDir)
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", serviceName,
				"--context", context,
			)

			// now check if listing the service using --context works
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, serviceName)
			})
			stdOut := helper.CmdShouldPass("odo", "service", "list", "--context", context)
			Expect(stdOut).To(ContainSubstring(serviceName))

			// now check if deleting the service using --context works
			stdOut = helper.CmdShouldPass("odo", "service", "delete", "-f", serviceName, "--context", context)
			Expect(stdOut).To(ContainSubstring(serviceName))
		})

		It("should be able to list services, as well as json list in a given app and project combination", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--app", app, "--project", project)

			// create a service from within a component directory
			helper.CmdShouldPass("odo", "service", "create", "dh-prometheus-apb", "--plan", "ephemeral",
				"--app", app, "--project", project,
			)
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-prometheus-apb")
			})

			// Listing the services should work as expected from within the component directory.
			// This means, it should not require --app or --project flags
			stdOut := helper.CmdShouldPass("odo", "service", "list")
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// Check json output
			stdOut = helper.CmdShouldPass("odo", "service", "list", "-o", "json")
			helper.MatchAllInOutput(stdOut, []string{"dh-prometheus-apb", "List"})

			// cd to a non-component directory and list services
			helper.Chdir(originalDir)
			stdOut = helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", project)
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// Check json output
			helper.Chdir(originalDir)
			stdOut = helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", project, "-o", "json")
			helper.MatchAllInOutput(stdOut, []string{"dh-prometheus-apb", "List"})
		})

		It("should be able to create, list and delete services without a context and using --app and --project flags instaed", func() {
			// create a service using only app and project flags
			// we do Chdir first because originalDir doesn't have a context
			helper.Chdir(originalDir)

			// create the service
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6",
				"--app", app, "--project", project)

			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			// list the service using app and project flags
			stdOut := helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", project)
			Expect(stdOut).To(ContainSubstring("dh-postgresql-apb"))

			// delete the service using app and project flags
			helper.CmdShouldPass("odo", "service", "delete", "-f", "dh-postgresql-apb", "--app", app, "--project", project)
		})
	})

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

	Context("When describing services", func() {
		JustBeforeEach(func() {
			preSetup()
		})
		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should succeed when we're describing service that could have integer value for default field", func() {
			// https://github.com/openshift/odo/issues/2488
			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "python", "component1", "--context", context, "--project", project)
			helper.Chdir(context)

			helper.CmdShouldPass("odo", "catalog", "describe", "service", "dh-es-apb")
			helper.CmdShouldPass("odo", "catalog", "describe", "service", "dh-import-vm-apb")
		})
	})

	Context("When the application is deleted", func() {
		JustBeforeEach(func() {
			preSetup()
			app = helper.RandString(6)
		})
		It("should delete the service(s) in the application as well", func() {
			helper.CmdShouldPass("odo", "service", "create", "--app", app, "-w", "dh-postgresql-apb", "--project", project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6")
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			helper.CmdShouldPass("odo", "app", "delete", app, "--project", project, "-f")

			ocArgs = []string{"get", "serviceinstances", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "No resources found")
			}, true)
		})
	})
})
