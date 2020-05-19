package e2escenarios

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo core beta flow", func() {
	var oc helper.OcRunner
	// path to odo binary
	var odo string
	var globals helper.Globals

	BeforeEach(func() {
		globals = helper.CommonBeforeEach()

		// initialize oc runner
		// right now it uses oc binary, but we should convert it to client-go
		oc = helper.NewOcRunner("oc")
		odo = "odo"
	})

	AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	// abstract main test to the function, to allow running the same test in a different context (slightly different arguments)
	TestBasicCreateConfigPush := func(extraArgs ...string) {
		createSession := helper.CmdShouldPass(odo, append([]string{"component", "create", "java:8", "mycomponent", "--app", "myapp", "--project", globals.Project}, extraArgs...)...)
		// output of the commands should point user to running "odo push"
		Expect(createSession).Should(ContainSubstring("odo push"))
		configFile := filepath.Join(globals.Context, ".odo", "config.yaml")
		Expect(configFile).To(BeARegularFile())
		helper.FileShouldContainSubstring(configFile, "Name: mycomponent")
		helper.FileShouldContainSubstring(configFile, "Type: java")
		helper.FileShouldContainSubstring(configFile, "Application: myapp")
		helper.FileShouldContainSubstring(configFile, "SourceType: local")
		// SourcePath should be relative
		//helper.FileShouldContainSubstring(configFile, "SourceLocation: .")
		helper.FileShouldContainSubstring(configFile, "Project: "+globals.Project)

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

		dcSession := oc.GetComponentDC("mycomponent", "myapp", globals.Project)
		Expect(dcSession).Should(ContainSubstring("app.kubernetes.io/instance: mycomponent"))
		Expect(dcSession).Should(ContainSubstring("app.kubernetes.io/component-source-type: local"))
		Expect(dcSession).Should(ContainSubstring("app.kubernetes.io/name: java"))
		Expect(dcSession).Should(ContainSubstring("app.kubernetes.io/part-of: myapp"))
		Expect(dcSession).Should(ContainSubstring("name: mycomponent-myapp"))
		// DC should have env variable
		Expect(dcSession).Should(ContainSubstring("name: FOO"))
		Expect(dcSession).Should(ContainSubstring("value: bar"))

		routeSession := oc.GetComponentRoutes("mycomponent", "myapp", globals.Project)
		// check that route is pointing gto right port and component
		Expect(routeSession).Should(ContainSubstring("targetPort: 8080"))
		Expect(routeSession).Should(ContainSubstring("name: mycomponent-myapp"))
		url := oc.GetFirstURL("mycomponent", "myapp", globals.Project)
		helper.HttpWaitFor("http://"+url, "Hello World from Javalin!", 10, 5)
	}

	Context("when component is in the current directory", func() {
		JustBeforeEach(func() {
			helper.Chdir(globals.Context)
		})

		It("'odo component' should fail if there already is .odo dir", func() {
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "--project", globals.Project)
			helper.CmdShouldFail("odo", "component", "create", "nodejs", "--project", globals.Project)
		})

		It("'odo config' should fail if there is no .odo dir", func() {
			helper.CmdShouldFail("odo", "config", "set", "memory", "2Gi")
		})

		It("create local java component and push code", func() {
			oc.ImportJavaIS(globals.Project)
			helper.CopyExample(filepath.Join("source", "openjdk"), globals.Context)
			TestBasicCreateConfigPush()
		})
	})

	Context("when --context flag is used", func() {
		It("odo component should fail if there already is .odo dir", func() {
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "--context", globals.Context, "--project", globals.Project)
			helper.CmdShouldFail("odo", "component", "create", "nodejs", "--context", globals.Context, "--project", globals.Project)
		})

		It("odo config should fail if there is no .odo dir", func() {
			helper.CmdShouldFail("odo", "config", "set", "memory", "2Gi", "--context", globals.Context)
		})

		It("create local java component and push code", func() {
			oc.ImportJavaIS(globals.Project)
			helper.CopyExample(filepath.Join("source", "openjdk"), globals.Context)
			TestBasicCreateConfigPush("--context", globals.Context)
		})
	})
})
