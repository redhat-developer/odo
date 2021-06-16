package devfile

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile app command tests", func() {

	var namespace string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		namespace = commonVar.Project
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("listing apps", func() {
		It("it should list, describe and delete the apps", func() {
			runner(namespace, false)
		})
	})

	Context("Testing URLs for OpenShift specific scenarios", func() {
		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})

		It("it should list, describe and delete the apps", func() {
			runner(namespace, true)
		})
	})
})

func runner(namespace string, s2i bool) {
	context0 := helper.CreateNewContext()
	context00 := helper.CreateNewContext()
	context1 := helper.CreateNewContext()

	defer func() {
		helper.DeleteDir(context0)
		helper.DeleteDir(context00)
		helper.DeleteDir(context1)
	}()

	app0 := helper.RandString(4)
	app1 := helper.RandString(4)

	component0 := helper.RandString(4)
	component00 := helper.RandString(4)
	component1 := helper.RandString(4)

	storage00 := helper.RandString(4)
	url00 := helper.RandString(4)

	helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, component0, "--context", context0, "--app", app0)
	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context0)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context0, "devfile.yaml"))
	helper.CmdShouldPass("odo", "push", "--context", context0)

	if s2i {
		helper.CopyExample(filepath.Join("source", "nodejs"), context00)
		helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", component00, "--app", app0, "--project", namespace, "--context", context00)
		helper.CmdShouldPass("odo", "storage", "create", storage00, "--path", "/data", "--size", "1Gi", "--context", context00)
		helper.CmdShouldPass("odo", "url", "create", url00, "--port", "8080", "--context", context00)
	} else {
		helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, component00, "--context", context00, "--app", app0)
		helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context00)
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context00, "devfile.yaml"))
		helper.CmdShouldPass("odo", "storage", "create", storage00, "--path", "/data", "--size", "1Gi", "--context", context00)
		helper.CmdShouldPass("odo", "url", "create", url00, "--port", "3000", "--context", context00, "--host", "com", "--ingress")
	}
	helper.CmdShouldPass("odo", "push", "--context", context00)

	helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, component1, "--context", context1, "--app", app1)
	helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context1)
	helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context1, "devfile.yaml"))
	helper.CmdShouldPass("odo", "push", "--context", context1)

	stdOut := helper.CmdShouldPass("odo", "app", "list", "--project", namespace)
	helper.MatchAllInOutput(stdOut, []string{app0, app1})

	// test the json output
	stdOut = helper.CmdShouldPass("odo", "app", "list", "--project", namespace, "-o", "json")
	helper.MatchAllInOutput(stdOut, []string{app0, app1})
	Expect(helper.IsJSON(stdOut)).To(BeTrue())

	stdOut = helper.CmdShouldPass("odo", "app", "describe", app0, "--project", namespace)
	helper.MatchAllInOutput(stdOut, []string{app0, component0, component00, storage00, url00})

	// test the json output
	stdOut = helper.CmdShouldPass("odo", "app", "describe", app0, "--project", namespace, "-o", "json")
	helper.MatchAllInOutput(stdOut, []string{app0, component0, component00})
	Expect(helper.IsJSON(stdOut)).To(BeTrue())

	stdOut = helper.CmdShouldPass("odo", "app", "delete", app0, "--project", namespace, "-f")
	helper.MatchAllInOutput(stdOut, []string{app0, component0, component00, url00, storage00})

	stdOut = helper.CmdShouldPass("odo", "app", "list", "--project", namespace)
	helper.MatchAllInOutput(stdOut, []string{app1})
}
