// +build !race

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gexec "github.com/onsi/gomega/gexec"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TODO: A neater way to provide odo path. Currently we assume \
// odo and oc in $PATH already.

var t = strconv.FormatInt(time.Now().Unix(), 10)
var projName = fmt.Sprintf("odo-%s", t)

func runCmd(cmdS string) string {
	cmd := exec.Command("/bin/sh", "-c", cmdS)
	fmt.Fprintf(GinkgoWriter, "Running command: %s\n", cmdS)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

	// wait for the command execution to complete
	<-session.Exited
	Expect(session.ExitCode()).To(Equal(0))
	Expect(err).NotTo(HaveOccurred())

	return string(session.Out.Contents())
}

func pingSvc(url string) {
	var ep bool = false
	pingTimeout := time.After(5 * time.Minute)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-pingTimeout:
			Fail("could not ping the specific service in given time: 5 minutes")

		case <-tick:
			httpTimeout := time.Duration(10 * time.Second)
			client := http.Client{
				Timeout: httpTimeout,
			}

			response, err := client.Get(url)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if response.Status == "200 OK" {
				ep = true

			} else {
				Fail("for service")
			}
		}

		if ep {
			break
		}
	}
}

// waitForCmdOut runs a command until it gets
// the expected output.
// It accepts 2 arguments, cmd (command to be run)
// & expOut (expected output).
// It times out if the command doesn't fetch the
// expected output  within the timeout period (1m).
func waitForCmdOut(cmd string, expOut string) bool {

	pingTimeout := time.After(1 * time.Minute)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-pingTimeout:
			Fail("Timeout out after 1 minute")

		case <-tick:
			out, err := exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil {
				Fail(err.Error())
			}

			if string(out) == expOut {
				return true
			}
		}
	}
}

func TestOdo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "odo test suite")
}

var _ = Describe("odoe2e", func() {

	tmpDir, err := ioutil.TempDir("", "odo")
	if err != nil {
		Fail(err.Error())
	}

	// TODO: Create component without creating application
	Context("odo project", func() {
		It("should create a new project", func() {
			session := runCmd("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})

		It("should get the project", func() {
			getProj := runCmd("odo project get --short")
			Expect(strings.TrimSpace(getProj)).To(Equal(projName))
		})
	})

	Context("creating component without an application", func() {
		It("should create the component in default application", func() {
			runCmd("odo create php testcmp")

			getCmp := runCmd("odo component get --short")
			Expect(getCmp).To(Equal("testcmp"))

			getApp := runCmd("odo app get --short")
			Expect(getApp).To(Equal("app"))
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
				appName := runCmd("odo app create usecase5")
				Expect(appName).To(ContainSubstring("usecase5"))
			})

			It("should get the current application", func() {
				appName := runCmd("odo app get --short")
				Expect(appName).To(Equal("usecase5"))
			})

			It("should be created within the project", func() {
				projName := runCmd("odo project get --short")
				Expect(projName).To(ContainSubstring(projName))
			})

			It("should be able to create another application", func() {
				appName := runCmd("odo app create usecase5-2")
				Expect(appName).To(ContainSubstring("usecase5-2"))
			})

			It("should be able to delete an application", func() {
				// Cleanup
				runCmd("odo app delete usecase5-2 -f")
			})

			It("should be able to set an application as current", func() {
				appName := runCmd("odo app set usecase5")
				Expect(appName).To(ContainSubstring("usecase5"))
			})
		})

		// TODO: Check if the application with the same name can be created
	})

	Describe("creating a component", func() {
		Context("when application exists", func() {
			It("should create a component", func() {
				runCmd("git clone https://github.com/openshift/nodejs-ex " +
					tmpDir + "/nodejs-ex")

				// TODO: add tests for --git
				runCmd("odo create nodejs --local " + tmpDir + "/nodejs-ex")
				runCmd("odo push")
			})

			It("should be the get the component created as active component", func() {
				cmp := runCmd("odo component get --short")
				Expect(cmp).To(Equal("nodejs"))
			})

			It("should create the component within the application", func() {
				getApp := runCmd("odo app get --short")
				Expect(getApp).To(Equal("usecase5"))
			})

			It("should list the components within the application", func() {
				cmpList := runCmd("odo list")
				Expect(cmpList).To(ContainSubstring("nodejs"))
			})

			It("should be able to create multiple components within the same application", func() {
				runCmd("odo create php")
			})

			It("should list the newly created second component", func() {
				cmpList := runCmd("odo list")
				Expect(cmpList).To(ContainSubstring("php"))
			})

			It("should get the application usecase5", func() {
				appGet := runCmd("odo app get --short")
				Expect(appGet).To(Equal("usecase5"))
			})

			It("should be able to set a component as active", func() {
				cmpSet := runCmd("odo component set nodejs")
				Expect(cmpSet).To(ContainSubstring("nodejs"))
			})
		})
	})

	Describe("pushing updates", func() {
		Context("When push is made", func() {
			It("should push the changes", func() {

				// Get IP and port
				getIP := runCmd("oc get svc nodejs -o go-template='{{.spec.clusterIP}}:{{(index .spec.ports 0).port}}'")
				pingUrl := fmt.Sprintf("http://%s", getIP)
				pingSvc(pingUrl)

				// Text before changes
				grepBeforePush := runCmd("curl -s " + pingUrl +
					" | grep 'Welcome to your Node.js application on OpenShift'")

				log.Printf("Text before change: %s", strings.TrimSpace(grepBeforePush))

				// Make changes to the html file
				runCmd("sed -i 's/Welcome to your Node.js application on OpenShift/Welcome to your Node.js on ODO/g' " + tmpDir + "/nodejs-ex/views/index.html")

				// Push the changes
				runCmd("odo push --local " + tmpDir + "/nodejs-ex")
			})

		})
	})

	Describe("Creating url", func() {
		Context("using odo url", func() {
			It("should create route", func() {
				runCmd("odo url create nodejs")
			})

			It("should be able to list the url", func() {
				getRoute := runCmd("odo url list  | sed -n '1!p' | awk '{ print $3 }'")
				getRoute = strings.TrimSpace(getRoute)
				Expect(getRoute).To(ContainSubstring("nodejs-" + projName))

				curlRoute := waitForCmdOut("curl -s "+getRoute+" | grep -i odo | wc -l | tr -d '\n'", "1")
				if curlRoute {
					grepAfterPush := runCmd("curl -s " + getRoute + " | grep -i odo")
					log.Printf("After change: %s", strings.TrimSpace(grepAfterPush))
				}
			})
		})
	})

	Describe("Adding storage", func() {
		Context("when storage is added", func() {
			It("should default to active component when no component name is passed", func() {
				storAdd := runCmd("odo storage create pv1 --path /mnt/pv1 --size 5Gi")
				Expect(storAdd).To(ContainSubstring("nodejs"))

				// Check against path and name against dc
				getDc := runCmd("oc get dc/nodejs -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv1"))

				// Check if the storage is added on the path provided
				getMntPath := runCmd("oc get dc/nodejs -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}}{{end}}{{end}}'")

				Expect(getMntPath).To(Equal("/mnt/pv1"))
			})

			It("should be able to list the storage added", func() {
				storList := runCmd("odo storage list")
				Expect(storList).To(ContainSubstring("pv1"))
			})

			// TODO: Verify if the storage removed using odo deletes pvc
			It("should be able to delete the storage added", func() {
				runCmd("odo storage delete pv1")

				storList := runCmd("odo storage list")
				Expect(storList).NotTo(ContainSubstring("pv1"))
			})

			It("should be able add storage to a component specified", func() {
				runCmd("odo storage create pv2 --path /mnt/pv2 --size 5Gi --component php")

				storList := runCmd("odo storage list --component php")
				Expect(storList).To(ContainSubstring("pv2"))

				// Verify with deploymentconfig
				getDc := runCmd("oc get dc/php -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(ContainSubstring("pv2"))

				// Check if the storage is added on the path provided
				getMntPath := runCmd("oc get dc/php -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}}{{end}}{{end}}'")

				Expect(getMntPath).To(Equal("/mnt/pv2"))
			})
		})
	})

	Context("deleting the application", func() {
		It("should delete application and component", func() {
			runCmd("odo app delete usecase5 -f")

			appGet := runCmd("odo app get --short")
			Expect(appGet).To(Equal(""))

			appList := runCmd("odo app list")
			Expect(appList).NotTo(ContainSubstring("usecase5"))

			cmpList := runCmd("odo list")
			Expect(cmpList).NotTo(ContainSubstring("nodejs"))

			// TODO: `odo project delete` once implemented
			runCmd("oc delete project " + projName)
		})
	})
})
