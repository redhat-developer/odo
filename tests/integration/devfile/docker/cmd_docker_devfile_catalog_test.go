package docker

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo docker devfile catalog command tests", func() {
	var context string
	var currentWorkingDirectory string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
	})

	Context("When executing catalog list components on Docker", func() {
		It("should list all supported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-springboot", "java-openliberty", "DefaultDevfileRegistry"})
			helper.DontMatchAllInOutput(output, []string{"SUPPORTED", "Odo OpenShift Components"})
		})
	})

	Context("When executing catalog list components with -o json flag", func() {
		It("should list devfile components in json format", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-o", "json")
			wantOutput := []string{
				"odo.dev/v1alpha1",
				"devfileItems",
				"java-openliberty",
				"java-springboot",
				"nodejs",
				"java-quarkus",
				"java-maven",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing catalog list components with registry that is not set up properly", func() {
		It("should list components from valid registry", func() {
			helper.CmdShouldPass("odo", "registry", "add", "fake", "http://fake")
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			helper.MatchAllInOutput(output, []string{
				"Odo Devfile Components",
				"java-springboot",
				"java-quarkus",
			})
			helper.CmdShouldPass("odo", "registry", "delete", "fake", "-f")
		})
	})
})
