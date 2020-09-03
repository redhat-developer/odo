package e2escenarios

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo core beta flow", func() {
	//new clean project and context for each test
	var project string
	var context string

	//  current directory and project (before any test is run) so it can restored  after all testing is done
	var originalDir string

	var oc helper.OcRunner
	// path to odo binary
	var odo string

	BeforeEach(func() {
		// Set default timeout for Eventually assertions
		// commands like odo push, might take a long time
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)

		// initialize oc runner
		// right now it uses oc binary, but we should convert it to client-go
		oc = helper.NewOcRunner("oc")
		odo = "odo"

		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	// abstract main test to the function, to allow running the same test in a different context (slightly different arguments)
	TestBasicCreateConfigPush := func(extraArgs ...string) {
		createSession := helper.CmdShouldPass(odo, append([]string{"component", "create", "--s2i", "java:8", "mycomponent", "--app", "myapp", "--project", project}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		Expect(createSession).Should(ContainSubstring("odo push"))
		configFile := filepath.Join(context, ".odo", "config.yaml")
		Expect(configFile).To(BeARegularFile())
		helper.FileShouldContainSubstring(configFile, "Name: mycomponent")
		helper.FileShouldContainSubstring(configFile, "Type: java")
		helper.FileShouldContainSubstring(configFile, "Application: myapp")
		helper.FileShouldContainSubstring(configFile, "SourceType: local")
		// SourcePath should be relative
		//helper.FileShouldContainSubstring(configFile, "SourceLocation: .")
		helper.FileShouldContainSubstring(configFile, "Project: "+project)

		configSession := helper.CmdShouldPass(odo, append([]string{"config", "set", "--env", "FOO=bar"}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		// currently failing
		Expect(configSession).Should(ContainSubstring("odo push"))
		helper.FileShouldContainSubstring(configFile, "Name: FOO")
		helper.FileShouldContainSubstring(configFile, "Value: bar")

		urlCreateSession := helper.CmdShouldPass(odo, append([]string{"url", "create", "--port", "8080"}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		Eventually(urlCreateSession).Should(ContainSubstring("odo push"))
		helper.FileShouldContainSubstring(configFile, "Url:")
		helper.FileShouldContainSubstring(configFile, "Port: 8080")

		helper.CmdShouldPass(odo, append([]string{"push"}, extraArgs...)...)

		dcSession := oc.GetComponentDC("mycomponent", "myapp", project)
		helper.MatchAllInOutput(dcSession, []string{
			"app.kubernetes.io/instance: mycomponent",
			"app.kubernetes.io/component-source-type: local",
			"app.kubernetes.io/name: java",
			"app.kubernetes.io/part-of: myapp",
			"name: mycomponent-myapp",
		})

		// DC should have env variable
		helper.MatchAllInOutput(dcSession, []string{"name: FOO", "value: bar"})

		routeSession := oc.GetComponentRoutes("mycomponent", "myapp", project)
		// check that route is pointing gto right port and component
		helper.MatchAllInOutput(routeSession, []string{"targetPort: 8080", "name: mycomponent-myapp"})
		url := oc.GetFirstURL("mycomponent", "myapp", project)
		helper.HttpWaitFor("http://"+url, "Hello World from Javalin!", 10, 5)
	}

	Context("when component is in the current directory", func() {
		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})
		// go back to original directory after each test
		JustAfterEach(func() {
			helper.Chdir(originalDir)
		})

		It("'odo component' should fail if there already is .odo dir", func() {
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "--project", project)
			helper.CmdShouldFail("odo", "component", "create", "--s2i", "nodejs", "--project", project)
		})

		It("'odo config' should fail if there is no .odo dir", func() {
			helper.CmdShouldFail("odo", "config", "set", "memory", "2Gi")
		})

		It("create local java component and push code", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("source", "openjdk"), context)
			TestBasicCreateConfigPush()
		})
	})

	Context("when --context flag is used", func() {
		It("odo component should fail if there already is .odo dir", func() {
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "--context", context, "--project", project)
			helper.CmdShouldFail("odo", "component", "create", "--s2i", "nodejs", "--context", context, "--project", project)
		})

		It("odo config should fail if there is no .odo dir", func() {
			helper.CmdShouldFail("odo", "config", "set", "memory", "2Gi", "--context", context)
		})

		It("create local java component and push code", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("source", "openjdk"), context)
			TestBasicCreateConfigPush("--context", context)
		})
	})
})
