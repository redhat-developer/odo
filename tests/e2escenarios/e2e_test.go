//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package e2escenarios

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("E2E Test", func() {
	var commonVar helper.CommonVar
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
	})
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	checkIfDevEnvIsUp := func(url, assertString string) {
		Eventually(func() string {
			resp, err := http.Get(fmt.Sprintf("http://%s", url))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			return string(body)
		}, 120*time.Second, 15*time.Second).Should(Equal(assertString))
	}

	Context("starting with empty Directory", func() {
		componentName := helper.RandString(6)
		var _ = BeforeEach(func() {
			helper.Chdir(commonVar.Context)
			Expect(helper.ListFilesInDir(commonVar.Context)).To(BeEmpty())
		})

		It("should verify developer workflow from empty Directory", func() {
			deploymentName := "my-component"
			serviceName := "my-cs"
			getDeployArgs := []string{"get", "deployment", "-n", commonVar.Project}
			getSVCArgs := []string{"get", "svc", "-n", commonVar.Project}

			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				helper.ExpectString(ctx, "Select language")
				helper.SendLine(ctx, "javascript")

				helper.ExpectString(ctx, "Select project type")
				helper.SendLine(ctx, "Node.js\n")

				helper.ExpectString(ctx, "Which starter project do you want to use")
				helper.SendLine(ctx, "nodejs-starter\n")

				helper.ExpectString(ctx, "Enter component name")
				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("server.js"))

			// "execute odo dev and add changes to application"
			var devSession helper.DevSession
			var ports map[string]string

			devSession, _, _, ports, err = helper.StartDevMode(nil)
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js", "from updated Node.js")
			Expect(err).ToNot(HaveOccurred())
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from updated Node.js Starter Application!")

			// "changes are made to the applications"
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from updated Node.js", "from Node.js app v2")
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should deploy new changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from Node.js app v2 Starter Application!")

			// "running odo list"
			stdout := helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Dev"})

			// "exit dev mode and run odo deploy"
			devSession.Stop()
			devSession.WaitEnd()

			// all resources should be deleted from the namespace
			services := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(services).To(BeEmpty())
			pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
			Expect(pvcs).To(BeEmpty())
			pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
			Expect(pods).To(BeEmpty())

			// "run odo deploy"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "nodejs-prj1-api-abhz", componentName)

			stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Your Devfile has been successfully deployed"))

			// should deploy new changes
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Deploy"})

			// start dev mode again
			devSession, _, _, ports, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())

			// making changes to the project again
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js app v2", "from Node.js app v3")
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from Node.js app v3 Starter Application!")

			// should list both dev,deploy
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Dev", "Deploy"})

			// "exit dev mode and run odo deploy"
			devSession.Stop()

			// "run odo deploy"
			stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Your Devfile has been successfully deployed"))

			// "run odo delete and check if the component is deleted"
			command = []string{"odo", "delete", "component"}
			_, err = helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Are you sure you want to delete \""+componentName+"\" and all its resources?")
				helper.SendLine(ctx, "y")
				helper.ExpectString(ctx, "successfully deleted")
			})
			Expect(err).To(BeNil())
			Eventually(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(deploymentName))
			Eventually(string(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(serviceName))
		})
	})

	Context("starting with non-empty Directory", func() {
		componentName := helper.RandString(6)
		var _ = BeforeEach(func() {
			helper.Chdir(commonVar.Context)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		It("should verify developer workflow from non-empty Directory", func() {
			deploymentName := "my-component"
			serviceName := "my-cs"
			getDeployArgs := []string{"get", "deployment", "-n", commonVar.Project}
			getSVCArgs := []string{"get", "svc", "-n", commonVar.Project}

			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				// helper.ExpectString(ctx, "Based on the files in the current directory odo detected")
				helper.ExpectString(ctx, "Language: javascript")
				helper.ExpectString(ctx, "Project type: nodejs")
				helper.ExpectString(ctx, "Is this correct")

				helper.SendLine(ctx, "\n")

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")

				helper.SendLine(ctx, "\n")

				helper.ExpectString(ctx, "Enter component name")

				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))

			// "execute odo dev and add changes to application"
			var devSession helper.DevSession
			var ports map[string]string

			devSession, _, _, ports, err = helper.StartDevMode(nil)
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js", "from updated Node.js")
			Expect(err).ToNot(HaveOccurred())

			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())

			// "should update the changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from updated Node.js Starter Application!")

			// "changes are made made to the applications"

			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from updated Node.js", "from Node.js app v2")
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())

			// "should deploy new changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from Node.js app v2 Starter Application!")

			// "running odo list"
			stdout := helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Dev"})

			// "exit dev mode and run odo deploy"
			devSession.Stop()
			devSession.WaitEnd()

			// all resources should be deleted from the namespace
			services := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(services).To(BeEmpty())
			pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
			Expect(pvcs).To(BeEmpty())
			pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
			Expect(pods).To(BeEmpty())

			// "run odo deploy"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "nodejs-prj1-api-abhz", componentName)
			stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Your Devfile has been successfully deployed"))

			// should deploy new changes
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Deploy"})

			// start dev mode again
			devSession, _, _, ports, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())

			// making changes to the project again
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js app v2", "from Node.js app v3")

			// "should update the changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from Node.js app v3 Starter Application!")

			// should list both dev,deploy
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Dev", "Deploy"})

			// "exit dev mode and run odo deploy"
			devSession.Stop()

			// "run odo deploy"
			stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Your Devfile has been successfully deployed"))

			command = []string{"odo", "delete", "component"}
			_, err = helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Are you sure you want to delete \""+componentName+"\" and all its resources?")
				helper.SendLine(ctx, "y")
				helper.ExpectString(ctx, "successfully deleted")
			})
			Expect(err).To(BeNil())
			Eventually(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(deploymentName))
			Eventually(string(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(serviceName))
		})
	})
})
