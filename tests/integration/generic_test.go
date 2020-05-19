package integration

import (
	"fmt"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo generic", func() {
	// TODO: A neater way to provide odo path. Currently we assume \
	// odo and oc in $PATH already.

	var oc helper.OcRunner
	var testPHPGitURL = "https://github.com/appuio/example-php-sti-helloworld"
	var testLongURLName = "long-url-name-long-url-name-long-url-name-long-url-name-long-url-name"
	var globals helper.Globals

	BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		globals = helper.CommonBeforeEach()
	})

	AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	Context("Check the help usage for odo", func() {

		It("Makes sure that we have the long-description when running odo and we dont error", func() {
			output := helper.CmdShouldPass("odo")
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})

		It("Make sure we have the full description when performing odo --help", func() {
			output := helper.CmdShouldPass("odo", "--help")
			Expect(output).To(ContainSubstring("Use \"odo [command] --help\" for more information about a command."))
		})

		It("Fail when entering an incorrect name for a component", func() {
			output := helper.CmdShouldFail("odo", "component", "foobar")
			Expect(output).To(ContainSubstring("Subcommand not found, use one of the available commands"))
		})

	})

	Context("When executing catalog list without component directory", func() {
		It("should list all component catalogs", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "list", "components")
			Expect(stdOut).To(ContainSubstring("dotnet"))
			Expect(stdOut).To(ContainSubstring("nginx"))
			Expect(stdOut).To(ContainSubstring("php"))
			Expect(stdOut).To(ContainSubstring("ruby"))
			Expect(stdOut).To(ContainSubstring("wildfly"))
		})

	})

	Context("check catalog component search functionality", func() {
		It("check that a component does not exist", func() {
			componentRandomName := helper.RandString(7)
			output := helper.CmdShouldFail("odo", "catalog", "search", "component", componentRandomName)
			Expect(output).To(ContainSubstring("no component matched the query: " + componentRandomName))
		})
	})

	// Test machine readable output
	Context("when creating project -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		JustAfterEach(func() {
			helper.DeleteProject(projectName)
		})

		// odo project create foobar -o json
		It("should be able to create project and show output in json format", func() {
			actual := helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"message":"Project '%s' is ready for use"}`, projectName, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("Creating same project twice with flag -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		JustAfterEach(func() {
			helper.DeleteProject(projectName)
		})
		// odo project create foobar -o json (x2)
		It("should fail along with proper machine readable output", func() {
			helper.CmdShouldPass("odo", "project", "create", projectName)
			actual := helper.CmdShouldFail("odo", "project", "create", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Error","apiVersion":"odo.dev/v1alpha1","metadata":{"creationTimestamp":null},"message":"unable to create new project: unable to create new project %s: project.project.openshift.io \"%s\" already exists"}`, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("Delete the project with flag -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})

		// odo project delete foobar -o json
		It("should be able to delete project and show output in json format", func() {
			helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")

			actual := helper.CmdShouldPass("odo", "project", "delete", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"message":"Deleted project : %s"}`, projectName, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	// Uncomment once https://github.com/openshift/odo/issues/1708 is fixed
	/*Context("odo machine readable output on empty project", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should be able to return project list", func() {
			actualProjectListJSON := helper.CmdShouldPass("odo", "project", "list", "-o", "json")
			partOfProjectListJSON := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},`, project)
			Expect(actualProjectListJSON).To(ContainSubstring(partOfProjectListJSON))
		})
	})*/

	Context("creating component with an application and url", func() {
		JustBeforeEach(func() {
			helper.Chdir(globals.Context)
		})

		It("should create the component in default application", func() {
			helper.CmdShouldPass("odo", "create", "php", "testcmp", "--app", "e2e-xyzk", "--project", globals.Project, "--git", testPHPGitURL)
			helper.CmdShouldPass("odo", "config", "set", "Ports", "8080/TCP", "-f")
			helper.CmdShouldPass("odo", "push")
			oc.VerifyCmpName("testcmp", globals.Project)
			oc.VerifyAppNameOfComponent("testcmp", "e2e-xyzk", globals.Project)
			helper.CmdShouldPass("odo", "app", "delete", "e2e-xyzk", "-f")
		})
	})

	Context("should list applications in other project", func() {
		It("should be able to create a php component with application created", func() {
			helper.CmdShouldPass("odo", "create", "php", "testcmp", "--app", "testing", "--project", globals.Project, "--ref", "master", "--git", testPHPGitURL, "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			currentProject := helper.CreateRandProject()
			currentAppNames := helper.CmdShouldPass("odo", "app", "list", "--project", currentProject)
			Expect(currentAppNames).To(ContainSubstring("There are no applications deployed in the project '" + currentProject + "'"))
			appNames := helper.CmdShouldPass("odo", "app", "list", "--project", globals.Project)
			Expect(appNames).To(ContainSubstring("testing"))
			helper.DeleteProject(currentProject)
		})
	})

	Context("when running odo push with flag --show-log", func() {
		It("should be able to push changes", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--project", globals.Project, "--context", globals.Context)

			// Push the changes with --show-log
			getLogging := helper.CmdShouldPass("odo", "push", "--show-log", "--context", globals.Context)
			Expect(getLogging).To(ContainSubstring("Building component"))
		})
	})

	Context("deploying a component with a specific image name", func() {
		It("should deploy the component", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "create", "nodejs:latest", "testversioncmp", "--project", globals.Project, "--context", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "delete", "-f", "--context", globals.Context+"/nodejs-ex")
		})
	})

	Context("When deleting two project one after the other", func() {
		It("should be able to delete sequentially", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
		})
	})

	Context("When deleting three project one after the other in opposite order", func() {
		It("should be able to delete", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()
			project3 := helper.CreateRandProject()

			helper.DeleteProject(project1)
			helper.DeleteProject(project2)
			helper.DeleteProject(project3)

		})
	})

	Context("when executing odo version command", func() {
		It("should show the version of odo major components including server login URL", func() {
			odoVersion := helper.CmdShouldPass("odo", "version")
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`)
			kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
			reServerURL := regexp.MustCompile(`Server:\s*https:\/\/(.+\.com|([0-9]+.){3}[0-9]+):[0-9]{4}`)
			serverURLStringMatch := reServerURL.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			Expect(kubernetesVersionStringMatch).Should(BeTrue())
			Expect(serverURLStringMatch).Should(BeTrue())
		})
	})

	Context("prevent the user from creating a URL with name that has more than 63 characters", func() {
		JustBeforeEach(func() {
			helper.Chdir(globals.Context)
		})
		It("should not allow creating a URL with long name", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project)
			stdOut := helper.CmdShouldFail("odo", "url", "create", testLongURLName, "--port", "8080")
			Expect(stdOut).To(ContainSubstring("url name must be shorter than 63 characters"))
		})
	})

	Context("when component's deployment config is deleted with oc", func() {
		var componentRandomName string

		JustBeforeEach(func() {
			componentRandomName = helper.RandString(6)
			helper.Chdir(globals.Context)
		})

		It("should delete all OpenShift objects except the component's imagestream", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", componentRandomName, "--project", globals.Project)
			helper.CmdShouldPass("odo", "push")

			// Delete the deployment config using oc delete
			dc := oc.GetDcName(componentRandomName, globals.Project)
			helper.CmdShouldPass("oc", "delete", "--wait", "dc", dc, "--namespace", globals.Project)

			// insert sleep because it takes a few seconds to delete *all*
			// objects owned by DC but we should be able to check if a service
			// got deleted in a second.
			time.Sleep(1 * time.Second)

			// now check if the service owned by the DC exists. Service name is
			// same as DC name for a given component.
			stdOut := helper.CmdShouldFail("oc", "get", "svc", dc, "--namespace", globals.Project)
			Expect(stdOut).To(ContainSubstring("NotFound"))

			// ensure that the image stream still exists
			helper.CmdShouldPass("oc", "get", "is", dc, "--namespace", globals.Project)
		})
	})
})
