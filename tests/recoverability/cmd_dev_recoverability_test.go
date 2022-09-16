//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package recoverability

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("Recoverability Test", func() {
	var commonVar helper.CommonVar
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
	})
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	checkIfDevEnvIsUp := func(url, assertString string) {
		Eventually(func() string {
			resp, err := http.Get(fmt.Sprintf("http://%s", url))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			return string(body)
		}, 120*time.Second, 15*time.Second).Should(Equal(assertString))
	}

	Context("starting with empty Directory", func() {
		componentName := helper.RandString(6)
		var listenPort string
		var _ = BeforeEach(func() {
			helper.Chdir(commonVar.Context)
			Expect(helper.ListFilesInDir(commonVar.Context)).To(BeEmpty())
		})

		It("should recover from network disconnection", func() {

			command := []string{"odo", "init"}
			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {

				helper.ExpectString(ctx, "Select language")
				helper.SendLine(ctx, "javascript")

				helper.ExpectString(ctx, "Select project type")
				helper.SendLine(ctx, "Node.js\n")

				helper.ExpectString(ctx, "Which starter project do you want to use")
				helper.SendLine(ctx, "nodejs-starter\n")

				helper.ExpectString(ctx, "Enter component name")
				helper.SendLine(ctx, componentName)

				helper.ExpectString(ctx, "Your new component '"+componentName+"' is ready in the current directory")

			})
			Expect(err).To(BeNil())
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("devfile.yaml"))
			Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElement("server.js"))

			// "execute odo dev and add changes to application"
			var devSession helper.DevSession
			var ports map[string]string

			devSession, _, _, ports, err = helper.StartDevMode(nil)
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js", "from updated Node.js")
			Expect(err).ToNot(HaveOccurred())
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should update the changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from updated Node.js Starter Application!")

			// "changes are made to the applications"
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from updated Node.js", "from Node.js app v2")
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should deploy new changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from Node.js app v2 Starter Application!")

			//Trigger network latency emulation
			//setup proxy to emulate network disconnect and / or latency
			//using https://github.com/Shopify/toxiproxy
			if os.Getenv("KUBERNETES") != "true" {
				listenPort = os.Getenv("IBM_OPENSHIFT_ENDPOINT")
			} else {
				listenPort = "6443"
			}
			toxiClient := toxiproxy.NewClient("localhost:8474")
			proxies, err := toxiClient.CreateProxy("recoverability", listenPort, "localhost:6443")
			if err != nil {
				fmt.Println("Failed to create toxiproxy", err)
			}
			proxies.AddToxic("latency_down", "latency", "downstream", 100.0, toxiproxy.Attributes{
				"latency": 100000,
			})

			// "changes are made to the applications"
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js app v2", "from Node.js app v3")
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())
			// "should not deploy new changes"
			checkIfDevEnvIsUp(ports["3000"], "Hello from Node.js app v3 Starter Application!")

			//Remove network latency
			proxies.Disable()
			proxies.Delete()

			// Make and push changes should now be successfull
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js app v2", "from Node.js app v3")
			_, _, _, err = devSession.WaitSync()
			Expect(err).ToNot(HaveOccurred())

			// exit dev mode
			devSession.Stop()
		})
	})

})
