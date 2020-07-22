package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile url command tests", func() {
	var context, currentWorkingDirectory, cmpName string
	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		// Stop all containers labeled with the component name
		label := "component=" + cmpName
		dockerClient.StopContainers(label)

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Creating urls", func() {
		It("create should pass", func() {
			var stdout string
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			stdout = helper.CmdShouldPass("odo", "url", "create")
			helper.MatchAllInOutput(stdout, []string{cmpName + "-3000", "created for component"})
			stdout = helper.CmdShouldPass("odo", "push")
			Expect(stdout).To(ContainSubstring("Changes successfully pushed to component"))
		})

		It("create with now flag should pass", func() {
			var stdout string
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			stdout = helper.CmdShouldPass("odo", "url", "create", url1, "--now")
			helper.MatchAllInOutput(stdout, []string{url1, "created for component", "Changes successfully pushed to component"})
		})

		It("create with same url name should fail", func() {
			var stdout string
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1)

			stdout = helper.CmdShouldFail("odo", "url", "create", url1)
			Expect(stdout).To(ContainSubstring("URL " + url1 + " already exists"))

		})

		It("should be able to do a GET on the URL after a successful push", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", cmpName)

			output := helper.CmdShouldPass("odo", "push")
			helper.MatchAllInOutput(output, []string{"Executing devbuild command", "Executing devrun command"})

			url := strings.TrimSpace(helper.ExtractSubString(output, "127.0.0.1", "created"))

			helper.HttpWaitFor("http://"+url, "Hello from Node.js Starter Application!", 30, 1)
		})
	})

	Context("Listing urls", func() {
		It("should list url with appropriate state", func() {
			var stdout string
			url1 := helper.RandString(5)
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))

			helper.CmdShouldPass("odo", "url", "create", url1)
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed"})

			helper.CmdShouldPass("odo", "push")
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")

			stdout = helper.CmdShouldPass("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("Locally Deleted"))
		})

		It("should be able to list url in machine readable json format", func() {
			var stdout string
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)
			helper.CmdShouldPass("odo", "url", "create", url1, "--exposed-port", freePort, "--now")
			// odo url list -o json
			helper.WaitForCmdOut("odo", []string{"url", "list", "-o", "json"}, 1, true, func(output string) bool {
				desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"host":"127.0.0.1","port": 3000,"secure":false,"externalport":%s},"status":{"state":"Pushed"}}]}`, url1, freePort)
				if strings.Contains(output, url1) {
					Expect(desiredURLListJSON).Should(MatchJSON(output))
					return true
				}
				return false
			})
		})
	})

	Context("Describing urls", func() {
		It("should describe URL with appropriate state", func() {
			var stdout string
			url1 := helper.RandString(5)
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1)

			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed"})

			helper.CmdShouldPass("odo", "push")
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted"})
		})

		It("should be able to describe url in machine readable json format", func() {
			var stdout string
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			stdout = helper.CmdShouldFail("odo", "url", "describe", url1)
			Expect(stdout).To(ContainSubstring("the url " + url1 + " does not exist"))

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)
			helper.CmdShouldPass("odo", "url", "create", url1, "--exposed-port", freePort, "--now")
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})

			desiredURLListJSON := fmt.Sprintf(`{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"host":"127.0.0.1","port": 3000,"secure":false,"externalport":%s},"status":{"state":"Pushed"}}`, url1, freePort)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "-o", "json")
			Expect(desiredURLListJSON).Should(MatchJSON(stdout))
		})
	})
})
