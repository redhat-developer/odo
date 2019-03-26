// +build !race

package e2e

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

// TODO: A neater way to provide odo path. Currently we assume \
// odo and oc in $PATH already.
var curProj string
var newProjName string
var testNamespacedImage = "https://raw.githubusercontent.com/bucharest-gold/centos7-s2i-nodejs/master/imagestreams/nodejs-centos7.json"
var testPHPGitURL = "https://github.com/appuio/example-php-sti-helloworld"

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
	// commenting this out to resolve e2e tests failures on OC 4
	// curProj = runCmdShouldPass("oc project -q")
})

func VerifyAppNameOfComponent(cmpName string, appName string) {
	session := runCmdShouldPass(fmt.Sprintf("oc get dc %s-%s --template={{.metadata.labels.'app'}}", cmpName, appName))
	Expect(session).To(ContainSubstring(appName))
}

func VerifyCmpName(cmpName string) {
	dcName := getDcName(cmpName)
	session := runCmdShouldPass(fmt.Sprintf("oc get dc %s -L app.kubernetes.io/component-name| awk '{print $6}'|sed -n 2p", dcName))
	//Expect(session).To(ContainSubstring(cmpName))
	Expect(session).To(Equal(cmpName))
}

func GetURL(contextDir string, port int) string {
	var url string
	currDir := runCmdShouldPass("pwd")
	currDir = strings.TrimSpace(currDir)
	runCmdShouldPass("cd " + contextDir)
	runCmdShouldPass(fmt.Sprintf("odo url create --port %d", port))
	url = determineRouteURL()
	runCmdShouldPass("cd " + currDir)
	return url
}

var _ = Describe("odoe2e", func() {
	projName := generateTimeBasedName("odo")
	const appTestName = "testing"
	const loginTestUserPassword = "developer"

	tmpDir, err := ioutil.TempDir("", "odo")
	if err != nil {
		Fail(err.Error())
	}

	/*
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
					runCmdShouldPass("odo create nodejs --project prj2 ")
					session := runCmdShouldFail("odo push")
					Expect(session).To(ContainSubstring("deploymentconfigs.apps.openshift.io is forbidden: User \"odonoprojectattemptscreate\" cannot list deploymentconfigs.apps.openshift.io in the namespace \"default\": no RBAC policy matched"))
					session = runCmdShouldFail("odo component create nodejs")
					Expect(session).To(ContainSubstring("User \"odonoprojectattemptscreate\" cannot list deploymentconfigs.apps.openshift.io in the namespace \"default\""))
					// Uncomment once storage related commands are fixed
						session = runCmdShouldFail("odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi")
						Expect(session).To(ContainSubstring("You dont have permission to project"))
						Expect(session).To(ContainSubstring("or it doesnt exist."))
						Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				})

				It("Should pass if user tries to create a project", func() {
					session := runCmdShouldPass("odo project create odonoprojectattemptscreateproject")
					Expect(session).To(ContainSubstring("New project created and now using project"))
					Expect(session).To(ContainSubstring("odonoprojectattemptscreateproject"))
					ocDeleteProject("odonoprojectattemptscreateproject")
				})
			})

			Context("Logs in as user with a project, deletes it and tries to create various objects", func() {
				It("Should login as a user and setup by creating a project, and then deleting it", func() {
					runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", "odosingleprojectattemptscreate", loginTestUserPassword))
					odoCreateProject("odosingleprojectattemptscreateproject")
					ocDeleteProject("odosingleprojectattemptscreateproject")
				})

				It("Should fail if user tries to create any object, other than project", func() {
					session := runCmdShouldFail("mkdir -p nodejs-no-perm; odo create nodejs --context ./nodejs-no-perm")
					Expect(session).To(ContainSubstring("deploymentconfigs.apps.openshift.io is forbidden: User \"odosingleprojectattemptscreate\" cannot list deploymentconfigs.apps.openshift.io in the namespace \"odosingleprojectattemptscreateproject\": no RBAC policy matched"))
					//Expect(session).To(ContainSubstring("or it doesnt exist"))
					//Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
					// Uncomment once storage commands are fixed to work with new config workflow and context
						session = runCmdShouldFail("odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi")
						Expect(session).To(ContainSubstring("You dont have permission to project"))
						Expect(session).To(ContainSubstring("or it doesnt exist"))
						Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				})

				It("Should pass if user tries to create a project", func() {
					session := runCmdShouldPass("odo project create odosingleprojectattemptscreateproject")
					Expect(session).To(ContainSubstring("New project created and now using project"))
					Expect(session).To(ContainSubstring("odosingleprojectattemptscreateproject"))
					ocDeleteProject("odosingleprojectattemptscreateproject")
				})
			})

			Context("Log back in as old user", func() {
				It("Should log back in as old user", func() {
					runCmdShouldPass(fmt.Sprintf("oc login --token %s", currentUserToken))
				})
			})
		})
	*/
	/*
		Context("odo service create", func() {
			It("should return error if the cluster has no service catalog deployed", func() {
				loginOutput := runCmdShouldPass("odo login --username developer --password developer")
				Expect(loginOutput).To(ContainSubstring("Login successful"))
				sessionErrOutput := runCmdShouldFail("odo service create")
				Expect(sessionErrOutput).To(ContainSubstring("unable to retrieve service classes"))
			})
		})
	*/
	// TODO: Create component without creating application
	/* Uncomment after project commands are fixed
	Context("odo project", func() {
		It("should create a new project", func() {
			loginOutput := runCmdShouldPass("odo login --username developer --password developer")
			Expect(loginOutput).To(ContainSubstring("Login successful"))
			//session := runCmdShouldPass("odo project create " + projName)
			//Expect(session).To(ContainSubstring(projName))
			runCmdShouldPass("oc new-project " + projName)
		})

		// Issue #630
		It("should list the project", func() {
			listProj := runCmdShouldPass("sleep 5s && odo project list")
			fmt.Println(listProj)
			Expect(listProj).To(ContainSubstring(projName))
		})
	})
	*/

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
					paramName:  "Type",
					paramValue: "java",
				},
				{
					paramName:  "Name",
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
				runCmdShouldPass(fmt.Sprintf("odo config set %s %s -f", testCase.paramName, testCase.paramValue))
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
					paramName:  "Type",
					paramValue: "java",
				},
				{
					paramName:  "Name",
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
			runCmdShouldPass("odo login --username developer --password developer")
			runCmdShouldPass("odo create php testcmp --app e2e-xyzk --git " + testPHPGitURL)
			runCmdShouldPass("odo push")

			VerifyCmpName("testcmp")

			VerifyAppNameOfComponent("testcmp", "e2e-xyzk")
		})
		// Uncommment after fixing the component delete once it has been modified to work with
		/*
			It("should be able to delete the component", func() {
				runCmdShouldPass("odo delete testcmp -f")

				getCmp := runCmdShouldPass("odo list")
				Expect(getCmp).NotTo(ContainSubstring("testcmp"))
			})
		*/
	})
	// Uncomment the tests below once app commands are fixed
	/*
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
	*/

	Context("should list applications in other project", func() {
		newProjName := strings.Replace(projName, "odo", "odo2", -1)
		It("should create a new project", func() {
			runCmdShouldPass("oc new-project " + newProjName)
			waitForCmdOut("oc get project", 2, true, func(output string) bool {
				return strings.Contains(output, newProjName)
			})
			//session := runCmdShouldPass("odo project create " + newProjName)
			//Expect(session).To(ContainSubstring(newProjName))
		})
		/*
			It("should show nice message when there is no application in project", func() {
				appNames := runCmdShouldPass("odo app list --project " + newProjName)
				Expect(strings.TrimSpace(appNames)).To(
					Equal("There are no applications deployed in the project '" + newProjName + "'."))
			})
		*/
		It("should be able to create a php component with application created", func() {
			runCmdShouldPass("odo create php testcmp --app " + appTestName + " --project " + newProjName + " --git " + testPHPGitURL)
			runCmdShouldPass("odo push")
		})

		It("should be able to list applications in other project", func() {
			appNames := runCmdShouldPass("odo app list --project " + newProjName)
			Expect(appNames).To(ContainSubstring(appTestName))
		})
	})

	Describe("creating a component", func() {
		Context("when application exists", func() {
			//var autoGenNodeJSCompName string
			/* Commented until this is fixed
			It("should be able to create new imagestream and find it in catalog list", func() {
				curProj = runCmdShouldPass("oc project -q")
				curProj = strings.TrimSuffix(curProj, "\n")
				cmd := fmt.Sprintf("oc create -f "+testNamespacedImage+" -n %s", curProj)
				runCmdShouldPass(cmd)
				cmpList := runCmdShouldPass("odo catalog list components")
				Expect(cmpList).To(ContainSubstring(curProj))
			})
			*/

			It("should create and push the contents of a named component excluding the contents in .odoignore file", func() {
				runCmdShouldPass("git clone https://github.com/openshift/nodejs-ex " +
					tmpDir + "/nodejs-ex")

				// TODO: add tests for --git
				curProj = runCmdShouldPass("oc project -q")
				curProj = strings.TrimSuffix(curProj, "\n")

				ignoreFilePath := tmpDir + "/nodejs-ex/.odoignore"

				if createFileAtPathWithContent(ignoreFilePath, ".git\ntests/\nREADME.md") != nil {
					fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
				}

				// Uncomment below line once image import problem is fixed
				// runCmdShouldPass("odo create " + curProj + "/nodejs nodejs --context " + tmpDir + "/nodejs-ex")
				runCmdShouldPass("odo create nodejs nodejs --context " + tmpDir + "/nodejs-ex")
				runCmdShouldPass("odo push --context " + tmpDir + "/nodejs-ex")

				// get the name of running pod
				podName := getRunningPodNameOfComp("nodejs")

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
				// runCmdShouldPass("odo create " + curProj + "/nodejs push-odoignore-flag-example --context " + tmpDir + "/nodejs-ex")
				runCmdShouldPass("odo create nodejs push-odoignore-flag-example --context " + tmpDir + "/nodejs-ex")
				// runCmdShouldPass("odo push --ignore tests/,README.md")
				runCmdShouldPass("odo push --context " + tmpDir + "/nodejs-ex")

				// get the name of running pod
				podName := getRunningPodNameOfComp("push-odoignore-flag-example")

				// verify that the views folder got pushed
				runCmdShouldPass("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep views")
				/*
					// verify that the tests was not pushed
					runCmdShouldFail("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep tests")

					// verify that the README.md file was not pushed
					runCmdShouldFail("oc exec " + podName + " -- ls -lai /opt/app-root/src | grep README.md")
				*/
			})

			It("should create a component with auto-generated name", func() {
				// runCmdShouldPass("odo create " + curProj + "/nodejs --context " + tmpDir + "/nodejs-ex --app " + appTestName)
				runCmdShouldPass("odo create nodejs --context " + tmpDir + "/nodejs-ex --app " + appTestName)
				runCmdShouldPass("odo push --context " + tmpDir + "/nodejs-ex")
			})

			It("should list the components within the application", func() {
				VerifyCmpName("nodejs")
				//cmpList := runCmdShouldPass("odo list")
				//Expect(cmpList).To(ContainSubstring("nodejs"))
			})

			It("should be able to create multiple components within the same application", func() {
				runCmdShouldPass("odo create php php --git " + testPHPGitURL + " --app " + appTestName)
				runCmdShouldPass("odo push")
			})

			It("should list the newly created second component", func() {
				// Uncomment below after --context is added to odo list
				//cmpList := runCmdShouldPass("odo list")
				//Expect(cmpList).To(ContainSubstring("php"))
				time.Sleep(20 * time.Second)
				VerifyCmpName("php")
			})

			It("should get the application "+appTestName, func() {
				VerifyAppNameOfComponent("php", appTestName)
			})

			// Uncomment below tests after odo logs is fixed as per new workflow
			/*
				It("should be able to retrieve logs", func() {
					runCmdShouldPass("odo log")
					runCmdShouldPass(fmt.Sprintf("odo log %s", autoGenNodeJSCompName))
				})
			*/

			/*
				It("should be able to create git component with required ports", func() {
					runCmdShouldPass("odo create nodejs nodejs-git --git https://github.com/openshift/nodejs-ex --port 8080/tcp,9100/udp --app " + appTestName)
					runCmdShouldPass("odo push")

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
					runCmdShouldPass("odo delete -f nodejs-git")

					getCmp := runCmdShouldPass("odo list")
					Expect(getCmp).NotTo(ContainSubstring("nodejs-git"))
				})


				It("should be able to create git component with required env vars", func() {
					runCmdShouldPass("odo create nodejs nodejs-git --git https://github.com/openshift/nodejs-ex --env key=value,key1=value1")
					runCmdShouldPass("odo push")

					// checking the values of the env vars pairs in bc
					envVars := runCmdShouldPass("oc get bc nodejs-git-" + appTestName + " -o go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
					Expect(envVars).To(Equal("keyvaluekey1value1"))

					// checking the values of the env vars pairs in dc
					envVars = runCmdShouldPass("oc get dc nodejs-git-" + appTestName + " -o go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
					Expect(envVars).To(Equal("keyvaluekey1value1"))

					// deleting the component
					runCmdShouldPass("odo delete -f nodejs-git")
				})
			*/
		})
	})
	/*
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
	*/

	Describe("pushing updates", func() {
		Context("When push is made", func() {
			// It("should push the changes", func() {
			// 	runCmdShouldPass("odo create nodejs --context " + tmpDir + "/nodejs-ex")
			// 	runCmdShouldPass("odo push --context " + tmpDir + "/nodejs-ex")

			// 	getRoute := GetURL(tmpDir+"/nodejs-ex", 8080)

			// 	responseStringMatchStatus := matchResponseSubString(getRoute, "Welcome to your Node.js application on OpenShift", 30, 1)
			// 	Expect(responseStringMatchStatus).Should(BeTrue())

			// 	// Make changes to the html file
			// 	replaceTextStatus := replaceTextInFile(tmpDir+"/nodejs-ex/views/index.html", "Welcome to your Node.js application on OpenShift", "Welcome to your Node.js application on ODO")
			// 	Expect(replaceTextStatus).To(BeNil())

			// 	// Push the changes
			// 	runCmdShouldPass("odo push --local " + tmpDir + "/nodejs-ex")

			// 	// Verify the changes
			// 	responseChangeStringStatus := matchResponseSubString(getRoute, "Welcome to your Node.js application on ODO", 30, 1)
			// 	Expect(responseChangeStringStatus).Should(BeTrue())
			// })

			/* Uncomment once url commands work with new workflow and context
			It("should be able to create the url with same name in different application", func() {
				appTestName_new := appTestName + "-1"
				runCmdShouldPass("odo create nodejs nodejs-1 --git https://github.com/sclorg/nodejs-ex --app " + appTestName_new)
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
				runCmdShouldPass("odo url delete nodejs -f")

				getRoute := getActiveElementFromCommandOutput("odo url list")
				Expect(getRoute).NotTo(ContainSubstring("nodejs-1-" + appTestName_new + "-" + projName))

				runCmdShouldPass("odo delete -f")
				runCmdShouldPass("odo app delete " + appTestName_new + " -f")
			})
			*/

		})
	})
	/*
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
	*/
	Context("deploying a component with a specific image name", func() {
		It("should deploy the component", func() {
			runCmdShouldPass("odo create nodejs:latest testversioncmp --context " + tmpDir + "/nodejs-ex")
			runCmdShouldPass("odo push --context " + tmpDir + "/nodejs-ex")
		})

		It("should delete the deployed image-specific component", func() {
			runCmdShouldPass("odo delete testversioncmp -f")
		})
	})

	Context("deleting the application", func() {
		/*
			// Check if url is deleted
			It("should be able to delete the url added to the component", func() {
				runCmdShouldPass("odo component set nodejs")
				runCmdShouldPass("odo url delete nodejs -f")

				urlList := getActiveElementFromCommandOutput("odo url list")
				Expect(urlList).NotTo(ContainSubstring("nodejs"))
			})
		*/

		It("should delete application and component", func() {

			runCmdShouldPass("odo app delete " + appTestName + " -f")

			appList := runCmdShouldPass("odo app list")
			Expect(appList).NotTo(ContainSubstring(appTestName))

			//cmpList := runCmdShouldFail("odo list --app " + appTestName)
			//Expect(cmpList).To(ContainSubstring("There are no components deployed"))

			ocDeleteProject(newProjName)
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
			reServerURL := regexp.MustCompile(`Server:\s*https:\/\/(.+\.com|([0-9]+.){3}[0-9]+):[0-9]{4}`)
			serverURLStringMatch := reServerURL.MatchString(odoVersion)
			Expect(serverURLStringMatch).Should(BeTrue())
		})
	})
})

/*
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
			ocDeleteProject(projectName)
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
*/
