package debug

import (
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

type OdoDebugFile struct {
	metav1.TypeMeta
	DebugProcessId int
	ProjectName    string
	AppName        string
	ComponentName  string
	RemotePort     int
	LocalPort      int
}

// GetDebugInfoFilePath gets the file path of the debug info file
func GetDebugInfoFilePath(client *occlient.Client, componentName, appName string) string {
	tempDir := os.TempDir()
	debugFileSuffix := "odo-debug.json"
	s := []string{client.Namespace, appName, componentName, debugFileSuffix}
	debugFileName := strings.Join(s, "-")
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

	odoDebugFile := OdoDebugFile{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OdoDebugInfo",
			APIVersion: "v1",
		},
		DebugProcessId: os.Getpid(),
		ProjectName:    f.client.Namespace,
		AppName:        f.appName,
		ComponentName:  f.componentName,
		RemotePort:     remotePort,
		LocalPort:      localPort,
	}
	odoDebugPathData, err := json.Marshal(odoDebugFile)
	if err != nil {
		return errors.New("error marshalling json data")
	}

	// writes the data to the debug info file
	file, err := fs.OpenFile(GetDebugInfoFilePath(f.client, f.componentName, f.appName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	_, err = file.Write(odoDebugPathData)
	if err != nil {
		return err
	}
	return nil
}

func GetDebugInfo(f *DefaultPortForwarder) (OdoDebugFile, bool) {
	return getDebugInfo(f, filesystem.DefaultFs{})
}

// getDebugInfo gets information regarding the debugging session of the component
// returns the OdoDebugFile from the debug info file
// returns true if debugging is running else false
func getDebugInfo(f *DefaultPortForwarder, fs filesystem.Filesystem) (OdoDebugFile, bool) {
	// gets the debug info file path and reads/unmarshals it
	debugInfoFilePath := GetDebugInfoFilePath(f.client, f.componentName, f.appName)
	readFile, err := fs.ReadFile(debugInfoFilePath)
	if err != nil {
		glog.V(4).Infof("the debug %v is not present", debugInfoFilePath)
		return OdoDebugFile{}, false
	}

	var odoDebugFileData OdoDebugFile
	err = json.Unmarshal(readFile, &odoDebugFileData)
	if err != nil {
		glog.V(4).Infof("couldn't unmarshal the debug file %v", debugInfoFilePath)
		return OdoDebugFile{}, false
	}

	// get the debug process id and send a signal 0 to check if it's alive or not
	// according to https://golang.org/pkg/os/#FindProcess
	// On Unix systems, FindProcess always succeeds and returns a Process for the given pid, regardless of whether the process exists.
	// thus this step will pass on Unix systems and so for those systems and some others supporting signals
	// we check if the process is alive or not by sending a signal 0 to the process
	processInfo, err := os.FindProcess(odoDebugFileData.DebugProcessId)
	if err != nil || processInfo == nil {
		glog.V(4).Infof("error getting the process info for pid %v", odoDebugFileData.DebugProcessId)
		return OdoDebugFile{}, false
	}

	// signal is not available on windows so we skip this step for windows
	if runtime.GOOS != "windows" {
		err = processInfo.Signal(syscall.Signal(0))
		if err != nil {
			glog.V(4).Infof("error sending signal 0 to pid %v, cause: %v", odoDebugFileData.DebugProcessId, err)
			return OdoDebugFile{}, false
		}
	}

	// gets the debug local port and tries to listen on it
	// if error doesn't occur the debug port was free and thus no debug process was using the port
	addressLook := "localhost:" + strconv.Itoa(odoDebugFileData.LocalPort)
	listener, err := net.Listen("tcp", addressLook)
	if err == nil {
		glog.V(4).Infof("the debug port %v is free, thus debug is not running", odoDebugFileData.LocalPort)
		err = listener.Close()
		if err != nil {
			glog.V(4).Infof("error occurred while closing the listener, cause :%v", err)
		}
		return OdoDebugFile{}, false
	}
	// returns the unmarshalled data
	return odoDebugFileData, true
}
