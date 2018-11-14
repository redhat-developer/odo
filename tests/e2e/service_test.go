package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"strings"
)

var _ = Describe("odoServiceE2e", func() {

	Context("odo service creation", func() {
		It("should be able to create a service", func() {
			runCmd("odo service create mysql-persistent")
			waitForCmdOut("oc get serviceinstance -o name", 1, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			cmd := serviceInstanceStatusCmd("mysql-persistent")
			waitForServiceStatusCmd(cmd, "ProvisionedSuccessfully")
		})

		It("should be able to list the service with correct status", func() {
			waitForCmdOut("odo service list | sed 1d", 1, func(output string) bool {
				return strings.Contains(output, "mysql-persistent") &&
					strings.Contains(output, "ProvisionedAndBound")
			})
		})

		It("should be able to delete a service", func() {
			runCmd("odo service delete mysql-persistent -f")
			cmd := serviceInstanceStatusCmd("mysql-persistent")
			waitForServiceStatusCmd(cmd, "Deprovisioning")
		})
	})
})

func serviceInstanceStatusCmd(serviceInstanceName string) string {
	return fmt.Sprintf("oc get serviceinstance %s -o go-template='{{ (index .status.conditions 0).reason}}'", serviceInstanceName)
}
