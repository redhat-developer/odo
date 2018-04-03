// +build !race

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	var cmd *exec.Cmd
	var out, stderr bytes.Buffer

	cmd = exec.Command("/bin/sh", "-c", cmdS)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		os.Exit(1)
	}

	fmt.Println(string(out.Bytes()))
	return string(out.Bytes())
}

func createApp(appName string) {
	runCmd("ocdev application create " + appName)
}

func getProj() string {
	proj := runCmd("ocdev project get --short")
	return strings.TrimSpace(proj)
}

func getApp() string {
	app := runCmd("ocdev application get --short")
	return app
}

func getCmp() string {
	cmp := runCmd("ocdev component get --short")
	return cmp
}

func pingSvc(url string) {
	var ep bool = false
	pingTimeout := time.After(5 * time.Minute)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-pingTimeout:
			log.Fatal("could not ping the specific service in given time: 10 minutes")
			os.Exit(1)

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
				log.Fatal("for service")
				os.Exit(1)
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

var _ = Describe("Usecase #5", func() {

	tmpDir, err := ioutil.TempDir("", "ocdev")
	if err != nil {
		log.Fatal(err)
	}

	runCmd("git clone https://github.com/openshift/nodejs-ex " + tmpDir + "/nodejs-ex")

	// TODO: Create component without creating application
	Context("ocdev project", func() {
		It("should create a new project", func() {
			runCmd("ocdev project create " + projName)
			Expect(getProj()).To(Equal(projName))
		})
	})

	Context("creating an application", func() {
		Context("when application by the same name doesn't exist", func() {
			It("should create an application", func() {
				createApp("usecase5")
				Expect(getApp()).To(Equal("usecase5"))
			})

			It("should be created within the project", func() {
				Expect(getProj()).To(Equal(projName))
			})

			It("should be able to set an application as current", func() {
				createApp("usecase5-2")
				Expect(getApp()).To(Equal("usecase5-2"))

				runCmd("ocdev application set usecase5")
				Expect(getApp()).To(Equal("usecase5"))
			})
		})

		// TODO: Check if the application with the same name can be created
	})

	Context("creating a component", func() {

		Context("when application exists", func() {
			It("should create a component", func() {
				runCmd("ocdev create nodejs --local " + tmpDir + "/nodejs-ex")
				Expect(getCmp()).To(Equal("nodejs"))
				time.Sleep(10)
			})

			It("should create the component within the application", func() {
				Expect(getApp()).To(Equal("usecase5"))
			})

			Context("when all the components are listed", func() {
				It("should list the components within the application", func() {
					cmpList := runCmd("ocdev list")
					Expect(cmpList).To(ContainSubstring("nodejs"))
				})
			})
		})

	})

	Context("When push is made", func() {
		It("should push reflect changes", func() {

			// Get IP and port
			getIP := runCmd("oc get svc nodejs -o go-template='{{.spec.clusterIP}}:{{(index .spec.ports 0).port}}'")
			pingUrl := fmt.Sprintf("http://%s", getIP)

			pingSvc(pingUrl)
			grepBeforePush := runCmd("curl -s " + pingUrl +
				" | grep 'Welcome to your Node.js application on OpenShift'")

			log.Printf("Text before change: %s", grepBeforePush)

			// Make changes to the html file
			runCmd("sed -i 's/Welcome to your Node.js application on OpenShift/Welcome to your Node.js on OCDEV/g' " +
				tmpDir + "/nodejs-ex/views/index.html")

			// Push the changes
			runCmd("ocdev push --local " + tmpDir + "/nodejs-ex")

			// Create route
			runCmd("ocdev url create nodejs")
			getRoute := runCmd("ocdev url list  | sed -n '1!p' | awk '{ print $3 }'")
			getRoute = strings.TrimSpace(getRoute)
			Expect(getRoute).To(ContainSubstring("nodejs-" + projName))

			for {
				pingCmd := "curl -s " + getRoute + " | grep -i ocdev | wc -l | tr -d '\n'"
				out, err := exec.Command("/bin/sh", "-c", pingCmd).Output()
				if err != nil {
					log.Fatal(err)
				}

				outInt, _ := strconv.Atoi(string(out))
				if outInt > 0 {
					grepAfterPush := runCmd("curl -s " + getRoute + " | grep -i ocdev")
					log.Printf("After change: %s", string(grepAfterPush))
					break
				}
				time.Sleep(5)
			}

		})
	})

	Context("deleting the application", func() {
		It("should delete application and component", func() {
			runCmd("ocdev application delete usecase5 -f")
			Expect(getApp()).To(Equal(""))

			cmpList := runCmd("ocdev list")
			Expect(cmpList).NotTo(ContainSubstring("nodejs"))
		})

		// TODO: `ocdev project delete` once implemented
	})
})
