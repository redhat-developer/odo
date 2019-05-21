package integration

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"

	"fmt"
)

var _ = Describe("odo generic", func() {
	// TODO: A neater way to provide odo path. Currently we assume \
	// odo and oc in $PATH already.
	var project string
	var context string
	var oc helper.OcRunner
	var err error
	var testPHPGitURL = "https://github.com/appuio/example-php-sti-helloworld"

	BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		SetDefaultEventuallyTimeout(10 * time.Minute)
	})

	Context("Executing catalog list without component directory", func() {
		It("All component catalogs are listed ", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "list", "components")
			Expect(stdOut).To(ContainSubstring("dotnet"))
			Expect(stdOut).To(ContainSubstring("nginx"))
			Expect(stdOut).To(ContainSubstring("php"))
			Expect(stdOut).To(ContainSubstring("ruby"))
			Expect(stdOut).To(ContainSubstring("wildfly"))
		})
	})

	Context("creating component without an application and url", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(".odo")
		})
		It("should create the component in default application", func() {
			helper.CmdShouldPass("odo", "create", "php", "testcmp", "--app", "e2e-xyzk", "--project", project, "--git", testPHPGitURL)
			helper.CmdShouldPass("odo", "config", "set", "Ports", "8080/TCP")
			helper.CmdShouldPass("odo", "push", "--config")
			helper.CmdShouldPass("odo", "url", "create", "myurl", "--port", "8080")
			helper.CmdShouldPass("odo", "push")
			oc.VerifyCmpName("testcmp", project)
			oc.VerifyAppNameOfComponent("testcmp", "e2e-xyzk", project)
			helper.CmdShouldPass("odo", "url", "delete", "myurl", "-f")
			helper.CmdShouldPass("odo", "app", "delete", "e2e-xyzk", "-f")
		})
	})

	Context("should list applications in other project", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(".odo")
		})
		It("should be able to create a php component with application created", func() {
			helper.CmdShouldPass("odo", "create", "php", "testcmp", "--app", "testing", "--project", project, "--ref", "master", "--git", testPHPGitURL)
			helper.CmdShouldPass("odo", "push")
			appNames := helper.CmdShouldPass("odo", "app", "list", "--project", project)
			Expect(appNames).To(ContainSubstring("testing"))
		})
	})

	Context("when .odoignore file exists", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			context = helper.CreateNewContext()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(context)
		})
		It("should create and push the contents of a named component excluding the contents in .odoignore file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			ignoreFilePath := filepath.Join(context, "nodejs-ex", ".odoignore")
			if helper.CreateFileWithContent(ignoreFilePath, ".git\ntests/\nREADME.md") != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}

			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--project", project, "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp("nodejs", project)

			// verify that the views folder got pushed
			stdOut1 := oc.ExecListDir(podName, project)
			Expect(stdOut1).To(ContainSubstring("views"))

			// verify that the tests was not pushed
			stdOut2 := oc.ExecListDir(podName, project)
			Expect(stdOut2).To(Not(ContainSubstring(("tests"))))

			// verify that the README.md file was not pushed
			stdOut3 := oc.ExecListDir(podName, project)
			Expect(stdOut3).To(Not(ContainSubstring(("README.md"))))

		})

		It("should be able to push changes while showing logging", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--context", context+"/nodejs-ex")

			// Push the changes with --show-log
			getLogging := helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")
			Expect(getLogging).To(ContainSubstring("Building component"))
		})

		It("should be able to spam odo push without anything breaking", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--context", context+"/nodejs-ex")

			// Iteration 1
			helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")

			// Iteration 2
			helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")

			// Iteration 3
			helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")
		})
	})

	Context("deploying a component with a specific image name", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
			context = helper.CreateNewContext()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(context)
		})
		It("should deploy the component", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "create", "nodejs:latest", "testversioncmp", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
		})

		It("should delete the deployed image-specific component", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "create", "nodejs:latest", "testversioncmp", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "delete", "-f", "--context", context+"/nodejs-ex")
		})
	})

	Context("project deletion", func() {
		var originalDir string

		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
			os.RemoveAll(context)
		})
		It("be able to delete two project one after the other", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
		})

		It("be able to delete three project one after the other in opposite order", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()
			project3 := helper.CreateRandProject()

			helper.DeleteProject(project1)
			helper.DeleteProject(project2)
			helper.DeleteProject(project3)

		})
	})

	Context("validate odo version cmd with other major components version", func() {
		It("should show the version of odo major components", func() {
			odoVersion := helper.CmdShouldPass("odo", "version")
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+\+\w+`)
			kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			Expect(kubernetesVersionStringMatch).Should(BeTrue())
		})

		It("should show server login URL", func() {
			odoVersion := helper.CmdShouldPass("odo", "version")
			reServerURL := regexp.MustCompile(`Server:\s*https:\/\/(.+\.com|([0-9]+.){3}[0-9]+):[0-9]{4}`)
			serverURLStringMatch := reServerURL.MatchString(odoVersion)
			Expect(serverURLStringMatch).Should(BeTrue())
		})
	})
})
