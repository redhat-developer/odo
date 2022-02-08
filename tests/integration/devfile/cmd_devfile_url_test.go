package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/odo/tests/helper"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile url command tests", func() {
	var componentName string
	var commonVar helper.CommonVar
	url1 := helper.RandString(5)
	url1Port := "5000"
	url2 := helper.RandString(5)
	url2Port := "9000"
	// Required to ensure duplicate ports are not used to create a URL
	nodejsDevfilePort := "3000"
	nodejsDevfileURLName := "3000-tcp"
	springbootDevfilePort := "8080"
	portNumber := "3030"
	host := helper.RandString(5) + ".com"

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should error out on devfile flag", func() {
		helper.Cmd("odo", "url", "create", "mynodejs", "--devfile", "invalid.yaml").ShouldFail()
		helper.Cmd("odo", "url", "delete", "mynodejs", "--devfile", "invalid.yaml").ShouldFail()
	})

	When("creating a Nodejs component", func() {
		stdout := ""

		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})

		It("should not allow creating an invalid host", func() {
			stdout = helper.Cmd("odo", "url", "create", "--host", "https://127.0.0.1:60104", "--port", portNumber, "--ingress").ShouldFail().Err()
			Expect(stdout).To(ContainSubstring("is not a valid host name"))
		})

		It("should not allow using tls secret if url is not secure", func() {
			stdout = helper.Cmd("odo", "url", "create", "--tls-secret", "foo", "--port", portNumber, "--ingress").ShouldFail().Err()
			Expect(stdout).To(ContainSubstring("TLS secret is only available for secure URLs of Ingress kind"))
		})

		It("should report multiple issues when it's the case", func() {
			stdout = helper.Cmd("odo", "url", "create", "--host", "https://127.0.0.1:60104", "--tls-secret", "foo", "--port", portNumber, "--ingress").ShouldFail().Err()
			Expect(stdout).To(And(ContainSubstring("is not a valid host name"), ContainSubstring("TLS secret is only available for secure URLs of Ingress kind")))
		})

		It("should not allow to create URL with duplicate port", func() {
			stdout = helper.Cmd("odo", "url", "create", "--port", nodejsDevfilePort).ShouldFail().Err()
			Expect(stdout).To(ContainSubstring("port 3000 already exists in devfile endpoint entry"))
		})

		It("should not allow creating under an invalid container", func() {
			containerName := helper.RandString(5)
			stdout = helper.Cmd("odo", "url", "create", "--host", "com", "--port", portNumber, "--container", containerName, "--ingress").ShouldFail().Err()
			helper.MatchAllInOutput(stdout, []string{"container", containerName, "not exist"})
		})

		It("should not allow creating an endpoint with same name", func() {
			stdout = helper.Cmd("odo", "url", "create", nodejsDevfileURLName, "--host", "com", "--port", nodejsDevfilePort, "--ingress").ShouldFail().Err()
			Expect(stdout).To(ContainSubstring(fmt.Sprintf("url %s already exists in devfile endpoint entry", nodejsDevfileURLName)))
		})

		When("creating ingress url1 with port flag and doing odo push", func() {

			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--secure", "--ingress").ShouldPass()
				helper.Cmd("odo", "push").ShouldPass()
			})

			It("should check state of url1 list", func() {
				stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "true", "ingress"})
			})
			It("should take user provided url name", func() {
				stdout = string(commonVar.CliRunner.Run("get", "svc", "--namespace", commonVar.Project, fmt.Sprintf("%v-%v", componentName, "app"), "-ojsonpath='{.spec.ports[*].name}'").Out.Contents())
				Expect(stdout).To(ContainSubstring(url1))
			})

			When("creating ingress url2 with port flag and doing odo push", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "url", "create", url2, "--port", url2Port, "--host", host, "--ingress").ShouldPass()
					stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				})
				It("should check state of url2", func() {
					helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "ingress"})
				})
				When("deleting url1", func() {
					BeforeEach(func() {
						helper.Cmd("odo", "url", "delete", url1, "-f").ShouldPass()
						stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
					})
					It("should check status of url1 and url2", func() {
						helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", "true", "ingress"})
						helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "ingress"})
					})
				})
			})
		})
		When("creating a url with -o json", func() {
			var createExpected []string
			var createValues []gjson.Result
			var createJSON string

			BeforeEach(func() {
				helper.Cmd("odo", "url", "delete", nodejsDevfileURLName, "-f").ShouldPass()
				createJSON = helper.Cmd("odo", "url", "create", url1, "--port", nodejsDevfilePort, "--host", host, "--ingress", "-o", "json").ShouldPass().Out()

			})

			It("should validate machine readable output for url create", func() {
				createValues = gjson.GetMany(createJSON, "kind", "metadata.name", "spec.port", "status.state")
				createExpected = []string{"URL", url1, nodejsDevfilePort, "Not Pushed"}
				Expect(helper.GjsonMatcher(createValues, createExpected)).To(Equal(true))
			})

			When("doing odo push", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
				})
				It("validate machine readable output for url list", func() {
					helper.WaitForCmdOut("odo", []string{"url", "list", "-o", "json"}, 1, true, func(output string) bool {
						if strings.Contains(output, url1) {
							values := gjson.GetMany(output, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.host", "items.0.status.state")
							expected := []string{"List", "URL", url1, url1, "Pushed"}
							Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
							return true
						}
						return false
					})
				})
			})

		})
		XWhen("creating a URL without port flag", func() {
			// marking it as pending, since we want to start using port number as a key to URLs in devfile,
			// FIXME: and odo should not allow using url create without a port number, this should be taken care of in v3-alpha2

			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--host", host, "--ingress").ShouldPass()
				stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
			})
			It("should create a URL without port flag if only one port exposed in devfile", func() {
				helper.MatchAllInOutput(stdout, []string{url1, "3000", "Not Pushed"})
			})
		})

		When("creating a secure URL and doing odo push", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--secure", "--ingress").ShouldPass()
				stdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should list secure URL", func() {
				helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host})
				stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host, "true"})
			})
		})

		When("create with now flag", func() {
			BeforeEach(func() {
				stdout = helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--now", "--ingress").ShouldPass().Out()
			})
			It("should check if url created for component", func() {
				// check the env for the runMode
				envOutput, err := helper.ReadFile(filepath.Join(commonVar.Context, ".odo/env/env.yaml"))
				Expect(err).To(BeNil())
				Expect(envOutput).To(ContainSubstring(" RunMode: run"))
				helper.MatchAllInOutput(stdout, []string{"URL " + url1 + " created for component", "http:", url1 + "." + host})
			})
			When("delete with now flag", func() {
				BeforeEach(func() {
					stdout = helper.Cmd("odo", "url", "delete", url1, "--now", "-f").ShouldPass().Out()
				})
				It("should check if successfully deleted", func() {
					helper.MatchAllInOutput(stdout, []string{"URL " + url1 + " successfully deleted", "Applying URL changes"})
				})
			})

		})

		When("creating a url and doing odo push twice", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--ingress").ShouldPass()
				helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
				stdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should be able to push again twice", func() {
				helper.DontMatchAllInOutput(stdout, []string{"successfully deleted", "created"})
				Expect(stdout).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))
			})
			When("deleting url and doing odo push twice", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "url", "delete", url1, "-f").ShouldPass()
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
					stdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				})
				It("should be able to push again twice", func() {
					helper.DontMatchAllInOutput(stdout, []string{"successfully deleted", "created"})
					Expect(stdout).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))
				})
			})
		})

		When("creating URL with path flag", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--path", "testpath", "--ingress").ShouldPass()
				stdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should create URL with path defined in Endpoint", func() {
				helper.MatchAllInOutput(stdout, []string{url1, "/testpath", "created"})
			})
		})

		When("executing odo outside of context", func() {
			subFolderContext := ""
			BeforeEach(func() {
				subFolderContext = filepath.Join(commonVar.Context, helper.RandString(6))
				helper.MakeDir(subFolderContext)
				helper.Chdir(subFolderContext)
			})

			It("should fail if url is created from non context dir", func() {
				stdout = helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--ingress").ShouldFail().Err()
				Expect(stdout).To(ContainSubstring("the current directory does not represent an odo component"))
			})
			It("should fail if host flag not provided with ingress flag", func() {
				stdout = helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--ingress", "--context", commonVar.Context).ShouldFail().Err()
				Expect(stdout).To(ContainSubstring("host must be provided"))
			})

			When("creating url1,url2 and doing odo push with context flag", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "url", "create", url2, "--port", url2Port, "--host", host, "--ingress", "--context", commonVar.Context).ShouldPass()
					helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--ingress", "--context", commonVar.Context).ShouldPass()
					stdout = helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
				})
				It("should successfully push url1 and url2", func() {
					helper.MatchAllInOutput(stdout, []string{url1 + "." + host, url2})
				})
				When("doing url list", func() {
					BeforeEach(func() {
						stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
					})
					It("should successfully push url1 and url2", func() {
						helper.MatchAllInOutput(stdout, []string{url1, url2, "Pushed", "false", "ingress"})
					})

				})
			})
		})
	})

	When("creating a java-springboot component", func() {
		var stdout string
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "springboot", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
		})

		When("create URLs under different container names", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--container", "runtime", "--ingress").ShouldPass()
				helper.Cmd("odo", "url", "create", url2, "--port", url2Port, "--host", host, "--container", "tools", "--ingress").ShouldPass()
				stdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should create URLs under different container names", func() {
				helper.MatchAllInOutput(stdout, []string{url1, url2, "created"})
			})

		})
		XWhen("create URLs under different container names with same port number", func() {
			// marking it as pending, since the current behavior of odo url create is buggy, this should be fixed in odo v3-alpha2
			BeforeEach(func() {
				stdout = helper.Cmd("odo", "url", "create", url1, "--port", "8080", "--host", host, "--container", "tools", "--ingress").ShouldFail().Err()
			})
			It("should not create URLs under different container names with same port number", func() {
				helper.MatchAllInOutput(stdout, []string{fmt.Sprintf("cannot set URL %s under container tools", url1), fmt.Sprintf("TargetPort %s is being used under container runtime", springbootDevfilePort)})
			})
		})
	})

	When("Creating nodejs component and url with .devfile.yaml", func() {
		var stdout string
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, ".devfile.yaml"))
			helper.Cmd("odo", "create", "--project", commonVar.Project, componentName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			stdout = helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--host", host, "--container", "runtime", "--ingress").ShouldPass().Out()
		})

		It("should create url", func() {
			helper.MatchAllInOutput(stdout, []string{url1, "created"})
		})
		When("listing url", func() {
			BeforeEach(func() {
				stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			})
			It("should list urls", func() {
				helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "false", "ingress"})
			})
		})
	})

	Context("Testing URLs for OpenShift specific scenarios", func() {

		stdout := ""
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})

		When("creating a nodejs component", func() {
			ingressurl := helper.RandString(5)

			BeforeEach(func() {
				helper.Cmd("odo", "create", "--project", commonVar.Project, componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			})

			It("should error out when a host is provided with a route on a openShift cluster", func() {
				output := helper.Cmd("odo", "url", "create", url1, "--host", "com", "--port", portNumber).ShouldFail().Err()
				Expect(output).To(ContainSubstring("host is not supported"))
			})

			When("creating multiple url with different state", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "url", "create", url1, "--port", url1Port, "--secure").ShouldPass()
					helper.Cmd("odo", "url", "create", ingressurl, "--port", portNumber, "--host", host, "--ingress").ShouldPass()
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
					helper.Cmd("odo", "url", "create", url2, "--port", url2Port).ShouldPass()
				})

				It("should list route and ingress urls with appropriate state", func() {
					stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
					helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "true", "route"})
					helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "route"})
					helper.MatchAllInOutput(stdout, []string{ingressurl, "Pushed", "false", "ingress"})
				})

				When("url1 is deleted locally", func() {

					BeforeEach(func() {
						helper.Cmd("odo", "url", "delete", url1, "-f").ShouldPass()
						stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
					})

					It("should list route and ingress urls with appropriate state", func() {
						helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", "true", "route"})
						helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "route"})
						helper.MatchAllInOutput(stdout, []string{ingressurl, "Pushed", "false", "ingress"})
					})
				})
			})

			XIt("should automatically create route on a openShift cluster", func() {
				// marking it as pending, since current odo url create behavior is buggy, and this should be taken care of in v3-alpha2
				helper.Cmd("odo", "url", "create", url1, "--port", nodejsDevfilePort).ShouldPass()
				fileOutput, err := helper.ReadFile(filepath.Join(commonVar.Context, "devfile.yaml"))
				Expect(err).To(BeNil())
				helper.MatchAllInOutput(fileOutput, []string{nodejsDevfileURLName, nodejsDevfilePort})
			})

			When("doing odo push twice and doing url list", func() {
				var pushstdout string

				BeforeEach(func() {
					helper.Cmd("odo", "url", "create", url1, "--port", url1Port).ShouldPass()
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
					pushstdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
					stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				})

				It("should push successfully", func() {
					helper.DontMatchAllInOutput(pushstdout, []string{"successfully deleted", "created"})
					Expect(pushstdout).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))
					Expect(stdout).Should(ContainSubstring(url1))
				})

				When("deleting the url1 and doing odo push twice and doing url list", func() {

					BeforeEach(func() {
						helper.Cmd("odo", "url", "delete", url1, "-f").ShouldPass()
						helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
						pushstdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
						stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
					})

					It("should push successfully", func() {
						helper.DontMatchAllInOutput(pushstdout, []string{"successfully deleted", "created"})
						Expect(pushstdout).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))
						Expect(stdout).ShouldNot(ContainSubstring(url1))
					})
				})
			})
			When("doing odo push", func() {
				BeforeEach(func() {
					stdout = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
					helper.MatchAllInOutput(stdout, []string{"URL 3000-tcp", "created"})
					stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				})

				It("should create a route on a openShift cluster without calling url create", func() {
					Expect(stdout).Should(ContainSubstring("3000-tcp"))
				})
			})
		})
	})

	When("creating a python component", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Chdir(commonVar.Context)
			helper.Cmd("odo", "create", "--project", commonVar.Project, componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "python", "devfile-registry.yaml")).ShouldPass()
		})

		When("creating a url and doing odo push", func() {
			var output string
			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", url1, "--port", "5000", "--host", "com", "--ingress").ShouldPass()
				helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
			})

			It("should create a url for a unsupported devfile component", func() {
				output = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				Expect(output).Should(ContainSubstring(url1))
			})
		})
	})

	XContext("Testing URLs for Kubernetes specific scenarios", func() {
		// TODO: marking it as pending, since current odo url create behavior is buggy, and this should be taken care of in v3-alpha2
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") != "true" {
				Skip("This is a Kubernetes specific scenario, skipping")
			}
		})

		When("creating a nodejs component", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "create", "--project", commonVar.Project, componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			})

			When("creating url", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "url", "create", "--host", "com", "--port", nodejsDevfilePort).ShouldPass()
				})

				When("creating a second url", func() {

					BeforeEach(func() {
						helper.Cmd("odo", "url", "create", url1, "--host", "com", "--port", url1Port).ShouldPass()
					})

					It("should use an existing URL when there are URLs with no host defined in the env file with same port", func() {
						fileOutput, err := helper.ReadFile(filepath.Join(commonVar.Context, "devfile.yaml"))
						Expect(err).To(BeNil())
						helper.MatchAllInOutput(fileOutput, []string{url1, url1Port})
						count := strings.Count(fileOutput, "targetPort")
						Expect(count).To(Equal(2))
					})
				})
				It("should verify", func() {
					fileOutput, err := helper.ReadFile(filepath.Join(commonVar.Context, "devfile.yaml"))
					Expect(err).To(BeNil())
					helper.MatchAllInOutput(fileOutput, []string{nodejsDevfileURLName, nodejsDevfilePort})
					count := strings.Count(fileOutput, "targetPort")
					Expect(count).To(Equal(1))
				})
			})
		})
	})
})
