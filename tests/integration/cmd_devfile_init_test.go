package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/odo/cli/messages"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	segmentContext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile init command tests", func() {
	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		var _ = Context("label "+label, Label(label), func() {

			var commonVar helper.CommonVar

			var _ = BeforeEach(func() {
				commonVar = helper.CommonBeforeEach()
				helper.Chdir(commonVar.Context)
				Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
			})

			var _ = AfterEach(func() {
				helper.CommonAfterEach(commonVar)
			})

			It("should fail", func() {
				By("running odo init with incomplete flags", func() {
					helper.Cmd("odo", "init", "--name", "aname").ShouldFail()
				})

				By("using an invalid component name", func() {
					helper.Cmd("odo", "init", "--devfile", "go", "--name", "123").ShouldFail()
				})

				By("running odo init with json and no other flags", func() {
					res := helper.Cmd("odo", "init", "-o", "json").ShouldFail()
					stdout, stderr := res.Out(), res.Err()
					Expect(stdout).To(BeEmpty())
					Expect(helper.IsJSON(stderr)).To(BeTrue())
					helper.JsonPathContentIs(stderr, "message", "parameters are expected to select a devfile")
				})

				By("running odo init with incomplete flags and JSON output", func() {
					res := helper.Cmd("odo", "init", "--name", "aname", "-o", "json").ShouldFail()
					stdout, stderr := res.Out(), res.Err()
					Expect(stdout).To(BeEmpty())
					Expect(helper.IsJSON(stderr)).To(BeTrue())
					helper.JsonPathContentContain(stderr, "message", "either --devfile or --devfile-path parameter should be specified")
				})

				By("keeping an empty directory when running odo init with wrong starter name", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go", "--starter", "wrongname").ShouldFail()
					files := helper.ListFilesInDir(commonVar.Context)
					Expect(len(files)).To(Equal(0))
				})
				By("using an invalid devfile name", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "invalid").ShouldFail()
				})
				By("running odo init in a directory containing a devfile.yaml", func() {
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"))
					defer os.Remove(filepath.Join(commonVar.Context, "devfile.yaml"))
					err := helper.Cmd("odo", "init").ShouldFail().Err()
					Expect(err).To(ContainSubstring("a devfile already exists in the current directory"))
				})

				By("running odo init in a directory containing a .devfile.yaml", func() {
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"),
						filepath.Join(commonVar.Context, ".devfile.yaml"))
					defer helper.DeleteFile(filepath.Join(commonVar.Context, ".devfile.yaml"))
					err := helper.Cmd("odo", "init").ShouldFail().Err()
					Expect(err).To(ContainSubstring("a devfile already exists in the current directory"))
				})

				By("running odo init with wrong local file path given to --devfile-path", func() {
					err := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "/some/path/devfile.yaml").ShouldFail().Err()
					Expect(err).To(ContainSubstring("unable to download devfile"))
				})
				By("running odo init with wrong URL path given to --devfile-path", func() {
					err := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "https://github.com/path/to/devfile.yaml").ShouldFail().Err()
					Expect(err).To(ContainSubstring("unable to download devfile"))
				})
				By("running odo init multiple times", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs").ShouldPass()
					defer helper.DeleteFile(filepath.Join(commonVar.Context, "devfile.yaml"))
					output := helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs").ShouldFail().Err()
					Expect(output).To(ContainSubstring("a devfile already exists in the current directory"))
				})

				By("running odo init with --devfile-path and --devfile-registry", func() {
					errOut := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "https://github.com/path/to/devfile.yaml", "--devfile-registry", "DefaultDevfileRegistry").ShouldFail().Err()
					Expect(errOut).To(ContainSubstring("--devfile-registry parameter cannot be used with --devfile-path"))
				})
				By("running odo init with invalid --devfile-registry value", func() {
					fakeRegistry := "fake"
					errOut := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "https://github.com/path/to/devfile.yaml", "--devfile-registry", fakeRegistry).ShouldFail().Err()
					Expect(errOut).To(ContainSubstring(fmt.Sprintf("%q not found", fakeRegistry)))
				})
			})

			Context("running odo init with valid flags", func() {
				When("using --devfile flag", func() {
					compName := "aname"
					var output string
					BeforeEach(func() {
						output = helper.Cmd("odo", "init", "--name", compName, "--devfile", "go").ShouldPass().Out()
					})

					It("should download a devfile.yaml file and correctly set the component name in it", func() {
						By("not showing the interactive mode notice message", func() {
							Expect(output).ShouldNot(ContainSubstring(messages.InteractiveModeEnabled))
						})
						files := helper.ListFilesInDir(commonVar.Context)
						Expect(files).To(SatisfyAll(
							HaveLen(2),
							ContainElements("devfile.yaml", util.DotOdoDirectory)))
						metadata := helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
						Expect(metadata.Name).To(BeEquivalentTo(compName))
					})
				})
				When("using --devfile flag and JSON output", func() {
					compName := "aname"
					var res *helper.CmdWrapper
					BeforeEach(func() {
						res = helper.Cmd("odo", "init", "--name", compName, "--devfile", "go", "-o", "json").ShouldPass()
					})

					It("should return correct values in output", func() {
						stdout, stderr := res.Out(), res.Err()
						Expect(stderr).To(BeEmpty())
						Expect(helper.IsJSON(stdout)).To(BeTrue())
						helper.JsonPathContentIs(stdout, "devfilePath", filepath.Join(commonVar.Context, "devfile.yaml"))
						helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata.name", compName)
						helper.JsonPathContentIs(stdout, "devfileData.supportedOdoFeatures.dev", "true")
						helper.JsonPathContentIs(stdout, "devfileData.supportedOdoFeatures.debug", "false")
						helper.JsonPathContentIs(stdout, "devfileData.supportedOdoFeatures.deploy", "false")
						helper.JsonPathContentIs(stdout, "managedBy", "odo")
					})
				})

				for _, ctx := range []struct {
					title, devfileVersion, requiredVersion string
				}{
					{
						title:           "to download the latest version",
						devfileVersion:  "latest",
						requiredVersion: "2.0.0",
					},
					{
						title:           "to download a specific version",
						devfileVersion:  "1.0.2",
						requiredVersion: "1.0.2",
					},
				} {
					ctx := ctx
					const (
						devfileName = "go"
					)
					When(fmt.Sprintf("using --devfile-version flag %s", ctx.title), func() {
						BeforeEach(func() {
							helper.Cmd("odo", "init", "--name", "aname", "--devfile", devfileName, "--devfile-version", ctx.devfileVersion).ShouldPass()
						})

						It("should download the devfile with the requested version", func() {
							files := helper.ListFilesInDir(commonVar.Context)
							Expect(files).To(ContainElements("devfile.yaml"))
							metadata := helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
							Expect(metadata.Version).To(BeEquivalentTo(ctx.requiredVersion))
						})
					})

					When(fmt.Sprintf("using --devfile-version flag and JSON output %s", ctx.title), func() {
						var res *helper.CmdWrapper
						BeforeEach(func() {
							res = helper.Cmd("odo", "init", "--name", "aname", "--devfile", devfileName, "--devfile-version", ctx.devfileVersion, "-o", "json").ShouldPass()
						})

						It("should show the requested devfile version", func() {
							stdout := res.Out()
							Expect(helper.IsJSON(stdout)).To(BeTrue())
							helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata.version", ctx.requiredVersion)
						})
					})
				}

				When("using --devfile-path flag with a local devfile", func() {
					var newContext string
					BeforeEach(func() {
						newContext = helper.CreateNewContext()
						newDevfilePath := filepath.Join(newContext, "devfile.yaml")
						helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-registry.yaml"), newDevfilePath)
						helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", newDevfilePath).ShouldPass()
					})
					AfterEach(func() {
						helper.DeleteDir(newContext)
					})
					It("should copy the devfile.yaml file", func() {
						files := helper.ListFilesInDir(commonVar.Context)
						Expect(files).To(SatisfyAll(
							HaveLen(2),
							ContainElements(util.DotOdoDirectory, "devfile.yaml")))
					})
				})
				When("using --devfile-path flag with a URL", func() {
					BeforeEach(func() {
						helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", "https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml").ShouldPass()
					})
					It("should copy the devfile.yaml file", func() {
						files := helper.ListFilesInDir(commonVar.Context)
						Expect(files).To(SatisfyAll(
							HaveLen(2),
							ContainElements("devfile.yaml", util.DotOdoDirectory)))
					})
				})
				When("using --devfile-registry flag", func() {
					It("should successfully run odo init if specified registry is valid", func() {
						helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go", "--devfile-registry", "DefaultDevfileRegistry").ShouldPass()
					})

				})
			})
			When("a dangling env file exists in the working directory", func() {
				BeforeEach(func() {
					helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				})
				It("should successfully create a devfile component and remove the dangling env file", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go").ShouldPass()
				})
			})

			When("a devfile is provided which has a starter that has its own devfile", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "init", "--name", "aname", "--starter", "nodejs-starter", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-starter-with-devfile.yaml")).ShouldPass()
				})
				It("should pass and keep the devfile in starter", func() {
					devfileContent, err := helper.ReadFile(filepath.Join(commonVar.Context, "devfile.yaml"))
					Expect(err).To(Not(HaveOccurred()))
					helper.MatchAllInOutput(devfileContent, []string{"2.2.0", "kubernetes-deploy", "deployk8s", "image-build"})
				})
			})

			When("running odo init with a devfile that has a subDir starter project", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-with-subDir.yaml"), "--starter", "springbootproject").ShouldPass()
				})

				It("should successfully extract the project in the specified subDir path", func() {
					var found, notToBeFound int
					pathsToValidate := map[string]bool{
						filepath.Join(commonVar.Context, "java", "com"):                                            true,
						filepath.Join(commonVar.Context, "java", "com", "example"):                                 true,
						filepath.Join(commonVar.Context, "java", "com", "example", "demo"):                         true,
						filepath.Join(commonVar.Context, "java", "com", "example", "demo", "DemoApplication.java"): true,
						filepath.Join(commonVar.Context, "resources", "application.properties"):                    true,
					}
					pathsNotToBePresent := map[string]bool{
						filepath.Join(commonVar.Context, "src"):  true,
						filepath.Join(commonVar.Context, "main"): true,
					}
					err := filepath.Walk(commonVar.Context, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if ok := pathsToValidate[path]; ok {
							found++
						}
						if ok := pathsNotToBePresent[path]; ok {
							notToBeFound++
						}
						return nil
					})
					Expect(err).To(BeNil())

					Expect(found).To(Equal(len(pathsToValidate)))
					Expect(notToBeFound).To(Equal(0))
				})
			})

			It("should successfully run odo init for devfile with starter project from the specified branch", func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-branch.yaml"), "--starter", "nodejs-starter").ShouldPass()
				expectedFiles := []string{"package.json", "package-lock.json", "README.md", "devfile.yaml", "test"}
				Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements(expectedFiles))
			})

			It("should successfully run odo init for devfile with starter project from the specified tag", func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-tag.yaml"), "--starter", "nodejs-starter").ShouldPass()
				expectedFiles := []string{"package.json", "package-lock.json", "README.md", "devfile.yaml", "app"}
				Expect(helper.ListFilesInDir(commonVar.Context)).To(ContainElements(expectedFiles))
			})

			It("should successfully run odo init for devfile with starter project on git with main default branch", func() {
				helper.Cmd("odo", "init",
					"--name", "vertx",
					"--devfile-path", helper.GetExamplePath("source", "devfiles", "java", "devfile-with-git-main-branch.yaml"),
					"--starter", "vertx-http-example-redhat",
				).ShouldPass()
			})

			When("running odo init from a directory with sources", func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				})
				It("should work without --starter flag", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs").ShouldPass()
				})
				It("should not accept --starter flag", func() {
					err := helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs", "--starter", "nodejs-starter").ShouldFail().Err()
					Expect(err).To(ContainSubstring("--starter parameter cannot be used when the directory is not empty"))
				})
			})
			Context("checking odo init final output message", func() {
				var newContext, devfilePath string

				BeforeEach(func() {
					newContext = helper.CreateNewContext()
					devfilePath = filepath.Join(newContext, "devfile.yaml")
				})

				AfterEach(func() {
					helper.DeleteDir(newContext)
				})

				When("the devfile used by `odo init` does not contain a deploy command", func() {
					var out string

					BeforeEach(func() {
						helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), devfilePath)
						out = helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", devfilePath).ShouldPass().Out()
					})

					It("should only show information about `odo dev`, and not `odo deploy`", func() {
						Expect(out).To(ContainSubstring("odo dev"))
						Expect(out).ToNot(ContainSubstring("odo deploy"))
					})

					It("should not show the interactive mode notice message", func() {
						Expect(out).ShouldNot(ContainSubstring(messages.InteractiveModeEnabled))
					})
				})

				When("the devfile used by `odo init` contains a deploy command", func() {
					var out string

					BeforeEach(func() {
						helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), devfilePath)
						out = helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", devfilePath).ShouldPass().Out()
					})

					It("should show information about both `odo dev`, and `odo deploy`", func() {
						Expect(out).To(ContainSubstring("odo dev"))
						Expect(out).To(ContainSubstring("odo deploy"))
					})

					It("should not show the interactive mode notice message", func() {
						Expect(out).ShouldNot(ContainSubstring(messages.InteractiveModeEnabled))
					})
				})
			})

			When("devfile contains parent URI", func() {
				var originalKeyList []string
				var srcDevfile string

				BeforeEach(func() {
					var err error
					srcDevfile = helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-parent.yaml")
					originalDevfileContent, err := ioutil.ReadFile(srcDevfile)
					Expect(err).To(BeNil())
					var content map[string]interface{}
					Expect(yaml.Unmarshal(originalDevfileContent, &content)).To(BeNil())
					for k := range content {
						originalKeyList = append(originalKeyList, k)
					}
				})

				It("should not replace the original devfile", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", srcDevfile).ShouldPass()
					devfileContent, err := ioutil.ReadFile(filepath.Join(commonVar.Context, "devfile.yaml"))
					Expect(err).To(BeNil())
					var content map[string]interface{}
					Expect(yaml.Unmarshal(devfileContent, &content)).To(BeNil())
					for k := range content {
						Expect(k).To(BeElementOf(originalKeyList))
					}
				})
			})

			When("source directory is empty", func() {
				BeforeEach(func() {
					Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
				})

				It("name in devfile is personalized in non-interactive mode", func() {
					helper.Cmd("odo", "init", "--name", "aname", "--devfile-path",
						filepath.Join(helper.GetExamplePath(), "source", "devfiles", "nodejs",
							"devfile-with-starter-with-devfile.yaml")).ShouldPass()

					metadata := helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
					Expect(metadata.Name).To(BeEquivalentTo("aname"))
					Expect(metadata.Language).To(BeEquivalentTo("nodejs"))
				})
			})

			Describe("telemetry", func() {

				for _, tt := range []struct {
					name string
					env  map[string]string
				}{
					{
						name: "ODO_DISABLE_TELEMETRY=true and ODO_TRACKING_CONSENT=yes",
						env: map[string]string{
							//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still want to test it
							segment.DisableTelemetryEnv: "true",
							segment.TrackingConsentEnv:  "yes",
						},
					},
					{
						name: "ODO_DISABLE_TELEMETRY=false and ODO_TRACKING_CONSENT=no",
						env: map[string]string{
							//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still want to test it
							segment.DisableTelemetryEnv: "false",
							segment.TrackingConsentEnv:  "no",
						},
					},
				} {
					tt := tt
					It("should error out if "+tt.name, func() {
						cmd := helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go")
						for k, v := range tt.env {
							cmd = cmd.AddEnv(fmt.Sprintf("%s=%s", k, v))
						}
						stderr := cmd.ShouldFail().Err()

						//lint:ignore SA1019 We deprecated this env var, but until it is removed, we still want to test it
						Expect(stderr).To(ContainSubstring("%s and %s values are in conflict.", segment.DisableTelemetryEnv, segment.TrackingConsentEnv))
					})
				}

				type telemetryTest struct {
					title         string
					env           map[string]string
					setupFunc     func(cfg preference.Client)
					callerChecker func(stdout, stderr string, data segment.TelemetryData)
				}
				allowedTelemetryCallers := []string{segmentContext.VSCode, segmentContext.IntelliJ, segmentContext.JBoss}
				telemetryTests := []telemetryTest{
					{
						title: "no caller env var",
						callerChecker: func(_, _ string, td segment.TelemetryData) {
							cmdProperties := td.Properties.CmdProperties
							Expect(cmdProperties).Should(HaveKey(segmentContext.Caller))
							Expect(cmdProperties[segmentContext.Caller]).To(BeEmpty())
						},
					},
					{
						title: "empty caller env var",
						env: map[string]string{
							helper.TelemetryCaller: "",
						},
						callerChecker: func(_, _ string, td segment.TelemetryData) {
							cmdProperties := td.Properties.CmdProperties
							Expect(cmdProperties).Should(HaveKey(segmentContext.Caller))
							Expect(cmdProperties[segmentContext.Caller]).To(BeEmpty())
						},
					},
					{
						title: "invalid caller env var",
						env: map[string]string{
							helper.TelemetryCaller: "an-invalid-caller",
						},
						callerChecker: func(stdout, stderr string, td segment.TelemetryData) {
							By("not disclosing list of allowed values", func() {
								helper.DontMatchAllInOutput(stdout, allowedTelemetryCallers)
								helper.DontMatchAllInOutput(stderr, allowedTelemetryCallers)
							})

							By("setting the value as caller property in telemetry even if it is invalid", func() {
								Expect(td.Properties.CmdProperties[segmentContext.Caller]).To(Equal("an-invalid-caller"))
							})
						},
					},
					{
						title: "ODO_TRACKING_CONSENT=yes env var should take precedence over ConsentTelemetry preference",
						env:   map[string]string{segment.TrackingConsentEnv: "yes"},
						callerChecker: func(_, _ string, td segment.TelemetryData) {
							cmdProperties := td.Properties.CmdProperties
							Expect(cmdProperties).Should(HaveKey(segmentContext.Caller))
							Expect(cmdProperties[segmentContext.Caller]).To(BeEmpty())
						},
						setupFunc: func(cfg preference.Client) {
							err := cfg.SetConfiguration(preference.ConsentTelemetrySetting, "false")
							Expect(err).ShouldNot(HaveOccurred())
						},
					},
				}
				for _, c := range allowedTelemetryCallers {
					c := c
					telemetryTests = append(telemetryTests, telemetryTest{
						title: fmt.Sprintf("valid caller env var: %s", c),
						env: map[string]string{
							helper.TelemetryCaller: c,
						},
						callerChecker: func(_, _ string, td segment.TelemetryData) {
							Expect(td.Properties.CmdProperties[segmentContext.Caller]).To(Equal(c))
						},
					})
				}
				for _, tt := range telemetryTests {
					tt := tt
					When("recording telemetry data with "+tt.title, func() {
						var stdout string
						var stderr string
						BeforeEach(func() {
							helper.EnableTelemetryDebug()

							ctx := context.Background()
							envConfig, err := config.GetConfiguration()
							Expect(err).To(BeNil())
							ctx = envcontext.WithEnvConfig(ctx, *envConfig)

							cfg, err := preference.NewClient(ctx)
							Expect(err).ShouldNot(HaveOccurred())
							if tt.setupFunc != nil {
								tt.setupFunc(cfg)
							}

							cmd := helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go")
							for k, v := range tt.env {
								cmd = cmd.AddEnv(fmt.Sprintf("%s=%s", k, v))
							}
							stdout, stderr = cmd.ShouldPass().OutAndErr()
						})

						AfterEach(func() {
							helper.ResetTelemetry()
						})

						It("should record the telemetry data correctly", func() {
							td := helper.GetTelemetryDebugData()
							Expect(td.Event).To(ContainSubstring("odo init"))
							Expect(td.Properties.Success).To(BeTrue())
							Expect(td.Properties.Error == "").To(BeTrue())
							Expect(td.Properties.ErrorType == "").To(BeTrue())
							Expect(td.Properties.CmdProperties[segmentContext.DevfileName]).To(ContainSubstring("aname"))
							Expect(td.Properties.CmdProperties[segmentContext.ComponentType]).To(ContainSubstring("Go"))
							Expect(td.Properties.CmdProperties[segmentContext.Language]).To(ContainSubstring("Go"))
							Expect(td.Properties.CmdProperties[segmentContext.ProjectType]).To(ContainSubstring("Go"))
							Expect(td.Properties.CmdProperties[segmentContext.Flags]).To(ContainSubstring("devfile name"))
							tt.callerChecker(stdout, stderr, td)
						})

					})
				}
			})
		})
	}
})
