package e2escenarios

import (
	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test Objective:
//    Test ODO devfile support features

// Scope:
//    Test debug support for the following components, making use of starter projects define in the corresponding devfile:
//    - nodejs
//    - java-springboot
//    - java-quarkus
//    - java-maven

var _ = Describe("odo devfile supported tests", func() {
	var componentName, projectDirPath string
	var projectDir = "/projectDir"
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		projectDirPath = commonVar.Context + projectDir
		helper.MakeDir(projectDirPath)
		helper.Chdir(projectDirPath)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	createStarterProjAndSetDebug := func(component, starter string) {
		workingDir := helper.Getwd()
		defer helper.Chdir(workingDir)
		helper.Chdir(projectDirPath)
		helper.Cmd("odo", "init", "--name", componentName, "--devfile", component, "--starter", starter).ShouldPass()
		session, _, _, _, err := helper.StartDevMode(nil)
		Expect(err).ToNot(HaveOccurred())
		session.Stop()
		session.WaitEnd()

		//		stopChannel := make(chan bool)
		//		go func() {
		//			helper.Cmd("odo", "debug", "port-forward", "--local-port", debugLocalPort, "--context", projectDirPath).WithTerminate(60*time.Second, stopChannel).ShouldRun()
		//		}()
		//
		//		// Make sure that the debug information output, outputs correctly.
		//		// We do *not* check the json output since the debugProcessID will be different each time.
		//		helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", projectDirPath}, 1, false, func(output string) bool {
		//			if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
		//				strings.Contains(output, fmt.Sprintf(`"localPort": %s`, debugLocalPort)) {
		//				return true
		//			}
		//			return false
		//		})
		//		stopChannel <- true
	}

	Context("odo debug support for devfile components", func() {
		It("Verify output debug information for nodeJS debug works", func() {
			createStarterProjAndSetDebug("nodejs", "nodejs-starter")
		})
		It("Verify output debug information for java-springboot works", func() {
			createStarterProjAndSetDebug("java-springboot", "springbootproject")
		})
		// enable it once https://github.com/redhat-developer/odo/issues/6181 is fixed
		// It("Verify output debug information for java-quarkus debug works", func() {
		// 	createStarterProjAndSetDebug("java-quarkus", "community")
		// })
		It("Verify output debug information for java-maven debug works", func() {
			createStarterProjAndSetDebug("java-maven", "springbootproject")
		})
	})

})
