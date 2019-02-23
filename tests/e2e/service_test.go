package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("odoServiceE2e", func() {
	// Uncomment once service commands are made to use component config and also use context flags
	/*
		Context("odo service creation", func() {
			It("should be able to create a service", func() {
				runCmdShouldPass("odo service create mysql-persistent -w")
				waitForCmdOut("oc get serviceinstance -o name", 1, true, func(output string) bool {
					return strings.Contains(output, "mysql-persistent")
				})
				cmd := serviceInstanceStatusCmd("mysql-persistent")
				waitForServiceStatusCmd(cmd, "ProvisionedSuccessfully")
			})

			It("should be able to list the service with correct status", func() {
				waitForCmdOut("odo service list | sed 1d", 1, true, func(output string) bool {
					return strings.Contains(output, "mysql-persistent") &&
						strings.Contains(output, "ProvisionedAndBound")
				})
			})

			It("should be able to delete a service", func() {
				runCmdShouldPass("odo service delete mysql-persistent -f")
				cmd := serviceInstanceStatusCmd("mysql-persistent")
				waitForServiceStatusCmd(cmd, "Deprovisioning")
			})
		})

		//we only execute the rest of the tests if the RUN_ALL_SERVICE_TESTS env var is set to 'true'
		if strings.ToUpper(os.Getenv("RUN_ALL_SERVICE_TESTS")) != "TRUE" {
			fmt.Println("To run all service catalog tests make sure the 'RUN_ALL_SERVICE' is set to true")
		} else {
			Context("odo service create with a spring boot application", func() {
				It("should be able to create postgresql", func() {
					runCmdShouldPass("odo service create dh-postgresql-apb --plan dev -p postgresql_user=luke -p postgresql_password=secret -p postgresql_database=my_data -p postgresql_version=9.6")
					waitForCmdOut("oc get serviceinstance -o name", 1, true, func(output string) bool {
						return strings.Contains(output, "dh-postgresql-apb")
					})
				})

				It("Should be able to deploy an openjdk source application", func() {
					importOpenJDKImage()

					runCmdShouldPass("odo create openjdk18 sb-app --local " + sourceExamples + "/openjdk-sb-postgresql/")

					// Push changes
					runCmdShouldPass("odo push")

					// Create a URL
					runCmdShouldPass("odo url create --port 8080")
				})

				It("Should be able to link the spring boot application to the postgresql DB", func() {
					runCmdShouldPass("odo link dh-postgresql-apb -w --wait-for-target")

					waitForCmdOut("odo service list | sed 1d", 1, true, func(output string) bool {
						return strings.Contains(output, "dh-postgresql-apb") &&
							strings.Contains(output, "ProvisionedAndLinked")
					})
				})

				It("The application should respond successfully", func() {
					routeURL := determineRouteURL()

					// Ping said URL
					responseStringMatchStatus := matchResponseSubString(routeURL, "Spring Boot", 30, 1)
					Expect(responseStringMatchStatus).Should(BeTrue())
				})

				It("Should be able to delete everything", func() {
					// Delete the component
					runCmdShouldPass("odo delete sb-app -f")

					// Delete the service
					runCmdShouldPass("odo service delete dh-postgresql-apb -f")
				})
			})

			Context("odo hides a hidden service in service catalog", func() {
				It("not show a hidden service in the catalog", func() {
					runCmdShouldPass("oc apply -f https://github.com/openshift/library/raw/master/official/sso/templates/sso72-https.json -n openshift")
					outputErr := runCmdShouldFail("odo catalog search service sso72-https")
					Expect(outputErr).To(ContainSubstring("No service matched the query: sso72-https"))
				})
			})
		}
	*/
})

func serviceInstanceStatusCmd(serviceInstanceName string) string {
	return fmt.Sprintf("oc get serviceinstance %s -o go-template='{{ (index .status.conditions 0).reason}}'", serviceInstanceName)
}
