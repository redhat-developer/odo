package devfile

/*
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/tests/helper"
	"github.com/redhat-developer/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile push command tests", func() {
	var cmpName string
	var sourcePath = "/projects"
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("creating a nodejs component", func() {
		output := ""
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})


		When("doing odo push and run command is not marked as hotReloadCapable", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push").ShouldPass().Out()
			})
			It("should restart the application", func() {
				// TODO: this is almost the same test as one below

				helper.Cmd("odo", "push", "-f").ShouldPass()
			})
		})

			It("should correctly execute PostStart commands", func() {

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()

				// Need to force so build and run get triggered again with the component already created.
				helper.Cmd("odo", "push", "--project", commonVar.Project, "-f").ShouldPass()
			})
		})
	})
})
*/
