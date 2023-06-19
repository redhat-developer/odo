package integration

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
	"net/http"
	"path/filepath"
)

var _ = Describe("odo dev command with api server tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	for _, podman := range []bool{false, true} {
		podman := podman
		for _, customPort := range []bool{false, true} {
			customPort := customPort
			When("the component is bootstrapped", helper.LabelPodmanIf(podman, func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"), cmpName)
				})
				When(fmt.Sprintf("odo dev is run with --api-server flag (custom api server port=%v)", customPort), func() {
					var (
						devSession helper.DevSession
						localPort  = helper.GetCustomStartPort()
					)
					BeforeEach(func() {
						opts := helper.DevSessionOpts{
							RunOnPodman:    podman,
							StartAPIServer: true,
							EnvVars:        []string{"ODO_EXPERIMENTAL_MODE=true"},
						}
						if customPort {
							opts.APIServerPort = localPort
						}
						var err error
						devSession, err = helper.StartDevMode(opts)
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})
					It("should start the Dev server when --api-server flag is passed", func() {
						if customPort {
							Expect(devSession.APIServerEndpoint).To(ContainSubstring(fmt.Sprintf("%d", localPort)))
						}
						url := fmt.Sprintf("http://%s/instance", devSession.APIServerEndpoint)
						resp, err := http.Get(url)
						Expect(err).ToNot(HaveOccurred())
						// TODO: Change this once it is implemented
						Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusNotImplemented))
					})
				})
			}))
		}
	}
})
