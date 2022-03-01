package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/tests/helper"
	"github.com/redhat-developer/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile status command tests", func() {
	var namespace, context, cmpName string
	var commonVar helper.CommonVar
	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		SetDefaultEventuallyTimeout(5 * time.Minute)
		cmpName = helper.RandString(6)
		namespace = commonVar.Project
		context = commonVar.Context
		helper.Chdir(commonVar.Context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// Function used to test context: "Verify URL status is correctly reported"
	testCombo := func(ingress bool, secure bool, name string) {
		defer GinkgoRecover()
		It("Verify that odo component status detects the URL status: "+name, func() {
			openshift := os.Getenv("KUBERNETES") != "true"
			if !ingress && !openshift {
				Skip("Route-based URLs is an OpenShift only scenario")
			}
			urlHost := helper.RandString(12) + ".com"

			urlParams := []string{"url", "create", "my-url", "--port", "4000"}
			if secure {
				urlParams = append(urlParams, "--secure")
			}

			if ingress {
				urlParams = append(urlParams, "--ingress")
				urlParams = append(urlParams, "--host", urlHost)
			}

			helper.Cmd("odo", urlParams...).ShouldPass()

			helper.Cmd("odo", "push", "-o", "json", "--project", namespace).ShouldPass()

			session := helper.CmdRunner("odo", "component", "status", "-o", "json", "--project", namespace, "--follow")

			helper.WaitForOutputToContain("urlReachable", 180, 10, session)

			stdoutContents := string(session.Out.Contents())

			entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
			Expect(err).NotTo(HaveOccurred())

			// Verify url status is present and correct
			urlReachableEntry := utils.GetMostRecentEventOfType(machineoutput.TypeURLReachable, entries, true).(*machineoutput.URLReachable)

			expectedKind := "ingress"
			if !ingress || openshift {
				expectedKind = "route"
			}

			Expect(urlReachableEntry.Kind).To(Equal(expectedKind))
			Expect(urlReachableEntry.Reachable).To(Equal(!ingress || openshift)) // On non-openshift, the ingress URL is using a random hostname, so should not be resolveable
			Expect(urlReachableEntry.Port).To(Equal(3000))
			Expect(urlReachableEntry.Secure).To(Equal(secure))

			utils.TerminateSession(session)

		})
	}

	When("Creating nodejs component using devfile", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", namespace, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
		})
		testCombo(false, false, "Route Nonsecure")
		testCombo(true, false, "Ingress Nonsecure")
		// testCombo(false, true, "Route Secure")   # Commented until issue https://github.com/redhat-developer/odo/issues/5217 gets fixed
		// testCombo(true, true, "Ingress Secure")  # Commented until issue https://github.com/redhat-developer/odo/issues/5217 gets fixed

		When("doing odo push", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "-o", "json", "--project", namespace).ShouldPass()
			})
			It("Verify that odo component status correctly reports supervisord status", func() {

				session := helper.CmdRunner("odo", "component", "status", "-o", "json", "--project", namespace, "--follow")

				helper.WaitForOutputToContain("supervisordStatus", 180, 10, session)

				stdoutContents := string(session.Out.Contents())

				entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
				Expect(err).NotTo(HaveOccurred())

				// Verify supervisord status is present and correct
				supervisordEntry := utils.GetMostRecentEventOfType(machineoutput.TypeSupervisordStatus, entries, true).(*machineoutput.SupervisordStatus)
				count := 0
				for _, status := range supervisordEntry.ProgramStatus {
					if status.Program == string(common.DefaultDevfileRunCommand) {
						Expect(status.Status).To(Equal("RUNNING"))
					} else if status.Program == string(common.DefaultDevfileDebugCommand) {
						Expect(status.Status).To(Equal("STOPPED"))
					} else {
						Fail(fmt.Sprintf("Unexpected program: %v", status.Program))
					}
					count++
				}
				Expect(count).To(Equal(2))

				// Kill the node processes within the container, returns true when complete
				Eventually(func() bool {
					podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

					contents := helper.GetCliRunner().Exec(podName, namespace, "--", "ps", "-ef")

					pids := []string{}
					for _, str := range strings.Split(contents, "\n") {

						if strings.Contains(str, "node") || strings.Contains(str, "npm") {

							fields := strings.Fields(str)
							if len(fields) >= 2 {
								pids = append(pids, fields[1])
							}
						}
					}

					for _, pid := range pids {
						helper.GetCliRunner().Exec(podName, namespace, "--", "kill", "-9", pid)
					}

					// We expect (at least) 2 node processes
					return len(pids) >= 2

				}, 180, 10).Should(Equal(true))

				// Wait for 'odo component status' to report that the programs are no longer RUNNING (EXITED or STOPPED)
				Eventually(func() bool {
					stdoutContents := string(session.Out.Contents())
					entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
					if err != nil {
						return false
					}

					supervisordStatus := utils.GetMostRecentEventOfType(machineoutput.TypeSupervisordStatus, entries, false).(*machineoutput.SupervisordStatus)
					if supervisordStatus == nil {
						return false
					}

					// All programs should be stopped, because we killed the node processes
					for _, programStatus := range supervisordStatus.ProgramStatus {
						if programStatus.Status == "RUNNING" {
							return false
						}
					}

					return true
				}, 180, 10).Should(Equal(true))

				utils.TerminateSession(session)

			})

			It("Verify that odo component status correctly detects component Kubernetes pods", func() {

				session := helper.CmdRunner("odo", "component", "status", "-o", "json", "--project", namespace, "--follow")

				// Returns true if 'odo component status' correctly reported the status of the expected pod, false otherwise
				Eventually(func() bool {

					stdoutContents := string(session.Out.Contents())
					entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
					if err != nil {
						return false
					}

					statusEntry := utils.GetMostRecentEventOfType(machineoutput.TypeKubernetesPodStatus, entries, false)

					if statusEntry == nil {
						return false
					}

					podStatus := statusEntry.(*machineoutput.KubernetesPodStatus)

					// Check if a pod is running and correct, returns "" if success, otherwise returns the reason why not
					checkPod := func(pod machineoutput.KubernetesPodStatusEntry) string {
						if len(pod.StartTime) == 0 {
							return "StartTime is empty"
						}

						if pod.Phase != "Running" {
							return "Phase is not running"
						}

						match := false
						for _, labelValue := range pod.Labels {
							if labelValue == cmpName {
								match = true
								break
							}
						}
						if !match {
							return "No matching labels"
						}

						match = false
						for _, container := range pod.Containers {
							if container.Name == "runtime" && container.State.Running != nil {
								match = true
								break
							}
						}
						if !match {
							return "Could not find runtime container"
						}
						return ""
					} // end checkpod

					for _, pod := range podStatus.Pods {
						failReason := checkPod(pod)
						if failReason == "" {
							return true
						}
						fmt.Println("pod", pod.Name, "did not satisfy condition:", failReason)
					}

					return false
				}, 180, 10).Should(Equal(true))

				// Delete the old pod, so that we can confirm that we can find the new one
				oldPodName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
				cliRunner.DeletePod(oldPodName, namespace)

				// Returns true if we correctly found the new pod that was launched by k8s after we deleted the old pod, false otherwise
				Eventually(func() bool {

					stdoutContents := string(session.Out.Contents())
					entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
					if err != nil {
						return false
					}

					statusEntry := utils.GetMostRecentEventOfType(machineoutput.TypeKubernetesPodStatus, entries, false)

					if statusEntry == nil {
						return false
					}

					podStatus := statusEntry.(*machineoutput.KubernetesPodStatus)

					// Check if a pod is running and correct, and different from the old pod; returns "" if success, otherwise returns the reason why not
					checkPod := func(pod machineoutput.KubernetesPodStatusEntry) string {

						if pod.Name == oldPodName {
							return "Skipping old pod"
						}

						if pod.Phase != "Running" {
							return "Phase is not running"
						}

						match := false
						for _, labelValue := range pod.Labels {
							if labelValue == cmpName {
								match = true
								break
							}
						}
						if !match {
							return "No matching labels"
						}

						match = false
						for _, container := range pod.Containers {
							if container.Name == "runtime" && container.State.Running != nil {
								match = true
								break
							}
						}
						if !match {
							return "Could not find runtime container"
						}
						return ""
					} // end checkPod

					for _, pod := range podStatus.Pods {
						failReason := checkPod(pod)
						if failReason == "" {
							return true
						}
						fmt.Println("Pod", pod.Name, "did not satisfy condition:", failReason)
					}

					return false
				}, 180, 10).Should(Equal(true))

				utils.TerminateSession(session)

			}) // end It
		})
	})

})
