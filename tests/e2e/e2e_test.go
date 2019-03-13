// +build !race

package e2e

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TODO: A neater way to provide odo path. Currently we assume \
// odo and oc in $PATH already.
var curProj string
var testNamespacedImage = "https://raw.githubusercontent.com/bucharest-gold/centos7-s2i-nodejs/master/imagestreams/nodejs-centos7.json"

// EnvVarTest checks the component container env vars in the build config for git and deployment config for git/binary/local
// appTestName is the app of the app
// sourceType is the type of the source of the component i.e git/binary/local
func EnvVarTest(resourceName string, sourceType string, envString string) {

	if sourceType == "git" {
		// checking the values of the env vars pairs in bc
		envVars := runCmdShouldPass("oc get bc " + resourceName + " -o go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
		Expect(envVars).To(Equal(envString))
	}

	// checking the values of the env vars pairs in dc
	envVars := runCmdShouldPass("oc get dc " + resourceName + " -o go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
	Expect(envVars).To(Equal(envString))
}

func TestOdo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "odo test suite")
}

var _ = BeforeSuite(func() {
	// Save the current project
	curProj = runCmdShouldPass("oc project -q")
})

var _ = Describe("odoe2e", func() {
	projName := generateTimeBasedName("odo")
	const appTestName = "testing"
	const loginTestUserPassword = "developer"

	tmpDir, err := ioutil.TempDir("", "odo")
	if err != nil {
		Fail(err.Error())
	}

	Describe("Check for failure if user tries to create or delete anything other than project, without active accessible project, with appropriate message", func() {
		var currentUserToken string
		Context("Knows who is currently logged in", func() {
			It("Should save token of current user to be able to log back in", func() {
				currentUserToken = runCmdShouldPass("oc whoami -t")
			})
		})

		Context("Logs in an new user without active project and tries to create various objects", func() {
			It("Login as test user without project", func() {
				runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", "odonoprojectattemptscreate", loginTestUserPassword))
			})

			It("Should fail if user tries to create anything other than project", func() {
				runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", "odonoprojectattemptscreate", loginTestUserPassword))
				session := runCmdShouldFail("odo create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'default' or it doesnt exist."))
				// The message should also give user apropriate command
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo component create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'default' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo application create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'default' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo application create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'default' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi")
				Expect(session).To(ContainSubstring("You dont have permission to project 'default' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
			})

			It("Should pass if user tries to create a project", func() {
				session := runCmdShouldPass("odo project create odonoprojectattemptscreateproject")
				Expect(session).To(ContainSubstring("New project created and now using project"))
				Expect(session).To(ContainSubstring("odonoprojectattemptscreateproject"))
				odoDeleteProject("odonoprojectattemptscreateproject")
			})
		})

		Context("Logs in as user with a project, deletes it and tries to create various objects", func() {
			It("Should login as a user and setup by creating a project, and then deleting it", func() {
				runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", "odosingleprojectattemptscreate", loginTestUserPassword))
				odoCreateProject("odosingleprojectattemptscreateproject")
				odoDeleteProject("odosingleprojectattemptscreateproject")
			})

			It("Should fail if user tries to create any object, other than project", func() {
				session := runCmdShouldFail("odo create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'odosingleprojectattemptscreateproject' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo component create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'odosingleprojectattemptscreateproject' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo application create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'odosingleprojectattemptscreateproject' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo application create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'odosingleprojectattemptscreateproject' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				session = runCmdShouldFail("odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi")
				Expect(session).To(ContainSubstring("You dont have permission to project 'odosingleprojectattemptscreateproject' or it doesnt exist."))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
			})

			It("Should pass if user tries to create a project", func() {
				session := runCmdShouldPass("odo project create odosingleprojectattemptscreateproject")
				Expect(session).To(ContainSubstring("New project created and now using project"))
				Expect(session).To(ContainSubstring("odosingleprojectattemptscreateproject"))
				odoDeleteProject("odosingleprojectattemptscreateproject")
			})
		})

		Context("Log back in as old user", func() {
			It("Should log back in as old user", func() {
				runCmdShouldPass(fmt.Sprintf("oc login --token %s", currentUserToken))
			})
		})
	})

	Context("odo service create", func() {
		It("should return error if the cluster has no service catalog deployed", func() {
			loginOutput := runCmdShouldPass("odo login --username developer --password developer")
			Expect(loginOutput).To(ContainSubstring("Login successful"))
			sessionErrOutput := runCmdShouldFail("odo service create")
			Expect(sessionErrOutput).To(ContainSubstring("unable to retrieve service classes"))
		})
	})

	// TODO: Create component without creating application
	Context("odo project", func() {
		It("should create a new project", func() {
			session := runCmdShouldPass("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})

		It("should get the project", func() {
			getProj := runCmdShouldPass("odo project get --short")
			Expect(getProj).To(Equal(projName))
		})

		// Issue #630
		It("should list the project", func() {
			listProj := runCmdShouldPass("sleep 5s && odo project list")
			Expect(listProj).To(ContainSubstring(projName))
		})
	})

	Context("odo config", func() {
		It("should get the default global config keys", func() {
			configOutput := runCmdShouldPass("odo preference view")
			Expect(configOutput).To(ContainSubstring("UpdateNotification"))
			Expect(configOutput).To(ContainSubstring("NamePrefix"))
			Expect(configOutput).To(ContainSubstring("Timeout"))
		})

		It("should be seeing global default empty config values as its not set", func() {
			updateNotificationValue := getPreferenceValue("UpdateNotification")
			Expect(updateNotificationValue).To(BeEmpty())
			namePrefixValue := getPreferenceValue("NamePrefix")
			Expect(namePrefixValue).To(BeEmpty())
			timeoutValue := getPreferenceValue("Timeout")
			Expect(timeoutValue).To(BeEmpty())
		})

		It("should be checking to see if global config values are the same as the configured ones", func() {
			runCmdShouldPass("odo preference set updatenotification false")
			runCmdShouldPass("odo preference set timeout 5")
			UpdateNotificationValue := getPreferenceValue("UpdateNotification")
			Expect(UpdateNotificationValue).To(ContainSubstring("false"))
			TimeoutValue := getPreferenceValue("Timeout")
			Expect(TimeoutValue).To(ContainSubstring("5"))
		})

		It("should be checking to see if local config values are the same as the configured ones", func() {
			cases := []struct {
				paramName  string
				paramValue string
			}{
				{
					paramName:  "ComponentType",
					paramValue: "java",
				},
				{
					paramName:  "ComponentName",
					paramValue: "odo-java",
				},
				{
					paramName:  "MinCPU",
					paramValue: "0.2",
				},
				{
					paramName:  "MinMemory",
					paramValue: "100M",
				},
			}
			for _, testCase := range cases {
				runCmdShouldPass(fmt.Sprintf("odo config set %s %s", testCase.paramName, testCase.paramValue))
				Value := getConfigValue(testCase.paramName)
				Expect(Value).To(ContainSubstring(testCase.paramValue))
			}
		})

		It("should allow unsetting a config locally", func() {
			cases := []struct {
				paramName  string
				paramValue string
			}{
				{
					paramName:  "ComponentType",
					paramValue: "java",
				},
				{
					paramName:  "ComponentName",
					paramValue: "odo-java",
				},
				{
					paramName:  "MinCPU",
					paramValue: "0.2",
				},
				{
					paramName:  "MinMemory",
					paramValue: "100M",
				},
			}

			for _, testCase := range cases {
				runCmdShouldPass(fmt.Sprintf("odo config set %s %s", testCase.paramName, testCase.paramValue))
				configOutput := runCmdShouldPass(fmt.Sprintf("odo config unset -f %s", testCase.paramName))
				Expect(configOutput).To(ContainSubstring("Local config was successfully updated."))
				Value := getConfigValue(testCase.paramName)
				Expect(Value).To(BeEmpty())
			}
		})

		It("should allow unsetting a config globally", func() {
			runCmdShouldPass("odo preference set timeout 5")
			runCmdShouldPass("odo preference unset -f timeout")
			timeoutValue := getPreferenceValue("Timeout")
			Expect(timeoutValue).To(BeEmpty())
		})
	})

	Context("creating component without an application", func() {
		It("should create the component in default application", func() {
			runCmdShouldPass("odo create php testcmp")

			getCmp := runCmdShouldPass("odo component get --short")
			Expect(getCmp).To(Equal("testcmp"))

			getApp := runCmdShouldPass("odo app get --short")
			Expect(getApp).To(ContainSubstring("e2e-"))
		})

		It("should be able to delete the component", func() {
			runCmdShouldPass("odo delete testcmp -f")

			getCmp := runCmdShouldPass("odo list")
			Expect(getCmp).NotTo(ContainSubstring("testcmp"))
		})
	})

	Describe("creating an application", func() {
		Context("when application by the same name doesn't exist", func() {
			It("should create an application", func() {
				appName := runCmdShouldPass("odo app create " + appTestName)
				Expect(appName).To(ContainSubstring(appTestName))
			})

			It("should get the current application", func() {
				appName := runCmdShouldPass("odo app get --short")
				Expect(appName).To(Equal(appTestName))
			})

			It("should be created within the project", func() {
				projName := runCmdShouldPass("odo project get --short")
				Expect(projName).To(Equal(projName))
			})

			It("should be able to create another application", func() {
				appName := runCmdShouldPass("odo app create " + appTestName + "-2")
				Expect(appName).To(ContainSubstring(appTestName + "-2"))
			})

			It("should be able to list applications in current project", func() {
				appNames := runCmdShouldPass("odo app list")
				Expect(appNames).To(ContainSubstring(appTestName))
				Expect(appNames).To(ContainSubstring(appTestName + "-2"))
			})

			It("should be able to delete an application", func() {
				// Cleanup
				runCmdShouldPass("odo app delete " + appTestName + "-2 -f")
			})

			It("should be able to set an application as current", func() {
				appName := runCmdShouldPass("odo app set " + appTestName)
				Expect(appName).To(ContainSubstring(appTestName))
			})
		})

		// TODO: Check if the application with the same name can be created
	})

	Context("should list applications in other project", func() {
		newProjName := strings.Replace(projName, "odo", "odo2", -1)
		It("should create a new project", func() {
			session := runCmdShouldPass("odo project create " + newProjName)
			Expect(session).To(ContainSubstring(newProjName))
		})

		It("should get the project", func() {
			getProj := runCmdShouldPass("odo project get --short")
			Expect(strings.TrimSpace(getProj)).To(Equal(newProjName))
		})

		It("should show nice message when there is no application in project", func() {
			appNames := runCmdShouldPass("odo app list")
			Expect(strings.TrimSpace(appNames)).To(
				Equal("There are no applications deployed in the project '" + newProjName + "'."))
		})

		It("should be able to list applications in other project", func() {
			appNames := runCmdShouldPass("odo app list --project " + projName)
			Expect(appNames).To(ContainSubstring(appTestName))
		})

		It("should set the other project as active", func() {
			setProj := runCmdShouldPass("odo project set --short " + projName)
			Expect(strings.TrimSpace(setProj)).To(Equal(projName))
		})

		It("should be able to set an application as current", func() {
			appName := runCmdShouldPass("odo app set " + appTestName)
			Expect(appName).To(ContainSubstring(appTestName))
		})
	})

	Describe("creating a component", func() {
		Context("when application exists", func() {
			var autoGenNodeJSCompName string

			It("should be able to create new imagestream and find it in catalog list", func() {
				curProj = runCmdShouldPass("oc project -q")
				curProj = strings.TrimSuffix(curProj, "\n")
				cmd := fmt.Sprintf("oc create -f "+testNamespacedImage+" -n %s", curProj)
				runCmdShouldPass(cmd)
				cmpList := runCmdShouldPass("odo catalog list components")
				Expect(cmpList).To(ContainSubstring(curProj))
			})

			It("should create and push the contents of a named component excluding the contents in .odoignore file", func() {
				runCmdShouldPass("git clone https://github.com/openshift/nodejs-ex " +
					tmpDir + "/nodejs-ex")

				// TODO: add tests for --git
				curProj = runCmdShouldPass("oc project -q")
				curProj = strings.TrimSuffix(curProj, "\n")
				// Sleep until status tags and their annotations are created
				time.Sleep(15 * time.Second)

				ignoreFilePath := tmpDir + "/nodejs-ex/.odoignore"

				if createFileAtPathWithContent(ignoreFilePath, ".git\ntests/\nREADME.md") != nil {
					fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
				}

				runCmdShouldPass("odo create " + curProj + "/nodejs nodejs --local " + tmpDir + "/nodejs-ex")
				runCmdShouldPass("odo push")

				// get the name of the pod
				podName := getPodNameOfComp("nodejs")

				// verify that the views folder got pushed
				runCmdShouldPass("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep views")

				// verify that the tests was not pushed
				runCmdShouldFail("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep tests")

				// verify that the README.md file was not pushed
				runCmdShouldFail("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep README.md")

				// remove the .odoignore file
				Expect(os.Remove(ignoreFilePath)).To(BeNil())
			})

			It("should create a component and push using the --ignore flag", func() {
				runCmdShouldPass("odo create " + curProj + "/nodejs push-odoignore-flag-example --local " + tmpDir + "/nodejs-ex")
				runCmdShouldPass("odo push --ignore tests/,README.md")

				// get the name of the pod
				podName := getPodNameOfComp("push-odoignore-flag-example")

				// verify that the views folder got pushed
				runCmdShouldPass("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep views")

				// verify that the tests was not pushed
				runCmdShouldFail("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep tests")

				// verify that the README.md file was not pushed
				runCmdShouldFail("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep README.md")
			})

			It("should create a component with auto-generated name", func() {
				runCmdShouldPass("odo create " + curProj + "/nodejs --local " + tmpDir + "/nodejs-ex")
				runCmdShouldPass("odo push")
			})

			It("should be the get the component created as active component", func() {
				autoGenNodeJSCompName = runCmdShouldPass("odo component get --short")
				Expect(autoGenNodeJSCompName).To(ContainSubstring(fmt.Sprintf("nodejs-ex-%s-nodejs", curProj)))
			})

			It("should create the component within the application", func() {
				getApp := runCmdShouldPass("odo app get --short")
				Expect(getApp).To(Equal(appTestName))
			})

			It("should list the components within the application", func() {
				cmpList := runCmdShouldPass("odo list")
				Expect(cmpList).To(ContainSubstring("nodejs"))
			})

			It("should be able to create multiple components within the same application", func() {
				runCmdShouldPass("odo create php php")
			})

			It("should list the newly created second component", func() {
				cmpList := runCmdShouldPass("odo list")
				Expect(cmpList).To(ContainSubstring("php"))
			})

			It("should get the application "+appTestName, func() {
				appGet := runCmdShouldPass("odo app get --short")
				Expect(appGet).To(Equal(appTestName))
			})

			It("should be able to set a component as active", func() {
				cmpSet := runCmdShouldPass(fmt.Sprintf("odo component set %s", autoGenNodeJSCompName))
				Expect(cmpSet).To(ContainSubstring(autoGenNodeJSCompName))
			})

			It("should be able to retrieve logs", func() {
				runCmdShouldPass("odo log")
				runCmdShouldPass(fmt.Sprintf("odo log %s", autoGenNodeJSCompName))
			})

			It("should be able to create git component with required ports", func() {
				runCmdShouldPass("odo create nodejs nodejs-git --git https://github.com/openshift/nodejs-ex --port 8080/tcp,9100/udp")

				// checking port names
				portsNames := runCmdShouldPass("oc get services nodejs-git-" + appTestName + " -o go-template='{{range .spec.ports}}{{.name}}{{end}}'")
				Expect(portsNames).To(ContainSubstring("8080-tcp"))
				Expect(portsNames).To(ContainSubstring("9100-udp"))

				// checking port numbers
				ports := runCmdShouldPass("oc get services nodejs-git-" + appTestName + " -o go-template='{{range .spec.ports}}{{.port}}{{end}}'")
				Expect(ports).To(ContainSubstring("8080"))
				Expect(ports).To(ContainSubstring("9100"))

				// checking protocols
				protocols := runCmdShouldPass("oc get services nodejs-git-" + appTestName + " -o go-template='{{range .spec.ports}}{{.protocol}}{{end}}'")
				Expect(protocols).To(ContainSubstring("TCP"))
				Expect(protocols).To(ContainSubstring("UDP"))

				// deleting the component
				runCmdShouldPass("odo delete -f")

				getCmp := runCmdShouldPass("odo list")
				Expect(getCmp).NotTo(ContainSubstring("nodejs-git"))
			})

			It("should be able to create git component with required env vars", func() {
				runCmdShouldPass("odo create nodejs nodejs-git --git https://github.com/openshift/nodejs-ex --env key=value,key1=value1")

				// checking the values of the env vars pairs in bc
				envVars := runCmdShouldPass("oc get bc nodejs-git-" + appTestName + " -o go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
				Expect(envVars).To(Equal("keyvaluekey1value1"))

				// checking the values of the env vars pairs in dc
				envVars = runCmdShouldPass("oc get dc nodejs-git-" + appTestName + " -o go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
				Expect(envVars).To(Equal("keyvaluekey1value1"))

				// deleting the component
				runCmdShouldPass("odo delete -f")
			})
		})
	})

	Describe("Creating odo url", func() {
		Context("using odo url", func() {
			It("should create route without url name provided", func() {
				runCmdShouldPass("odo component set nodejs")
				getURLOut := runCmdShouldPass("odo url create")
				Expect(getURLOut).To(ContainSubstring("nodejs-8080-" + appTestName + "-" + projName))

				// check the port number of the created URL
				port := runCmdShouldPass("oc get route nodejs-8080-" + appTestName + " -o go-template='{{index .spec.port.targetPort}}'")
				Expect(port).To(Equal("8080"))

				// delete the url
				runCmdShouldPass("odo url delete nodejs-8080 -f")
			})

			It("should create route without port in case of single service port component", func() {
				runCmdShouldPass("odo component set nodejs")
				getURLOut := runCmdShouldPass("odo url create nodejs")
				Expect(getURLOut).To(ContainSubstring("nodejs-" + appTestName + "-" + projName))

				// check the port number of the created URL
				port := runCmdShouldPass("oc get route nodejs-" + appTestName + " -o go-template='{{index .spec.port.targetPort}}'")
				Expect(port).To(Equal("8080"))
			})

			It("should be able to list the url", func() {
				getRoute := getActiveElementFromCommandOutput("odo url list")
				Expect(getRoute).To(ContainSubstring("nodejs-" + appTestName + "-" + projName))

				// Check the labels in `oc get route`
				routeName := "nodejs-" + appTestName
				getRouteLabel := runCmdShouldPass("oc get route/" + routeName + " -o jsonpath='" +
					"{.metadata.labels.app\\.kubernetes\\.io/component-name}'")
				Expect(getRouteLabel).To(Equal("nodejs"))
			})

			It("should create route with required port", func() {
				runCmdShouldPass("odo create httpd httpd-test --git https://github.com/openshift/httpd-ex.git")
				getURLOut := runCmdShouldPass("odo url create example-url --port 8443")
				Expect(getURLOut).To(ContainSubstring("example-url-" + appTestName + "-" + projName))

				// check the port number of the created URL
				port := runCmdShouldPass("oc get route example-url-" + appTestName + " -o go-template='{{index .spec.port.targetPort}}'")
				Expect(port).To(Equal("8443"))

				// delete the component
				runCmdShouldPass("odo delete httpd-test -f")
			})
		})
	})

	Describe("pushing updates", func() {
		Context("When push is made", func() {
			It("should push the changes", func() {
				// Switch to nodejs component
				runCmdShouldPass("odo component set nodejs")

				getRoute := determineRouteURL()
				responseStringMatchStatus := matchResponseSubString(getRoute, "Welcome to your Node.js application on OpenShift", 30, 1)
				Expect(responseStringMatchStatus).Should(BeTrue())

				// Make changes to the html file
				replaceTextStatus := replaceTextInFile(tmpDir+"/nodejs-ex/views/index.html", "Welcome to your Node.js application on OpenShift", "Welcome to your Node.js application on ODO")
				Expect(replaceTextStatus).To(BeNil())

				// Push the changes
				runCmdShouldPass("odo push --local " + tmpDir + "/nodejs-ex")

				// Verify the changes
				responseChangeStringStatus := matchResponseSubString(getRoute, "Welcome to your Node.js application on ODO", 30, 1)
				Expect(responseChangeStringStatus).Should(BeTrue())
			})

			It("should be able to create the url with same name in different application", func() {
				appTestName_new := appTestName + "-1"
				runCmdShouldPass("odo app create " + appTestName_new)
				runCmdShouldPass("odo create nodejs nodejs-1 --git https://github.com/sclorg/nodejs-ex")
				runCmdShouldPass("odo url create nodejs --port 8080")

				getRoute := getActiveElementFromCommandOutput("odo url list")
				Expect(getRoute).To(ContainSubstring("nodejs-" + appTestName_new + "-" + projName))

				// Check the labels in `oc get route`
				routeName := "nodejs-" + appTestName_new
				getRouteLabel := runCmdShouldPass("oc get route/" + routeName + " -o jsonpath='" +
					"{.metadata.labels.app\\.kubernetes\\.io/component-name}'")
				Expect(getRouteLabel).To(Equal("nodejs-1"))
			})

			// Check if url is deleted
			It("should be able to delete the url added", func() {
				appTestName_new := appTestName + "-1"
				runCmdShouldPass("odo app set " + appTestName_new)
				runCmdShouldPass("odo component set nodejs-1")
				runCmdShouldPass("odo url delete nodejs -f")

				getRoute := getActiveElementFromCommandOutput("odo url list")
				Expect(getRoute).NotTo(ContainSubstring("nodejs-1-" + appTestName_new + "-" + projName))

				runCmdShouldPass("odo delete -f")
				runCmdShouldPass("odo app delete " + appTestName_new + " -f")
			})

		})
	})

	Describe("Adding storage", func() {
		Context("when storage is added", func() {
			It("should default to active component when no component name is passed", func() {

				runCmdShouldPass("odo app set " + appTestName)
				runCmdShouldPass("odo component set nodejs")

				storAdd := runCmdShouldPass("odo storage create pv1 --path /mnt/pv1 --size 5Gi")
				Expect(storAdd).To(ContainSubstring("nodejs"))

				// Check against path and name against dc
				getDc := runCmdShouldPass("oc get dc/nodejs-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv1"))

				// Check if the storage is added on the path provided
				getMntPath := runCmdShouldPass("oc get dc/nodejs-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

				Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))
			})

			It("should be able to list the storage added", func() {
				storList := runCmdShouldPass("odo storage list")
				Expect(storList).To(ContainSubstring("pv1"))
			})

			It("should be able add storage to a component specified", func() {
				runCmdShouldPass("odo storage create pv2 --path /mnt/pv2 --size 5Gi --component php")

				storList := runCmdShouldPass("odo storage list --component php")
				Expect(storList).To(ContainSubstring("pv2"))

				// Verify with deploymentconfig
				getDc := runCmdShouldPass("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv2"))

				// Check if the storage is added on the path provided
				getMntPath := runCmdShouldPass("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

				Expect(getMntPath).To(ContainSubstring("/mnt/pv2"))
			})

			It("should be able to list all storage in all components", func() {
				storList := runCmdShouldPass("odo storage list --all")
				Expect(storList).To(ContainSubstring("pv1"))
				Expect(storList).To(ContainSubstring("pv2"))
			})

			// TODO: Verify if the storage removed using odo deletes pvc
			It("should be able to delete the storage added", func() {
				runCmdShouldPass("odo storage delete pv1 -f")

				storList := runCmdShouldPass("odo storage list")
				Expect(storList).NotTo(ContainSubstring("pv1"))
			})

			It("should be able to unmount the storage using the storage name", func() {
				runCmdShouldPass("odo storage unmount pv2 --component php")

				// Verify with deploymentconfig
				getDc := runCmdShouldPass("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).NotTo(ContainSubstring("pv2"))
			})

			It("should be able to mount the storage to the path specified", func() {
				runCmdShouldPass("odo storage mount pv2 --path /mnt/pv2 --component php")

				// Verify with deploymentconfig
				getDc := runCmdShouldPass("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv2"))

				// Check if the storage is added on the path provided
				getMntPath := runCmdShouldPass("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

				Expect(getMntPath).To(ContainSubstring("/mnt/pv2"))
			})

			It("should be able to unmount the storage", func() {
				runCmdShouldPass("odo storage unmount /mnt/pv2 --component php")

				// Verify with deploymentconfig
				getDc := runCmdShouldPass("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).NotTo(ContainSubstring("pv2"))
			})
		})
	})

	Context("deploying a component with a specific image name", func() {
		It("should deploy the component", func() {
			runCmdShouldPass("odo create nodejs:latest testversioncmp")
		})

		It("should delete the deployed image-specific component", func() {
			runCmdShouldPass("odo delete testversioncmp -f")
		})
	})

	Context("deleting the application", func() {
		// Check if url is deleted
		It("should be able to delete the url added to the component", func() {
			runCmdShouldPass("odo component set nodejs")
			runCmdShouldPass("odo url delete nodejs -f")

			urlList := getActiveElementFromCommandOutput("odo url list")
			Expect(urlList).NotTo(ContainSubstring("nodejs"))
		})

		It("should delete application and component", func() {

			runCmdShouldPass("odo app delete " + appTestName + " -f")

			appGet := runCmdShouldPass("odo app get --short")
			Expect(appGet).NotTo(ContainSubstring(appTestName))

			appList := runCmdShouldPass("odo app list")
			Expect(appList).NotTo(ContainSubstring(appTestName))

			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).NotTo(ContainSubstring("nodejs"))

			odoDeleteProject(projName)
		})

		It("should auto switch when deleting applications", func() {
			// create two new projects

			odoCreateProject(projName + "-auto-0")

			runCmdShouldPass("odo app create app-1")
			runCmdShouldPass("odo app create app-2")
			runCmdShouldPass("odo app create app-3")

			odoCreateProject(projName + "-auto-1")
			runCmdShouldPass("odo app create app-4")
			runCmdShouldPass("odo app create app-5")

			// delete app in some other project which is not active
			// the current app in the active project should not switch
			runCmdShouldPass("odo app delete --project " + projName + "-auto-0 app-1 -f")
			Expect(getActiveApplication()).To(Equal("app-5"))

			// delete in the active project
			// the current app should switch
			runCmdShouldPass("odo app delete -f")
			Expect(getActiveApplication()).To(Equal("app-4"))

			// deleting the last app in the active project
			runCmdShouldPass("odo app delete -f")
			Expect(getActiveApplication()).To(Equal(""))

			// clean up
			odoDeleteProject(projName + "-auto-0")
			odoDeleteProject(projName + "-auto-1")
		})
	})

	Context("validate odo version cmd with other major components version", func() {
		It("should show the version of odo major components", func() {
			odoVersion := runCmdShouldPass("odo version")
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+\+\w+`)
			kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			Expect(kubernetesVersionStringMatch).Should(BeTrue())
		})

		It("should show server login URL", func() {
			odoVersion := runCmdShouldPass("odo version")
			reServerURL := regexp.MustCompile(`Server:\s*https:\/\/([0-9]+.){3}[0-9]+:8443`)
			serverURLStringMatch := reServerURL.MatchString(odoVersion)
			Expect(serverURLStringMatch).Should(BeTrue())
		})
	})
})

var _ = Describe("updateE2e", func() {
	projName := generateTimeBasedName("odo")
	const appTestName = "testing"

	const bootStrapSupervisorURI = "https://github.com/kadel/bootstrap-supervisored-s2i"
	const initContainerName = "copy-files-to-volume"
	const wildflyURI1 = "https://github.com/marekjelen/katacoda-odo-backend"
	const wildflyURI2 = "https://github.com/mik-dass/katacoda-odo-backend"
	const appRootVolumeName = "-testing-s2idata"

	tmpDir, err := ioutil.TempDir("", "odo")
	if err != nil {
		Fail(err.Error())
	}

	Describe("creating the project", func() {
		Context("odo project", func() {
			It("should create a new project", func() {
				session := runCmdShouldPass("odo project create " + projName + "-1")
				Expect(session).To(ContainSubstring(projName))
			})

			It("should get the project", func() {
				getProj := runCmdShouldPass("odo project get --short")
				Expect(strings.TrimSpace(getProj)).To(Equal(projName + "-1"))
			})
		})
	})

	Describe("creating an application", func() {
		Context("when application by the same name doesn't exist", func() {
			It("should create an application", func() {
				appName := runCmdShouldPass("odo app create " + appTestName)
				Expect(appName).To(ContainSubstring(appTestName))
			})
		})
	})

	Context("updating the component", func() {
		It("should be able to create binary component", func() {
			runCmdShouldPass("curl -L -o " + tmpDir + "/sample-binary-testing-1.war " +
				"https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/e5bc575ac8b14ba2b23d66b5cb4873657e1a1489/sample.war")
			runCmdShouldPass("odo create wildfly --binary " + tmpDir + "/sample-binary-testing-1.war  --env key=value,key1=value1")
			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).To(ContainSubstring("wildfly"))

			runCmdShouldPass("oc get dc")
			runCmdShouldPass("oc get bc")
		})

		It("should update component from binary to binary", func() {
			runCmdShouldPass("curl -L -o " + tmpDir + "/sample-binary-testing-2.war " +
				"'https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/85354d9ee8583a9c1e64a331425eede235b07a9e/sample%2520(1).war'")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-2.war")

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-2.war")
			EnvVarTest("wildfly-"+appTestName, "binary", "keyvaluekey1value1")
		})

		It("should update component from binary to local", func() {
			runCmdShouldPass("git clone " + wildflyURI1 + " " +
				tmpDir + "/katacoda-odo-backend-1")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
			EnvVarTest("wildfly-"+appTestName, "local", "keyvaluekey1value1")
		})

		It("should update component from local to local", func() {
			runCmdShouldPass("git clone " + wildflyURI2 + " " +
				tmpDir + "/katacoda-odo-backend-2")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-2")

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-2")
			EnvVarTest("wildfly-"+appTestName, "local", "keyvaluekey1value1")
		})

		It("should update component from local to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --git " + wildflyURI1)

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyURI1))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyURI1)
			EnvVarTest("wildfly-"+appTestName, "git", "keyvaluekey1value1")
		})

		It("should update component from git to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --git " + wildflyURI2)

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyURI2))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyURI2)
			EnvVarTest("wildfly-"+appTestName, "git", "keyvaluekey1value1")
		})

		It("should update component from git to binary", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
			EnvVarTest("wildfly-"+appTestName, "binary", "keyvaluekey1value1")
		})

		It("should update component from binary to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --git " + wildflyURI1)

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyURI1))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyURI1)
			EnvVarTest("wildfly-"+appTestName, "git", "keyvaluekey1value1")
		})

		It("should update component from git to local", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
			EnvVarTest("wildfly-"+appTestName, "local", "keyvaluekey1value1")
		})

		It("should update component from local to binary", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmdShouldPass("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// checking bc for updates
			getBc := runCmdShouldPass("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmdShouldPass("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
			EnvVarTest("wildfly-"+appTestName, "binary", "keyvaluekey1value1")
		})
	})
	Context("odo login", func() {
		It("should login with username and password", func() {
			runCmdShouldPass("oc logout")
			runCmdShouldPass("odo login --username developer --password developer")
			userName := runCmdShouldPass("oc whoami")
			Expect(userName).To(ContainSubstring("developer"))
		})
		It("should login with token", func() {
			userToken := runCmdShouldPass("oc whoami -t")
			runCmdShouldPass("oc logout")
			runCmdShouldPass("odo login -t" + userToken)
			token := runCmdShouldPass("oc whoami -t")
			Expect(token).To(ContainSubstring(userToken))
		})
	})

	Context("Deleting application, with component, should list affected children", func() {
		projectName := generateTimeBasedName("project")
		appName := generateTimeBasedName("app")
		componentName := generateTimeBasedName("component")
		urlName := generateTimeBasedName("url")

		It("Should setup for the tests ,by creating dummy projects to test against", func() {
			odoCreateProject(projectName)
			runCmdShouldPass("odo app create " + appName)
			runCmdShouldPass("odo component create nodejs " + componentName)
			runCmdShouldPass("odo url create " + urlName)
		})

		It("Should list affected child objects", func() {
			session := runCmdShouldPass("odo app delete -f " + appName)
			Expect(session).To(ContainSubstring("This application has following components that will be deleted"))
			Expect(session).To(ContainSubstring(componentName))
			Expect(session).To(ContainSubstring(urlName))
		})

		It("Should delete project", func() {
			odoDeleteProject(projectName)
		})
	})

	Context("Deleting project, with application should list affected children", func() {
		projectName := generateTimeBasedName("project")
		appName := generateTimeBasedName("app")
		componentName := generateTimeBasedName("component")
		urlName := generateTimeBasedName("url")

		It("Should setup for the tests ,by creating dummy projects to test against", func() {
			odoCreateProject(projectName)
			runCmdShouldPass("odo app create " + appName)
			runCmdShouldPass("odo component create nodejs " + componentName)
			runCmdShouldPass("odo url create " + urlName)
		})

		It("Should list affected child objects", func() {
			session := runCmdShouldPass("odo project delete -f " + projectName)
			Expect(session).To(ContainSubstring("This project contains the following applications, which will be deleted"))
			Expect(session).To(ContainSubstring(appName))
			Expect(session).To(ContainSubstring(componentName))
			Expect(session).To(ContainSubstring(urlName))
		})
	})

	Context("logout of the cluster", func() {
		// test for odo logout
		It("should logout the user from the cluster", func() {
			logoutMsg := runCmdShouldPass("odo logout")
			Expect(logoutMsg).To(ContainSubstring("Logged"))
			Expect(logoutMsg).To(ContainSubstring("out on"))
			// validate using oc whoami
			outputErr := runCmdShouldFail("oc whoami")
			Expect(outputErr).To(ContainSubstring("cannot get users.user.openshift.io at the cluster scope"))
		})
		It("Logout should throw error if user is not logged in", func() {
			logoutErrMsg := runCmdShouldFail("odo logout")
			Expect(logoutErrMsg).To(Equal("Please log in to the cluster\n"))
		})
	})
})
