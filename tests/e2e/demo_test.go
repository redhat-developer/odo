// +build !race

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"log"
	"strings"
)

var _ = Describe("katacodaDemo", func() {

	tmpDir, err := ioutil.TempDir("", "odoDemo")
	if err != nil {
		Fail(err.Error())
	}

	// TODO: Create component without creating application
	Context("odo project", func() {
		It("should create a new project", func() {
			session := runCmd("odo project create sample-proj")
			Expect(session).To(ContainSubstring("sample-proj"))
		})

		It("should create a new application", func() {
			appCreate := runCmd("odo app create sample")
			Expect(appCreate).To(ContainSubstring("sample"))
		})

		It("should list the application created", func() {
			appList := runCmd("odo app list")
			Expect(appList).To(ContainSubstring("sample"))
		})
	})

	Context("odo component creation", func() {
		It("should list the components in the catalog", func() {
			getProj := runCmd("odo catalog list")
			Expect(getProj).To(ContainSubstring("wildfly"))
			Expect(getProj).To(ContainSubstring("ruby"))
		})

		It("should be able to create the component", func() {
			runCmd("git clone https://github.com/marekjelen/katacoda-odo-backend " + tmpDir + "/backend")
			runCmd("cd " + tmpDir + "/backend && mvn package")

			runCmd("odo create wildfly backend --binary " + tmpDir + "/backend/target/ROOT.war")

			// Push the changes
			runCmd("odo push")
		})

		It("should list the component", func() {
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("wildfly"))
		})

		It("should add storage to the component", func() {
			storAdd := runCmd("odo storage create pv1 --path=/data --size=1G")
			Expect(storAdd).To(ContainSubstring("pv1"))
			Expect(storAdd).To(ContainSubstring("backend"))
		})

		It("should create the frontend component", func() {
			runCmd("odo create php frontend")

			runCmd("git clone https://github.com/marekjelen/katacoda-odo-frontend " +
				tmpDir + "/frontend")
			runCmd("odo push --local " + tmpDir + "/frontend")

			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("php"))

			cmpGet := runCmd("odo component get --short")
			Expect(cmpGet).To(Equal("frontend"))
		})

		It("should link the frontend and backend components", func() {
			runCmd("odo link backend --component frontend")
		})

		It("should create a url for frontend", func() {
			runCmd("odo url create frontend")

			getUrl := runCmd("odo url list")
			Expect(getUrl).To(ContainSubstring("frontend-sample-proj"))
		})
	})

	Context("edit and pushing changes to the component", func() {
		It("should push changes", func() {
			// TODO: Get watch working in the background on travis
			runCmd("sed -i 's/<h1 class=\"text-center\">/<h1 class=\"text-center\">Counter: /g' " + tmpDir + "/frontend/index.php")

			runCmd("odo push --local " + tmpDir + "/frontend")
		})

		It("should fetch the updated changes", func() {
			getRoute := runCmd("odo url list  | sed -n '1!p' | awk '{ print $3 }'")
			getRoute = strings.TrimSpace(getRoute)

			curlRoute := waitForCmdOut("curl -s "+getRoute+" | grep 'Counter' | wc -l | tr -d '\n'", "1")
			if curlRoute {
				log.Printf("Push successful")
			}
		})
	})

	Context("cleaning up", func() {
		It("should delete the application", func() {
			runCmd("odo app delete sample -f")

			runCmd("oc delete project sample-proj")
		})
	})
})
