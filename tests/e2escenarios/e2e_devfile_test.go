package e2escenarios

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
)

// Test Objective:
//    Test ODO devfile support features

// Scope:
//    Test debug support for the following components, making use of starter projects define in the corresponding devfile:
//    - nodejs
//    - java-springboot
//    - java-quarkus
//    - java-openliberty
//    - java-maven, no starter project available at the time this script was developed, so skiping this component for now

var _ = Describe("odo devfile supported tests", func() {
	var componentName, projectDirPath string
	var projectDir = "/projectDir"
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		projectDirPath = commonVar.Context + projectDir
	})

	preSetup := func() {
		helper.MakeDir(projectDirPath)
		helper.Chdir(projectDirPath)
	}

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("odo debug on a nodeJS starterProject component", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		It("Verify output debug information for nodeJS debug works", func() {
			fmt.Println("********************************************************")
			fmt.Println("*   Test devfile debug support for nodeJS component    *")
			fmt.Println("********************************************************")
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter", "--project", commonVar.Project, componentName, "--context", projectDirPath)
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5858", "--context", projectDirPath)
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", projectDirPath}, 1, false, func(output string) bool {
				if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
					strings.Contains(output, `"localPort": 5858`) {
					return true
				}
				return false
			})

			stopChannel <- true
		})

	})
	Context("odo debug on a java-springboot starterProject component", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		It("Verify output debug information for nodeJS java-springboot works", func() {
			fmt.Println("*****************************************************************")
			fmt.Println("*   Test devfile debug support for java-springboot component    *")
			fmt.Println("*****************************************************************")
			helper.CmdShouldPass("odo", "create", "java-springboot", "--starter", "--project", commonVar.Project, componentName, "--context", projectDirPath)
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5858", "--context", projectDirPath)
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", projectDirPath}, 1, false, func(output string) bool {
				if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
					strings.Contains(output, `"localPort": 5858`) {
					return true
				}
				return false
			})

			stopChannel <- true
		})

	})

	Context("odo debug on a java-openliberty starterProject component", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		It("Verify output debug information for java-openliberty debug works", func() {
			fmt.Println("******************************************************************")
			fmt.Println("*   Test devfile debug support for java-openliberty component    *")
			fmt.Println("******************************************************************")
			helper.CmdShouldPass("odo", "create", "java-openliberty", "--starter", "--project", commonVar.Project, componentName, "--context", projectDirPath)
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5858", "--context", projectDirPath)
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", projectDirPath}, 1, false, func(output string) bool {
				if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
					strings.Contains(output, `"localPort": 5858`) {
					return true
				}
				return false
			})

			stopChannel <- true
		})

	})

	Context("odo debug on a java-quarkus starterProject component", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		It("Verify output debug information for java-quarkus debug works", func() {
			fmt.Println("**************************************************************")
			fmt.Println("*   Test devfile debug support for java-quarkus component    *")
			fmt.Println("**************************************************************")
			helper.CmdShouldPass("odo", "create", "java-quarkus", "--starter", "--project", commonVar.Project, componentName, "--context", projectDirPath)
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5858", "--context", projectDirPath)
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", projectDirPath}, 1, false, func(output string) bool {
				if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
					strings.Contains(output, `"localPort": 5858`) {
					return true
				}
				return false
			})

			stopChannel <- true
		})

	})

})
