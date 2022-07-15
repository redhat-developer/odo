package integration

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile build-images command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("using a devfile.yaml containing an Image component", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-outerloop.yaml")).ShouldPass()
			helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
		})
		It("should run odo build-images without push", func() {
			stdout := helper.Cmd("odo", "build-images").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
		})

		It("should run odo build-images --push", func() {
			stdout := helper.Cmd("odo", "build-images", "--push").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
			Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
		})
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
			Expect(stdout).To(ContainSubstring("--unknown-flag value"))
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
				Expect(lines[nbLines-2]).To(BeEquivalentTo(
					fmt.Sprintf("build -t %s -f %s %s", containerImage, dockerfilePath, buildCtx)))
			})
		}
	})

	When("using a devfile.yaml containing an Image component with no build context", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs",
					"issue-5600-devfile-with-image-component-and-no-buildContext.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"))
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
				Expect(lines[nbLines-2]).To(BeEquivalentTo(
					fmt.Sprintf("build -t %s -f %s %s", containerImage, dockerfilePath, buildCtx)))
			})
		}
	})
})
