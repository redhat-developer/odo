package devfile

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo devfile storage command tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When devfile storage create command is executed", func() {

		It("should create the storage and mount it on the container", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/" + storageNames[1]}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context)
			// check storage create without the path name
			helper.CmdShouldPass("odo", "storage", "create", storageNames[1], "--size", sizes[1], "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			volumesMatched := 0

			// check the volume name and mount paths for the containers
			deploymentName, err := util.NamespaceKubernetesObject(cmpName, "app")
			Expect(err).To(BeNil())
			volNamesAndPaths := commonVar.CliRunner.GetVolumeMountNamesandPathsFromContainer(deploymentName, "runtime", commonVar.Project)
			volNamesAndPathsArr := strings.Fields(volNamesAndPaths)
			for _, volNamesAndPath := range volNamesAndPathsArr {
				volNamesAndPathArr := strings.Split(volNamesAndPath, ":")

				for i, storageName := range storageNames {
					if strings.Contains(volNamesAndPathArr[0], storageName) && volNamesAndPathArr[1] == pathNames[i] {
						volumesMatched++
					}
				}
			}

			Expect(volumesMatched).To(Equal(2))

			for i, storageName := range storageNames {
				// Verify the pvc size
				storageSize := commonVar.CliRunner.GetPVCSize(cmpName, storageName, commonVar.Project)
				Expect(storageSize).To(ContainSubstring(sizes[i]))
			}
		})

		It("should create storage and attach to specified container successfully and list it correctly", func() {
			args := []string{"create", "java-springboot", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageName := helper.RandString(5)
			pathName := "/data1"
			size := "1Gi"
			helper.CmdShouldPass("odo", "storage", "create", storageName, "--path", pathName, "--context", commonVar.Context, "--container", "tools", "--size", size)
			storageList := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(storageList, []string{pathName, "tools", storageName, size})
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			storageList = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(storageList, []string{pathName, "tools", storageName})

			// check the volume name and mount paths for the funtime container
			deploymentName, err := util.NamespaceKubernetesObject(cmpName, "app")
			Expect(err).To(BeNil())

			volumesMatched := 0
			volNamesAndPaths := commonVar.CliRunner.GetVolumeMountNamesandPathsFromContainer(deploymentName, "tools", commonVar.Project)
			volNamesAndPathsArr := strings.Fields(volNamesAndPaths)
			for _, volNamesAndPath := range volNamesAndPathsArr {
				volNamesAndPathArr := strings.Split(volNamesAndPath, ":")
				if strings.Contains(volNamesAndPathArr[0], storageName) && volNamesAndPathArr[1] == pathName {
					volumesMatched++
				}
			}
			Expect(volumesMatched).To(Equal(1))

			// check the volume name and mount path Not present in runtime container
			volumesMatched = 0
			volNamesAndPaths = commonVar.CliRunner.GetVolumeMountNamesandPathsFromContainer(deploymentName, "runtime", commonVar.Project)
			volNamesAndPathsArr = strings.Fields(volNamesAndPaths)
			for _, volNamesAndPath := range volNamesAndPathsArr {
				volNamesAndPathArr := strings.Split(volNamesAndPath, ":")
				if strings.Contains(volNamesAndPathArr[0], storageName) && volNamesAndPathArr[1] == pathName {
					volumesMatched++
				}
			}
			Expect(volumesMatched).To(Equal(0))

			helper.CmdShouldPass("odo", "storage", "delete", "-f", "--context", commonVar.Context, storageName)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			storageList = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.DontMatchAllInOutput(storageList, []string{pathName, "tools", storageName, size})

			storageName2 := helper.RandString(5)
			helper.CmdShouldPass("odo", "storage", "create", storageName2, "--path", pathName, "--context", commonVar.Context, "--container", "runtime", "--size", size)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "delete", "-f", "--context", commonVar.Context, storageName2)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
		})

		It("should create a storage with default size when --size is not provided", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageName := helper.RandString(5)

			helper.CmdShouldPass("odo", "storage", "create", storageName, "--path", "/data", "--context", commonVar.Context)

			args = []string{"push", "--context", commonVar.Context}
			helper.CmdShouldPass("odo", args...)

			// Verify the pvc size
			storageSize := commonVar.CliRunner.GetPVCSize(cmpName, storageName, commonVar.Project)
			Expect(storageSize).To(ContainSubstring("1Gi"))
		})

		It("should create a storage when storage is not provided", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "storage", "create", "--path", "/data", "--context", commonVar.Context)

			args = []string{"push", "--context", commonVar.Context}
			helper.CmdShouldPass("odo", args...)

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
			Expect(len(PVCs)).To(Equal(1))
		})

		It("should create and output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			actualJSONStorage := helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context, "-o", "json")
			values := gjson.GetMany(actualJSONStorage, "kind", "metadata.name", "spec.size", "spec.path")
			expected := []string{"storage", "mystorage", "1Gi", "/opt/app-root/src/storage/"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))

		})
	})

	Context("When devfile storage list command is executed", func() {
		It("should list the storage with the proper states", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context)
			stdOut := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Not Pushed", cmpName})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "storage", "create", storageNames[1], "--path", pathNames[1], "--size", sizes[1], "--context", commonVar.Context)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
			helper.MatchAllInOutput(stdOut, []string{storageNames[1], pathNames[1], sizes[1], "Not Pushed"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "storage", "delete", storageNames[0], "-f", "--context", commonVar.Context)
			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Locally Deleted"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "delete", "-f", "--context", commonVar.Context)

			// since we don't have `wait` for `odo delete` at this moment
			// we need to wait for the pod to be in the terminating state or it has been deleted from the cluster
			commonVar.CliRunner.WaitAndCheckForTerminatingState("pods", commonVar.Project, 1)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)

			helper.MatchAllInOutput(stdOut, []string{"Not Pushed"})
			// since `Pushed` is a sub string of `Not Pushed`, we count the occurrence of `Pushed`
			count := strings.Count(stdOut, "Pushed")
			Expect(count).To(Equal(1))
		})

		It("should list the storage with the proper states and container names", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdOut := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Not Pushed", "CONTAINER", "runtime", "runtime2"})

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Pushed", "CONTAINER", "runtime", "runtime2"})

			helper.CmdShouldPass("odo", "storage", "delete", "firstvol", "-f", "--context", commonVar.Context)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Pushed", "Locally Deleted", "CONTAINER", "runtime", "runtime2"})
		})

		It("should list output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context)

			actualStorageList := helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context, "-o", "json")
			valuesSL := gjson.GetMany(actualStorageList, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.size", "items.0.spec.containerName", "items.0.status")
			expectedSL := []string{"List", "storage", "mystorage", "1Gi", "runtime", "Not Pushed"}
			Expect(helper.GjsonMatcher(valuesSL, expectedSL)).To(Equal(true))

		})
	})

	Context("When devfile storage commands are invalid", func() {
		It("should error if same storage name is provided again", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageName := helper.RandString(5)
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageName, "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context)
			helper.CmdShouldFail("odo", "storage", "create", storageName, "--path", pathNames[1], "--size", sizes[1], "--context", commonVar.Context)
		})

		It("should error if same path is provided again", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathName := "/data"
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathName, "--size", sizes[0], "--context", commonVar.Context)
			helper.CmdShouldFail("odo", "storage", "create", storageNames[1], "--path", pathName, "--size", sizes[1], "--context", commonVar.Context)
		})

		It("should throw error if no storage is present", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldFail("odo", "storage", "delete", helper.RandString(5), "--context", commonVar.Context, "-f")
		})
	})

	Context("When ephemeral is set to true in preference.yaml", func() {
		It("should not create a pvc to store source code", func() {

			helper.CmdShouldPass("odo", "preference", "set", "ephemeral", "true")

			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

			Expect(len(PVCs)).To(Equal(0))
		})
	})

	Context("When ephemeral is set to false in preference.yaml", func() {
		It("should create a pvc to store source code", func() {

			helper.CmdShouldPass("odo", "preference", "set", "ephemeral", "false")

			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

			Expect(len(PVCs)).To(Not(Equal(0)))
		})
	})

	Context("When ephemeral is not set in preference.yaml", func() {
		It("should not create a pvc to store source code  (default is ephemeral=false)", func() {

			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			helper.CmdShouldPass("odo", "preference", "view")

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

			Expect(len(PVCs)).To(Equal(0))
		})
	})
})
