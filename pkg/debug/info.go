package debug

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// Info contains the information about the current Debug session
type Info struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              InfoSpec `json:"spec"`
}

type InfoSpec struct {
	App            string `json:"app,omitempty"`
	DebugProcessID int    `json:"debugProcessID"`
	RemotePort     int    `json:"remotePort"`
	LocalPort      int    `json:"localPort"`
}

// GetDebugInfoFilePath gets the file path of the debug info file
func GetDebugInfoFilePath(componentName, appName string, projectName string) string {
	tempDir := os.TempDir()
	debugFileSuffix := "odo-debug.json"
	var arr []string
	if appName == "" {
		arr = []string{projectName, componentName, debugFileSuffix}
	} else {
		arr = []string{projectName, appName, componentName, debugFileSuffix}
	}
	debugFileName := strings.Join(arr, "-")
	return filepath.Join(tempDir, debugFileName)
}

func CreateDebugInfoFile(f *DefaultPortForwarder, portPair string) error {
	return createDebugInfoFile(f, portPair, filesystem.DefaultFs{})
}

// createDebugInfoFile creates a file in the temp directory with information regarding the debugging session of a component
func createDebugInfoFile(f *DefaultPortForwarder, portPair string, fs filesystem.Filesystem) error {
	portPairs := strings.Split(portPair, ":")
	if len(portPairs) != 2 {
		return errors.New("port pair should be of the format localPort:RemotePort")
	}

	localPort, err := strconv.Atoi(portPairs[0])
	if err != nil {
		return errors.New("local port should be a int")
	}
	remotePort, err := strconv.Atoi(portPairs[1])
	if err != nil {
		return errors.New("remote port should be a int")
	}

	debugFile := Info{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OdoDebugInfo",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.componentName,
			Namespace: f.projectName,
		},
		Spec: InfoSpec{
			App:            f.appName,
			DebugProcessID: os.Getpid(),
			RemotePort:     remotePort,
			LocalPort:      localPort,
		},
	}
	odoDebugPathData, err := json.Marshal(debugFile)
	if err != nil {
		return errors.New("error marshalling json data")
	}

	// writes the data to the debug info file
	file, err := fs.OpenFile(GetDebugInfoFilePath(f.componentName, f.appName, f.projectName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close() // #nosec G307

	_, err = file.Write(odoDebugPathData)
	if err != nil {
		return err
	}
	return nil
}

// GetInfo gathers the information with regards to debugging information
func GetInfo(f *DefaultPortForwarder) (Info, bool) {
	return getInfo(f, filesystem.DefaultFs{})
}

// getInfo gets information regarding the debugging session of the component
// returns the OdoDebugFile from the debug info file
// returns true if debugging is running else false
func getInfo(f *DefaultPortForwarder, fs filesystem.Filesystem) (Info, bool) {
	// gets the debug info file path and reads/unmarshalls it
	debugInfoFilePath := GetDebugInfoFilePath(f.componentName, f.appName, f.projectName)
	readFile, err := fs.ReadFile(debugInfoFilePath)
	if err != nil {
		klog.V(4).Infof("the debug %v is not present", debugInfoFilePath)
		return Info{}, false
	}

	var info Info
	err = json.Unmarshal(readFile, &info)
	if err != nil {
		klog.V(4).Infof("couldn't unmarshal the debug file %v", debugInfoFilePath)
		return Info{}, false
	}

	// get the debug process id and send a signal 0 to check if it's alive or not
	// according to https://golang.org/pkg/os/#FindProcess
	// On Unix systems, FindProcess always succeeds and returns a Process for the given pid, regardless of whether the process exists.
	// thus this step will pass on Unix systems and so for those systems and some others supporting signals
	// we check if the process is alive or not by sending a signal 0 to the process
	processInfo, err := os.FindProcess(info.Spec.DebugProcessID)
	if err != nil || processInfo == nil {
		klog.V(4).Infof("error getting the process info for pid %v", info.Spec.DebugProcessID)
		return Info{}, false
	}

	// signal is not available on windows so we skip this step for windows
	if runtime.GOOS != "windows" {
		err = processInfo.Signal(syscall.Signal(0))
		if err != nil {
			klog.V(4).Infof("error sending signal 0 to pid %v, cause: %v", info.Spec.DebugProcessID, err)
			return Info{}, false
		}
	}

	// gets the debug local port and tries to listen on it
	// if error doesn't occur the debug port was free and thus no debug process was using the port
	addressLook := "localhost:" + strconv.Itoa(info.Spec.LocalPort)
	listener, err := net.Listen("tcp", addressLook)
	if err == nil {
		klog.V(4).Infof("the debug port %v is free, thus debug is not running", info.Spec.LocalPort)
		err = listener.Close()
		if err != nil {
			klog.V(4).Infof("error occurred while closing the listener, cause :%v", err)
		}
		return Info{}, false
	}

	return info, true
}
