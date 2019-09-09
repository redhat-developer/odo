package benchmark

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("Basic benchmark", func() {

	// how many times each benchmark will be executed
	const numberOfRuns int = 1

	// new clean project and context for each test
	var project string
	var context string

	// store current directory and project (before any test is run) so it can restored after all testing is done
	var originalDir string
	var originalProject string

	var oc helper.OcRunner
	// path to odo binary
	var odo string

	BeforeEach(func() {
		// Set default timeout for Eventually assertions
		// commands like odo push, might take a long time
		SetDefaultEventuallyTimeout(10 * time.Minute)

		// initialize oc runner
		// right now it uses oc binary, but we should convert it to client-go
		oc = helper.NewOcRunner("oc")
		odo = "odo"

		originalProject = oc.GetCurrentProject()
		originalDir = helper.Getwd()

		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		oc.SwitchProject(project)
		helper.Chdir(context)

	})

	AfterEach(func() {
		oc.SwitchProject(originalProject)
		helper.Chdir(originalDir)
		helper.DeleteProject(project)
		helper.DeleteDir(context)
	})

	Measure("Simple Java (Javalin) component", func(b Benchmarker) {
		helper.CopyExample(filepath.Join("source", "openjdk"), context)
		oc.ImportJavaIS(project)

		b.Time("create component", func() {
			helper.CmdShouldPass(odo, "component", "create", "java:8", "javacomponent", "--app", "myapp")
		})
		b.Time("crate url", func() {
			helper.CmdShouldPass(odo, "url", "create", "--port", "8080")
		})
		b.Time("first time push", func() {
			helper.CmdShouldPass(odo, "push")
		})

		url := oc.GetFirstURL("javacomponent", "myapp", project)

		b.Time("running app after push", func() {
			helper.HttpWaitFor("http://"+url, "Hello World from Javalin!", 60, 1)
		})

		helper.ReplaceString("./src/main/java/MessageProducer.java", "Hello World from Javalin!", "UPDATED!")

		b.Time("push after file change", func() {
			helper.CmdShouldPass(odo, "push")
		})
		b.Time("change in app after push", func() {
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)
		})
	}, numberOfRuns)

	Measure("Simple NodeJS component", func(b Benchmarker) {
		helper.CopyExample(filepath.Join("source", "nodejs"), context)

		b.Time("create component", func() {
			helper.CmdShouldPass(odo, "component", "create", "nodejs", "nodejscomponent", "--app", "myapp")
		})
		b.Time("crate url", func() {
			helper.CmdShouldPass(odo, "url", "create", "--port", "8080")
		})
		b.Time("first time push", func() {
			helper.CmdShouldPass(odo, "push")
		})

		url := oc.GetFirstURL("nodejscomponent", "myapp", project)

		b.Time("running app after push", func() {
			helper.HttpWaitFor("http://"+url, "Hello world from node.js!", 60, 1)
		})

		helper.ReplaceString("./server.js", "Hello world from node.js!", "UPDATED!")

		b.Time("push after file change", func() {
			helper.CmdShouldPass(odo, "push")
		})
		b.Time("change in app after push", func() {
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)
		})
	}, numberOfRuns)

})
