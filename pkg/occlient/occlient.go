package occlient

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// getOcBinary returns full path to oc binary
// first it looks for env variable KUBECTL_PLUGINS_CALLER (run as oc plugin)
// than looks for env variable OC_BIN (set manualy by user)
// at last it tries to find oc in default PATH
func getOcBinary() (string, error) {
	log.Debug("getOcBinary - searching for oc binary")

	var ocPath string

	envKubectlPluginCaller := os.Getenv("KUBECTL_PLUGINS_CALLER")
	envOcBin := os.Getenv("OC_BIN")

	log.Debugf("envKubectlPluginCaller = %s\n", envKubectlPluginCaller)
	log.Debugf("envOcBin = %s\n", envOcBin)

	if len(envKubectlPluginCaller) > 0 {
		log.Debug("using path from KUBECTL_PLUGINS_CALLER")
		ocPath = envKubectlPluginCaller
	} else if len(envOcBin) > 0 {
		log.Debug("using path from OC_BIN")
		ocPath = envOcBin
	} else {
		path, err := exec.LookPath("oc")
		if err != nil {
			log.Debug("oc binary not found in PATH")
			return "", err
		}
		log.Debug("using oc from PATH")
		ocPath = path
	}
	log.Debug("using oc from %s", ocPath)

	if _, err := os.Stat(ocPath); err != nil {
		return "", err
	}

	return ocPath, nil
}

// runOcCommands executes oc
// args - command line arguments to be passed to oc ('-o json' is added by default if data is not nil)
// data - is a pointer to a string, if set than data is given to command to stdin ('-f -' is added to args as default)
func runOcComamnd(args []string, data *string) ([]byte, error) {

	ocpath, err := getOcBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ocpath, args...)

	// if data is not set assume that it is get command
	if data == nil {
		cmd.Args = append(cmd.Args, "-o", "json")
	} else {
		// data is given, assume this is crate or apply command
		// that takes data from stdin
		cmd.Args = append(cmd.Args, "-f", "-")

		// Read from stdin
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		// Write to stdin
		go func() {
			defer stdin.Close()
			_, err := io.WriteString(stdin, *data)
			if err != nil {
				fmt.Printf("can't write to stdin %v\n", err)
			}
		}()
	}

	// Execute the actual command
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	log.Debugf("running oc command with arguments: %s\n", strings.Join(cmd.Args, " "))

	err = cmd.Run()
	if err != nil {
		outputMessage := ""
		if stdErr.Len() != 0 {
			outputMessage = stdErr.String()
		}
		if stdOut.Len() != 0 {
			outputMessage = fmt.Sprintf("\n%s", stdErr.String())
		}

		if outputMessage != "" {
			return nil, fmt.Errorf("failed to execute oc command\n %s", outputMessage)
		}
		return nil, err
	}

	if stdErr.Len() != 0 {
		return nil, fmt.Errorf("Error output:\n%s", stdErr.String())
	}

	return stdOut.Bytes(), nil

}

// // GetDeploymentConfig returns information about DeploymentConfig
// func (occlient *OcClient) GetDeploymentConfig(name string) (*ov1.DeploymentConfig, error) {
// 	args := []string{
// 		"get",
// 		"deploymentconfig",
// 		name,
// 	}

// 	output, err := occlient.runOcComamnd(args, nil)

// 	if err != nil {
// 		return nil, err
// 	}

// 	dc := &ov1.DeploymentConfig{}

// 	err = json.Unmarshal(output, dc)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return dc, nil
// }

// // CreateDeploymentConfig creates new DeploymentConfig
// func (occlient *OcClient) CreateDeploymentConfig(dc *ov1.DeploymentConfig) error {
// 	data, err := json.Marshal(dc)
// 	if err != nil {
// 		return err
// 	}

// 	args := []string{
// 		"create",
// 	}

// 	stringData := string(data[:])
// 	output, err := occlient.runOcComamnd(args, &stringData)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(string(output[:]))
// 	return nil
// }
