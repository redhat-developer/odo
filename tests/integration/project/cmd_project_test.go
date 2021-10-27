package project

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo project command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should display machine ouptut when getting help for odo project list", func() {
		output := helper.Cmd("odo", "project", "list", "--help").ShouldPass().Out()
		Expect(output).To(ContainSubstring("Specify output format, supported format: json"))
	})

	It("should be able to get project", func() {
		projectGetJSON := helper.Cmd("odo", "project", "get", "-o", "json").ShouldPass().Out()
		getOutputJSON, err := helper.Unindented(projectGetJSON)
		Expect(err).Should(BeNil())
		valuesJSON := gjson.GetMany(getOutputJSON, "kind", "metadata.name", "status.active")
		expectedJSON := []string{"Project", commonVar.Project, "true"}
		Expect(helper.GjsonMatcher(valuesJSON, expectedJSON)).To(Equal(true))

	})

	It("should display the help of project command", func() {
		projectHelp := helper.Cmd("odo", "project", "-h").ShouldPass().Out()
		Expect(projectHelp).To(ContainSubstring("Perform project operations"))
	})

	It("should display only the project name when running command with -q flag", func() {
		projectName := helper.Cmd("odo", "project", "get", "-q").ShouldPass().Out()
		Expect(projectName).Should(Equal(commonVar.Project))
	})

	It("should list current empty project in json format", func() {
		helper.WaitForCmdOut("odo", []string{"project", "list", "-o", "json"}, 5, true, func(output string) bool {
			return strings.Contains(output, commonVar.Project)
		})
		projectListJSON := helper.Cmd("odo", "project", "list", "-o", "json").ShouldPass().Out()
		valuesJSON := gjson.GetMany(projectListJSON, "kind")
		expectedJSON := []string{"List"}
		Expect(helper.GjsonMatcher(valuesJSON, expectedJSON)).To(Equal(true))

		items := gjson.Get(projectListJSON, "items").Array()
		found := false
		for _, item := range items {
			kind := item.Get("kind").String()
			name := item.Get("metadata.name").String()
			active := item.Get("status.active").String()
			if kind == "Project" && name == commonVar.Project && active == "true" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	})

	It("should list current empty project", func() {
		helper.WaitForCmdOut("odo", []string{"project", "list"}, 1, true, func(output string) bool {
			return strings.Contains(output, commonVar.Project)
		})
	})

	When("creating a new project", func() {
		var projectName string

		BeforeEach(func() {
			projectName = "cmd-project-" + helper.RandString(6)
			helper.Cmd("odo", "project", "create", projectName).ShouldPass()
		})

		It("should delete a project with --wait", func() {
			output := helper.Cmd("odo", "project", "delete", projectName, "-f", "--wait").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Waiting for project to be deleted"))
		})
	})

	When("user is logged out", func() {
		var token string
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("Openshift specific scenario.")
			}
			ocRunner := helper.NewOcRunner("oc")
			token = ocRunner.GetToken()
			helper.Cmd("odo", "logout").ShouldPass()
		})
		AfterEach(func() {
			helper.Cmd("odo", "login", "--token", token).ShouldPass()
		})
		It("should show login message when setting project and not login", func() {
			err := helper.Cmd("odo", "project", "set", "something").ShouldFail().Err()
			Expect(err).To(ContainSubstring("Unauthorized to access the cluster"))
		})
		It("should show login message when deleting project and not login", func() {
			err := helper.Cmd("odo", "project", "delete", "something").ShouldFail().Err()
			Expect(err).To(ContainSubstring("Unauthorized to access the cluster"))
		})
	})

	When("creating a new project with -o json", func() {
		var projectName string
		var output string

		BeforeEach(func() {
			projectName = "cmd-project-" + helper.RandString(6)
			output = helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldPass().Out()
		})

		AfterEach(func() {
			helper.Cmd("odo", "project", "delete", "-f", projectName)
		})

		It("should display information of created project", func() {
			values := gjson.GetMany(output, "kind", "metadata.name", "status.active")
			expected := []string{"Project", projectName, "true"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})

		It("should delete project and show output in json format", func() {
			actual := helper.Cmd("odo", "project", "delete", projectName, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(actual, "kind", "message")
			expected := []string{"Status", "Deleted project :"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})
	})
})
