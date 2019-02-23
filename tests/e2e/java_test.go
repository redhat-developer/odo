package e2e

import (
	//"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const javaFiles = "examples/binary/java/"

var _ = Describe("odoJavaE2e", func() {
	//const t = "java"
	//var projName = fmt.Sprintf("odo-%s", t)

	// contains a minimal javaee app
	const warGitRepo = "https://github.com/lordofthejars/book-insultapp"

	// contains a minimal javalin app
	const jarGitRepo = "https://github.com/geoand/javalin-helloworld"

	// Create a separate project for Java
	Context("create java project", func() {
		It("should create a new java project", func() {
			session := runCmdShouldPass("odo project create odo-java")
			Expect(session).To(ContainSubstring("odo-java"))
			// waitForCmdOut("odo project set "+projName, 4, false, func(output string) bool {
			// 	return strings.Contains(output, "Already on project : "+projName)
			// })
		})
	})

	// Test Java
	Context("odo component creation", func() {
		// It("Should be able to deploy a git repo that contains a wildfly application", func() {

		// 	// Deploy the git repo / wildfly example
		// 	cmpCreateLog := runCmdShouldPass("odo create wildfly javaee-git-test --git " + warGitRepo + " -w")
		// 	Expect(cmpCreateLog).ShouldNot(ContainSubstring("This may take a few moments to be ready"))
		// 	buildName := getBuildName("javaee-git-test")
		// 	Expect(buildName).To(ContainSubstring("javaee-git-test"))
		// 	buildStatus := runCmdShouldPass("oc get build " + buildName)
		// 	Expect(buildStatus).To(ContainSubstring("Complete"))

		// 	cmpList := runCmdShouldPass("odo list")
		// 	Expect(cmpList).To(ContainSubstring("javaee-git-test"))

		// 	// Push changes
		// 	runCmdShouldPass("odo push")

		// 	// Create a URL
		// 	runCmdShouldPass("odo url create")
		// 	routeURL := determineRouteURL()

		// 	// Ping said URL
		// 	responsePing := matchResponseSubString(routeURL, "Insult", 90, 1)
		// 	Expect(responsePing).Should(BeTrue())

		// 	// Delete the component
		// 	runCmdShouldPass("odo delete javaee-git-test -f")
		// })

		It("Should be able to deploy a git repo that contains a wildfly application without wait flag", func() {
			// Deploy the git repo / wildfly example
			runCmdShouldPass("odo create wildfly wo-wait-javaee-git-test --project odo-java --git " + warGitRepo)

			// buildName := getBuildName("wo-wait-javaee-git-test")
			// Expect(buildName).To(ContainSubstring("wo-wait-javaee-git-test"))
			// buildStatus := getBuildParameterValues("wo-wait-javaee-git-test")
			// Expect(buildStatus).To(ContainSubstring("Pending"))

			// dcName := getDcName("wo-wait-javaee-git-test")
			// // For waiting until the deployment starts
			// for {
			// 	time.Sleep(5 * time.Second)
			// 	dcStatus := getDcStatusValue("wo-wait-javaee-git-test")
			// 	if dcStatus == "1" {
			// 		break
			// 	}
			// }
			// // following the logs and waiting for the build to finish
			// runCmdShouldPass("oc logs --version=1 dc/" + dcName)

			// cmpList := runCmdShouldPass("odo list")
			// Expect(cmpList).To(ContainSubstring("wo-wait-javaee-git-test"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create")
			routeURL := determineRouteURL()

			// Ping said URL
			responsePing := matchResponseSubString(routeURL, "Insult", 90, 1)
			Expect(responsePing).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo delete wo-wait-javaee-git-test -f")
		})

		It("Should be able to deploy a .war file using wildfly", func() {
			runCmdShouldPass("odo create wildfly javaee-war-test --project odo-java --binary " + javaFiles + "/wildfly/ROOT.war")

			// dcName := getDcName("javaee-war-test")

			// // Following the logs
			// runCmdShouldPass("oc logs --version=1 dc/" + dcName)

			// cmpList := runCmdShouldPass("odo list")
			// Expect(cmpList).To(ContainSubstring("javaee-war-test"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create")
			routeURL := determineRouteURL()

			// Ping said URL
			responsePing := matchResponseSubString(routeURL, "Sample", 90, 1)
			Expect(responsePing).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo delete javaee-war-test -f")
		})

		It("Should be able to deploy a git repo that contains a java uberjar application using openjdk", func() {
			importOpenJDKImage()

			// Deploy the git repo / wildfly example
			runCmdShouldPass("odo create openjdk18 uberjar-git-test --project odo-java --git " + jarGitRepo)
			// buildName := getBuildName("uberjar-git-test")
			// Expect(buildName).To(ContainSubstring("uberjar-git-test"))
			// buildStatus := runCmdShouldPass("oc get build " + buildName)
			// Expect(buildStatus).To(ContainSubstring("Complete"))

			// cmpList := runCmdShouldPass("odo list")
			// Expect(cmpList).To(ContainSubstring("uberjar-git-test"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			responsePing := matchResponseSubString(routeURL, "Hello World", 90, 1)
			Expect(responsePing).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo delete uberjar-git-test -f")
		})

		It("Should be able to deploy a spring boot uberjar file using openjdk", func() {
			importOpenJDKImage()

			runCmdShouldPass("odo create openjdk18 sb-jar-test --project odo-java --binary " + javaFiles + "/openjdk/sb.jar")

			// dcName := getDcName("sb-jar-test")
			// Expect(dcName).To(ContainSubstring("sb-jar-test"))

			// cmpList := runCmdShouldPass("odo list")
			// Expect(cmpList).To(ContainSubstring("sb-jar-test"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			responsePing := matchResponseSubString(routeURL, "HTTP Booster", 90, 1)
			Expect(responsePing).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo delete sb-jar-test -f")
		})

	})

	// Delete the project
	Context("java project delete", func() {
		It("should delete java project", func() {
			odoDeleteProject("odo-java")
		})
	})
})

func importOpenJDKImage() {
	// we need to import the openjdk image which is used for jars because it's not available by default
	runCmdShouldPass("oc import-image openjdk18 --from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift:1.5 --confirm")
	runCmdShouldPass("oc annotate istag/openjdk18:latest tags=builder --overwrite")
}
