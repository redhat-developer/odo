package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo storage command tests", func() {
	var oc helper.OcRunner
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("when running help for storage command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "storage", "-h")
			Expect(appHelp).To(ContainSubstring("Perform storage operations"))
		})
	})

	Context("when using storage command with default flag values", func() {
		It("should add a storage, list and delete it", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "nodejs", "--app", "nodeapp", "--project", commonVar.Project, "--context", commonVar.Context)
			// Default flag value
			// --app string         Application, defaults to active application
			// --component string   Component, defaults to active component.
			// --project string     Project, defaults to active project
			storAdd := helper.CmdShouldPass("odo", "storage", "create", "pv1", "--path", "/mnt/pv1", "--size", "1Gi", "--context", commonVar.Context)
			Expect(storAdd).To(ContainSubstring("nodejs"))
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			dcName := oc.GetDcName("nodejs", commonVar.Project)

			// Check against the volume name against dc
			getDcVolumeMountName := oc.GetVolumeMountName(dcName, commonVar.Project)
			Expect(getDcVolumeMountName).To(ContainSubstring("pv1"))

			// Check if the storage is added on the path provided
			getMntPath := oc.GetVolumeMountPath(dcName, commonVar.Project)
			Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))

			storeList := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			Expect(storeList).To(ContainSubstring("pv1"))

			// delete the storage
			helper.CmdShouldPass("odo", "storage", "delete", "pv1", "--context", commonVar.Context, "-f")
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			storeList = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			Expect(storeList).NotTo(ContainSubstring("pv1"))

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			getDcVolumeMountName = oc.GetVolumeMountName(dcName, commonVar.Project)
			Expect(getDcVolumeMountName).NotTo(ContainSubstring("pv1"))
		})
	})

	Context("when using storage command with specified flag values", func() {
		It("should add a storage, list and delete it", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "python", "python", "--app", "pyapp", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			storAdd := helper.CmdShouldPass("odo", "storage", "create", "pv1", "--path", "/mnt/pv1", "--size", "1Gi", "--context", commonVar.Context)
			Expect(storAdd).To(ContainSubstring("python"))
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			dcName := oc.GetDcName("python", commonVar.Project)

			// Check against the volume name against dc
			getDcVolumeMountName := oc.GetVolumeMountName(dcName, commonVar.Project)
			Expect(getDcVolumeMountName).To(ContainSubstring("pv1"))

			// Check if the storage is added on the path provided
			getMntPath := oc.GetVolumeMountPath(dcName, commonVar.Project)
			Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))

			storeList := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			Expect(storeList).To(ContainSubstring("pv1"))

			// delete the storage
			helper.CmdShouldPass("odo", "storage", "delete", "pv1", "--context", commonVar.Context, "-f")
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			storeList = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)

			Expect(storeList).NotTo(ContainSubstring("pv1"))

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			getDcVolumeMountName = oc.GetVolumeMountName(dcName, commonVar.Project)
			Expect(getDcVolumeMountName).NotTo(ContainSubstring("pv1"))
		})
	})

	Context("when using storage command with -o json", func() {
		It("should create and list output in json format", func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "wildfly", "wildfly", "--app", "wildflyapp", "--project", commonVar.Project, "--context", commonVar.Context)
			actualJSONStorage := helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context, "-o", "json")

			valuesStoreC := gjson.GetMany(actualJSONStorage, "kind", "metadata.name", "spec.size", "spec.path")
			expectedStoreC := []string{"storage", "mystorage", "1Gi", "/opt/app-root/src/storage/"}
			Expect(helper.GjsonMatcher(valuesStoreC, expectedStoreC)).To(Equal(true))

			actualStorageList := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context, "-o", "json")

			valuesStoreL := gjson.GetMany(actualStorageList, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.size")
			expectedStoreL := []string{"List", "storage", "mystorage", "1Gi"}
			Expect(helper.GjsonMatcher(valuesStoreL, expectedStoreL)).To(Equal(true))

			helper.CmdShouldPass("odo", "storage", "delete", "mystorage", "--context", commonVar.Context, "-f")
		})
	})

	Context("when running storage list command to check state", func() {
		It("should list storage with correct state", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "nodejs", "--app", "nodeapp", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// create storage, list storage should have state "Not Pushed"
			helper.CmdShouldPass("odo", "storage", "create", "pv1", "--path=/tmp1", "--size=1Gi", "--context", commonVar.Context)
			StorageList := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			Expect(StorageList).To(ContainSubstring("Not Pushed"))

			// Push storage, list storage should have state "Pushed"
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			StorageList = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			Expect(StorageList).To(ContainSubstring("Pushed"))

			// Delete storage, list storage should have state "Locally Deleted"
			helper.CmdShouldPass("odo", "storage", "delete", "pv1", "-f", "--context", commonVar.Context)
			StorageList = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			Expect(StorageList).To(ContainSubstring("Locally Deleted"))

		})
	})

})
