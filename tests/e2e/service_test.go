package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoServiceE2e", func() {

	Context("odo service creation", func() {
		It("should be able to create a service", func() {
			runCmd("odo service create mysql-persistent")
			cmd := "oc get serviceinstance mysql-persistent -o go-template='{{ (index .status.conditions 0).reason}}'"
			waitForServiceStatusCmd(cmd, "ProvisionedSuccessfully")
		})

		It("should be able to list the service", func() {
			out := runCmd("odo service list | sed 1d")
			Expect(out).To(ContainSubstring("mysql-persistent"))
			Expect(out).To(ContainSubstring("ProvisionedSuccessfully"))
		})

		It("should be able to delete a service", func() {
			runCmd("odo service delete mysql-persistent -f")
			cmd := "oc get serviceinstance mysql-persistent -o go-template='{{ (index .status.conditions 0).reason}}'"
			waitForServiceStatusCmd(cmd, "Deprovisioning")
		})
	})
})
