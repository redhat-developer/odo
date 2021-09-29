package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo service command tests for OperatorHub", func() {

	var commonVar helper.CommonVar

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Operators are installed in the cluster", func() {

		BeforeEach(func() {
			// wait till odo can see that all operators installed by setup script in the namespace
			odoArgs := []string{"catalog", "list", "services"}
			operators := []string{"redis-operator"}
			if os.Getenv("KUBERNETES") != "true" {
				operators = append(operators, "service-binding-operator")
			}
			for _, operator := range operators {
				helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
					return strings.Contains(output, operator)
				})
			}
		})

		It("should not allow creating service without valid context", func() {
			stdOut := helper.Cmd("odo", "service", "create").ShouldFail().Err()
			Expect(stdOut).To(ContainSubstring("service can be created/deleted from a valid component directory only"))
		})

		Context("a namespace specific operator is installed", func() {

			var postgresOperator string
			var postgresDatabase string
			var projectName string

			BeforeEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					Skip("This is a OpenShift specific scenario, skipping")
				}
				projectName = util.GetEnvWithDefault("REDHAT_POSTGRES_OPERATOR_PROJECT", "odo-operator-test")
				helper.GetCliRunner().SetProject(projectName)
				operators := helper.Cmd("odo", "catalog", "list", "services").ShouldPass().Out()
				postgresOperator = regexp.MustCompile(`postgresql-operator\.*[a-z][0-9]\.[0-9]\.[0-9]`).FindString(operators)
				postgresDatabase = fmt.Sprintf("%s/Database", postgresOperator)
			})

			When("a nodejs component is created", func() {

				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					// change the app name to avoid conflicts
					appName := helper.RandString(5)
					helper.Cmd("odo", "create", "nodejs", "--app", appName, "--context", commonVar.Context).ShouldPass().Out()
					helper.Cmd("odo", "config", "set", "Memory", "300M", "-f", "--context", commonVar.Context).ShouldPass()
				})

				AfterEach(func() {
					// we do this because for these specific tests we dont delete the project
					helper.Cmd("odo", "delete", "--all", "-f", "--context", commonVar.Context).ShouldPass().Out()
				})

				It("should try to create a service in dry run mode with some provided params", func() {
					serviceName := helper.RandString(10)
					output := helper.Cmd("odo", "service", "create", postgresDatabase, serviceName, "-p",
						"databaseName=odo", "-p", "size=1", "-p", "databaseUser=odo", "-p",
						"databaseStorageRequest=1Gi", "-p", "databasePassword=odopasswd", "--dry-run", "--context", commonVar.Context).ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{fmt.Sprintf("name: %s", serviceName), "odo", "odopasswd", "1Gi"})
				})

				When("creating a postgres operand with params", func() {
					var operandName string

					BeforeEach(func() {
						operandName = helper.RandString(10)
						helper.Cmd("odo", "service", "create", postgresDatabase, operandName, "-p",
							"databaseName=odo", "-p", "size=1", "-p", "databaseUser=odo", "-p",
							"databaseStorageRequest=1Gi", "-p", "databasePassword=odopasswd", "--context", commonVar.Context).ShouldPass().Out()

					})

					AfterEach(func() {
						helper.Cmd("odo", "service", "delete", fmt.Sprintf("Database/%s", operandName), "-f", "--context", commonVar.Context).ShouldPass().Out()
						helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
					})

					When("odo push is executed", func() {
						BeforeEach(func() {
							helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
						})

						It("should create pods in running state", func() {
							commonVar.CliRunner.PodsShouldBeRunning(projectName, fmt.Sprintf(`%s-.[\-a-z0-9]*`, operandName))
						})

						It("should list the service", func() {
							// now test listing of the service using odo
							stdOut := helper.Cmd("odo", "service", "list", "--context", commonVar.Context).ShouldPass().Out()
							Expect(stdOut).To(ContainSubstring(fmt.Sprintf("Database/%s", operandName)))
						})
					})

				})

			})
		})

		Context("a specific operator is installed", func() {
			var redisOperator string
			var redisCluster string

			BeforeEach(func() {
				commonVar.CliRunner.CreateSecret("redis-secret", "password", commonVar.Project)
				operators := helper.Cmd("odo", "catalog", "list", "services").ShouldPass().Out()
				redisOperator = regexp.MustCompile(`redis-operator\.*[a-z][0-9]\.[0-9]\.[0-9]`).FindString(operators)
				redisCluster = fmt.Sprintf("%s/Redis", redisOperator)
			})

			It("should describe the operator with human-readable output", func() {
				output := helper.Cmd("odo", "catalog", "describe", "service", redisCluster).ShouldPass().Out()
				Expect(output).To(MatchRegexp("KIND: *Redis"))
				Expect(output).To(MatchRegexp(`redisExporter\.image *\(string\) *-required-`))
			})

			It("should describe the example of the operator", func() {
				output := helper.Cmd("odo", "catalog", "describe", "service", redisCluster, "--example").ShouldPass().Out()
				Expect(output).To(ContainSubstring("kind: Redis"))
				helper.MatchAllInOutput(output, []string{"apiVersion", "kind"})
			})

			It("should describe the example of the operator as json", func() {
				outputJSON := helper.Cmd("odo", "catalog", "describe", "service", redisCluster, "--example", "-o", "json").ShouldPass().Out()
				value := gjson.Get(outputJSON, "spec.kind")
				Expect(value.String()).To(Equal("Redis"))
			})

			It("should describe the operator with json output", func() {
				outputJSON := helper.Cmd("odo", "catalog", "describe", "service", redisCluster, "-o", "json").ShouldPass().Out()
				values := gjson.GetMany(outputJSON, "spec.kind", "spec.displayName", "spec.schema.type", "spec.schema.properties.redisExporter.properties.image.type")
				expected := []string{"Redis", "Redis", "object", "string"}
				Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
			})

			It("should find the services by keyword", func() {
				stdOut := helper.Cmd("odo", "catalog", "search", "service", "redis").ShouldPass().Out()
				helper.MatchAllInOutput(stdOut, []string{"redis-operator", "Redis"})

				stdOut = helper.Cmd("odo", "catalog", "search", "service", "Redis").ShouldPass().Out()
				helper.MatchAllInOutput(stdOut, []string{"redis-operator", "Redis"})

				stdOut = helper.Cmd("odo", "catalog", "search", "service", "dummy").ShouldFail().Err()
				Expect(stdOut).To(ContainSubstring("no service matched the query: dummy"))
			})

			It("should list the operator in JSON output", func() {
				jsonOut := helper.Cmd("odo", "catalog", "list", "services", "-o", "json").ShouldPass().Out()
				helper.MatchAllInOutput(jsonOut, []string{"redis-operator"})
			})

			When("a nodejs component is created", func() {

				var cmpName string
				BeforeEach(func() {
					cmpName = helper.RandString(4)
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.Cmd("odo", "create", "nodejs", cmpName).ShouldPass()
					helper.Cmd("odo", "config", "set", "Memory", "300M", "-f").ShouldPass()
				})

				It("should fail for interactive mode", func() {
					stdOut := helper.Cmd("odo", "service", "create").ShouldFail().Err()
					Expect(stdOut).To(ContainSubstring("odo doesn't support interactive mode for creating Operator backed service"))
				})

				It("should define the CR output of the operator instance in dryRun mode", func() {
					stdOut := helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), "--dry-run", "--project", commonVar.Project).ShouldPass().Out()
					helper.MatchAllInOutput(stdOut, []string{"apiVersion", "kind"})
				})

				It("should fail creating a service with wrong parameters", func() {
					helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), "-p", "a=b").ShouldFail()
				})

				It("should create a service with parameters of different types", func() {
					helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator),
						"-p", "kubernetesConfig.image=quay.io/opstree/redis:v6.2.5",
						"-p", "kubernetesConfig.imagePullPolicy=IfNotPresent",
						"-p", "kubernetesConfig.serviceType=ClusterIP",
						"-p", "redisExporter.enabled=false",
						"-p", "redisExporter.image=quay.io/opstree/redis-exporter:1.0",
						"-p", "securityContext.runAsUser=1000").ShouldPass()
				})

				When("odo push is executed", func() {
					BeforeEach(func() {
						helper.Cmd("odo", "push").ShouldPass()
					})

					It("should fail if the provided service doesn't exist in the namespace", func() {
						stdOut := helper.Cmd("odo", "link", "Redis/redis-standalone").ShouldFail().Err()
						Expect(stdOut).To(ContainSubstring("couldn't find service named %q", "Redis/redis-standalone"))
					})
				})

				When("a Redis instance definition copied from example file", func() {

					var fileName string

					BeforeEach(func() {
						randomFileName := helper.RandString(6) + ".yaml"
						fileName = filepath.Join(os.TempDir(), randomFileName)
						helper.CopyExampleFile(filepath.Join("operators", "redis.yaml"), filepath.Join(fileName))
					})

					AfterEach(func() {
						os.Remove(fileName)
					})

					When("a service is created from the output of the dryRun command with no name and odo push is executed", func() {
						BeforeEach(func() {
							helper.Cmd("odo", "service", "create", "--from-file", fileName, "--project", commonVar.Project).ShouldPass()
							helper.Cmd("odo", "push").ShouldPass()
						})

						AfterEach(func() {
							helper.Cmd("odo", "service", "delete", "Redis/redis-standalone", "-f").ShouldPass()
							helper.Cmd("odo", "push").ShouldPass()
						})

						It("should create pods in running state", func() {
							commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, `redis.[a-z0-9-]*`)
						})
					})

					When("a service is created from the output of the dryRun command with a specific name and odo push is executed", func() {

						var name string
						var svcFullName string
						BeforeEach(func() {
							name = helper.RandString(6)
							svcFullName = strings.Join([]string{"Redis", name}, "/")
							helper.Cmd("odo", "service", "create", "--from-file", fileName, name, "--project", commonVar.Project).ShouldPass()
							helper.Cmd("odo", "push").ShouldPass()
						})

						AfterEach(func() {
							helper.Cmd("odo", "service", "delete", svcFullName, "-f").ShouldPass()
							helper.Cmd("odo", "push").ShouldPass()
						})

						It("should fail to create a service again with the same name", func() {
							stdOut := helper.Cmd("odo", "service", "create", "--from-file", fileName, name, "--project", commonVar.Project).ShouldFail().Err()
							Expect(stdOut).To(ContainSubstring("please provide a different name or delete the existing service first"))
						})

						It("should create pods in running state", func() {
							commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, name+`-.[a-z0-9-]*`)
						})
					})
				})

				When("a Redis instance is created with no name and inlined flag is used", func() {
					var stdOut string
					BeforeEach(func() {
						stdOut = helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), "--project", commonVar.Project, "--inlined").ShouldPass().Out()
						Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))
					})

					It("should insert service definition in devfile.yaml when the inlined flag is used", func() {
						devfilePath := filepath.Join(commonVar.Context, "devfile.yaml")
						content, err := ioutil.ReadFile(devfilePath)
						Expect(err).To(BeNil())
						matchInOutput := []string{"kubernetes", "inlined", "Redis", "redis"}
						helper.MatchAllInOutput(string(content), matchInOutput)
					})

					It("should list the service in JSON format", func() {
						jsonOut := helper.Cmd("odo", "service", "list", "-o", "json").ShouldPass().Out()
						helper.MatchAllInOutput(jsonOut, []string{"\"apiVersion\": \"redis.redis.opstreelabs.in/v1beta1\"", "\"kind\": \"Redis\"", "\"name\": \"redis\""})
					})

					When("odo push is executed", func() {

						BeforeEach(func() {
							helper.Cmd("odo", "push").ShouldPass()
						})

						It("should create pods in running state", func() {
							commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, `redis.[a-z0-9-]*`)
						})

						It("should list the service", func() {
							// now test listing of the service using odo
							stdOut := helper.Cmd("odo", "service", "list").ShouldPass().Out()
							Expect(stdOut).To(ContainSubstring("Redis/redis"))
						})

						It("should list the service in JSON format", func() {
							jsonOut := helper.Cmd("odo", "service", "list", "-o", "json").ShouldPass().Out()
							helper.MatchAllInOutput(jsonOut, []string{"\"apiVersion\": \"redis.redis.opstreelabs.in/v1beta1\"", "\"kind\": \"Redis\"", "\"name\": \"redis\""})
						})

						When("a link is created with the service", func() {
							var stdOut string
							BeforeEach(func() {
								stdOut = helper.Cmd("odo", "link", "Redis/redis").ShouldPass().Out()
							})

							It("should display a successful message", func() {
								Expect(stdOut).To(ContainSubstring("Successfully created link between component"))
							})

							It("Should fail to link it again", func() {
								stdOut = helper.Cmd("odo", "link", "Redis/redis").ShouldFail().Err()
								Expect(stdOut).To(ContainSubstring("already linked with the service"))
							})

							When("the link is deleted", func() {
								BeforeEach(func() {
									stdOut = helper.Cmd("odo", "unlink", "Redis/redis").ShouldPass().Out()
								})

								It("should display a successful message", func() {
									Expect(stdOut).To(ContainSubstring("Successfully unlinked component"))
								})

								It("should fail to delete it again", func() {
									stdOut = helper.Cmd("odo", "unlink", "Redis/redis").ShouldFail().Err()
									Expect(stdOut).To(ContainSubstring("failed to unlink the service"))
								})
							})
						})

						When("the service is deleted", func() {
							BeforeEach(func() {
								helper.Cmd("odo", "service", "delete", "Redis/redis", "-f").ShouldPass()
							})

							It("should delete service definition from devfile.yaml", func() {
								// read the devfile.yaml to check if service definition was deleted
								devfilePath := filepath.Join(commonVar.Context, "devfile.yaml")
								content, err := ioutil.ReadFile(devfilePath)
								Expect(err).To(BeNil())
								matchInOutput := []string{"kubernetes", "inlined", "Redis", "redis"}
								helper.DontMatchAllInOutput(string(content), matchInOutput)
							})

							It("should fail to delete the service again", func() {
								stdOut = helper.Cmd("odo", "service", "delete", "Redis/redis", "-f").ShouldFail().Err()
								Expect(stdOut).To(ContainSubstring("couldn't find service named"))
							})

							When("odo push is executed", func() {
								BeforeEach(func() {
									helper.Cmd("odo", "push").ShouldPass()
								})

								It("Should fail listing the services", func() {
									out := helper.Cmd("odo", "service", "list").ShouldFail().Err()
									msg := fmt.Sprintf("no operator backed services found in namespace: %s", commonVar.Project)
									Expect(out).To(ContainSubstring(msg))
								})

								It("Should fail listing the services in JSON format", func() {
									jsonOut := helper.Cmd("odo", "service", "list", "-o", "json").ShouldFail().Err()
									msg := fmt.Sprintf("no operator backed services found in namespace: %s", commonVar.Project)
									msgWithQuote := fmt.Sprintf("\"message\": \"no operator backed services found in namespace: %s\"", commonVar.Project)
									helper.MatchAllInOutput(jsonOut, []string{msg, msgWithQuote})
								})
							})
						})

						When("a second service is created and odo push is executed", func() {
							BeforeEach(func() {
								stdOut = helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), "myredis2", "--project", commonVar.Project).ShouldPass().Out()
								Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))
								helper.Cmd("odo", "push").ShouldPass()
							})

							It("should list both services", func() {
								stdOut = helper.Cmd("odo", "service", "list").ShouldPass().Out()
								// first service still here
								Expect(stdOut).To(ContainSubstring("Redis/redis"))
								// second service created
								Expect(stdOut).To(ContainSubstring("Redis/myredis2"))
							})
						})
					})
				})

				When("a Redis instance is created with a specific name", func() {

					var name string
					var svcFullName string

					BeforeEach(func() {
						name = helper.RandString(6)
						svcFullName = strings.Join([]string{"Redis", name}, "/")
						helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), name, "--project", commonVar.Project).ShouldPass()
					})

					AfterEach(func() {
						helper.Cmd("odo", "service", "delete", svcFullName, "-f").ShouldRun()
					})

					It("should not insert service definition in devfile.yaml when the inlined flag is not used", func() {
						devfilePath := filepath.Join(commonVar.Context, "devfile.yaml")
						content, err := ioutil.ReadFile(devfilePath)
						Expect(err).To(BeNil())
						matchInOutput := []string{"redis", "Redis", "inlined"}
						helper.DontMatchAllInOutput(string(content), matchInOutput)
					})

					It("should be listed as Not pushed", func() {
						stdOut := helper.Cmd("odo", "service", "list").ShouldPass().Out()
						helper.MatchAllInOutput(stdOut, []string{svcFullName, "Not pushed"})
					})

					When("odo push is executed", func() {

						BeforeEach(func() {
							helper.Cmd("odo", "push").ShouldPass()
						})

						It("should create pods in running state", func() {
							commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, name+`-.[a-z0-9-]*`)
						})

						It("should fail to create a service again with the same name", func() {
							stdOut := helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), name, "--project", commonVar.Project).ShouldFail().Err()
							Expect(stdOut).To(ContainSubstring(fmt.Sprintf("service %q already exists", svcFullName)))
						})

						It("should be listed as Pushed", func() {
							stdOut := helper.Cmd("odo", "service", "list").ShouldPass().Out()
							helper.MatchAllInOutput(stdOut, []string{svcFullName, "Pushed"})
						})

						When("the redisCluster instance is deleted", func() {
							BeforeEach(func() {
								helper.Cmd("odo", "service", "delete", svcFullName, "-f").ShouldPass()
							})

							It("should be listed as Deleted locally", func() {
								stdOut := helper.Cmd("odo", "service", "list").ShouldPass().Out()
								helper.MatchAllInOutput(stdOut, []string{svcFullName, "Deleted locally"})
							})

							When("odo push is executed", func() {
								BeforeEach(func() {
									helper.Cmd("odo", "push").ShouldPass()
								})

								It("should not be listed anymore", func() {
									stdOut := helper.Cmd("odo", "service", "list").ShouldRun().Out()
									Expect(strings.Contains(stdOut, svcFullName)).To(BeFalse())
								})
							})
						})

						When("a link is created with a specific name", func() {

							var linkName string

							BeforeEach(func() {
								linkName = "link-" + helper.RandString(6)
								helper.Cmd("odo", "link", "Redis/"+name, "--name", linkName).ShouldPass()
								helper.Cmd("odo", "push").ShouldPass()
							})

							AfterEach(func() {
								// delete the link
								helper.Cmd("odo", "unlink", "Redis/"+name).ShouldPass()
								helper.Cmd("odo", "push").ShouldPass()
							})

							It("should create the link with the specified name", func() {
								envFromValues := commonVar.CliRunner.GetEnvRefNames(cmpName, "app", commonVar.Project)
								envFound := false
								for i := range envFromValues {
									if strings.Contains(envFromValues[i], linkName) {
										envFound = true
									}
								}
								Expect(envFound).To(BeTrue())
							})
						})

						When("a link is created with a specific name and bind-as-files flag", func() {

							var linkName string

							BeforeEach(func() {
								linkName = "link-" + helper.RandString(6)
								helper.Cmd("odo", "link", "Redis/"+name, "--name", linkName, "--bind-as-files").ShouldPass()
								helper.Cmd("odo", "push").ShouldPass()
							})

							AfterEach(func() {
								// delete the link
								helper.Cmd("odo", "unlink", "Redis/"+name).ShouldPass()
								helper.Cmd("odo", "push").ShouldPass()
							})

							It("should create a link with bindAsFiles set to true", func() {
								// check the volume name and mount paths for the container
								deploymentName, err := util.NamespaceKubernetesObject(cmpName, "app")
								if err != nil {
									Expect(err).To(BeNil())
								}
								volNamesAndPaths := commonVar.CliRunner.GetVolumeMountNamesandPathsFromContainer(deploymentName, "runtime", commonVar.Project)
								mountFound := false
								volNamesAndPathsArr := strings.Fields(volNamesAndPaths)
								for _, volNamesAndPath := range volNamesAndPathsArr {
									volNamesAndPathArr := strings.Split(volNamesAndPath, ":")
									if strings.Contains(volNamesAndPathArr[0], linkName) {
										mountFound = true
									}
								}
								Expect(mountFound).To(BeTrue())
							})
						})
					})
				})

				When("a Redis instance is created with a specific name and json output", func() {

					var name string
					var svcFullName string
					var output string

					testServiceInfo := func(serviceName string, text string) {
						values := gjson.GetMany(text, "kind", "metadata.name", "manifest.kind", "manifest.metadata.name")
						expected := []string{"Service", "Redis/" + serviceName, "Redis", serviceName}
						Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
					}

					testClusterInfo := func(serviceName string, text string, inDevfile bool, deployed bool) {

						values := gjson.GetMany(text, "inDevfile", "deployed")
						expected := []string{fmt.Sprintf("%v", inDevfile), fmt.Sprintf("%v", deployed)}
						Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))

						tsValue := gjson.Get(text, "manifest.metadata.creationTimestamp")
						if deployed {
							Expect(tsValue.Str).NotTo(BeEmpty())
						} else {
							Expect(tsValue.Str).To(BeEmpty())
						}
					}

					BeforeEach(func() {
						name = helper.RandString(6)
						svcFullName = strings.Join([]string{"Redis", name}, "/")
						output = helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), name, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
					})

					AfterEach(func() {
						helper.Cmd("odo", "service", "delete", svcFullName, "-f").ShouldRun()
					})

					It("should display valid information in output of create command", func() {
						By("displaying service information", func() {
							testServiceInfo(name, output)
						})

						By("not containing cluster specific information", func() {
							testClusterInfo(name, output, true, false)
						})
					})

					When("executing odo service describe", func() {
						var descOutput string
						BeforeEach(func() {
							descOutput = helper.Cmd("odo", "service", "describe", "Redis/"+name, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
						})

						It("should display valid information in output of create command", func() {
							By("displaying service information", func() {
								testServiceInfo(name, descOutput)
							})

							By("not containing cluster specific information", func() {
								testClusterInfo(name, descOutput, true, false)
							})
						})
					})

					When("odo push is executed", func() {

						BeforeEach(func() {
							helper.Cmd("odo", "push").ShouldPass()
						})

						When("executing odo service describe", func() {
							var descOutput string
							BeforeEach(func() {
								descOutput = helper.Cmd("odo", "service", "describe", "Redis/"+name, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
							})

							It("should display valid information in output of create command", func() {
								By("displaying service information", func() {
									testServiceInfo(name, descOutput)
								})

								By("containing cluster specific information", func() {
									testClusterInfo(name, descOutput, true, true)
								})
							})
						})

						When("service is deleted from devfile", func() {
							BeforeEach(func() {
								helper.Cmd("odo", "service", "delete", "Redis/"+name, "--project", commonVar.Project, "-f").ShouldPass()
							})

							When("executing odo service describe", func() {
								var descOutput string
								BeforeEach(func() {
									descOutput = helper.Cmd("odo", "service", "describe", "Redis/"+name, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
								})

								It("should display valid information in output of create command", func() {
									By("displaying service information", func() {
										testServiceInfo(name, descOutput)
									})

									By("containing cluster specific information", func() {
										testClusterInfo(name, descOutput, false, true)
									})
								})
							})

							When("odo push is executed", func() {
								BeforeEach(func() {
									helper.Cmd("odo", "push").ShouldPass()
								})

								It("should not describe the service anymore", func() {
									helper.Cmd("odo", "service", "describe", "Redis/"+name, "--project", commonVar.Project, "-o", "json").ShouldFail()
								})
							})
						})
					})
				})

				Context("Invalid service templates exist", func() {

					var tmpContext string
					var noMetaFileName string
					var invalidFileName string

					BeforeEach(func() {
						tmpContext = helper.CreateNewContext()

						// TODO write helpers to create such files
						noMetadata := `
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
spec:`
						noMetaFile := helper.RandString(6) + ".yaml"
						noMetaFileName = filepath.Join(tmpContext, noMetaFile)
						if err := ioutil.WriteFile(noMetaFileName, []byte(noMetadata), 0644); err != nil {
							fmt.Printf("Could not write yaml spec to file %s because of the error %v", noMetaFileName, err.Error())
						}

						invalidMetadata := `
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
  noname: noname
spec:`
						invalidMetaFile := helper.RandString(6) + ".yaml"
						invalidFileName = filepath.Join(tmpContext, invalidMetaFile)
						if err := ioutil.WriteFile(invalidFileName, []byte(invalidMetadata), 0644); err != nil {
							fmt.Printf("Could not write yaml spec to file %s because of the error %v", invalidFileName, err.Error())
						}

					})

					AfterEach(func() {
						helper.DeleteDir(tmpContext)
					})

					It("should fail to create a service based on a template without metadata", func() {
						stdOut := helper.Cmd("odo", "service", "create", "--from-file", noMetaFileName, "--project", commonVar.Project).ShouldFail().Err()
						Expect(stdOut).To(ContainSubstring("couldn't find \"metadata\" in the yaml"))
					})

					It("should fail to create a service based on a template with invalid metadata", func() {
						stdOut := helper.Cmd("odo", "service", "create", "--from-file", invalidFileName, "--project", commonVar.Project).ShouldFail().Err()
						Expect(stdOut).To(ContainSubstring("couldn't find metadata.name in the yaml"))
					})
				})
			})
		})
	})
})
