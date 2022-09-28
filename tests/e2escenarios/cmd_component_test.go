package e2escenarios

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo dev command tests", func() {
	fmt.Println("Hello, World!")
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// automate test for python components
	Context("e2e test for the python component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "python", "project"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "python", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8080"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Hello World!"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
	// automate test for go components
	Context("e2e test for the go component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "go", "project"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "go", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8080"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Hello, !"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
	// automate test for dotnet60 components
	Context("e2e test for the dotnet60 component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "dotnet60", "project"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "dotnet60", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8080"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Welcome"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
	// automate test for Java-Maven components
	Context("e2e test for the java-maven component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "java-maven", "project"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "java-maven", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8080"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Hello World!"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
	// automate test for python-django components
	Context("e2e test for the python-django component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "python-django", "project"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "python-django", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8000"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"The install worked successfully! Congratulations!"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
	// automate test for springboot components
	Context("e2e test for the springboot component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project-e2e"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-e2e.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8080"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Hello World!"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
	// automate test for php-laravel components
	Context("e2e test for the php-laravel component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "php-laravel", "project"), commonVar.Context)
				helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "php-laravel", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					url := fmt.Sprintf("http://%s", ports["8000"])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Laravel"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})

})
