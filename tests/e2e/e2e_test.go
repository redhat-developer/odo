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

// TODO: A neater way to provide ocdev path. Currently we assume \
// ocdev and oc in $PATH already.

var t = strconv.FormatInt(time.Now().Unix(), 10)
var projName = fmt.Sprintf("ocdev-%s", t)

func runCmd(cmdS string) string {
	cmd := exec.Command("/bin/sh", "-c", cmdS)
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
			Fail("could not ping the specific service in given time: 10 minutes")

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

func TestOCdev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ocdev test suite")
}

var _ = Describe("ocdev", func() {

	tmpDir, err := ioutil.TempDir("", "ocdev")
	if err != nil {
		Fail(err.Error())
	}

	// TODO: Create component without creating application
	Context("ocdev project", func() {
		It("should create a new project", func() {
			session := runCmd("ocdev project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})

		It("should get the project", func() {
			getProj := runCmd("ocdev project get --short")
			Expect(strings.TrimSpace(getProj)).To(Equal(projName))
		})
	})

	Describe("creating an application", func() {
		Context("when application by the same name doesn't exist", func() {
			It("should create an application", func() {
				appName := runCmd("ocdev application create usecase5")
				Expect(appName).To(ContainSubstring("usecase5"))
			})

			It("should get the current application", func() {
				appName := runCmd("ocdev application get --short")
				Expect(appName).To(Equal("usecase5"))
			})

			It("should be created within the project", func() {
				projName := runCmd("ocdev project get --short")
				Expect(projName).To(ContainSubstring(projName))
			})

			It("should be able to create another application", func() {
				appName := runCmd("ocdev application create usecase5-2")
				Expect(appName).To(ContainSubstring("usecase5-2"))
			})

			It("should be able to delete an application", func() {
				// Cleanup
				runCmd("ocdev application delete usecase5-2 -f")
			})

			It("should be able to set an application as current", func() {
				appName := runCmd("ocdev application set usecase5")
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
				runCmd("ocdev create nodejs --local " + tmpDir + "/nodejs-ex")
			})

			It("should be the get the component created as active component", func() {
				cmp := runCmd("ocdev component get --short")
				Expect(cmp).To(Equal("nodejs"))
			})

			It("should create the component within the application", func() {
				getApp := runCmd("ocdev application get --short")
				Expect(getApp).To(Equal("usecase5"))
			})

			It("should list the components within the application", func() {
				cmpList := runCmd("ocdev list")
				Expect(cmpList).To(ContainSubstring("nodejs"))
			})

			It("should be able to create multiple components within the same application", func() {
				runCmd("ocdev create php")
			})

			It("should list the newly created second component", func() {
				cmpList := runCmd("ocdev list")
				Expect(cmpList).To(ContainSubstring("php"))
			})

			It("should get the application usecase5", func() {
				appGet := runCmd("ocdev application get --short")
				Expect(appGet).To(Equal("usecase5"))
			})

			It("should be able to set a component as active", func() {
				cmpSet := runCmd("ocdev component set nodejs")
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
				runCmd("sed -i 's/Welcome to your Node.js application on OpenShift/Welcome to your Node.js on OCDEV/g' " + tmpDir + "/nodejs-ex/views/index.html")

				// Push the changes
				runCmd("ocdev push --local " + tmpDir + "/nodejs-ex")
			})

		})
	})

	Describe("Creating url", func() {
		Context("using ocdev url", func() {
			It("should create route", func() {
				runCmd("ocdev url create nodejs")
			})

			It("should be able to list the url", func() {
				getRoute := runCmd("ocdev url list  | sed -n '1!p' | awk '{ print $3 }'")
				getRoute = strings.TrimSpace(getRoute)
				Expect(getRoute).To(ContainSubstring("nodejs-" + projName))

				for {
					pingCmd := "curl -s " + getRoute + " | grep -i ocdev | wc -l | tr -d '\n'"
					out, err := exec.Command("/bin/sh", "-c", pingCmd).Output()
					if err != nil {
						Fail(err.Error())
					}

					outInt, _ := strconv.Atoi(string(out))
					if outInt > 0 {
						grepAfterPush := runCmd("curl -s " + getRoute + " | grep -i ocdev")
						log.Printf("After change: %s", strings.TrimSpace(grepAfterPush))
						break
					}
					time.Sleep(5)
				}
			})
		})
	})

	Describe("Adding storage", func() {
		Context("when storage is added", func() {
			It("should default to active component when no component name is passed", func() {
				storAdd := runCmd("ocdev storage add pv1 --path /mnt/pv1 --size 5Gi")
				Expect(storAdd).To(ContainSubstring("nodejs"))

				// Check against path and name against dc
				getDc := runCmd("oc get dc/nodejs -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(Equal("pv1"))

				// Check if the storage is added on the path provided
				getMntPath := runCmd("oc get dc/nodejs -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.mountPath}}{{end}}{{end}}'")

				Expect(getMntPath).To(Equal("/mnt/pv1"))
			})

			It("should be able to list the storage added", func() {
				storList := runCmd("ocdev storage list")
				Expect(storList).To(ContainSubstring("pv1"))
			})

			// TODO: Verify if the storage removed using ocdev deletes pvc
			It("should be able to delete the storage added", func() {
				runCmd("ocdev storage remove pv1")

				storList := runCmd("ocdev storage list")
				Expect(storList).NotTo(ContainSubstring("pv1"))
			})

			It("should be able add storage to a component specified", func() {
				runCmd("ocdev storage add pv2 --path /mnt/pv2 --size 5Gi --component php")

				storList := runCmd("ocdev storage list --component php")
				Expect(storList).To(ContainSubstring("pv2"))

				// Verify with deploymentconfig
				getDc := runCmd("oc get dc/php -o go-template='" +
					"{{range .spec.template.spec.containers}}" +
					"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

				Expect(getDc).To(Equal("pv2"))

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
			runCmd("ocdev application delete usecase5 -f")

			appGet := runCmd("ocdev application get --short")
			Expect(appGet).To(Equal(""))

			appList := runCmd("ocdev application list")
			Expect(appList).NotTo(ContainSubstring("usecase5"))

			cmpList := runCmd("ocdev list")
			Expect(cmpList).NotTo(ContainSubstring("nodejs"))

			// TODO: `ocdev project delete` once implemented
			runCmd("oc delete project " + projName)
		})
	})
})
