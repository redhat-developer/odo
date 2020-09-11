package devfile

import (
	"fmt"
	"github.com/openshift/odo/tests/helper"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile storage command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig string

	// Using program command according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)

		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("When devfile storage create command is executed", func() {

		It("should create the storage and mount it on the container", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", context)
			helper.CmdShouldPass("odo", "storage", "create", storageNames[1], "--path", pathNames[1], "--size", sizes[1], "--context", context)

			args = []string{"push", "--context", context}
			helper.CmdShouldPass("odo", args...)

			volumesMatched := 0

			// check the volume name and mount paths for the containers
			volNamesAndPaths := cliRunner.GetVolumeMountNamesandPathsFromContainer(cmpName, "runtime", namespace)
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
				storageSize := cliRunner.GetPVCSize(cmpName, storageName, namespace)
				Expect(storageSize).To(ContainSubstring(sizes[i]))
			}
		})

		It("should create a storage with default size when --size is not provided", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			storageName := helper.RandString(5)

			helper.CmdShouldPass("odo", "storage", "create", storageName, "--path", "/data", "--context", context)

			args = []string{"push", "--context", context}
			helper.CmdShouldPass("odo", args...)

			// Verify the pvc size
			storageSize := cliRunner.GetPVCSize(cmpName, storageName, namespace)
			Expect(storageSize).To(ContainSubstring("1Gi"))
		})

		It("should create a storage when storage is not provided", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "storage", "create", "--path", "/data", "--context", context)

			args = []string{"push", "--context", context}
			helper.CmdShouldPass("odo", args...)

			// Verify the pvc size
			PVCs := cliRunner.GetAllPVCNames(namespace)
			Expect(len(PVCs)).To(Equal(1))
		})

		It("should create and output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			actualJSONStorage := helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", context, "-o", "json")
			desiredJSONStorage := `{"kind":"storage","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi","path":"/opt/app-root/src/storage/"}}`
			Expect(desiredJSONStorage).Should(MatchJSON(actualJSONStorage))
		})
	})

	Context("When devfile storage delete command is executed", func() {
		It("should delete the storage and unmount it on the container", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", context)
			helper.CmdShouldPass("odo", "storage", "create", storageNames[1], "--path", pathNames[1], "--size", sizes[1], "--context", context)

			args = []string{"push", "--context", context}
			helper.CmdShouldPass("odo", args...)

			helper.CmdShouldPass("odo", "storage", "delete", storageNames[0], "-f", "--context", context)
			helper.CmdShouldPass("odo", "storage", "delete", storageNames[1], "-f", "--context", context)

			args = []string{"push", "--context", context}
			helper.CmdShouldPass("odo", args...)

			volumesMatched := 0

			// check the volume name and mount paths for the containers
			volNamesAndPaths := cliRunner.GetVolumeMountNamesandPathsFromContainer(cmpName, "runtime", namespace)
			volNamesAndPathsArr := strings.Fields(volNamesAndPaths)
			for _, volNamesAndPath := range volNamesAndPathsArr {
				volNamesAndPathArr := strings.Split(volNamesAndPath, ":")

				for i, storageName := range storageNames {
					if strings.Contains(volNamesAndPathArr[0], storageName) && volNamesAndPathArr[1] == pathNames[i] {
						volumesMatched++
					}
				}
			}

			Expect(volumesMatched).To(Equal(0))

			pvcNames := cliRunner.GetAllPVCNames(namespace)
			Expect(len(pvcNames)).To(Equal(0))
		})

		It("should delete the storage and output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "storage", "create", "store-1", "--path", "/path", "--size", "1Gi", "--context", context)

			actualJSONStorage := helper.CmdShouldPass("odo", "storage", "delete", "store-1", "-f", "--context", context, "-o", "json")
			desiredJSONStorage := fmt.Sprintf(`{"kind": "storage","apiVersion": "odo.dev/v1alpha1","metadata": {"name": "store-1","creationTimestamp": null},"message": "Deleted storage store-1 from %s"}`, cmpName)
			Expect(desiredJSONStorage).Should(MatchJSON(actualJSONStorage))
		})
	})

	Context("When devfile storage list command is executed", func() {
		It("should list the storage with the proper states", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", context)
			stdOut := helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Not Pushed", cmpName})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "push", "--context", context)
			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "storage", "create", storageNames[1], "--path", pathNames[1], "--size", sizes[1], "--context", context)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
			helper.MatchAllInOutput(stdOut, []string{storageNames[1], pathNames[1], sizes[1], "Not Pushed"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})

			helper.CmdShouldPass("odo", "storage", "delete", storageNames[0], "-f", "--context", context)
			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Locally Deleted"})
			helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})
		})

		It("should list the storage with the proper states and container names", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(context, "devfile.yaml"))

			stdOut := helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "Not Pushed", "CONTAINER", "runtime", "runtime2"})

			helper.CmdShouldPass("odo", "push", "--context", context)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "Pushed", "CONTAINER", "runtime", "runtime2"})

			helper.CmdShouldPass("odo", "storage", "delete", "firstvol", "-f", "--context", context)

			stdOut = helper.CmdShouldPass("odo", "storage", "list", "--context", context)
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "Pushed", "Locally Deleted", "CONTAINER", "runtime", "runtime2"})
		})

		It("should list output in json format", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", context)

			actualStorageList := helper.CmdShouldPass("odo", "storage", "list", "--context", context, "-o", "json")
			desiredStorageList := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"storage","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi","path":"/opt/app-root/src/storage/","containerName":"runtime"},"status":"Not Pushed"}]}`
			Expect(desiredStorageList).Should(MatchJSON(actualStorageList))
		})
	})

	Context("When devfile storage commands are invalid", func() {
		It("should error if same storage name is provided again", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			storageName := helper.RandString(5)
			pathNames := []string{"/data", "/data-1"}
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageName, "--path", pathNames[0], "--size", sizes[0], "--context", context)
			helper.CmdShouldFail("odo", "storage", "create", storageName, "--path", pathNames[1], "--size", sizes[1], "--context", context)
		})

		It("should error if same path is provided again", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			storageNames := []string{helper.RandString(5), helper.RandString(5)}
			pathName := "/data"
			sizes := []string{"5Gi", "1Gi"}

			helper.CmdShouldPass("odo", "storage", "create", storageNames[0], "--path", pathName, "--size", sizes[0], "--context", context)
			helper.CmdShouldFail("odo", "storage", "create", storageNames[1], "--path", pathName, "--size", sizes[1], "--context", context)
		})

		It("should throw error if no storage is present", func() {
			args := []string{"create", "nodejs", cmpName, "--context", context, "--project", namespace}
			helper.CmdShouldPass("odo", args...)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldFail("odo", "storage", "delete", helper.RandString(5), "--context", context, "-f")
		})
	})
})
