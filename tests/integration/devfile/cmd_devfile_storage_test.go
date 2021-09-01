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
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When devfile storage create command is executed", func() {

		It("should create the storage and mount it on the container", func() {
			helper.Cmd("odo", "create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/" + storageNames[1]}
			sizes := []string{"5Gi", "1Gi"}

			helper.Cmd("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context).ShouldPass()
			// check storage create without the path name
			helper.Cmd("odo", "storage", "create", storageNames[1], "--size", sizes[1], "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

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
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageName := helper.RandString(5)
			pathName := "/data1"
			size := "1Gi"
			helper.Cmd("odo", "storage", "create", storageName, "--path", pathName, "--context", commonVar.Context, "--container", "tools", "--size", size).ShouldPass()
			storageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(storageList, []string{pathName, "tools", storageName, size})
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			storageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
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

			helper.Cmd("odo", "storage", "delete", "-f", "--context", commonVar.Context, storageName).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			storageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.DontMatchAllInOutput(storageList, []string{pathName, "tools", storageName, size})

			storageName2 := helper.RandString(5)
			helper.Cmd("odo", "storage", "create", storageName2, "--path", pathName, "--context", commonVar.Context, "--container", "runtime", "--size", size).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "delete", "-f", "--context", commonVar.Context, storageName2).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		})

		It("should create a storage with default size when --size is not provided", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageName := helper.RandString(5)

			helper.Cmd("odo", "storage", "create", storageName, "--path", "/data", "--context", commonVar.Context).ShouldPass()

			args = []string{"push", "--context", commonVar.Context}
			helper.Cmd("odo", args...).ShouldPass()

			// Verify the pvc size
			storageSize := commonVar.CliRunner.GetPVCSize(cmpName, storageName, commonVar.Project)
			Expect(storageSize).To(ContainSubstring("1Gi"))
		})

		It("should create a storage when storage is not provided", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "storage", "create", "--path", "/data", "--context", commonVar.Context).ShouldPass()

			args = []string{"push", "--context", commonVar.Context}
			helper.Cmd("odo", args...).ShouldPass()

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
			Expect(len(PVCs)).To(Equal(1))
		})

		It("should create and output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			actualJSONStorage := helper.Cmd("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(actualJSONStorage, "kind", "metadata.name", "spec.size", "spec.path")
			expected := []string{"storage", "mystorage", "1Gi", "/opt/app-root/src/storage/"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))

		})
	})

	Context("When devfile storage list command is executed", func() {
		It("should list the storage with the proper states", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.Cmd("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context).ShouldPass()
			stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Not Pushed", cmpName})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.Cmd("odo", "storage", "create", storageNames[1], "--path", pathNames[1], "--size", sizes[1], "--context", commonVar.Context).ShouldPass()

			stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
			helper.MatchAllInOutput(stdOut, []string{storageNames[1], pathNames[1], sizes[1], "Not Pushed"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.Cmd("odo", "storage", "delete", storageNames[0], "-f", "--context", commonVar.Context).ShouldPass()
			stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Locally Deleted"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()

			// since we don't have `wait` for `odo delete` at this moment
			// we need to wait for the pod to be in the terminating state or it has been deleted from the cluster
			commonVar.CliRunner.WaitAndCheckForTerminatingState("pods", commonVar.Project, 1)

			stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()

			helper.MatchAllInOutput(stdOut, []string{"Not Pushed"})
			// since `Pushed` is a sub string of `Not Pushed`, we count the occurrence of `Pushed`
			count := strings.Count(stdOut, "Pushed")
			Expect(count).To(Equal(1))
		})

		It("should list the storage with the proper states and container names", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Not Pushed", "CONTAINER", "runtime", "runtime2"})

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Pushed", "CONTAINER", "runtime", "runtime2"})

			helper.Cmd("odo", "storage", "delete", "firstvol", "-f", "--context", commonVar.Context).ShouldPass()

			stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Pushed", "Locally Deleted", "CONTAINER", "runtime", "runtime2"})
		})

		It("should list output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context).ShouldPass()

			actualStorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
			valuesSL := gjson.GetMany(actualStorageList, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.size", "items.0.spec.containerName", "items.0.status")
			expectedSL := []string{"List", "storage", "mystorage", "1Gi", "runtime", "Not Pushed"}
			Expect(helper.GjsonMatcher(valuesSL, expectedSL)).To(Equal(true))

		})
	})

	Context("When devfile storage commands are invalid", func() {
		It("should error if same storage name is provided again", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageName := helper.RandString(5)
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.Cmd("odo", "storage", "create", storageName, "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", storageName, "--path", pathNames[1], "--size", sizes[1], "--context", commonVar.Context).ShouldFail()
		})

		It("should error if same path is provided again", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathName := "/data"
			sizes := []string{"5Gi", "1Gi"}

			helper.Cmd("odo", "storage", "create", storageNames[0], "--path", pathName, "--size", sizes[0], "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", storageNames[1], "--path", pathName, "--size", sizes[1], "--context", commonVar.Context).ShouldFail()
		})

		It("should throw error if no storage is present", func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "storage", "delete", helper.RandString(5), "--context", commonVar.Context, "-f").ShouldFail()
		})
	})

	Context("When ephemeral is set to true in preference.yaml", func() {
		It("should not create a pvc to store source code", func() {

			helper.Cmd("odo", "preference", "set", "ephemeral", "true").ShouldPass()

			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

			Expect(len(PVCs)).To(Equal(0))
			output := commonVar.CliRunner.GetVolumeNamesFromDeployment(cmpName, "app", commonVar.Project)
			found := false
			for key, value := range output {
				if key == "odo-projects" {
					if value == "emptyDir" {
						found = true
						break
					}
				}
			}
			Expect(found).To(BeTrue())
		})
	})

	Context("When ephemeral is set to false in preference.yaml", func() {
		It("should create a pvc to store source code", func() {

			helper.Cmd("odo", "preference", "set", "ephemeral", "false").ShouldPass()

			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

			Expect(len(PVCs)).To(Not(Equal(0)))

			output := commonVar.CliRunner.GetVolumeNamesFromDeployment(cmpName, "app", commonVar.Project)
			found := false
			for key, value := range output {
				if key == "odo-projects" {
					if len(PVCs) > 0 && value == PVCs[0] {
						found = true
						break
					}
				}
			}
			Expect(found).To(BeTrue())
			Expect(len(output)).To(Equal(2))

			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()

			// check if the owner reference is set on the source code PVC properly or not
			commonVar.CliRunner.WaitAndCheckForTerminatingState("pvc", commonVar.Project, 1)
		})
	})

	Context("When ephemeral is not set in preference.yaml", func() {
		It("should not create a pvc to store source code  (default is ephemeral=true)", func() {

			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project}
			helper.Cmd("odo", args...).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			helper.Cmd("odo", "preference", "view").ShouldPass()

			// Verify the pvc size
			PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

			Expect(len(PVCs)).To(Equal(0))

			output := commonVar.CliRunner.GetVolumeNamesFromDeployment(cmpName, "app", commonVar.Project)

			found := false
			for key, value := range output {
				if key == "odo-projects" {
					if value == "emptyDir" {
						found = true
						break
					}
				}
			}
			Expect(found).To(BeTrue())
		})
	})
})
