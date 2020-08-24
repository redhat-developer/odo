package e2escenarios

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo core beta flow", func() {
	var oc helper.OcRunner
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		// initialize oc runner
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// abstract main test to the function, to allow running the same test in a different context (slightly different arguments)
	TestBasicCreateConfigPush := func(extraArgs ...string) {
		createSession := helper.CmdShouldPass("odo", append([]string{"component", "create", "java:8", "mycomponent", "--app", "myapp", "--project", commonVar.Project}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		Expect(createSession).Should(ContainSubstring("odo push"))
		configFile := filepath.Join(commonVar.Context, ".odo", "config.yaml")
		Expect(configFile).To(BeARegularFile())
		helper.FileShouldContainSubstring(configFile, "Name: mycomponent")
		helper.FileShouldContainSubstring(configFile, "Type: java")
		helper.FileShouldContainSubstring(configFile, "Application: myapp")
		helper.FileShouldContainSubstring(configFile, "SourceType: local")
		// SourcePath should be relative
		//helper.FileShouldContainSubstring(configFile, "SourceLocation: .")
		helper.FileShouldContainSubstring(configFile, "Project: "+commonVar.Project)

		configSession := helper.CmdShouldPass("odo", append([]string{"config", "set", "--env", "FOO=bar"}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		// currently failing
		Expect(configSession).Should(ContainSubstring("odo push"))
		helper.FileShouldContainSubstring(configFile, "Name: FOO")
		helper.FileShouldContainSubstring(configFile, "Value: bar")

		urlCreateSession := helper.CmdShouldPass("odo", append([]string{"url", "create", "--port", "8080"}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		Eventually(urlCreateSession).Should(ContainSubstring("odo push"))
		helper.FileShouldContainSubstring(configFile, "Url:")
		helper.FileShouldContainSubstring(configFile, "Port: 8080")

		helper.CmdShouldPass("odo", append([]string{"push"}, extraArgs...)...)

		dcSession := oc.GetComponentDC("mycomponent", "myapp", commonVar.Project)
		helper.MatchAllInOutput(dcSession, []string{
			"app.kubernetes.io/instance: mycomponent",
			"app.kubernetes.io/component-source-type: local",
			"app.kubernetes.io/name: java",
			"app.kubernetes.io/part-of: myapp",
			"name: mycomponent-myapp",
		})

		// DC should have env variable
		helper.MatchAllInOutput(dcSession, []string{"name: FOO", "value: bar"})

		routeSession := oc.GetComponentRoutes("mycomponent", "myapp", commonVar.Project)
		// check that route is pointing gto right port and component
		helper.MatchAllInOutput(routeSession, []string{"targetPort: 8080", "name: mycomponent-myapp"})
		url := oc.GetFirstURL("mycomponent", "myapp", commonVar.Project)
		helper.HttpWaitFor("http://"+url, "Hello World from Javalin!", 10, 5)
	}

	Context("when component is in the current directory", func() {
		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("'odo component' should fail if there already is .odo dir", func() {
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "--project", commonVar.Project)
			helper.CmdShouldFail("odo", "component", "create", "nodejs", "--project", commonVar.Project)
		})

		It("'odo config' should fail if there is no .odo dir", func() {
			helper.CmdShouldFail("odo", "config", "set", "memory", "2Gi")
		})

		It("create local java component and push code", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("source", "openjdk"), commonVar.Context)
			TestBasicCreateConfigPush()
		})
	})

	Context("when --context flag is used", func() {
		It("odo component should fail if there already is .odo dir", func() {
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project)
			helper.CmdShouldFail("odo", "component", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project)
		})

		It("odo config should fail if there is no .odo dir", func() {
			helper.CmdShouldFail("odo", "config", "set", "memory", "2Gi", "--context", commonVar.Context)
		})

		It("create local java component and push code", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("source", "openjdk"), commonVar.Context)
			TestBasicCreateConfigPush("--context", commonVar.Context)
		})
	})
})
