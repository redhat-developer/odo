package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo docker devfile status command tests", func() {
	var context, currentWorkingDirectory, cmpName string

	dockerClient := helper.NewDockerRunner("docker")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Local devfile push requires experimental mode to be set and the pushtarget set to docker
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		// Stop all containers labeled with the component name
		label := "component=" + cmpName
		dockerClient.StopContainers(label)

		dockerClient.RemoveVolumesByComponent(cmpName)

		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Verify devfile status works", func() {

		It("Verify that odo component status correctly reports supervisord status", func() {

			helper.CmdShouldPass("odo", "create", "java-springboot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--context", context)

			session := helper.CmdRunner("odo", "component", "status", "-o", "json", "--context", context, "--follow")

			utils.WaitForOutputToContain("supervisordStatus", session)

			// return true if the supervisord status value is as expected, false otherwise
			Eventually(func() bool {
				stdoutContents := string(session.Out.Contents())

				entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
				if err != nil {
					return false
				}

				for _, entry := range entries {
					// We only want supervisord status
					supervisordEntry, success := entry.(*machineoutput.SupervisordStatus)
					if !success {
						continue
					}

					// The build container contains an entry called 'OCI' which we want to skip
					containsOCI := false
					for _, programStatus := range supervisordEntry.ProgramStatus {
						if programStatus.Program == "OCI" {
							containsOCI = true
							break
						}
					}
					if containsOCI {
						continue
					}

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

				}

				return true
			}, 180, 3).Should(Equal(true))

			// Kill the java process within the container, return true when complete
			Eventually(func() bool {

				containers := dockerClient.GetContainersByCompAlias(cmpName, "runtime", true)
				Expect(len(containers)).To(Equal(1))

				contents := dockerClient.ExecContainer(containers[0], "ps -ef")

				pids := []string{}
				for _, str := range strings.Split(contents, "\n") {

					if strings.Contains(str, "java -jar") {

						fields := strings.Fields(str)
						if len(fields) >= 2 {
							pids = append(pids, fields[1])
						}
					}
				}

				for _, pid := range pids {
					dockerClient.ExecContainer(containers[0], "kill -9 "+pid)
				}

				// Return true if we found and killed it
				return len(pids) == 1

			}, 180, 10).Should(Equal(true))

			// Wait for 'odo component status' to report that the programs are no longer RUNNING (EXITED or STOPPED)
			Eventually(func() bool {
				stdoutContents := string(session.Out.Contents())
				entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
				if err != nil {
					return false
				}

				supervisordStatusMostRecent := utils.GetMostRecentEventOfType(machineoutput.TypeSupervisordStatus, entries, false)

				if supervisordStatusMostRecent == nil {
					return false
				}

				supervisordStatus := supervisordStatusMostRecent.(*machineoutput.SupervisordStatus)

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

		It("Verify that odo component status correctly detects containers", func() {

			helper.CmdShouldPass("odo", "create", "java-springboot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--context", context)

			session := helper.CmdRunner("odo", "component", "status", "-o", "json", "--context", context, "--follow")

			desiredStatus := (*string)(nil)

			// Returns true if component status has reported the container has 'desiredStatus', false otherwise
			checkDesiredStatus := func() bool {

				stdoutContents := string(session.Out.Contents())
				entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
				if err != nil {
					return false
				}

				statusEntry := utils.GetMostRecentEventOfType(machineoutput.TypeContainerStatus, entries, false)

				if statusEntry == nil {
					return false
				}

				containerStatus := statusEntry.(*machineoutput.ContainerStatus)

				containers := dockerClient.GetContainersByCompAlias(cmpName, "runtime", false)
				if len(containers) != 1 {
					return false
				}

				if desiredStatus == nil {
					return false
				}

				// The runtime container should have status 'running'
				match := false
				for _, entry := range containerStatus.Status {
					if strings.Contains(entry.ID, containers[0]) && entry.Status == *desiredStatus {
						match = true
					}
				}

				return match
			}

			// Verify that odo component status reports the container is running
			runningStatus := "running"
			desiredStatus = &runningStatus
			Eventually(checkDesiredStatus, 180, 10).Should(Equal(true))

			// Stop the container, then verify that odo component status reports that the container is stopped
			label := "component=" + cmpName
			dockerClient.StopContainers(label)
			exitedStatus := "exited"
			desiredStatus = &exitedStatus
			Eventually(checkDesiredStatus, 180, 10).Should(Equal(true))

			utils.TerminateSession(session)

		}) // end It

	}) // End Context

	Context("Verify URL status is correctly reported", func() {

		It("Verify that odo component status detects the URL status", func() {

			helper.CmdShouldPass("odo", "create", "java-springboot", cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV1", "springboot", "devfile-init.yaml"), filepath.Join(context, "devfile.yaml"))

			urlParams := []string{"url", "create", "my-url", "--port", "8080"}

			helper.CmdShouldPass("odo", urlParams...)

			helper.CmdShouldPass("odo", "push", "--context", context)

			session := helper.CmdRunner("odo", "component", "status", "-o", "json", "--context", context, "--follow")

			// Return true if the urlReachable value was found with the expected values, false otherwise.
			Eventually(func() bool {

				stdoutContents := string(session.Out.Contents())

				entries, err := utils.ParseMachineEventJSONLines(stdoutContents)
				if err != nil {
					return false
				}

				// Verify url status is present and correct
				urlReachablEntryResult := utils.GetMostRecentEventOfType(machineoutput.TypeURLReachable, entries, false)
				if urlReachablEntryResult == nil {
					return false
				}

				urlReachableEntry := urlReachablEntryResult.(*machineoutput.URLReachable)

				if urlReachableEntry.Kind != "docker" {
					return false
				}

				if urlReachableEntry.Reachable != true {
					return false
				}

				if urlReachableEntry.Port <= 0 {
					return false
				}

				if urlReachableEntry.Secure == true {
					return false
				}

				return true
			}, 180, 10).Should(Equal(true))

			utils.TerminateSession(session)

		})

	}) // end context
})
