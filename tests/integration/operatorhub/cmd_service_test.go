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
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo service command tests for OperatorHub", func() {

	var commonVar helper.CommonVar
	var oc helper.OcRunner

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		oc = helper.NewOcRunner("oc")
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Operators are installed in the cluster", func() {

		JustBeforeEach(func() {
			// wait till odo can see that all operators installed by setup script in the namespace
			odoArgs := []string{"catalog", "list", "services"}
			operators := []string{"etcdoperator", "service-binding-operator"}
			for _, operator := range operators {
				helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
					return strings.Contains(output, operator)
				})
			}
		})

		JustAfterEach(func() {
			helper.DeleteProject(commonVar.Project)
		})

		It("should not allow creating service without valid context", func() {
			stdOut := helper.CmdShouldFail("odo", "service", "create")
			Expect(stdOut).To(ContainSubstring("service can be created/deleted from a valid component directory only"))
		})

		Context("a specific operator is installed", func() {
			var etcdOperator string
			var etcdCluster string

			JustBeforeEach(func() {
				operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
				etcdOperator = regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
				etcdCluster = fmt.Sprintf("%s/EtcdCluster", etcdOperator)
			})

			It("should describe the operator with human-readable output", func() {
				output := helper.CmdShouldPass("odo", "catalog", "describe", "service", etcdCluster)
				Expect(output).To(ContainSubstring("Kind: EtcdCluster"))
			})

			It("should describe the operator with json output", func() {
				outputJSON := helper.CmdShouldPass("odo", "catalog", "describe", "service", etcdCluster, "-o", "json")
				values := gjson.GetMany(outputJSON, "spec.kind", "spec.displayName")
				expected := []string{"EtcdCluster", "etcd Cluster"}
				Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
			})

			It("should find the services by keyword", func() {
				stdOut := helper.CmdShouldPass("odo", "catalog", "search", "service", "etcd")
				helper.MatchAllInOutput(stdOut, []string{"etcdoperator", "EtcdCluster"})

				stdOut = helper.CmdShouldPass("odo", "catalog", "search", "service", "EtcdCluster")
				helper.MatchAllInOutput(stdOut, []string{"etcdoperator", "EtcdCluster"})

				stdOut = helper.CmdShouldFail("odo", "catalog", "search", "service", "dummy")
				Expect(stdOut).To(ContainSubstring("no service matched the query: dummy"))
			})

			It("should list the operator in JSON output", func() {
				jsonOut := helper.CmdShouldPass("odo", "catalog", "list", "services", "-o", "json")
				helper.MatchAllInOutput(jsonOut, []string{"etcdoperator"})
			})

			When("a nodejs component is created", func() {

				JustBeforeEach(func() {
					helper.CmdShouldPass("odo", "create", "nodejs")
				})

				It("should fail for interactive mode", func() {
					stdOut := helper.CmdShouldFail("odo", "service", "create")
					Expect(stdOut).To(ContainSubstring("odo doesn't support interactive mode for creating Operator backed service"))
				})

				It("should fail if service name doesn't adhere to <service-type>/<service-name> format", func() {
					if os.Getenv("KUBERNETES") == "true" {
						Skip("This is a OpenShift specific scenario, skipping")
					}
					helper.CmdShouldFail("odo", "link", "EtcdCluster")
					helper.CmdShouldFail("odo", "link", "EtcdCluster/")
					helper.CmdShouldFail("odo", "link", "/example")
				})

				When("odo push is executed", func() {
					JustBeforeEach(func() {
						helper.CmdShouldPass("odo", "push")
					})

					It("should fail if the provided service doesn't exist in the namespace", func() {
						if os.Getenv("KUBERNETES") == "true" {
							Skip("This is a OpenShift specific scenario, skipping")
						}
						stdOut := helper.CmdShouldFail("odo", "link", "EtcdCluster/example")
						Expect(stdOut).To(ContainSubstring("Couldn't find service named %q", "EtcdCluster/example"))
					})
				})

				When("an EtcdCluster instance is created in dryRun mode", func() {

					var stdOut string

					JustBeforeEach(func() {
						stdOut = helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--dry-run", "--project", commonVar.Project)
					})

					It("should only output the definition of the CR that will be used to start service", func() {
						helper.MatchAllInOutput(stdOut, []string{"apiVersion", "kind"})
					})

					When("the output of the command is stored in a file", func() {

						var fileName string

						JustBeforeEach(func() {
							randomFileName := helper.RandString(6) + ".yaml"
							fileName = filepath.Join(os.TempDir(), randomFileName)
							if err := ioutil.WriteFile(fileName, []byte(stdOut), 0644); err != nil {
								fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
							}
						})

						JustAfterEach(func() {
							os.Remove(fileName)
						})

						When("a service is created from the output of the dryRun command with no name", func() {
							JustBeforeEach(func() {
								helper.CmdShouldPass("odo", "service", "create", "--from-file", fileName, "--project", commonVar.Project)
							})

							JustAfterEach(func() {
								helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")
							})

							When("odo push is executed", func() {
								JustBeforeEach(func() {
									helper.CmdShouldPass("odo", "push")
								})

								It("should create pods in running state", func() {
									oc.PodsShouldBeRunning(commonVar.Project, `example-.[a-z0-9]*`)
								})
							})
						})

						When("a service is created from the output of the dryRun command with a specific name", func() {

							var name string
							var svcFullName string
							JustBeforeEach(func() {
								name = helper.RandString(6)
								svcFullName = strings.Join([]string{"EtcdCluster", name}, "/")
								helper.CmdShouldPass("odo", "service", "create", "--from-file", fileName, name, "--project", commonVar.Project)
							})

							JustAfterEach(func() {
								helper.CmdShouldPass("odo", "service", "delete", svcFullName, "-f")
							})

							When("odo push is executed", func() {

								JustBeforeEach(func() {
									helper.CmdShouldPass("odo", "push")
								})

								It("should fail to create a service again with the same name", func() {
									stdOut = helper.CmdShouldFail("odo", "service", "create", "--from-file", fileName, name, "--project", commonVar.Project)
									Expect(stdOut).To(ContainSubstring("please provide a different name or delete the existing service first"))
								})

								It("should create pods in running state", func() {
									oc.PodsShouldBeRunning(commonVar.Project, name+`-.[a-z0-9]*`)
								})
							})
						})
					})
				})

				When("an EtcdCluster instance is created with no name", func() {
					var stdOut string
					JustBeforeEach(func() {
						stdOut = helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--project", commonVar.Project)
						Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))
					})

					It("should insert service definition in devfile.yaml", func() {
						devfilePath := filepath.Join(commonVar.Context, "devfile.yaml")
						content, err := ioutil.ReadFile(devfilePath)
						Expect(err).To(BeNil())
						matchInOutput := []string{"kubernetes", "inlined", "EtcdCluster", "example"}
						helper.MatchAllInOutput(string(content), matchInOutput)
					})

					When("odo push is executed", func() {

						JustBeforeEach(func() {
							helper.CmdShouldPass("odo", "push")
						})

						It("should create pods in running state", func() {
							oc.PodsShouldBeRunning(commonVar.Project, `example-.[a-z0-9]*`)
						})

						It("should list the service", func() {
							// now test listing of the service using odo
							stdOut := helper.CmdShouldPass("odo", "service", "list")
							Expect(stdOut).To(ContainSubstring("EtcdCluster/example"))
						})

						It("should list the service in JSON format", func() {
							jsonOut := helper.CmdShouldPass("odo", "service", "list", "-o", "json")
							helper.MatchAllInOutput(jsonOut, []string{"\"apiVersion\": \"etcd.database.coreos.com/v1beta2\"", "\"kind\": \"EtcdCluster\"", "\"name\": \"example\""})
						})

						When("a link is created with the service", func() {
							var stdOut string
							JustBeforeEach(func() {
								stdOut = helper.CmdShouldPass("odo", "link", "EtcdCluster/example")
							})

							It("should display a successful message", func() {
								if os.Getenv("KUBERNETES") == "true" {
									Skip("This is a OpenShift specific scenario, skipping")
								}
								Expect(stdOut).To(ContainSubstring("Successfully created link between component"))
							})

							It("Should fail to link it again", func() {
								if os.Getenv("KUBERNETES") == "true" {
									Skip("This is a OpenShift specific scenario, skipping")
								}
								stdOut = helper.CmdShouldFail("odo", "link", "EtcdCluster/example")
								Expect(stdOut).To(ContainSubstring("already linked with the service"))
							})

							When("the link is deleted", func() {
								JustBeforeEach(func() {
									stdOut = helper.CmdShouldPass("odo", "unlink", "EtcdCluster/example")
								})

								It("should display a successful message", func() {
									if os.Getenv("KUBERNETES") == "true" {
										Skip("This is a OpenShift specific scenario, skipping")
									}
									Expect(stdOut).To(ContainSubstring("Successfully unlinked component"))
								})

								It("should fail to delete it again", func() {
									if os.Getenv("KUBERNETES") == "true" {
										Skip("This is a OpenShift specific scenario, skipping")
									}
									stdOut = helper.CmdShouldFail("odo", "unlink", "EtcdCluster/example")
									Expect(stdOut).To(ContainSubstring("failed to unlink the service"))
								})
							})
						})

						When("the service is deleted", func() {
							JustBeforeEach(func() {
								helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")
							})

							It("should delete service definition from devfile.yaml", func() {
								// read the devfile.yaml to check if service definition was deleted
								devfilePath := filepath.Join(commonVar.Context, "devfile.yaml")
								content, err := ioutil.ReadFile(devfilePath)
								Expect(err).To(BeNil())
								matchInOutput := []string{"kubernetes", "inlined", "EtcdCluster", "example"}
								helper.DontMatchAllInOutput(string(content), matchInOutput)
							})

							It("should fail to delete the service again", func() {
								stdOut = helper.CmdShouldFail("odo", "service", "delete", "EtcdCluster/example", "-f")
								Expect(stdOut).To(ContainSubstring("couldn't find service named"))
							})

							When("odo push is executed", func() {
								JustBeforeEach(func() {
									helper.CmdShouldPass("odo", "push")
								})

								It("Should fail listing the services", func() {
									out := helper.CmdShouldFail("odo", "service", "list")
									msg := fmt.Sprintf("no operator backed services found in namespace: %s", commonVar.Project)
									Expect(out).To(ContainSubstring(msg))
								})

								It("Should fail listing the services in JSON format", func() {
									jsonOut := helper.CmdShouldFail("odo", "service", "list", "-o", "json")
									msg := fmt.Sprintf("no operator backed services found in namespace: %s", commonVar.Project)
									msgWithQuote := fmt.Sprintf("\"message\": \"no operator backed services found in namespace: %s\"", commonVar.Project)
									helper.MatchAllInOutput(jsonOut, []string{msg, msgWithQuote})
								})
							})
						})

						When("a second service is created", func() {
							JustBeforeEach(func() {
								stdOut = helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "myetcd2", "--project", commonVar.Project)
								Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))
							})

							When("odo push is executed", func() {
								JustBeforeEach(func() {
									helper.CmdShouldPass("odo", "push")
								})

								It("should list both services", func() {
									stdOut = helper.CmdShouldPass("odo", "service", "list")
									// first service still here
									Expect(stdOut).To(ContainSubstring("EtcdCluster/example"))
									// second service created
									Expect(stdOut).To(ContainSubstring("EtcdCluster/myetcd2"))
								})
							})
						})
					})
				})

				When("an EtcdCluster instance is created with a specific name", func() {

					var name string
					var svcFullName string

					JustBeforeEach(func() {
						name = helper.RandString(6)
						svcFullName = strings.Join([]string{"EtcdCluster", name}, "/")
						helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), name, "--project", commonVar.Project)
					})

					JustAfterEach(func() {
						helper.Cmd("odo", "service", "delete", svcFullName, "-f").ShouldRun()
					})

					It("should be listed as Not pushed", func() {
						stdOut := helper.CmdShouldPass("odo", "service", "list")
						helper.MatchAllInOutput(stdOut, []string{svcFullName, "Not pushed"})
					})

					When("odo push is executed", func() {

						JustBeforeEach(func() {
							helper.CmdShouldPass("odo", "push")
						})

						It("should create pods in running state", func() {
							oc.PodsShouldBeRunning(commonVar.Project, name+`-.[a-z0-9]*`)
						})

						It("should fail to create a service again with the same name", func() {
							stdOut := helper.CmdShouldFail("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), name, "--project", commonVar.Project)
							Expect(stdOut).To(ContainSubstring(fmt.Sprintf("service %q already exists", svcFullName)))
						})

						It("should be listed as Pushed", func() {
							stdOut := helper.CmdShouldPass("odo", "service", "list")
							helper.MatchAllInOutput(stdOut, []string{svcFullName, "Pushed"})
						})

						When("the etcdCluster instance is deleted", func() {
							JustBeforeEach(func() {
								helper.CmdShouldPass("odo", "service", "delete", svcFullName, "-f")
							})

							It("should be listed as Deleted locally", func() {
								stdOut := helper.CmdShouldPass("odo", "service", "list")
								helper.MatchAllInOutput(stdOut, []string{svcFullName, "Deleted locally"})
							})

							When("odo push is executed", func() {
								JustBeforeEach(func() {
									helper.CmdShouldPass("odo", "push")
								})

								It("should not be listed anymore", func() {
									stdOut := helper.Cmd("odo", "service", "list").ShouldRun().Out()
									Expect(strings.Contains(stdOut, svcFullName)).To(BeFalse())
								})
							})
						})

						When("a link is created with a specific name", func() {

							var linkName string

							JustBeforeEach(func() {
								linkName = "link-" + helper.RandString(6)
								helper.CmdShouldPass("odo", "link", "EtcdCluster/"+name, "--name", linkName)
								// for the moment, odo push is not necessary to deploy the link
							})

							JustAfterEach(func() {
								// delete the link
								helper.CmdShouldPass("odo", "unlink", "EtcdCluster/"+name)
							})

							It("should create the link with the specified name", func() {
								ocArgs := []string{"get", "servicebinding", linkName, "-n", commonVar.Project}
								helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
									return strings.Contains(output, linkName)
								})
							})
						})

						When("a link is created with a specific name and bind-as-files flag", func() {

							var linkName string

							JustBeforeEach(func() {
								linkName = "link-" + helper.RandString(6)
								helper.CmdShouldPass("odo", "link", "EtcdCluster/"+name, "--name", linkName, "--bind-as-files")
								// for the moment, odo push is not necessary to deploy the link
							})

							JustAfterEach(func() {
								// delete the link
								helper.CmdShouldPass("odo", "unlink", "EtcdCluster/"+name)
							})

							It("should create a servicebinding resource with bindAsFiles set to true", func() {
								ocArgs := []string{"get", "servicebinding", linkName, "-o", "jsonpath='{.spec.bindAsFiles}'", "-n", commonVar.Project}
								helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
									return strings.Contains(output, "true")
								})
							})
						})
					})
				})

				Context("Invalid service templates exist", func() {

					var tmpContext string
					var noMetaFileName string
					var invalidFileName string

					JustBeforeEach(func() {
						tmpContext = helper.CreateNewContext()

						// TODO write helpers to create such files
						noMetadata := `
apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
spec:
  size: 3
  version: 3.2.13`
						noMetaFile := helper.RandString(6) + ".yaml"
						noMetaFileName = filepath.Join(tmpContext, noMetaFile)
						if err := ioutil.WriteFile(noMetaFileName, []byte(noMetadata), 0644); err != nil {
							fmt.Printf("Could not write yaml spec to file %s because of the error %v", noMetaFileName, err.Error())
						}

						invalidMetadata := `
apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
metadata:
  noname: noname
spec:
  size: 3
  version: 3.2.13`
						invalidMetaFile := helper.RandString(6) + ".yaml"
						invalidFileName = filepath.Join(tmpContext, invalidMetaFile)
						if err := ioutil.WriteFile(invalidFileName, []byte(invalidMetadata), 0644); err != nil {
							fmt.Printf("Could not write yaml spec to file %s because of the error %v", invalidFileName, err.Error())
						}

					})

					JustAfterEach(func() {
						helper.DeleteDir(tmpContext)
					})

					It("should fail to create a service based on a template without metadata", func() {
						stdOut := helper.CmdShouldFail("odo", "service", "create", "--from-file", noMetaFileName, "--project", commonVar.Project)
						Expect(stdOut).To(ContainSubstring("couldn't find \"metadata\" in the yaml"))
					})

					It("should fail to create a service based on a template with invalid metadata", func() {
						stdOut := helper.CmdShouldFail("odo", "service", "create", "--from-file", invalidFileName, "--project", commonVar.Project)
						Expect(stdOut).To(ContainSubstring("couldn't find metadata.name in the yaml"))
					})
				})
			})
		})

	})
})
