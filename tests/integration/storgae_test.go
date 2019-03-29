package integration

import (
	"os"
	"path/filepath"
	"time"

	//. "github.com/Benjamintf1/unmarshalledmatchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoStorageE2e", func() {
	var project string
	var context string

	appName := "app"
	cmpName := "nodejs"

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		oc = helper.NewOcRunner("oc")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		os.RemoveAll(".odo")
	})

	Context("Add, list and delete storage", func() {
		helper.CopyExample(filepath.Join("source", "nodejs"), context)
		helper.Chdir(context)

		helper.CmdShouldPass("odo component create nodejs " + cmpName + " --app " + appName + " --project " + project)
		storAdd := helper.CmdShouldPass("odo storage create pv1 --path /mnt/pv1 --size 5Gi --context " + context)
		Expect(storAdd).To(ContainSubstring("nodejs"))
		helper.CmdShouldPass("odo push")

		helper.CmdShouldPass("oc get dc")
		dcName := helper.NewOcRunner("oc").GetDcName(cmpName, project)

		// Check against the volume name against dc
		getDcVolumeMountName := helper.NewOcRunner("oc").GetVolumeMountName(dcName)

		Expect(getDcVolumeMountName).To(ContainSubstring("pv1"))

		// Check if the storage is added on the path provided
		getMntPath := helper.NewOcRunner("oc").GetVolumeMountPath(dcName)
		Expect(getMntPath).To(ContainSubstring("/mnt/pv1"))

		storeList := helper.CmdShouldPass("odo storage list")
		Expect(storeList).To(ContainSubstring("pv1"))

		// delete the storage
		helper.CmdShouldPass("odo storage delete pv1 -f")

		storeList = helper.CmdShouldPass("odo storage list")
		Expect(storeList).NotTo(ContainSubstring("pv1"))
		helper.CmdShouldPass("odo push")

		getDcVolumeMountName = helper.NewOcRunner("oc").GetVolumeMountName(dcName)
		Expect(getDcVolumeMountName).NotTo(ContainSubstring("pv1"))

	})

})
