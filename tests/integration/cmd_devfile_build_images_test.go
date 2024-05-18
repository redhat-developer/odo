package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile build-images command tests", Label(helper.LabelSkipOnOpenShift), func() {

	var commonVar helper.CommonVar
	var cmpName string

	var _ = BeforeEach(func() {
		cmpName = helper.RandString(6)
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		var _ = Context("label "+label, Label(label), func() {

			When("using a devfile.yaml containing an Image component", func() {

				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-outerloop.yaml")).ShouldPass()
					helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				})

				for _, tt := range []struct {
					name          string
					args          []string
					env           []string
					shouldPass    bool
					checkOutputFn func(stdout, stderr string)
				}{
					{
						name:       "should run odo build-images without push",
						env:        []string{"PODMAN_CMD=echo"},
						shouldPass: true,
						checkOutputFn: func(stdout, stderr string) {
							Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f %s %s",
								filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
						},
					},
					{
						name:       "should run odo build-images --push",
						args:       []string{"--push"},
						env:        []string{"PODMAN_CMD=echo"},
						shouldPass: true,
						checkOutputFn: func(stdout, stderr string) {
							Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f %s %s",
								filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
							Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
						},
					},
					{
						name: "should pass extra args to Podman",
						env: []string{
							"PODMAN_CMD=echo",
							"ODO_IMAGE_BUILD_ARGS=--platform=linux/amd64;--build-arg=MY_ARG=my_value",
						},
						shouldPass: true,
						checkOutputFn: func(stdout, stderr string) {
							Expect(stdout).To(ContainSubstring("build --platform=linux/amd64 --build-arg=MY_ARG=my_value -t quay.io/unknown-account/myimage -f %s %s",
								filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
						},
					},
					{
						name: "should pass extra args to Docker",
						env: []string{
							"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
							"DOCKER_CMD=echo",
							"ODO_IMAGE_BUILD_ARGS=--platform=linux/amd64;--build-arg=MY_ARG=my_value",
						},
						shouldPass: true,
						checkOutputFn: func(stdout, stderr string) {
							Expect(stdout).To(ContainSubstring("build --platform=linux/amd64 --build-arg=MY_ARG=my_value -t quay.io/unknown-account/myimage -f %s %s",
								filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
						},
					},
				} {
					tt := tt
					It(tt.name, func() {
						args := []string{"build-images"}
						args = append(args, tt.args...)
						env := []string{"PODMAN_CMD=echo"}
						env = append(env, tt.env...)

						cmd := helper.Cmd("odo", args...).AddEnv(env...)
						if tt.shouldPass {
							cmd = cmd.ShouldPass()
						} else {
							cmd = cmd.ShouldFail()
						}
						stdout, stderr := cmd.OutAndErr()
						tt.checkOutputFn(stdout, stderr)
					})
				}
			})

			When("using a devfile.yaml with no Image component", func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.Cmd("odo", "init", "--name", "aname",
						"--devfile-path",
						helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
					helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				})
				It("should not be able to run odo build-images", func() {
					stdout, stderr := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldFail().OutAndErr()
					// Make sure no "{podman,docker} build -t ..." command gets executed
					imageBuildCmd := "build -t "
					Expect(stdout).ShouldNot(ContainSubstring(imageBuildCmd))
					Expect(stderr).ShouldNot(ContainSubstring(imageBuildCmd))
					Expect(stderr).To(ContainSubstring("no component with type \"Image\" found in Devfile"))
				})
			})

			When("using a devfile.yaml containing an Image component with Dockerfile args", func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-outerloop-args.yaml")).ShouldPass()
					helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				})

				It("should use args to build image when running odo build-images", func() {
					stdout := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
					Expect(stdout).To(ContainSubstring("build -t myimage -f %s %s --unknown-flag value",
						filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
				})

				It("should be able pass extra flags to Podman/Docker build command", func() {
					stdout := helper.Cmd("odo", "build-images").AddEnv(
						"PODMAN_CMD=echo",
						"ODO_IMAGE_BUILD_ARGS=--platform=linux/amd64;--build-arg=MY_ARG=my_value",
					).ShouldPass().Out()
					Expect(stdout).To(ContainSubstring("build --platform=linux/amd64 --build-arg=MY_ARG=my_value -t myimage -f %s %s --unknown-flag value",
						filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
				})

			})

			When("using a devfile.yaml containing an Image component with a build context", func() {

				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.Cmd("odo", "init", "--name", "aname",
						"--devfile-path",
						helper.GetExamplePath("source", "devfiles", "nodejs",
							"devfile-outerloop-project_source-in-docker-build-context.yaml")).ShouldPass()
					helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				})

				for _, scope := range []struct {
					name    string
					envvars []string
				}{
					{
						name:    "Podman",
						envvars: []string{"PODMAN_CMD=echo"},
					},
					{
						name: "Docker",
						envvars: []string{
							"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
							"DOCKER_CMD=echo",
						},
					},
				} {
					// this is a workaround to ensure that the for loop works with `It` blocks
					scope := scope
					It(fmt.Sprintf("should build image via %s if build context references PROJECT_SOURCE env var", scope.name), func() {
						stdout := helper.Cmd("odo", "build-images").AddEnv(scope.envvars...).ShouldPass().Out()
						lines, err := helper.ExtractLines(stdout)
						Expect(err).ShouldNot(HaveOccurred())
						nbLines := len(lines)
						Expect(nbLines).To(BeNumerically(">", 2))
						containerImage := "localhost:5000/devfile-nodejs-deploy:0.1.0" // from Devfile yaml file
						dockerfilePath := filepath.Join(commonVar.Context, "Dockerfile")
						buildCtx := commonVar.Context
						Expect(stdout).To(ContainSubstring(
							fmt.Sprintf("build -t %s -f %s %s", containerImage, dockerfilePath, buildCtx)))
					})
				}
			})

			When("using a devfile.yaml containing an Image component with no build context", func() {

				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "issue-5600-devfile-with-image-component-and-no-buildContext.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						cmpName)
					helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				})

				for _, scope := range []struct {
					name    string
					envvars []string
				}{
					{
						name:    "Podman",
						envvars: []string{"PODMAN_CMD=echo"},
					},
					{
						name: "Docker",
						envvars: []string{
							"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
							"DOCKER_CMD=echo",
						},
					},
				} {
					// this is a workaround to ensure that the for loop works with `It` blocks
					scope := scope
					It(fmt.Sprintf("should build image via %s by defaulting build context to devfile path", scope.name), func() {
						stdout := helper.Cmd("odo", "build-images").AddEnv(scope.envvars...).ShouldPass().Out()
						lines, err := helper.ExtractLines(stdout)
						Expect(err).ShouldNot(HaveOccurred())
						nbLines := len(lines)
						Expect(nbLines).To(BeNumerically(">", 2))
						containerImage := "localhost:5000/devfile-nodejs-deploy:0.1.0" // from Devfile yaml file
						dockerfilePath := filepath.Join(commonVar.Context, "Dockerfile")
						buildCtx := commonVar.Context
						Expect(stdout).To(ContainSubstring(
							fmt.Sprintf("build -t %s -f %s %s", containerImage, dockerfilePath, buildCtx)))
					})
				}
			})

			for _, env := range [][]string{
				{"PODMAN_CMD=echo"},
				{
					"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
					"DOCKER_CMD=echo",
				},
			} {
				env := env
				Describe("using a Devfile with an image component using a remote Dockerfile", func() {

					BeforeEach(func() {
						helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
						helper.CopyExampleDevFile(
							filepath.Join("source", "devfiles", "nodejs", "devfile-outerloop.yaml"),
							path.Join(commonVar.Context, "devfile.yaml"),
							cmpName)
					})

					When("remote server returns an error", func() {
						var server *httptest.Server
						var url string
						BeforeEach(func() {
							server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								w.WriteHeader(http.StatusNotFound)
							}))
							url = server.URL

							helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "./Dockerfile", url)
						})

						AfterEach(func() {
							server.Close()
						})

						It("should not build images", func() {
							cmdWrapper := helper.Cmd("odo", "build-images").AddEnv(env...).ShouldFail()
							stderr := cmdWrapper.Err()
							stdout := cmdWrapper.Out()
							Expect(stderr).To(ContainSubstring("failed to retrieve " + url))
							Expect(stdout).NotTo(ContainSubstring("build -t quay.io/unknown-account/myimage -f "))
						})

						It("should not run 'odo build-images --push'", func() {
							cmdWrapper := helper.Cmd("odo", "build-images", "--push").AddEnv(env...).ShouldFail()
							stderr := cmdWrapper.Err()
							stdout := cmdWrapper.Out()
							Expect(stderr).To(ContainSubstring("failed to retrieve " + url))
							Expect(stdout).NotTo(ContainSubstring("build -t quay.io/unknown-account/myimage -f "))
							Expect(stdout).NotTo(ContainSubstring("push quay.io/unknown-account/myimage"))
						})
					})

					When("remote server returns a valid file", func() {
						var buildRegexp string
						var server *httptest.Server
						var url string

						BeforeEach(func() {
							buildRegexp = regexp.QuoteMeta("build -t quay.io/unknown-account/myimage -f ") +
								".*\\.dockerfile " + regexp.QuoteMeta(commonVar.Context)
							server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								fmt.Fprintf(w, `# Dockerfile
FROM node:8.11.1-alpine
COPY . /app
WORKDIR /app
RUN npm install
CMD ["npm", "start"]
`)
							}))
							url = server.URL

							helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "./Dockerfile", url)
						})

						AfterEach(func() {
							server.Close()
						})

						It("should build images", func() {
							stdout := helper.Cmd("odo", "build-images").AddEnv(env...).ShouldPass().Out()
							lines, _ := helper.ExtractLines(stdout)
							_, ok := helper.FindFirstElementIndexMatchingRegExp(lines, buildRegexp)
							Expect(ok).To(BeTrue(), "build regexp not found in output: "+buildRegexp)
						})

						It("should run 'odo build-images --push'", func() {
							stdout := helper.Cmd("odo", "build-images", "--push").AddEnv(env...).ShouldPass().Out()
							lines, _ := helper.ExtractLines(stdout)
							_, ok := helper.FindFirstElementIndexMatchingRegExp(lines, buildRegexp)
							Expect(ok).To(BeTrue(), "build regexp not found in output: "+buildRegexp)
							Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
						})
					})
				})
			}
		})
	}

	// More details on https://github.com/devfile/api/issues/852#issuecomment-1211928487
	When("starting with Devfile with autoBuild or deployByDefault components", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-autobuild-deploybydefault.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})

		When("building images", func() {
			var stdout string

			BeforeEach(func() {
				stdout = helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			})

			It("should build all Image components regardless of autoBuild", func() {
				for _, tag := range []string{
					"autobuild-true-and-referenced",
					"autobuild-true-and-not-referenced",
					"autobuild-false-and-referenced",
					"autobuild-false-and-not-referenced",
					"autobuild-not-set-and-referenced",
					"autobuild-not-set-and-not-referenced",
				} {
					Expect(stdout).Should(ContainSubstring("Building Image: localhost:5000/odo-dev/node:%s", tag))
				}
			})
		})

		When("building and pushing images", func() {
			var stdout string

			BeforeEach(func() {
				stdout = helper.Cmd("odo", "build-images", "--push").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			})

			It("should build and push all Image components regardless of autoBuild", func() {
				for _, tag := range []string{
					"autobuild-true-and-referenced",
					"autobuild-true-and-not-referenced",
					"autobuild-false-and-referenced",
					"autobuild-false-and-not-referenced",
					"autobuild-not-set-and-referenced",
					"autobuild-not-set-and-not-referenced",
				} {
					Expect(stdout).Should(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:%s", tag))
				}
			})
		})
	})

	When("using a Devfile with variable image names", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})

		checkOutput := func(stdout string, images []string, push bool) {
			var matchers []types.GomegaMatcher
			for _, img := range images {
				msg := "Building"
				if push {
					msg += " & Pushing"
				}
				matchers = append(matchers, ContainSubstring("%s Image: %s", msg, img))
				matchers = append(matchers, ContainSubstring("build -t %s -f %s %s", img, filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context))
				if push {
					matchers = append(matchers, ContainSubstring("push %s", img))
				}
			}
			Expect(stdout).Should(SatisfyAll(matchers...))
		}

		for _, push := range []bool{false, true} {
			push := push
			initialArgs := []string{"build-images"}
			if push {
				initialArgs = append(initialArgs, "--push")
			}
			It(fmt.Sprintf("should build images with default variable values (push=%v)", push), func() {
				args := initialArgs
				stdout := helper.Cmd("odo", args...).AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				checkOutput(stdout, []string{"my-image-1:1.2.3-rc4", "my-image-2:2.3.4-alpha5"}, push)
			})

			It(fmt.Sprintf("should build images with --var (push=%v)", push), func() {
				args := initialArgs
				args = append(args, "--var", "VARIABLE_CONTAINER_IMAGE_2=my-image-2-overridden:next")
				stdout := helper.Cmd("odo", args...).
					AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				checkOutput(stdout, []string{"my-image-1:1.2.3-rc4", "my-image-2-overridden:next"}, push)
			})

			It(fmt.Sprintf("should build images with --var-file (push=%v)", push), func() {
				var varFilename = filepath.Join(commonVar.Context, "vars.txt")
				err := helper.CreateFileWithContent(varFilename, `VARIABLE_CONTAINER_IMAGE_1=my-image-1-overridden-from-file:next
VARIABLE_CONTAINER_IMAGE_2=my-image-2-overridden-from-file:next
`)
				Expect(err).ShouldNot(HaveOccurred())

				args := initialArgs
				args = append(args, "--var-file", varFilename)
				stdout := helper.Cmd("odo", args...).
					AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				checkOutput(stdout, []string{"my-image-1-overridden-from-file:next", "my-image-2-overridden-from-file:next"}, push)
			})
		}
	})
})
