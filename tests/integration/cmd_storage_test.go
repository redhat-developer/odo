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

	Context("when running help for storage command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "storage", "-h").ShouldPass().Out()
			Expect(appHelp).To(ContainSubstring("Perform storage operations"))
		})
	})

	Context("when using storage command with -o json", func() {
		It("should create and list output in json format", func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "wildfly", "wildfly", "--app", "wildflyapp", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			actualJSONStorage := helper.Cmd("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()

			valuesStoreC := gjson.GetMany(actualJSONStorage, "kind", "metadata.name", "spec.size", "spec.path")
			expectedStoreC := []string{"storage", "mystorage", "1Gi", "/opt/app-root/src/storage/"}
			Expect(helper.GjsonMatcher(valuesStoreC, expectedStoreC)).To(Equal(true))

			actualStorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()

			valuesStoreL := gjson.GetMany(actualStorageList, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.size")
			expectedStoreL := []string{"List", "storage", "mystorage", "1Gi"}
			Expect(helper.GjsonMatcher(valuesStoreL, expectedStoreL)).To(Equal(true))

			helper.Cmd("odo", "storage", "delete", "mystorage", "--context", commonVar.Context, "-f").ShouldPass()
		})
	})

	Context("when running storage list command to check state", func() {
		It("should list storage with correct state", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", "nodejs", "--app", "nodeapp", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// create storage, list storage should have state "Not Pushed"
			helper.Cmd("odo", "storage", "create", "pv1", "--path=/tmp1", "--size=1Gi", "--context", commonVar.Context).ShouldPass()
			StorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(StorageList).To(ContainSubstring("Not Pushed"))

			// Push storage, list storage should have state "Pushed"
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			StorageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(StorageList).To(ContainSubstring("Pushed"))

			// Delete storage, list storage should have state "Locally Deleted"
			helper.Cmd("odo", "storage", "delete", "pv1", "-f", "--context", commonVar.Context).ShouldPass()
			StorageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(StorageList).To(ContainSubstring("Locally Deleted"))

		})
	})

})
