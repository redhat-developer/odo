package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// TODO: A neater way to provide ocdev path. Currently we assume \
// ocdev and oc in $PATH already.

var t = strconv.FormatInt(time.Now().Unix(), 10)
var projName = fmt.Sprintf("ocdev-%s", t)

func runCmd(cmdS string) string {
	cmd, err := exec.Command("/bin/sh", "-c", cmdS).Output()
	if err != nil {
		log.Fatalf("Error running command: %s: %v", cmdS, err)
		os.Exit(1)
	}
	return string(cmd)
}

func createApp(appName string) {
	runCmd("ocdev application create " + appName)
}

func getApp() string {
	app := runCmd("ocdev application get -q")
	return app
}

func getCmp() string {
	cmp := runCmd("ocdev component get --short")
	return cmp
}

func pingSvc(url string) {
	var ep bool = false
	pingTimeout := time.After(10 * time.Minute)
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

var _ = BeforeSuite(func() {
	runCmd("oc new-project " + projName)
})

var _ = AfterSuite(func() {
	runCmd("oc delete project " + projName)
})

var _ = Describe("Usecase #5", func() {

	TmpDir, err := ioutil.TempDir("", "ocdev")
	if err != nil {
		log.Fatal(err)
	}

	It("create application", func() {
		createApp("usecase5")
		Expect(getApp()).To(Equal("usecase5"))
	})

	It("create component", func() {
		runCmd("git clone https://github.com/openshift/nodejs-ex " + TmpDir + "/nodejs-ex")
		runCmd("ocdev component create nodejs --dir " + TmpDir + "/nodejs-ex")
		Expect(getCmp()).To(Equal("nodejs"))
		time.Sleep(10)
	})

	It("push changes to component", func() {

		// Get IP and Port
		getIP := runCmd("oc get svc nodejs -o go-template='{{.spec.clusterIP}}:{{(index .spec.ports 0).port}}'")
		pingUrl := fmt.Sprintf("http://%s", getIP)
		pingSvc(pingUrl)

		grepBeforePush := runCmd("curl -s " + pingUrl +
			" | grep 'Welcome to your Node.js application on OpenShift'")

		log.Printf("Text before change: %s", grepBeforePush)

		// Make changes to the html file
		runCmd("sed -i 's/Welcome to your Node.js application on OpenShift/Welcome to your Node.js on OCDEV/g' " + TmpDir + "/nodejs-ex/views/index.html")

		// Push the changes
		runCmd("ocdev push --dir " + TmpDir + "/nodejs-ex")

		// ping the ip
		pingSvc(pingUrl)

		for {
			pingCmd := "curl -s " + pingUrl + " | grep -i ocdev | wc -l | tr -d '\n'"
			out, err := exec.Command("/bin/sh", "-c", pingCmd).Output()
			if err != nil {
				log.Print(err)
			}

			outInt, _ := strconv.Atoi(string(out))
			if outInt > 0 {
				grepAfterPush := runCmd("curl -s " + pingUrl + " | grep 'OCDEV'")
				log.Printf("After change: %s", string(grepAfterPush))
				break
			}
			time.Sleep(2)
		}

	})

	It("delete application", func() {
		runCmd("ocdev application delete usecase5")
		os.RemoveAll(TmpDir)
	})

})
