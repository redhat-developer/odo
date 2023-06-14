package e2escenarios

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("E2E Test", func() {
	var commonVar helper.CommonVar
	var componentName string
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
	})
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	waitRemoteApp := func(urlInContainer, assertString, containerName string) {
		cmp := helper.NewComponent(componentName, "app", "Dev", commonVar.Project, commonVar.CliRunner)
		helper.WaitAppReadyInContainer(cmp, containerName, []string{"curl", urlInContainer}, 5*time.Second, 120*time.Second, ContainSubstring(assertString), nil)
	}

	checkIfDevEnvIsUp := func(url, assertString string) {
		Eventually(func() string {
			resp, err := http.Get(fmt.Sprintf("http://%s", url))
			if err != nil {
				fmt.Fprintf(GinkgoWriter, "error while trying to GET %q: %v\n", url, err)
				return ""
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			return string(body)
		}, 120*time.Second, 15*time.Second).Should(ContainSubstring(assertString))
	}

	Context("starting with empty Directory", func() {
		var hasMultipleVersions bool
		var _ = BeforeEach(func() {
			helper.Chdir(commonVar.Context)
			Expect(helper.ListFilesInDir(commonVar.Context)).To(BeEmpty())
			out := helper.Cmd("odo", "registry", "--devfile", "nodejs", "--devfile-registry", "DefaultDevfileRegistry").ShouldPass().Out()
			// Version pattern has always been in the form of X.X.X
			vMatch := regexp.MustCompile(`(?:\d.\d.\d)`)
			if matches := vMatch.FindAll([]byte(out), -1); len(matches) > 1 {
				hasMultipleVersions = true
			}
		})

		It("should verify developer workflow from empty Directory", func() {
			deploymentName := "my-component"
			serviceName := "my-cs"
			getDeployArgs := []string{"get", "deployment", "-n", commonVar.Project}
			getSVCArgs := []string{"get", "svc", "-n", commonVar.Project}

			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				helper.ExpectString(ctx, "Select language")
				helper.SendLine(ctx, "JavaScript")

				helper.ExpectString(ctx, "Select project type")
				helper.SendLine(ctx, "Node.js")

				if hasMultipleVersions {
					helper.ExpectString(ctx, "Select version: ")
					helper.SendLine(ctx, "")
				}

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Which starter project do you want to use")
				helper.SendLine(ctx, "nodejs-starter")

				helper.ExpectString(ctx, "Enter component name")
				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("server.js"))

			// "execute odo dev and add changes to application"
			var devSession helper.DevSession

			devSession, err = helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())
			waitRemoteApp("http://127.0.0.1:3000", "Hello from Node.js Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints["3000"], "Hello from Node.js Starter Application!")

			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js", "from updated Node.js")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			waitRemoteApp("http://127.0.0.1:3000", "Hello from updated Node.js Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints["3000"], "Hello from updated Node.js Starter Application!")

			// "changes are made to the applications"
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from updated Node.js", "from Node.js app v2")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should deploy new changes"
			waitRemoteApp("http://127.0.0.1:3000", "Hello from Node.js app v2 Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints["3000"], "Hello from Node.js app v2 Starter Application!")

			// "running odo list"
			stdout := helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "Node.js", "Dev"})

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
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"),
				path.Join(commonVar.Context, "devfile.yaml"),
				componentName)

			stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Your Devfile has been successfully deployed"))

			// should deploy new changes
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "nodejs", "Deploy"})

			// start dev mode again
			devSession, err = helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())

			// making changes to the project again
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js app v2", "from Node.js app v3")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			waitRemoteApp("http://127.0.0.1:3000", "Hello from Node.js app v3 Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints["3000"], "Hello from Node.js app v3 Starter Application!")

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
		const (
			AppPort     = "8080"
			AppLocalURL = "http://localhost:8080"
		)
		var _ = BeforeEach(func() {
			helper.Chdir(commonVar.Context)
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
		})
		It("should verify developer workflow from non-empty Directory", func() {
			deploymentName := "my-component"
			serviceName := "my-cs"
			getDeployArgs := []string{"get", "deployment", "-n", commonVar.Project}
			getSVCArgs := []string{"get", "svc", "-n", commonVar.Project}

			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				// helper.ExpectString(ctx, "Based on the files in the current directory odo detected")
				helper.ExpectString(ctx, "Language: Java")
				helper.ExpectString(ctx, "Project type: springboot")
				helper.ExpectString(ctx, "Is this correct")

				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")

				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Enter component name")

				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))

			// "execute odo dev and add changes to application"
			var devSession helper.DevSession
			devSession, err = helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())
			Expect(devSession.StdOut).ToNot(BeEmpty())
			waitRemoteApp(AppLocalURL, "Hello World!", "tools")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello World!")

			helper.ReplaceString(filepath.Join(commonVar.Context, "src", "main", "java", "com", "example", "demo", "DemoApplication.java"), "Hello World!", "Hello updated World!")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			waitRemoteApp(AppLocalURL, "Hello updated World!", "tools")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello updated World!")

			// "changes are made to the applications"
			helper.ReplaceString(filepath.Join(commonVar.Context, "src", "main", "java", "com", "example", "demo", "DemoApplication.java"), "Hello updated World!", "Hello from an updated World!")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should deploy new changes"
			waitRemoteApp(AppLocalURL, "Hello from an updated World!", "tools")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello from an updated World!")

			// "running odo list"
			stdout := helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "springboot", "Dev"})

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
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "springboot", "devfile-deploy.yaml"),
				path.Join(commonVar.Context, "devfile.yaml"),
				componentName)

			stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("Your Devfile has been successfully deployed"))

			// should deploy new changes
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "springboot", "Deploy"})

			// start dev mode again
			devSession, err = helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())

			// making changes to the project again
			helper.ReplaceString(filepath.Join(commonVar.Context, "src", "main", "java", "com", "example", "demo", "DemoApplication.java"), "Hello from an updated World!", "Hello from an updated v2 World!")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())

			// "should update the changes"
			waitRemoteApp(AppLocalURL, "Hello from an updated v2 World!", "tools")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello from an updated v2 World!")

			// should list both dev,deploy
			stdout = helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "springboot", "Dev", "Deploy"})

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

	Context("starting with non-empty Directory test debugging", func() {
		// We use a devfile that does not require an external debugger client like in the case of Java Devfiles.
		// Node.js is simple and good for testing debugging feature
		const (
			LocalAppURL = "http://127.0.0.1:3000"
			AppPort     = "3000"
			DebugPort   = "5858"
		)
		var _ = BeforeEach(func() {
			helper.Chdir(commonVar.Context)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		It("should verify developer workflow from non-empty Directory", func() {
			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				// helper.ExpectString(ctx, "Based on the files in the current directory odo detected")
				helper.ExpectString(ctx, "Language: JavaScript")
				helper.ExpectString(ctx, "Project type: Node.js")
				helper.ExpectString(ctx, "Is this correct")

				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")
				helper.SendLine(ctx, "runtime")

				// Personalize the Devfile to use the debug envvar defined inside package.json
				helper.ExpectString(ctx, "What configuration do you want change")
				helper.SendLine(ctx, "Delete environment variable \"DEBUG_PORT\"")

				helper.ExpectString(ctx, "What configuration do you want change")
				helper.SendLine(ctx, "Add new environment variable")

				helper.ExpectString(ctx, "Enter new environment variable name")
				helper.SendLine(ctx, "DEBUG_PORT_PROJECT")

				helper.ExpectString(ctx, "Enter value for \"DEBUG_PORT_PROJECT\" environment variable")
				helper.SendLine(ctx, "5858")

				helper.ExpectString(ctx, "What configuration do you want change")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Enter component name")

				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))

			// "execute odo dev and add changes to application"
			var devSession helper.DevSession

			devSession, err = helper.StartDevMode(helper.DevSessionOpts{
				CmdlineArgs: []string{"--debug"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(devSession.StdOut).ToNot(BeEmpty())

			waitRemoteApp(LocalAppURL, "Hello from Node.js Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello from Node.js Starter Application!")
			checkIfDevEnvIsUp(devSession.Endpoints[DebugPort], "WebSockets request was expected")

			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js", "from updated Node.js")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			waitRemoteApp(LocalAppURL, "Hello from updated Node.js Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello from updated Node.js Starter Application!")
			checkIfDevEnvIsUp(devSession.Endpoints[DebugPort], "WebSockets request was expected")

			// "changes are made to the applications"
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from updated Node.js", "from Node.js app v2")
			err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should deploy new changes"
			waitRemoteApp(LocalAppURL, "Hello from Node.js app v2 Starter Application!", "runtime")
			checkIfDevEnvIsUp(devSession.Endpoints[AppPort], "Hello from Node.js app v2 Starter Application!")
			checkIfDevEnvIsUp(devSession.Endpoints[DebugPort], "WebSockets request was expected")

			// "running odo list"
			stdout := helper.Cmd("odo", "list", "component").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "Node.js", "Dev"})

			// "exit dev mode"
			devSession.Stop()
			devSession.WaitEnd()

			// all resources should be deleted from the namespace
			services := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(services).To(BeEmpty())
			pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
			Expect(pvcs).To(BeEmpty())
			pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
			Expect(pods).To(BeEmpty())

		})
	})

	Context("starting with non-empty Directory add Binding", func() {
		sendDataEntry := func(url string) map[string]interface{} {
			values := map[string]interface{}{"name": "joe",
				"location": "tokyo",
				"age":      23,
			}
			json_data, err := json.Marshal(values)
			Expect(err).To(BeNil())
			resp, err := http.Post(fmt.Sprintf("http://%s/api/newuser", url), "application/json", bytes.NewBuffer(json_data))
			Expect(err).To(BeNil())
			var res map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&res)
			Expect(err).To(BeNil())
			return res
		}

		receiveData := func(url string) (string, error) {
			resp, err := http.Get(fmt.Sprintf("http://%s", url))
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			Expect(err).To(BeNil())
			return string(body), nil
		}

		var _ = BeforeEach(func() {
			commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
			commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
			Eventually(func() string {
				out, _ := commonVar.CliRunner.GetBindableKinds()
				return out
			}, 120, 3).Should(ContainSubstring("Cluster"))
			helper.Chdir(commonVar.Context)
			helper.CopyExample(filepath.Join("source", "devfiles", "go"), commonVar.Context)
			addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("source", "devfiles", "go", "cluster.yaml"))
			Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
		})

		It("should verify developer workflow of using binding as env in innerloop", func() {
			bindingName := helper.RandString(6)

			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				// helper.ExpectString(ctx, "Based on the files in the current directory odo detected")
				helper.ExpectString(ctx, "Language: Go")
				helper.ExpectString(ctx, "Project type: Go")
				helper.ExpectString(ctx, "Is this correct")

				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Select container for which you want to change configuration?")

				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "Enter component name")

				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))

			// // "execute odo dev and add changes to application"
			var devSession helper.DevSession

			devSession, err = helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())

			// "send data"
			_, err = receiveData(fmt.Sprintf(devSession.Endpoints["8080"] + "/api/user"))
			Expect(err).ToNot(BeNil()) // should fail as application is not connected to DB

			//add binding information (binding as ENV)
			helper.Cmd("odo", "add", "binding", "--name", bindingName, "--service", "cluster-example-initdb", "--bind-as-files=false").ShouldPass()

			// Get new random port after restart
			err = devSession.WaitRestartPortforward()
			Expect(err).ToNot(HaveOccurred())

			// "send data"
			waitRemoteApp("http://127.0.0.1:8080/ping", "pong", "runtime")
			data := sendDataEntry(devSession.Endpoints["8080"])
			Expect(data["message"]).To(Equal("User created successfully"))

			// "get all data"
			rec, err := receiveData(fmt.Sprintf(devSession.Endpoints["8080"] + "/api/user"))
			Expect(err).To(BeNil())
			helper.MatchAllInOutput(rec, []string{"id", "1", "name", "joe", "location", "tokyo", "age", "23"})

			// check odo describe to check for env
			stdout := helper.Cmd("odo", "describe", "binding").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{"Available binding information:", "CLUSTER_HOST", "CLUSTER_PASSWORD", "CLUSTER_USERNAME"})

			// "running odo list"
			stdout = helper.Cmd("odo", "list").ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{componentName, "Go", "Dev", bindingName})

			// remove bindings and check devfile to not contain binding info
			helper.Cmd("odo", "remove", "binding", "--name", bindingName).ShouldPass()

			err = devSession.WaitSync()
			Expect(err).To(BeNil())
			Eventually(func() string { return helper.Cmd("odo", "describe", "binding").ShouldRun().Out() }).
				WithTimeout(120 * time.Second).
				WithPolling(5 * time.Second).
				Should(ContainSubstring("No ServiceBinding used by the current component"))

			devSession.Stop()
			devSession.WaitEnd()

			// all resources should be deleted from the namespace
			services := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(services).NotTo(ContainSubstring(componentName))
			pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
			Expect(pvcs).NotTo(ContainElement(componentName)) //To(Not(ContainSubstring(componentName)))
			pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
			Expect(pods).NotTo(ContainElement(componentName))
		})
	})
})
