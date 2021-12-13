package devfile

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/tests/helper"
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

	When("creating a nodejs component", func() {
		storageNames := []string{helper.RandString(5), helper.RandString(5)}
		pathNames := []string{"/data", "/" + storageNames[1]}
		sizes := []string{"5Gi", "1Gi"}

		BeforeEach(func() {
			helper.Cmd("odo", "create", cmpName, "--context", commonVar.Context, "--project", commonVar.Project, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})

		It("should throw error if no storage is present", func() {
			helper.Cmd("odo", "storage", "delete", helper.RandString(5), "--context", commonVar.Context, "-f").ShouldFail()
		})

		When("ephemeral is set to true in preference.yaml and doing odo push", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "preference", "set", "ephemeral", "true").ShouldPass()
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})

			It("should not create a pvc to store source code", func() {
				// Verify the pvc size
				PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

				Expect(len(PVCs)).To(Equal(0))
				output := commonVar.CliRunner.GetVolumeNamesFromDeployment(cmpName, "app", commonVar.Project)
				value, ok := output["odo-projects"]
				Expect(ok).To(BeTrue())
				Expect(value).Should(Equal(("emptyDir")))
			})
		})

		When("ephemeral is set to false in preference.yaml", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "preference", "set", "ephemeral", "false").ShouldPass()
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})
			It("should create a pvc to store source code", func() {
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

		When("storage create command is executed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "storage", "create", storageNames[0], "--path", pathNames[0], "--size", sizes[0], "--context", commonVar.Context).ShouldPass()
			})

			It("should error if same path or same storage is provided again", func() {
				By("same path is provided again", func() {
					helper.Cmd("odo", "storage", "create", storageNames[1], "--path", pathNames[0], "--size", sizes[1], "--context", commonVar.Context).ShouldFail()
				})
				By("same storage is provided again", func() {
					helper.Cmd("odo", "storage", "create", storageNames[0], "--path", pathNames[1], "--size", sizes[1], "--context", commonVar.Context).ShouldFail()
				})
			})

			It("should list output in json format", func() {
				actualStorageList := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
				valuesSL := gjson.GetMany(actualStorageList, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.size", "items.0.spec.path", "items.0.spec.containerName", "items.0.status")
				expectedSL := []string{"List", "Storage", storageNames[0], sizes[0], pathNames[0], "runtime", "Not Pushed"}
				Expect(helper.GjsonMatcher(valuesSL, expectedSL)).To(Equal(true))
			})

			It("should list storage in not pushed state", func() {
				stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Not Pushed", cmpName})
				helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})
			})

			When("odo push is executed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
				})

				It("should list storage in pushed state", func() {
					stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
					helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
					helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})
				})

				When("creating new storage", func() {
					BeforeEach(func() {
						helper.Cmd("odo", "storage", "create", storageNames[1], "--path", pathNames[1], "--size", sizes[1], "--context", commonVar.Context).ShouldPass()
					})

					It("should list storage in correct state", func() {
						stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
						helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Pushed"})
						helper.MatchAllInOutput(stdOut, []string{storageNames[1], pathNames[1], sizes[1], "Not Pushed"})
						helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})
					})
					When("deleting pushed storage", func() {
						BeforeEach(func() {
							helper.Cmd("odo", "storage", "delete", storageNames[0], "-f", "--context", commonVar.Context).ShouldPass()
						})

						It("should list it as locally deleted", func() {
							stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
							helper.MatchAllInOutput(stdOut, []string{storageNames[0], pathNames[0], sizes[0], "Locally Deleted"})
							helper.DontMatchAllInOutput(stdOut, []string{"CONTAINER", "runtime"})
						})

						When("doing odo push, odo delete -f", func() {
							BeforeEach(func() {
								helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
								helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
								// since we don't have `wait` for `odo delete` at this moment
								// we need to wait for the pod to be in the terminating state or it has been deleted from the cluster
								commonVar.CliRunner.WaitAndCheckForTerminatingState("pods", commonVar.Project, 1)
							})
							It("should list storage with correct state", func() {
								stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
								helper.MatchAllInOutput(stdOut, []string{"Not Pushed"})
								// since `Pushed` is a sub string of `Not Pushed`, we count the occurrence of `Pushed`
								count := strings.Count(stdOut, "Pushed")
								Expect(count).To(Equal(1))
							})
						})
					})

					When("pushing the new storage", func() {
						BeforeEach(func() {
							helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
						})
						It("should mount it on the container with correct path and size", func() {
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
					})
				})

				When("creating with output as json format", func() {
					var values []gjson.Result
					BeforeEach(func() {
						actualJSONStorage := helper.Cmd("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
						values = gjson.GetMany(actualJSONStorage, "kind", "metadata.name", "spec.size", "spec.path")
					})
					It("should create storage", func() {
						expected := []string{"Storage", "mystorage", "1Gi", "/opt/app-root/src/storage/"}
						Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
					})
					When("doing storage delete with output as json", func() {
						BeforeEach(func() {
							deleteJSONStorage := helper.Cmd("odo", "storage", "delete", "mystorage", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
							values = gjson.GetMany(deleteJSONStorage, "kind", "status", "message", "details.name", "details.kind")

						})
						It("should delete storage", func() {
							deleteExpected := []string{"Status", "Success", "Deleted storage", "mystorage", "Storage"}
							Expect(helper.GjsonMatcher(values, deleteExpected)).To(Equal(true))
						})
					})
				})
			})
		})

		When("ephemeral is not set in preference.yaml", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
				value := helper.GetPreferenceValue("Ephemeral")
				Expect(value).To(BeEmpty())
			})
			It("should not create a pvc to store source code  (default is ephemeral=true)", func() {
				helper.Cmd("odo", "preference", "view").ShouldPass().Out()

				// Verify the pvc size
				PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)

				Expect(len(PVCs)).To(Equal(0))

				output := commonVar.CliRunner.GetVolumeNamesFromDeployment(cmpName, "app", commonVar.Project)

				value, found := output["odo-projects"]
				Expect(found).To(BeTrue())
				Expect(value).Should(Equal("emptyDir"))
			})
		})

		When("creating storage  without --size and pushed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "storage", "create", storageNames[0], "--path", "/data", "--context", commonVar.Context).ShouldPass()
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})
			It("should create a storage with default size", func() {
				// Verify the pvc size
				storageSize := commonVar.CliRunner.GetPVCSize(cmpName, storageNames[0], commonVar.Project)
				Expect(storageSize).To(ContainSubstring("1Gi"))
			})
		})

		When("creating storage without storage name and pushed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "storage", "create", "--path", "/data", "--context", commonVar.Context).ShouldPass()
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})
			It("should create a storage", func() {
				// Verify the pvc size
				PVCs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
				Expect(len(PVCs)).To(Equal(1))
			})
		})
	})

	When("creating the storage with proper states and container names set in the devfile", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", cmpName, "--context", commonVar.Context, "--project", commonVar.Project, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

		})

		It("should list the storage with the proper states and container names", func() {
			stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Not Pushed", "CONTAINER", "runtime", "runtime2"})
		})

		When("doing odo push", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})

			It("should list the storage with the proper states and container names", func() {
				stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Pushed", "CONTAINER", "runtime", "runtime2"})
			})

			When("deleting storage push and doing storage list", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "storage", "delete", "firstvol", "-f", "--context", commonVar.Context).ShouldPass()
				})

				It("should list the storage with the proper states and container names", func() {
					stdOut := helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
					helper.MatchAllInOutput(stdOut, []string{"firstvol", "secondvol", "/secondvol", "/data", "/data2", "Pushed", "Locally Deleted", "CONTAINER", "runtime", "runtime2"})
				})
			})
		})
	})

	When("creating a springboot component", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", cmpName, "--context", commonVar.Context, "--project", commonVar.Project, "--devfile", helper.GetExamplePath("source", "devfiles", "springboot", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
		})
		When("creating storage ", func() {

			var storageList string
			storageName := helper.RandString(5)
			pathName := "/data1"
			size := "1Gi"

			BeforeEach(func() {
				helper.Cmd("odo", "storage", "create", storageName, "--path", pathName, "--context", commonVar.Context, "--container", "tools", "--size", size).ShouldPass()

			})
			It("should list the storage attached to the specified container", func() {
				storageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(storageList, []string{pathName, "tools", storageName, size})
			})
			When("doing odo push", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
				})
				It("should list the storage attached to the specified container", func() {
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
				})

				When("deleting storage and doing odo push", func() {

					BeforeEach(func() {
						helper.Cmd("odo", "storage", "delete", "-f", "--context", commonVar.Context, storageName).ShouldPass()
						helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
					})

					It("should list the deleted storage with odo list", func() {
						storageList = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
						helper.DontMatchAllInOutput(storageList, []string{pathName, "tools", storageName, size})
					})
					It("should be able to create and push storage at same path", func() {
						storageName2 := helper.RandString(5)
						helper.Cmd("odo", "storage", "create", storageName2, "--path", pathName, "--context", commonVar.Context, "--container", "runtime", "--size", size).ShouldPass()
						helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
					})
				})
			})
		})
	})
})
