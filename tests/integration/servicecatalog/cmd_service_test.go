package integration

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo service command tests", func() {
	//new clean globals.Project and context for each test
	var context1, context2 string
	var app string
	var serviceName string

	var oc helper.OcRunner
	var globals helper.Globals

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	BeforeEach(func() {
		globals = helper.CommonBeforeEach()
		oc = helper.NewOcRunner("oc")
	})

	AfterEach(func() {
		helper.CommonAfterEeach(globals)
	})

	preSetup := func() {
		context1 = helper.CreateNewContext()
		context2 = helper.CreateNewContext()
	}

	cleanPreSetup := func() {
		helper.DeleteDir(context1)
		helper.DeleteDir(context2)
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
			app = helper.RandString(7)
		})

		It("should be able to create postgresql with env", func() {
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", globals.Project, "--app", app,
				"--plan", "dev", "-p", "postgresql_user=lukecage", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w")
			// there is only a single pod in the project
			ocArgs := []string{"describe", "pod", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "lukecage")
			})

			// Delete the service
			helper.CmdShouldPass("odo", "service", "delete", "dh-postgresql-apb", "-f", "--app", app, "--project", globals.Project)
		})

		It("should be able to create postgresql with env multiple times", func() {
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", globals.Project, "--app", app,
				"--plan", "dev", "-p", "postgresql_user=lukecage", "-p", "postgresql_user=testworker", "-p", "postgresql_password=secret",
				"-p", "postgresql_password=universe", "-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w")
			// there is only a single pod in the project
			ocArgs := []string{"describe", "pod", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "testworker")
			})

			// Delete the service
			helper.CmdShouldPass("odo", "service", "delete", "dh-postgresql-apb", "-f", "--app", app, "--project", globals.Project)
		})
	})

	Context("When creating with a spring boot application", func() {
		JustBeforeEach(func() {

			helper.Chdir(globals.Context)
		})

		It("should be able to create postgresql and link it with springboot", func() {
			oc.ImportJavaIS(globals.Project)
			helper.CopyExample(filepath.Join("source", "openjdk-sb-postgresql"), globals.Context)

			// Local config needs to be present in order to create service https://github.com/openshift/odo/issues/1602
			helper.CmdShouldPass("odo", "create", "java:8", "sb-app", "--project", globals.Project)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080")

			// push
			helper.CmdShouldPass("odo", "push")

			// create the postgres service
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--project", globals.Project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", "-w")

			// link the service
			helper.CmdShouldPass("odo", "link", "dh-postgresql-apb", "--project", globals.Project, "-w", "--wait-for-target")
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
		var originalDir string
		JustBeforeEach(func() {
			originalDir = helper.Getwd()
			app = helper.RandString(7)
			serviceName = "odo-postgres-service"
			helper.Chdir(globals.Context)
		})

		It("should be able to create, list and delete a service using a given value for --context", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "python"), globals.Context)
			helper.CmdShouldPass("odo", "create", "python", "--app", app, "--project", globals.Project)

			// cd to the originalDir to create service using --context
			helper.Chdir(originalDir)
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6", serviceName,
				"--context", globals.Context,
			)

			// now check if listing the service using --context works
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, serviceName)
			})
			stdOut := helper.CmdShouldPass("odo", "service", "list", "--context", globals.Context)
			Expect(stdOut).To(ContainSubstring(serviceName))

			// now check if deleting the service using --context works
			stdOut = helper.CmdShouldPass("odo", "service", "delete", "-f", serviceName, "--context", globals.Context)
			Expect(stdOut).To(ContainSubstring(serviceName))
		})

		It("should be able to list services, as well as json list in a given app and project combination", func() {
			// create a component by copying the example
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--app", app, "--project", globals.Project)

			// create a service from within a component directory
			helper.CmdShouldPass("odo", "service", "create", "dh-prometheus-apb", "--plan", "ephemeral",
				"--app", app, "--project", globals.Project,
			)
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-prometheus-apb")
			})

			// Listing the services should work as expected from within the component directory.
			// This means, it should not require --app or --project flags
			stdOut := helper.CmdShouldPass("odo", "service", "list")
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// Check json output
			stdOut = helper.CmdShouldPass("odo", "service", "list", "-o", "json")
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))
			Expect(stdOut).To(ContainSubstring("ServiceList"))

			// cd to a non-component directory and list services
			helper.Chdir(originalDir)
			stdOut = helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", globals.Project)
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))

			// Check json output
			helper.Chdir(originalDir)
			stdOut = helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", globals.Project, "-o", "json")
			Expect(stdOut).To(ContainSubstring("dh-prometheus-apb"))
			Expect(stdOut).To(ContainSubstring("ServiceList"))

		})

		It("should be able to create, list and delete services without a context and using --app and --project flags instaed", func() {
			// create a service using only app and project flags
			// we do Chdir first because originalDir doesn't have a context
			helper.Chdir(originalDir)

			// create the service
			helper.CmdShouldPass("odo", "service", "create", "dh-postgresql-apb", "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6",
				"--app", app, "--project", globals.Project)

			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			// list the service using app and project flags
			stdOut := helper.CmdShouldPass("odo", "service", "list", "--app", app, "--project", globals.Project)
			Expect(stdOut).To(ContainSubstring("dh-postgresql-apb"))

			// delete the service using app and project flags
			helper.CmdShouldPass("odo", "service", "delete", "-f", "dh-postgresql-apb", "--app", app, "--project", globals.Project)
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
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", globals.Project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", globals.Project)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			helper.CmdShouldPass("odo", "link", "backend", "--context", context1)
			// Switching to context2 dir because --context flag is not supported with service command
			helper.Chdir(context2)
			helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

			ocArgs := []string{"get", "serviceinstance", "-n", globals.Project, "-o", "name"}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			helper.CmdShouldPass("odo", "link", "mysql-persistent", "--wait-for-target", "--component", "backend", "--project", globals.Project)
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("backend", "app", globals.Project)
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
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", globals.Project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", globals.Project)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			helper.CmdShouldPass("odo", "link", "backend", "--context", context1)
			helper.Chdir(context2)
			helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

			ocArgs := []string{"get", "serviceinstance", "-n", globals.Project, "-o", "name"}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			helper.CmdShouldPass("odo", "service", "delete", "mysql-persistent", "-f")
			// ensure that the backend no longer has an envFrom value
			backendEnvFromOutput := oc.GetEnvFromEntry("backend", "app", globals.Project)
			Expect(backendEnvFromOutput).To(Equal("''"))
			// ensure that the frontend envFrom was not changed
			frontEndEnvFromOutput := oc.GetEnvFromEntry("frontend", "app", globals.Project)
			Expect(frontEndEnvFromOutput).To(ContainSubstring("backend"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--component", "frontend", "--project", globals.Project)
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app", globals.Project)
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
			helper.CmdShouldPass("odo", "create", "python", "component1", "--context", context1, "--project", globals.Project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "nodejs"), context2)
			helper.CmdShouldPass("odo", "create", "nodejs", "component2", "--context", context2, "--project", globals.Project)
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
			helper.CmdShouldPass("odo", "service", "create", "-w", "dh-postgresql-apb", "--project", globals.Project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6")
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", globals.Project}
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

	Context("When describing services", func() {
		JustBeforeEach(func() {
			preSetup()
		})
		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should succeed when we're describing service that could have integer value for default field", func() {
			// https://github.com/openshift/odo/issues/2488
			helper.CopyExample(filepath.Join("source", "python"), globals.Context)
			helper.CmdShouldPass("odo", "create", "python", "component1", "--context", globals.Context, "--project", globals.Project)
			helper.Chdir(globals.Context)

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
			helper.CmdShouldPass("odo", "service", "create", "--app", app, "-w", "dh-postgresql-apb", "--project", globals.Project, "--plan", "dev",
				"-p", "postgresql_user=luke", "-p", "postgresql_password=secret",
				"-p", "postgresql_database=my_data", "-p", "postgresql_version=9.6")
			ocArgs := []string{"get", "serviceinstance", "-o", "name", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "dh-postgresql-apb")
			})

			helper.CmdShouldPass("odo", "app", "delete", app, "--project", globals.Project, "-f")

			ocArgs = []string{"get", "serviceinstances", "-n", globals.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "No resources found")
			}, true)
		})
	})
})
