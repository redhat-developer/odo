package integration

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo storage command tests", func() {
	var project string
	var context string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
		oc = helper.NewOcRunner("oc")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for storage command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "storage", "-h")
			Expect(appHelp).To(ContainSubstring("Perform storage operations"))
		})
	})

	Context("when running storage command without required flag(s)", func() {
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "nodejs", "--app", "nodeapp", "--project", project, "--context", context)
			stdErr := helper.CmdShouldFail("odo", "storage", "create", "pv1")
			Expect(stdErr).To(ContainSubstring("Required flag"))
			//helper.CmdShouldFail("odo", "storage", "create", "pv1", "-o", "json")
		})
	})

	Context("when using storage command with default flag values", func() {
		It("should add a storage, list and delete it", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)

			helper.CmdShouldPass("odo", "component", "create", "nodejs", "nodejs", "--app", "nodeapp", "--project", project, "--context", context)
			// Default flag value
			// --app string         Application, defaults to active application
			// --component string   Component, defaults to active component.
			// --project string     Project, defaults to active project
			storAdd := helper.CmdShouldPass("odo", "storage", "create", "pv1", "--path", "/mnt/pv1", "--size", "1Gi", "--context", context)
			Expect(storAdd).To(ContainSubstring("nodejs"))
			helper.CmdShouldPass("odo", "push", "--context", context)

			dcName := oc.GetDcName("nodejs", project)

			// Check against the volume name against dc
			getDcVolumeMountName := oc.GetVolumeMountName(dcName, project)
			Expect(getDcVolumeMountName).To(ContainSubstring("pv1"))

			// Check if the storage is added on the path provided
			getMntPath := oc.GetVolumeMountPath(dcName, project)
			Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))

			storeList := helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			Expect(storeList).To(ContainSubstring("pv1"))

			// delete the storage
			helper.CmdShouldPass("odo", "storage", "delete", "pv1", "--context", context, "-f")
			helper.CmdShouldPass("odo", "push", "--context", context)

			storeList = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			Expect(storeList).NotTo(ContainSubstring("pv1"))

			helper.CmdShouldPass("odo", "push", "--context", context)
			getDcVolumeMountName = oc.GetVolumeMountName(dcName, project)
			Expect(getDcVolumeMountName).NotTo(ContainSubstring("pv1"))
		})
	})

	Context("when using storage command with specified flag values", func() {
		It("should add a storage, list and delete it", func() {
			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.CmdShouldPass("odo", "component", "create", "python", "python", "--app", "pyapp", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			storAdd := helper.CmdShouldPass("odo", "storage", "create", "pv1", "--path", "/mnt/pv1", "--size", "1Gi", "--context", context)
			Expect(storAdd).To(ContainSubstring("python"))
			helper.CmdShouldPass("odo", "push", "--context", context)

			dcName := oc.GetDcName("python", project)

			// Check against the volume name against dc
			getDcVolumeMountName := oc.GetVolumeMountName(dcName, project)
			Expect(getDcVolumeMountName).To(ContainSubstring("pv1"))

			// Check if the storage is added on the path provided
			getMntPath := oc.GetVolumeMountPath(dcName, project)
			Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))

			storeList := helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			Expect(storeList).To(ContainSubstring("pv1"))

			// delete the storage
			helper.CmdShouldPass("odo", "storage", "delete", "pv1", "--context", context, "-f")
			helper.CmdShouldPass("odo", "push", "--context", context)

			storeList = helper.CmdShouldPass("odo", "storage", "list", "--context", context)

			Expect(storeList).NotTo(ContainSubstring("pv1"))

			helper.CmdShouldPass("odo", "push", "--context", context)
			getDcVolumeMountName = oc.GetVolumeMountName(dcName, project)
			Expect(getDcVolumeMountName).NotTo(ContainSubstring("pv1"))
		})
	})

	Context("when using storage command with -o json", func() {
		It("should create and list output in json format", func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), context)
			helper.CmdShouldPass("odo", "component", "create", "wildfly", "wildfly", "--app", "wildflyapp", "--project", project, "--context", context)
			actualJSONStorage := helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", context, "-o", "json")
			desiredJSONStorage := `{"kind":"storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi"},"status":{"path":"/opt/app-root/src/storage/"}}`
			Expect(desiredJSONStorage).Should(MatchJSON(actualJSONStorage))

			actualStorageList := helper.CmdShouldPass("odo", "storage", "list", "--context", context, "-o", "json")
			desiredStorageList := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi"},"status":{"path":"/opt/app-root/src/storage/"}, "state":"Not Pushed"}]}`
			Expect(desiredStorageList).Should(MatchJSON(actualStorageList))

			helper.CmdShouldPass("odo", "storage", "delete", "mystorage", "--context", context, "-f")
		})
	})

	Context("when running storage list command to check state", func() {
		It("should list storage with correct state", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "nodejs", "--app", "nodeapp", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// create storage, list storage should have state "Not Pushed"
			helper.CmdShouldPass("odo", "storage", "create", "pv1", "--path=/tmp1", "--size=1Gi", "--context", context)
			StorageList := helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			Expect(StorageList).To(ContainSubstring("Not Pushed"))

			// Push storage, list storage should have state "Pushed"
			helper.CmdShouldPass("odo", "push", "--context", context)
			StorageList = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			Expect(StorageList).To(ContainSubstring("Pushed"))

			// Delete storage, list storage should have state "Locally Deleted"
			helper.CmdShouldPass("odo", "storage", "delete", "pv1", "-f", "--context", context)
			StorageList = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			Expect(StorageList).To(ContainSubstring("Locally Deleted"))

		})
	})

})
