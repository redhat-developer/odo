package occlient

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
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

type OcCommand struct {
	args   []string
	data   *string
	format string
}

// runOcCommands executes oc
// args - command line arguments to be passed to oc ('-o json' is added by default if data is not nil)
// data - is a pointer to a string, if set than data is given to command to stdin ('-f -' is added to args as default)
func runOcComamnd(command *OcCommand) ([]byte, error) {

	ocpath, err := getOcBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ocpath, command.args...)

	// if data is not set assume that it is get command
	if len(command.format) > 0 {
		cmd.Args = append(cmd.Args, "-o", command.format)
	}
	if command.data != nil {
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
			_, err := io.WriteString(stdin, *command.data)
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

func GetCurrentProjectName() (string, error) {
	// We need to run `oc project` because it returns an error when project does
	// not exist, while `oc project -q` does not return an error, it simply
	// returns the project name
	_, err := runOcComamnd(&OcCommand{
		args: []string{"project"},
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to get current project info")
	}

	output, err := runOcComamnd(&OcCommand{
		args: []string{"project", "-q"},
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to get current project name")
	}

	return string(output), nil
}

func CreateNewProject(name string) error {
	_, err := runOcComamnd(&OcCommand{
		args: []string{"new-project", name},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create new project")
	}
	return nil
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
