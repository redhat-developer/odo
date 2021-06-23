package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo app command tests", func() {
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

	It("should display the help for app command", func() {
		appHelp := helper.Cmd("odo", "app", "-h").ShouldPass().Out()
		// Trimmed the end of the message string to make it compatible across clusters
		Expect(appHelp).To(ContainSubstring("Performs application operations related to"))
	})

	Context("on a fresh new project", func() {

		BeforeEach(func() {
			appList := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appList).To(ContainSubstring("There are no applications deployed"))
			actual := helper.Cmd("odo", "app", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			values := gjson.GetMany(actual, "kind", "metadata", "items")
			expected := []string{"List", "{}", "[]"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})

		It("should fail deleting non existing app", func() {
			appDelete := helper.Cmd("odo", "app", "delete", "test", "--project", commonVar.Project, "-f").ShouldFail().Err()
			Expect(appDelete).To(ContainSubstring("test app does not exists"))
		})

		It("should fail describing non existing app", func() {
			appDescribe := helper.Cmd("odo", "app", "describe", "test", "--project", commonVar.Project).ShouldFail().Err()
			Expect(appDescribe).To(ContainSubstring("test app does not exists"))
		})

	})

	When("creating a new component", func() {

		appName := "app"
		cmpName := "nodejs"

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
		})

		When("odo push is executed", func() {

			BeforeEach(func() {
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})

			It("should fail describing app without app parameter", func() {
				helper.Cmd("odo", "app", "describe", "--project", commonVar.Project).ShouldFail()
			})

			It("should fail deleting an app without app parameter", func() {
				helper.Cmd("odo", "app", "delete", "-f", "--project", commonVar.Project).ShouldFail()
			})

			It("should list apps", func() {
				appListOutput := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
				Expect(appListOutput).To(ContainSubstring(appName))
			})

			It("should list apps in JSON format", func() {
				actualCompListJSON := helper.Cmd("odo", "app", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
				valuesList := gjson.GetMany(actualCompListJSON, "kind", "items.#.metadata.name", "items.#.metadata.namespace")
				expectedList := []string{"List", "app", commonVar.Project}
				Expect(helper.GjsonMatcher(valuesList, expectedList)).To(Equal(true))
			})

			It("should describe specific app", func() {
				helper.Cmd("odo", "app", "describe", appName, "--project", commonVar.Project).ShouldPass()
			})

			It("should describe specific app in JSON format", func() {
				actualDesAppJSON := helper.Cmd("odo", "app", "describe", appName, "--project", commonVar.Project, "-o", "json").ShouldPass().Out()
				valuesDes := gjson.GetMany(actualDesAppJSON, "kind", "metadata.name", "metadata.namespace")
				expectedDes := []string{"Application", appName, commonVar.Project}
				Expect(helper.GjsonMatcher(valuesDes, expectedDes)).To(Equal(true))
			})

			When("changing to context directory", func() {

				BeforeEach(func() {
					helper.Chdir(commonVar.Context)
				})

				It("should list apps", func() {
					appListOutput := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
					Expect(appListOutput).To(ContainSubstring(appName))
				})

				It("should list apps in json format", func() {
					actualCompListJSON := helper.Cmd("odo", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
					valuesL := gjson.GetMany(actualCompListJSON, "kind", "devfileComponents.0.metadata.name", "devfileComponents.0.metadata.namespace")
					expectedL := []string{"List", "nodejs", commonVar.Project}
					Expect(helper.GjsonMatcher(valuesL, expectedL)).To(Equal(true))
				})

				It("should decsribe app", func() {
					helper.Cmd("odo", "app", "describe", "--project", commonVar.Project).ShouldPass()
					actualDesAppJSON := helper.Cmd("odo", "app", "describe", "app", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
					valuesDes := gjson.GetMany(actualDesAppJSON, "kind", "metadata.name", "metadata.namespace")
					expectedDes := []string{"Application", "app", commonVar.Project}
					Expect(helper.GjsonMatcher(valuesDes, expectedDes)).To(Equal(true))
				})
			})
		})

		When("adding storage", func() {

			mountPath := "/data"
			size := "1Gi"

			BeforeEach(func() {
				helper.Cmd("odo", "storage", "create", "storage-one", "--context", commonVar.Context, "--path", mountPath, "--size", size).ShouldPass()
			})

			When("odo push is executed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
				})

				It("should successfully execute describe", func() {
					helper.Cmd("odo", "app", "describe", appName, "--project", commonVar.Project).ShouldPass()
				})
			})
		})
	})
})
