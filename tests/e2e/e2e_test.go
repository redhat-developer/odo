// +build !race

package e2e

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/odo/pkg/config"
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
		envVars := runCmd("oc get bc " + resourceName + " -o go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
		Expect(envVars).To(Equal(envString))
	}

	// checking the values of the env vars pairs in dc
	envVars := runCmd("oc get dc " + resourceName + " -o go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
	Expect(envVars).To(Equal(envString))
}

func TestOdo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "odo test suite")
}

var _ = BeforeSuite(func() {
	// Save the current project
	curProj = runCmd("oc project -q")
})

var _ = Describe("odoe2e", func() {
	var t = strconv.FormatInt(time.Now().Unix(), 10)
	var projName = fmt.Sprintf("odo-%s", t)
	const appTestName = "testing"

	tmpDir, err := ioutil.TempDir("", "odo")
	if err != nil {
		Fail(err.Error())
	}

	Context("odo service create", func() {
		It("should return error if the cluster has no service catalog deployed", func() {
			loginOutput := runCmd("odo login --username developer --password developer")
			Expect(loginOutput).To(ContainSubstring("Login successful"))
			sessionOutput := runFailCmd("odo service create", 1)
			Expect(sessionOutput).To(ContainSubstring("unable to retrieve service classes"))
		})
	})

	// TODO: Create component without creating application
	Context("odo project", func() {
		It("should create a new project", func() {
			session := runCmd("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})

		It("should get the project", func() {
			getProj := runCmd("odo project get --short")
			Expect(getProj).To(Equal(projName))
		})

		// Issue #630
		It("should list the project", func() {
			listProj := runCmd("sleep 5s && odo project list")
			Expect(listProj).To(ContainSubstring(projName))
		})
	})

	Context("odo utils config", func() {
		It("should get true for updatenotification by default globally", func() {
			configOutput := runCmd("odo utils config view --global")
			Expect(configOutput).To(ContainSubstring("true"))
			Expect(configOutput).To(ContainSubstring(config.UpdateNotificationSetting))
			Expect(configOutput).To(ContainSubstring(config.NamePrefixSetting))
			Expect(configOutput).To(ContainSubstring(config.TimeoutSetting))
		})
		It("should be checking to see if timeout is the same as the constant globally", func() {
			configOutput := runCmd("odo utils config view --global|grep Timeout")
			Expect(configOutput).To(ContainSubstring(fmt.Sprintf("%d", config.DefaultTimeout)))
		})
		It("should be checking to see if global config values are the same as the configured ones", func() {
			runCmd("odo utils config set --global updatenotification false")
			runCmd("odo utils config set --global timeout 5")
			configOutput := runCmd("odo utils config view --global |grep UpdateNotification")
			Expect(configOutput).To(ContainSubstring("false"))
			Expect(configOutput).To(ContainSubstring("UpdateNotification"))
			configOutput = runCmd("odo utils config view --global |grep Timeout")
			Expect(configOutput).To(ContainSubstring("5"))
		})
		It("should be checking to see if local config values are the same as the configured ones", func() {
			runCmd("odo utils config set componenttype java")
			configOutput := runCmd("odo utils config view|grep ComponentType")
			Expect(configOutput).To(ContainSubstring("java"))
			Expect(configOutput).To(ContainSubstring("ComponentType"))
		})

	})

	Context("creating component without an application", func() {
		It("should create the component in default application", func() {
			runCmd("odo create php testcmp")

			getCmp := runCmd("odo component get --short")
			Expect(getCmp).To(Equal("testcmp"))

			getApp := runCmd("odo app get --short")
			Expect(getApp).To(ContainSubstring("e2e-"))
		})

		It("should be able to delete the component", func() {
			runCmd("odo delete testcmp -f")

			getCmp := runCmd("odo list")
			Expect(getCmp).NotTo(ContainSubstring("testcmp"))
		})
	})

	Describe("creating an application", func() {
		Context("when application by the same name doesn't exist", func() {
			It("should create an application", func() {
				appName := runCmd("odo app create " + appTestName)
				Expect(appName).To(ContainSubstring(appTestName))
			})

			It("should get the current application", func() {
				appName := runCmd("odo app get --short")
				Expect(appName).To(Equal(appTestName))
			})

			It("should be created within the project", func() {
				projName := runCmd("odo project get --short")
				Expect(projName).To(Equal(projName))
			})

			It("should be able to create another application", func() {
				appName := runCmd("odo app create " + appTestName + "-2")
				Expect(appName).To(ContainSubstring(appTestName + "-2"))
			})

			It("should be able to list applications in current project", func() {
				appNames := runCmd("odo app list")
				Expect(appNames).To(ContainSubstring(appTestName))
				Expect(appNames).To(ContainSubstring(appTestName + "-2"))
			})

			It("should be able to delete an application", func() {
				// Cleanup
				runCmd("odo app delete " + appTestName + "-2 -f")
			})

			It("should be able to set an application as current", func() {
				appName := runCmd("odo app set " + appTestName)
				Expect(appName).To(ContainSubstring(appTestName))
			})
		})

		// TODO: Check if the application with the same name can be created
	})

	Context("should list applications in other project", func() {
		newProjName := strings.Replace(projName, "odo", "odo2", -1)
		It("should create a new project", func() {
			session := runCmd("odo project create " + newProjName)
			Expect(session).To(ContainSubstring(newProjName))
		})

		It("should get the project", func() {
			getProj := runCmd("odo project get --short")
			Expect(strings.TrimSpace(getProj)).To(Equal(newProjName))
		})

		It("should show nice message when there is no application in project", func() {
			appNames := runCmd("odo app list")
			Expect(strings.TrimSpace(appNames)).To(
				Equal("There are no applications deployed in the project '" + newProjName + "'."))
		})

		It("should be able to list applications in other project", func() {
			appNames := runCmd("odo app list --project " + projName)
			Expect(appNames).To(ContainSubstring(appTestName))
		})

		It("should set the other project as active", func() {
			setProj := runCmd("odo project set --short " + projName)
			Expect(strings.TrimSpace(setProj)).To(Equal(projName))
		})

		It("should be able to set an application as current", func() {
			appName := runCmd("odo app set " + appTestName)
			Expect(appName).To(ContainSubstring(appTestName))
		})
	})

	Describe("creating a component", func() {
		Context("when application exists", func() {
			var autoGenNodeJSCompName string

			It("should be able to create new imagestream and find it in catalog list", func() {
				curProj = strings.TrimSuffix(runCmd("oc project -q"), "\n")
				cmd := fmt.Sprintf("oc create -f "+testNamespacedImage+" -n %s", curProj)
				runCmd(cmd)
				cmpList := runCmd("odo catalog list components")
				Expect(cmpList).To(ContainSubstring(curProj))
			})

			It("should create a named component", func() {
				runCmd("git clone https://github.com/openshift/nodejs-ex " +
					tmpDir + "/nodejs-ex")

				// TODO: add tests for --git
				curProj = strings.TrimSuffix(runCmd("oc project -q"), "\n")
				// Sleep until status tags and their annotations are created
				time.Sleep(15 * time.Second)
				runCmd("odo create " + curProj + "/nodejs nodejs --local " + tmpDir + "/nodejs-ex")
				runCmd("odo push")
			})

			It("should create a component with auto-generated name", func() {
				runCmd("odo create " + curProj + "/nodejs --local " + tmpDir + "/nodejs-ex")
				runCmd("odo push")
			})

			It("should be the get the component created as active component", func() {
				autoGenNodeJSCompName = runCmd("odo component get --short")
				Expect(autoGenNodeJSCompName).To(ContainSubstring(fmt.Sprintf("nodejs-ex-%s-nodejs", curProj)))
			})

			It("should create the component within the application", func() {
				getApp := runCmd("odo app get --short")
				Expect(getApp).To(Equal(appTestName))
			})

			It("should list the components within the application", func() {
				cmpList := runCmd("odo list")
				Expect(cmpList).To(ContainSubstring("nodejs"))
			})

			It("should be able to create multiple components within the same application", func() {
				runCmd("odo create php php")
			})

			It("should list the newly created second component", func() {
				cmpList := runCmd("odo list")
				Expect(cmpList).To(ContainSubstring("php"))
			})

			It("should get the application "+appTestName, func() {
				appGet := runCmd("odo app get --short")
				Expect(appGet).To(Equal(appTestName))
			})

			It("should be able to set a component as active", func() {
				cmpSet := runCmd(fmt.Sprintf("odo component set %s", autoGenNodeJSCompName))
				Expect(cmpSet).To(ContainSubstring(autoGenNodeJSCompName))
			})

			It("should be able to retrieve logs", func() {
				runCmd("odo log")
				runCmd(fmt.Sprintf("odo log %s", autoGenNodeJSCompName))
			})

			It("should be able to create git component with required ports", func() {
				runCmd("odo create nodejs nodejs-git --git https://github.com/openshift/nodejs-ex --port 8080/tcp,9100/udp")

				// checking port names
				portsNames := runCmd("oc get services nodejs-git-" + appTestName + " -o go-template='{{range .spec.ports}}{{.name}}{{end}}'")
				Expect(portsNames).To(ContainSubstring("8080-tcp"))
				Expect(portsNames).To(ContainSubstring("9100-udp"))

				// checking port numbers
				ports := runCmd("oc get services nodejs-git-" + appTestName + " -o go-template='{{range .spec.ports}}{{.port}}{{end}}'")
				Expect(ports).To(ContainSubstring("8080"))
				Expect(ports).To(ContainSubstring("9100"))

				// checking protocols
				protocols := runCmd("oc get services nodejs-git-" + appTestName + " -o go-template='{{range .spec.ports}}{{.protocol}}{{end}}'")
				Expect(protocols).To(ContainSubstring("TCP"))
				Expect(protocols).To(ContainSubstring("UDP"))

				// deleting the component
				runCmd("odo delete -f")

				getCmp := runCmd("odo list")
				Expect(getCmp).NotTo(ContainSubstring("nodejs-git"))
			})

			It("should be able to create git component with required env vars", func() {
				runCmd("odo create nodejs nodejs-git --git https://github.com/openshift/nodejs-ex --env key=value,key1=value1")

				// checking the values of the env vars pairs in bc
				envVars := runCmd("oc get bc nodejs-git-" + appTestName + " -o go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
				Expect(envVars).To(Equal("keyvaluekey1value1"))

				// checking the values of the env vars pairs in dc
				envVars = runCmd("oc get dc nodejs-git-" + appTestName + " -o go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
				Expect(envVars).To(Equal("keyvaluekey1value1"))

				// deleting the component
				runCmd("odo delete -f")
			})
		})
	})

	Describe("Creating odo url", func() {
		Context("using odo url", func() {
			It("should create route without url name provided", func() {
				runCmd("odo component set nodejs")
				getURLOut := runCmd("odo url create")
				Expect(getURLOut).To(ContainSubstring("nodejs-8080-" + appTestName + "-" + projName))

				// check the port number of the created URL
				port := runCmd("oc get route nodejs-8080-" + appTestName + " -o go-template='{{index .spec.port.targetPort}}'")
				Expect(port).To(Equal("8080"))

				// delete the url
				runCmd("odo url delete nodejs-8080 -f")
			})

			It("should create route without port in case of single service port component", func() {
				runCmd("odo component set nodejs")
				getURLOut := runCmd("odo url create nodejs")
				Expect(getURLOut).To(ContainSubstring("nodejs-" + appTestName + "-" + projName))

				// check the port number of the created URL
				port := runCmd("oc get route nodejs-" + appTestName + " -o go-template='{{index .spec.port.targetPort}}'")
				Expect(port).To(Equal("8080"))
			})

			It("should be able to list the url", func() {
				getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
				getRoute = strings.TrimSpace(getRoute)
				Expect(getRoute).To(ContainSubstring("nodejs-" + appTestName + "-" + projName))

				// Check the labels in `oc get route`
				routeName := "nodejs-" + appTestName
				getRouteLabel := runCmd("oc get route/" + routeName + " -o jsonpath='" +
					"{.metadata.labels.app\\.kubernetes\\.io/component-name}'")
				Expect(getRouteLabel).To(Equal("nodejs"))
			})

			It("should create route with required port", func() {
				runCmd("odo create httpd httpd-test --git https://github.com/openshift/httpd-ex.git")
				getURLOut := runCmd("odo url create example-url --port 8443")
				Expect(getURLOut).To(ContainSubstring("example-url-" + appTestName + "-" + projName))

				// check the port number of the created URL
				port := runCmd("oc get route example-url-" + appTestName + " -o go-template='{{index .spec.port.targetPort}}'")
				Expect(port).To(Equal("8443"))

				// delete the component
				runCmd("odo delete httpd-test -f")
			})
		})
	})

	Describe("pushing updates", func() {
		Context("When push is made", func() {
			It("should push the changes", func() {
				// Switch to nodejs component
				runCmd("odo component set nodejs")

				getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
				getRoute = strings.TrimSpace(getRoute)

				curlRoute := waitForEqualCmd("curl -s "+getRoute+" | grep 'Welcome to your Node.js application on OpenShift' | wc -l | tr -d '\n'", "1", 10)
				if curlRoute {
					grepBeforePush := runCmd("curl -s " + getRoute + " | grep 'Welcome to your Node.js application on OpenShift'")
					log.Printf("Text before odo push: %s", strings.TrimSpace(grepBeforePush))
				}

				// Make changes to the html file
				runCmd("sed -i 's/Welcome to your Node.js application on OpenShift/Welcome to your Node.js on ODO/g' " + tmpDir + "/nodejs-ex/views/index.html")

				// Push the changes
				runCmd("odo push --local " + tmpDir + "/nodejs-ex")
			})

			It("should reflect the changes pushed", func() {

				getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
				getRoute = strings.TrimSpace(getRoute)

				curlRoute := waitForEqualCmd("curl -s "+getRoute+" | grep -i odo | wc -l | tr -d '\n'", "1", 10)
				if curlRoute {
					grepAfterPush := runCmd("curl -s " + getRoute + " | grep -i odo")
					log.Printf("Text after odo push: %s", strings.TrimSpace(grepAfterPush))
					Expect(grepAfterPush).To(ContainSubstring("ODO"))
				}
			})

			It("should be able to create the url with same name in different application", func() {
				appTestName_new := appTestName + "-1"
				runCmd("odo app create " + appTestName_new)
				runCmd("odo create nodejs nodejs-1 --git https://github.com/sclorg/nodejs-ex")
				runCmd("odo url create nodejs --port 8080")

				getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
				getRoute = strings.TrimSpace(getRoute)
				Expect(getRoute).To(ContainSubstring("nodejs-" + appTestName_new + "-" + projName))

				// Check the labels in `oc get route`
				routeName := "nodejs-" + appTestName_new
				getRouteLabel := runCmd("oc get route/" + routeName + " -o jsonpath='" +
					"{.metadata.labels.app\\.kubernetes\\.io/component-name}'")
				Expect(getRouteLabel).To(Equal("nodejs-1"))
			})

			// Check if url is deleted
			It("should be able to delete the url added", func() {
				appTestName_new := appTestName + "-1"
				runCmd("odo app set " + appTestName_new)
				runCmd("odo component set nodejs-1")
				runCmd("odo url delete nodejs -f")

				getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
				getRoute = strings.TrimSpace(getRoute)
				Expect(getRoute).NotTo(ContainSubstring("nodejs-1-" + appTestName_new + "-" + projName))

				runCmd("odo delete -f")
				runCmd("odo app delete " + appTestName_new + " -f")
			})

		})
	})

	Describe("Adding storage", func() {
		Context("when storage is added", func() {
			It("should default to active component when no component name is passed", func() {

				runCmd("odo app set " + appTestName)
				runCmd("odo component set nodejs")

				storAdd := runCmd("odo storage create pv1 --path /mnt/pv1 --size 5Gi")
				Expect(storAdd).To(ContainSubstring("nodejs"))

				// Check against path and name against dc
				getDc := runCmd("oc get dc/nodejs-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv1"))

				// Check if the storage is added on the path provided
				getMntPath := runCmd("oc get dc/nodejs-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

				Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))
			})

			It("should be able to list the storage added", func() {
				storList := runCmd("odo storage list")
				Expect(storList).To(ContainSubstring("pv1"))
			})

			It("should be able add storage to a component specified", func() {
				runCmd("odo storage create pv2 --path /mnt/pv2 --size 5Gi --component php")

				storList := runCmd("odo storage list --component php")
				Expect(storList).To(ContainSubstring("pv2"))

				// Verify with deploymentconfig
				getDc := runCmd("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv2"))

				// Check if the storage is added on the path provided
				getMntPath := runCmd("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

				Expect(getMntPath).To(ContainSubstring("/mnt/pv2"))
			})

			It("should be able to list all storage in all components", func() {
				storList := runCmd("odo storage list --all")
				Expect(storList).To(ContainSubstring("pv1"))
				Expect(storList).To(ContainSubstring("pv2"))
			})

			// TODO: Verify if the storage removed using odo deletes pvc
			It("should be able to delete the storage added", func() {
				runCmd("odo storage delete pv1 -f")

				storList := runCmd("odo storage list")
				Expect(storList).NotTo(ContainSubstring("pv1"))
			})

			It("should be able to unmount the storage using the storage name", func() {
				runCmd("odo storage unmount pv2 --component php")

				// Verify with deploymentconfig
				getDc := runCmd("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).NotTo(ContainSubstring("pv2"))
			})

			It("should be able to mount the storage to the path specified", func() {
				runCmd("odo storage mount pv2 --path /mnt/pv2 --component php")

				// Verify with deploymentconfig
				getDc := runCmd("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv2"))

				// Check if the storage is added on the path provided
				getMntPath := runCmd("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

				Expect(getMntPath).To(ContainSubstring("/mnt/pv2"))
			})

			It("should be able to unmount the storage", func() {
				runCmd("odo storage unmount /mnt/pv2 --component php")

				// Verify with deploymentconfig
				getDc := runCmd("oc get dc/php-" + appTestName + " -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).NotTo(ContainSubstring("pv2"))
			})
		})
	})

	Context("deploying a component with a specific image name", func() {
		It("should deploy the component", func() {
			runCmd("odo create nodejs:latest testversioncmp")
		})

		It("should delete the deployed image-specific component", func() {
			runCmd("odo delete testversioncmp")
		})
	})

	Context("deleting the application", func() {
		// Check if url is deleted
		It("should be able to delete the url added to the component", func() {
			runCmd("odo component set nodejs")
			runCmd("odo url delete nodejs -f")

			urlList := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
			Expect(urlList).NotTo(ContainSubstring("nodejs"))
		})

		It("should delete application and component", func() {

			runCmd("odo app delete " + appTestName + " -f")

			appGet := runCmd("odo app get --short")
			Expect(appGet).NotTo(ContainSubstring(appTestName))

			appList := runCmd("odo app list")
			Expect(appList).NotTo(ContainSubstring(appTestName))

			cmpList := runCmd("odo list")
			Expect(cmpList).NotTo(ContainSubstring("nodejs"))

			runCmd("odo project delete " + projName + " -f")
			waitForDeleteCmd("odo project list", projName)
		})
	})

	Context("validate odo version cmd with other major components version", func() {
		It("should show the version of odo major components", func() {
			odoVersion := runCmd("odo version")
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+\+\w+`)
			kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			Expect(kubernetesVersionStringMatch).Should(BeTrue())
		})

		It("should show server login URL", func() {
			odoVersion := runCmd("odo version")
			reServerURL := regexp.MustCompile(`Server:\s*https:\/\/([0-9]+.){3}[0-9]+:8443`)
			serverURLStringMatch := reServerURL.MatchString(odoVersion)
			Expect(serverURLStringMatch).Should(BeTrue())
		})
	})
})

var _ = Describe("updateE2e", func() {
	var t = strconv.FormatInt(time.Now().Unix(), 10)
	var projName = fmt.Sprintf("odo-%s", t)
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
				session := runCmd("odo project create " + projName + "-1")
				Expect(session).To(ContainSubstring(projName))
			})

			It("should get the project", func() {
				getProj := runCmd("odo project get --short")
				Expect(strings.TrimSpace(getProj)).To(Equal(projName + "-1"))
			})
		})
	})

	Describe("creating an application", func() {
		Context("when application by the same name doesn't exist", func() {
			It("should create an application", func() {
				appName := runCmd("odo app create " + appTestName)
				Expect(appName).To(ContainSubstring(appTestName))
			})
		})
	})

	Context("updating the component", func() {
		It("should be able to create binary component", func() {
			runCmd("curl -L -o " + tmpDir + "/sample-binary-testing-1.war " +
				"https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/e5bc575ac8b14ba2b23d66b5cb4873657e1a1489/sample.war")
			runCmd("odo create wildfly --binary " + tmpDir + "/sample-binary-testing-1.war  --env key=value,key1=value1")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("wildfly"))

			runCmd("oc get dc")
			runCmd("oc get bc")
		})

		It("should update component from binary to binary", func() {
			runCmd("curl -L -o " + tmpDir + "/sample-binary-testing-2.war " +
				"'https://gist.github.com/mik-dass/f95bd818ddba508ff76a386f8d984909/raw/85354d9ee8583a9c1e64a331425eede235b07a9e/sample%2520(1).war'")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-2.war")

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-2.war")
			EnvVarTest("wildfly-"+appTestName, "binary", "keyvaluekey1value1")
		})

		It("should update component from binary to local", func() {
			runCmd("git clone " + wildflyURI1 + " " +
				tmpDir + "/katacoda-odo-backend-1")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
			EnvVarTest("wildfly-"+appTestName, "local", "keyvaluekey1value1")
		})

		It("should update component from local to local", func() {
			runCmd("git clone " + wildflyURI2 + " " +
				tmpDir + "/katacoda-odo-backend-2")

			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-2")

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-2")
			EnvVarTest("wildfly-"+appTestName, "local", "keyvaluekey1value1")
		})

		It("should update component from local to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --git " + wildflyURI1)

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyURI1))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyURI1)
			EnvVarTest("wildfly-"+appTestName, "git", "keyvaluekey1value1")
		})

		It("should update component from git to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --git " + wildflyURI2)

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyURI2))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyURI2)
			EnvVarTest("wildfly-"+appTestName, "git", "keyvaluekey1value1")
		})

		It("should update component from git to binary", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
			EnvVarTest("wildfly-"+appTestName, "binary", "keyvaluekey1value1")
		})

		It("should update component from binary to git", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --git " + wildflyURI1)

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(wildflyURI1))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).NotTo(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "git", wildflyURI1)
			EnvVarTest("wildfly-"+appTestName, "git", "keyvaluekey1value1")
		})

		It("should update component from git to local", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --local " + tmpDir + "/katacoda-odo-backend-1")

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "local", "file://"+tmpDir+"/katacoda-odo-backend-1")
			EnvVarTest("wildfly-"+appTestName, "local", "keyvaluekey1value1")
		})

		It("should update component from local to binary", func() {
			waitForDCOfComponentToRolloutCompletely("wildfly")
			runCmd("odo update wildfly --binary " + tmpDir + "/sample-binary-testing-1.war")

			// checking bc for updates
			getBc := runCmd("oc get bc wildfly-" + appTestName + " -o go-template={{.spec.source.git.uri}}")
			Expect(getBc).To(Equal(bootStrapSupervisorURI))

			// checking for init containers
			getDc := runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.initContainers}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring(initContainerName))

			// checking for volumes
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.volumes}}" +
				"{{.name}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			// checking for volumes mounts
			getDc = runCmd("oc get dc wildfly-" + appTestName + " -o go-template='" +
				"{{range .spec.template.spec.containers}}{{range .volumeMounts}}{{.name}}" +
				"{{.name}}{{end}}{{end}}'")
			Expect(getDc).To(ContainSubstring("wildfly" + appRootVolumeName))

			SourceTest(appTestName, "binary", "file://"+tmpDir+"/sample-binary-testing-1.war")
			EnvVarTest("wildfly-"+appTestName, "binary", "keyvaluekey1value1")
		})
	})
	Context("odo login", func() {
		It("should login with username and password", func() {
			runCmd("oc logout")
			runCmd("odo login --username developer --password developer")
			userName := runCmd("oc whoami")
			Expect(userName).To(ContainSubstring("developer"))
		})
		It("should login with token", func() {
			userToken := runCmd("oc whoami -t")
			runCmd("oc logout")
			runCmd("odo login -t" + userToken)
			token := runCmd("oc whoami -t")
			Expect(token).To(ContainSubstring(userToken))
		})
	})
	Context("logout of the cluster", func() {
		// test for odo logout
		It("should logout the user from the cluster", func() {
			logoutMsg := runCmd("odo logout")
			Expect(logoutMsg).To(ContainSubstring("Logged"))
			Expect(logoutMsg).To(ContainSubstring("out on"))
			// validate using oc whoami
			runFailCmd("oc whoami", 1)
		})
		It("should throw error if user is not logged in", func() {
			logoutMsg := runFailCmd("odo logout", 1)
			Expect(logoutMsg).To(Equal("Please log in to the cluster\n"))
		})
	})
})
