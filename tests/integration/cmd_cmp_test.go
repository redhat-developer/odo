// +build !race

package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/tests/helper"
)

func componentTestsNoSub() {
	componentTests()
	Context("when components are not created/managed by odo", func() {
		var commonVar helper.CommonVar

		JustBeforeEach(func() {
			/*
				Copy the deployment-label.yaml and dc-label.yaml file.
				Create a deployment with deployment-label.yaml.
				Create a dc with dc-label.yaml.
			*/
			commonVar = helper.CommonBeforeEach()
			helper.CopyManifestFile("dc-label.yaml", filepath.Join(commonVar.Context, "dc-label.yaml"))
			helper.CopyManifestFile("deployment-label.yaml", filepath.Join(commonVar.Context, "deployment-label.yaml"))
			helper.CmdShouldPass("oc", "apply", "-f", "dc-label.yaml")
			helper.CmdShouldPass("oc", "apply", "-f", "deployment-label.yaml")
		})
		JustAfterEach(func() {
			//DeleteNonOdoComponent
			client, _ := genericclioptions.Client()
			label := map[string]string{
				labels.ManagedBy: "!odo",
			}
			Expect(client.Delete(label, true)).To(BeNil())
			Expect(client.GetKubeClient().DeleteDeployment(label)).To(BeNil())
			helper.CommonAfterEach(commonVar)
		})
		FIt("should list the components", func() {
			output := helper.CmdShouldPass("odo", "list")
			Expect(output).To(ContainSubstring("Other Components running on the cluster(read-only)"))
		})
		It("should list the components with --all-apps flag", func() {
			output := helper.CmdShouldPass("odo", "list", "--all-apps")
			Expect(output).To(ContainSubstring("Other Components running on the cluster(read-only)"))

		})
		It("should list the components with --app flag", func() {
			output := helper.CmdShouldPass("odo", "list", "--app", "httpd")
			Expect(output).To(ContainSubstring("Other Components running on the cluster(read-only)"))
		})
		It("should list the components in json format with -o json flag", func() {
			output := helper.CmdShouldPass("odo", "list", "-o", "json")
			Expect(output).To(ContainSubstring("Other Components running on the cluster(read-only)"))
		})
		When("executing odo list from other project", func() {
			JustBeforeEach(func() {
				helper.CmdShouldPass("odo", "project", "set", "default")
			})
			JustAfterEach(func() {
				helper.CmdShouldPass("odo", "project", "set", commonVar.Project)
			})
			It("should list the components with --project flag", func() {
				output := helper.CmdShouldPass("odo", "list", "--project", commonVar.Project)
				Expect(output).To(ContainSubstring("Other Components running on the cluster(read-only)"))
			})

		})
	})
}

var _ = Describe("odo component command tests", componentTestsNoSub)
