package integration

import (
	"path"
	"path/filepath"
	"regexp"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/tidwall/gjson"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo list with devfile", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	Context("listing non-odo managed components", func() {
		When("a non-odo managed component is deployed", func() {
			const (
				// hard coded names from the deployment-app-label.yaml
				deploymentName = "example-deployment"
				managedBy      = "some-tool"
			)
			BeforeEach(func() {
				commonVar.CliRunner.Run("create", "-f", helper.GetExamplePath("manifests", "deployment-app-label.yaml"))
			})
			AfterEach(func() {
				commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "deployment-app-label.yaml"))
			})
			It("should list the component with odo list", func() {
				output := helper.Cmd("odo", "list", "component").ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{deploymentName, "Unknown", "None", managedBy})
			})
			It("should list the component in JSON", func() {
				output := helper.Cmd("odo", "list", "component", "-o", "json").ShouldPass().Out()
				helper.JsonPathContentIs(output, "components.0.name", deploymentName)
				Expect(gjson.Get(output, "components.0.runningIn").String()).To(BeEmpty())
				helper.JsonPathContentIs(output, "components.0.projectType", "Unknown")
				helper.JsonPathContentIs(output, "components.0.managedBy", managedBy)
			})
		})
		When("a non-odo managed component without the managed-by label is deployed", func() {
			const (
				// hard coded names from the deployment-without-managed-by-label.yaml
				deploymentName = "java-springboot-basic"
			)
			BeforeEach(func() {
				commonVar.CliRunner.Run("create", "-f", helper.GetExamplePath("manifests", "deployment-without-managed-by-label.yaml"))
			})
			AfterEach(func() {
				commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "deployment-without-managed-by-label.yaml"))
			})
			It("should list the component with odo list", func() {
				output := helper.Cmd("odo", "list", "component").ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{deploymentName, "Unknown", "None", "Unknown"})
			})
			It("should list the component in JSON", func() {
				output := helper.Cmd("odo", "list", "component", "-o", "json").ShouldPass().Out()
				helper.JsonPathContentContain(output, "components.0.name", deploymentName)
				Expect(gjson.Get(output, "components.0.runningIn").String()).To(BeEmpty())
				helper.JsonPathContentContain(output, "components.0.projectType", "Unknown")
				helper.JsonPathContentContain(output, "components.0.managedBy", "")
			})
		})
		When("an operator managed deployment(without instance and managed-by label) is deployed", func() {
			deploymentName := "nginx"
			BeforeEach(func() {
				commonVar.CliRunner.Run("create", "deployment", deploymentName, "--image=nginx")
			})
			AfterEach(func() {
				commonVar.CliRunner.Run("delete", "deployment", deploymentName)
			})
			It("should not be listed in the odo list output", func() {
				output := helper.Cmd("odo", "list", "component").ShouldRun().Out()
				Expect(output).ToNot(ContainSubstring(deploymentName))

			})
		})
	})

	When("a component created in 'app' application", func() {

		var devSession helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.Chdir(commonVar.Context)
			var err error
			devSession, _, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			devSession.Stop()
			devSession.WaitEnd()
		})

		var checkList = func(componentType string) {
			By("checking the normal output", func() {
				stdOut := helper.Cmd("odo", "list", "component").ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring(componentType))
			})
		}
		Context("verifying the managedBy Version in the odo list output", func() {
			var version string
			BeforeEach(func() {
				versionOut := helper.Cmd("odo", "version").ShouldPass().Out()
				reOdoVersion := regexp.MustCompile(`v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?`)
				version = reOdoVersion.FindString(versionOut)

			})
			It("should show managedBy Version", func() {
				By("checking the normal output", func() {
					stdout := helper.Cmd("odo", "list", "component").ShouldPass().Out()
					Expect(stdout).To(ContainSubstring(version))
				})
				By("checking the JSON output", func() {
					stdout := helper.Cmd("odo", "list", "component", "-o", "json").ShouldPass().Out()
					helper.JsonPathContentContain(stdout, "components.0.managedByVersion", version)
				})
			})
		})

		It("show an odo deploy or dev in the list", func() {
			By("should display the component as 'Dev' in odo list", func() {
				checkList("Dev")
			})

			By("should display the component as 'Dev' in odo list -o json", func() {
				res := helper.Cmd("odo", "list", "component", "-o", "json").ShouldPass()
				stdout, stderr := res.Out(), res.Err()
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentContain(stdout, "components.0.runningIn.dev", "true")
				helper.JsonPathContentContain(stdout, "components.0.runningIn.deploy", "")
			})

			// Fake the odo deploy image build / push passing in "echo" to PODMAN
			stdout := helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			By("building and pushing image to registry", func() {
				Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage"))
				Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
			})

			By("should display the component as 'Deploy' in odo list", func() {
				checkList("Dev, Deploy")
			})

			By("should display the component as 'Dev, Deploy' in odo list -o json", func() {
				res := helper.Cmd("odo", "list", "component", "-o", "json").ShouldPass()
				stdout, stderr := res.Out(), res.Err()
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentContain(stdout, "components.0.runningIn.dev", "true")
				helper.JsonPathContentContain(stdout, "components.0.runningIn.deploy", "true")
			})

		})

	})

	Context("devfile has missing metadata", Label(helper.LabelSerial), func() {
		// Note: We will be using SpringBoot example here because it helps to distinguish between language and projectType.
		// In terms of SpringBoot, spring is the projectType and java is the language; see https://github.com/redhat-developer/odo/issues/4815
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
		})
		var metadata devfilepkg.DevfileMetadata

		// checkList checks the list output (both normal and json) to see if it contains the expected componentType
		var checkList = func(componentType string) {
			By("checking the normal output", func() {
				stdOut := helper.Cmd("odo", "list", "component").ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring(componentType))
			})

			By("checking the JSON output", func() {
				res := helper.Cmd("odo", "list", "component", "-o", "json").ShouldPass()
				stdout, stderr := res.Out(), res.Err()
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentContain(stdout, "components.0.projectType", componentType)
			})
		}

		When("projectType is missing", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-with-missing-projectType-metadata.yaml")).ShouldPass()
				helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				metadata = helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			})

			It("should show the language for 'Type' in odo list", func() {
				checkList(metadata.Language)
			})
			When("the component is pushed in dev mode", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(nil)
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should show the language for 'Type' in odo list", func() {
					checkList(metadata.Language)
				})
			})
		})

		When("projectType and language is missing", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-with-missing-projectType-and-language-metadata.yaml")).ShouldPass()
				helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				metadata = helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			})
			It("should show 'Not available' for 'Type' in odo list", func() {
				checkList("Not available")
			})
			When("the component is pushed", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(nil)
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})
				It("should show 'nodejs' for 'Type' in odo list", func() {
					checkList("Not available")
				})
			})
		})
	})
})
