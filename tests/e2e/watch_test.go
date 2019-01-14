package e2e

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoWatchE2e", func() {

	const appTestName = "testing"
	const wildflyURI = "https://github.com/marekjelen/katacoda-odo-backend"
	const pythonURI = "https://github.com/OpenShiftDemos/os-sample-python"
	const nodejsURI = "https://github.com/openshift/nodejs-ex"
	const openjdkURI = "https://github.com/geoand/javalin-helloworld"

	var t = strconv.FormatInt(time.Now().Unix(), 10)
	var projName = fmt.Sprintf("odowatch-%s", t)

	tmpDir, err := ioutil.TempDir("", "odoCmp")
	if err != nil {
		Fail(err.Error())
	}

	Context("watch python component created from local source", func() {
		It("should create the project and application", func() {
			runCmd("odo project create " + projName)
			runCmd("odo app create " + appTestName)
		})

		It("should watch python component local sources for any changes", func() {
			runCmd("git clone " + pythonURI + " " + tmpDir + "/os-sample-python")
			runCmd("odo create python python-watch --local " + tmpDir + "/os-sample-python --memory 400Mi")
			runCmd("odo push -v 4")
			// Test multiple push so as to avoid regressions like: https://github.com/redhat-developer/odo/issues/1054
			runCmd("odo push -v 4")
			runCmd("odo url create")

			startSimulationCh := make(chan bool)
			go func() {
				startMsg := <-startSimulationCh
				if startMsg {
					fmt.Println("Received signal, starting file modification simulation")
					fileModificationCmd := fmt.Sprintf("sed -i 's/World/odo/g' %s", filepath.Join(tmpDir, "os-sample-python", "wsgi.py"))
					runCmd(fileModificationCmd)
					fmt.Printf("Triggered file modification %s\n\n", fileModificationCmd)
					runCmd(fmt.Sprintf("mkdir -p %s/os-sample-python/tests", tmpDir))
					runCmd(fmt.Sprintf("touch %s/os-sample-python/tests/test_1.py", tmpDir))
					runCmd(fmt.Sprintf("rm -fr %s/os-sample-python/tests", tmpDir))
					if err != nil {
						fmt.Printf("Failed performing file operation with error %v", err)
					}
				}
			}()
			success, err := pollNonRetCmdStdOutForString("odo watch python-watch -v 4", time.Duration(5)*time.Minute, func(output string) bool {
				url := runCmd("odo url list | grep `odo component get -q` | grep 8080 | awk '{print $2}'")
				curlOp := runCmd(fmt.Sprintf("curl %s", url))
				return strings.Contains(curlOp, "Hello odo")
			}, startSimulationCh, func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})
			Expect(success).To(Equal(true))
			Expect(err).To(BeNil())

			// Verify memory limits to be same as configured
			getMemoryLimit := runCmd("oc get dc python-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
			)
			Expect(getMemoryLimit).To(ContainSubstring("400Mi"))
			getMemoryRequest := runCmd("oc get dc python-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
			)
			Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
		})

		It("watch wildfly component created from local source", func() {
			runCmd("git clone " + wildflyURI + " " + tmpDir + "/katacoda-odo-backend")
			runCmd("odo create wildfly wildfly-watch --local " + tmpDir + "/katacoda-odo-backend --min-memory 400Mi --max-memory 700Mi")
			runCmd("odo push -v 4")
			// Test multiple push so as to avoid regressions like: https://github.com/redhat-developer/odo/issues/1054
			runCmd("odo push -v 4")
			runCmd("odo url create --port 8080")

			startSimulationCh := make(chan bool)
			go func() {
				startMsg := <-startSimulationCh
				if startMsg {
					fmt.Println("Received signal, starting file modification simulation")
					fileModificationPath := filepath.Join(
						tmpDir,
						"katacoda-odo-backend",
						"src",
						"main",
						"java",
						"eu",
						"mjelen",
						"katacoda",
						"odo",
						"BackendServlet.java",
					)
					fmt.Printf("Triggering filemodification @; %s\n", fileModificationPath)
					fileModificationCmd := fmt.Sprintf(
						"sed -i 's/response.getWriter().println(counter.toString())/response.getWriter().println(\"Hello odo\" + counter.toString())/g' %s",
						fileModificationPath,
					)
					runCmd(fileModificationCmd)

					runCmd(fmt.Sprintf("mkdir -p %s/katacoda-odo-backend/tests", tmpDir))
					runCmd(fmt.Sprintf("touch %s/katacoda-odo-backend/tests/test_1.java", tmpDir))
					runCmd(fmt.Sprintf("rm -fr %s/katacoda-odo-backend/tests", tmpDir))

					fmt.Printf("Triggered file modification %s\n\n", fileModificationCmd)
					if err != nil {
						fmt.Printf("Failed performing file operation with error %v", err)
					}
				}
			}()
			success, err := pollNonRetCmdStdOutForString("odo watch wildfly-watch -v 4", time.Duration(5)*time.Minute, func(output string) bool {
				url := runCmd("odo url list | grep `odo component get -q` | grep 8080 | awk '{print $2}' | tr -d '\n'")
				url = fmt.Sprintf("%s/counter", url)
				curlOp := runCmd(fmt.Sprintf("curl %s", url))
				return strings.Contains(curlOp, "Hello odo")
			}, startSimulationCh, func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})
			Expect(success).To(Equal(true))
			Expect(err).To(BeNil())

			// Verify memory limits to be same as configured
			getMemoryLimit := runCmd("oc get dc wildfly-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
			)
			Expect(getMemoryLimit).To(ContainSubstring("700Mi"))
			getMemoryRequest := runCmd("oc get dc wildfly-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
			)
			Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
		})

		It("watch openjdk component created from local source", func() {
			importOpenJDKImage()
			runCmd("git clone " + openjdkURI + " " + tmpDir + "/javalin-helloworld")
			runCmd("odo create openjdk18 openjdk-watch --local " + tmpDir + "/javalin-helloworld --min-memory 400Mi --max-memory 700Mi")
			runCmd("odo push -v 4")
			// Test multiple push so as to avoid regressions like: https://github.com/redhat-developer/odo/issues/1054
			runCmd("odo push -v 4")
			runCmd("odo url create --port 8080")

			startSimulationCh := make(chan bool)
			go func() {
				startMsg := <-startSimulationCh
				if startMsg {
					fmt.Println("Received signal, starting file modification simulation")
					fileModificationPath := filepath.Join(
						tmpDir,
						"javalin-helloworld",
						"src",
						"main",
						"java",
						"Application.java",
					)
					fileModificationCmd := fmt.Sprintf(
						"sed -i 's/Hello World/Hello odo/g' %s",
						fileModificationPath,
					)
					runCmd(fileModificationCmd)

					runCmd(fmt.Sprintf("mkdir -p %s/javalin-helloworld/tests", tmpDir))
					runCmd(fmt.Sprintf("touch %s/javalin-helloworld/tests/test_1.java", tmpDir))
					runCmd(fmt.Sprintf("rm -fr %s/javalin-helloworld/tests", tmpDir))
				}
			}()
			success, err := pollNonRetCmdStdOutForString("odo watch openjdk-watch -v 4", time.Duration(5)*time.Minute, func(output string) bool {
				url := runCmd("odo url list | grep `odo component get -q` | grep 8080 | awk '{print $2}' | tr -d '\n'")
				curlOp := runCmd(fmt.Sprintf("curl %s", url))
				return strings.Contains(curlOp, "Hello odo")
			}, startSimulationCh, func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})
			Expect(success).To(Equal(true))
			Expect(err).To(BeNil())

			// Verify memory limits to be same as configured
			getMemoryLimit := runCmd("oc get dc openjdk-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
			)
			Expect(getMemoryLimit).To(ContainSubstring("700Mi"))
			getMemoryRequest := runCmd("oc get dc openjdk-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
			)
			Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
		})

		It("watch openjdk component created from local binary", func() {
			importOpenJDKImage()
			runCmd("git clone " + openjdkURI + " " + tmpDir + "/javalin-helloworld")
			runCmd("mvn clean package -f " + tmpDir + "/javalin-helloworld")
			runCmd("odo create openjdk18 openjdk-watch --binary " + tmpDir + "/javalin-helloworld/target/javalin-hello-world-0.1-SNAPSHOT.jar --min-memory 400Mi --max-memory 700Mi")
			runCmd("odo push -v 4")
			// Test multiple push so as to avoid regressions like: https://github.com/redhat-developer/odo/issues/1054
			runCmd("odo push -v 4")
			runCmd("odo url create --port 8080")

			startSimulationCh := make(chan bool)
			go func() {
				startMsg := <-startSimulationCh
				if startMsg {
					fmt.Println("Received signal, starting file modification simulation")
					fileModificationPath := filepath.Join(
						tmpDir,
						"javalin-helloworld",
						"src",
						"main",
						"java",
						"Application.java",
					)
					fileModificationCmd := fmt.Sprintf(
						"sed -i 's/Hello World/Hello odo/g' %s",
						fileModificationPath,
					)
					runCmd(fileModificationCmd)

					runCmd(fmt.Sprintf("mkdir -p %s/javalin-helloworld/tests", tmpDir))
					runCmd(fmt.Sprintf("touch %s/javalin-helloworld/tests/test_1.java", tmpDir))
					runCmd(fmt.Sprintf("rm -fr %s/javalin-helloworld/tests", tmpDir))
					runCmd("mvn clean package -f " + tmpDir + "/javalin-helloworld")
				}
			}()
			success, err := pollNonRetCmdStdOutForString("odo watch openjdk-watch -v 4", time.Duration(5)*time.Minute, func(output string) bool {
				url := runCmd("odo url list | grep `odo component get -q` | grep 8080 | awk '{print $2}' | tr -d '\n'")
				curlOp := runCmd(fmt.Sprintf("curl %s", url))
				return strings.Contains(curlOp, "Hello odo")
			}, startSimulationCh, func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})
			Expect(success).To(Equal(true))
			Expect(err).To(BeNil())

			// Verify memory limits to be same as configured
			getMemoryLimit := runCmd("oc get dc openjdk-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
			)
			Expect(getMemoryLimit).To(ContainSubstring("700Mi"))
			getMemoryRequest := runCmd("oc get dc openjdk-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
			)
			Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
		})

		It("watch nodejs component created from local source", func() {
			runCmd("git clone " + nodejsURI + " " + tmpDir + "/nodejs-ex")
			runCmd("odo create nodejs nodejs-watch --local " + tmpDir + "/nodejs-ex --min-memory 400Mi --max-memory 700Mi")
			runCmd("odo push -v 4")
			// Test multiple push so as to avoid regressions like: https://github.com/redhat-developer/odo/issues/1054
			runCmd("odo push -v 4")
			runCmd("odo url create --port 8080")

			startSimulationCh := make(chan bool)
			go func() {
				startMsg := <-startSimulationCh
				if startMsg {
					fmt.Println("Received signal, starting file modification simulation")
					fileModificationPath := filepath.Join(
						tmpDir,
						"nodejs-ex",
						"server.js",
					)
					fmt.Printf("Triggering filemodification @; %s\n", fileModificationPath)
					fileModificationCmd := fmt.Sprintf(
						"sed -i \"s/res.send('{ pageCount: -1 }')/res.send('{ pageCount: -1, message: Hello odo }')/g\" %s",
						fileModificationPath,
					)
					runCmd(fileModificationCmd)
					fmt.Printf("Triggered file modification %s\n\n", fileModificationCmd)
					if err != nil {
						fmt.Printf("Failed performing file operation with error %v", err)
					}

					runCmd(fmt.Sprintf("mkdir -p %s/nodejs-ex/tests/sample-tests", tmpDir))
					runCmd(fmt.Sprintf("touch %s/nodejs-ex/tests/sample-tests/test_1.js", tmpDir))
					runCmd(fmt.Sprintf("rm -fr %s/nodejs-ex/tests/sample-tests", tmpDir))
				}
			}()
			success, err := pollNonRetCmdStdOutForString("odo watch nodejs-watch -v 4", time.Duration(20)*time.Minute, func(output string) bool {
				url := runCmd("odo url list | grep `odo component get -q` | grep 8080 | awk '{print $2}' | tr -d '\n'")
				url = fmt.Sprintf("%s/pagecount", url)
				curlOp := runCmd(fmt.Sprintf("curl %s", url))
				return strings.Contains(curlOp, "Hello odo")
			}, startSimulationCh, func(output string) bool {
				return strings.Contains(output, "Waiting for something to change")
			})
			Expect(success).To(Equal(true))
			Expect(err).To(BeNil())

			// Verify memory limits to be same as configured
			getMemoryLimit := runCmd("oc get dc nodejs-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'",
			)
			Expect(getMemoryLimit).To(ContainSubstring("700Mi"))
			getMemoryRequest := runCmd("oc get dc nodejs-watch-" +
				appTestName +
				" -o go-template='{{range .spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'",
			)
			Expect(getMemoryRequest).To(ContainSubstring("400Mi"))
		})
	})
})
