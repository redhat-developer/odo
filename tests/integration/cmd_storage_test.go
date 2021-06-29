package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo storage command tests", func() {
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

	It("should display the help for storage command", func() {
		appHelp := helper.Cmd("odo", "storage", "-h").ShouldPass().Out()
		Expect(appHelp).To(ContainSubstring("Perform storage operations"))
	})

	When("creating a new component", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "wildfly", "wildfly", "--app", "wildflyapp", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
		})

		When("creating storage", func() {

			actualJSONStorage := ""

			BeforeEach(func() {
				// create storage
				actualJSONStorage = helper.Cmd("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
			})

			AfterEach(func() {
				helper.Cmd("odo", "storage", "delete", "mystorage", "--context", commonVar.Context, "-f").ShouldPass()
			})

			It("should create", func() {
				valuesStoreC := gjson.GetMany(actualJSONStorage, "kind", "metadata.name", "spec.size", "spec.path")
				expectedStoreC := []string{"storage", "mystorage", "1Gi", "/opt/app-root/src/storage/"}
				Expect(helper.GjsonMatcher(valuesStoreC, expectedStoreC)).To(Equal(true))

			})

			When("listing storage in json", func() {

				actualStorageList := ""

				BeforeEach(func() {
					actualStorageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
				})

				It("should list output in json format", func() {
					valuesStoreL := gjson.GetMany(actualStorageList, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.size")
					expectedStoreL := []string{"List", "storage", "mystorage", "1Gi"}
					Expect(helper.GjsonMatcher(valuesStoreL, expectedStoreL)).To(Equal(true))

				})
			})
		})

	})

	When("component is created and pushed", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", "nodejs", "--app", "nodeapp", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		})

		When("creating storage", func() {
			BeforeEach(func() {
				// create storage
				helper.Cmd("odo", "storage", "create", "pv1", "--path=/tmp1", "--size=1Gi", "--context", commonVar.Context).ShouldPass()
			})

			It("should list storage as Not Pushed", func() {
				// list storage should have state "Not Pushed"
				StorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
				Expect(StorageList).To(ContainSubstring("Not Pushed"))
			})
			When("storage is pushed", func() {
				BeforeEach(func() {
					// Push storage, list storage should have state "Pushed"
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
				})
				It("should have state push", func() {
					// list storage should have state "Pushed"
					StorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
					Expect(StorageList).To(ContainSubstring("Pushed"))
				})

				When("storage is deleted", func() {
					BeforeEach(func() {
						// Delete storage
						helper.Cmd("odo", "storage", "delete", "pv1", "-f", "--context", commonVar.Context).ShouldPass()
					})
					It("should have state Locally Deleted", func() {
						// list storage should have state "Locally Deleted"
						StorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
						Expect(StorageList).To(ContainSubstring("Locally Deleted"))
					})
				})
			})
		})
	})
})
