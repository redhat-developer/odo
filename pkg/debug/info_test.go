package debug

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/util"

	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeOdoDebugFileString creates a json string of a fake OdoDebugFile
func fakeOdoDebugFileString(typeMeta metav1.TypeMeta, processID int, projectName, appName, componentName string, remotePort, localPort int) (string, error) {
	file := Info{
		TypeMeta: typeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Namespace: projectName,
			Name:      componentName,
		},
		Spec: InfoSpec{
			App:            appName,
			DebugProcessID: processID,
			RemotePort:     remotePort,
			LocalPort:      localPort,
		},
	}

	data, err := json.Marshal(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Test_createDebugInfoFile(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()

	type args struct {
		defaultPortForwarder *DefaultPortForwarder
		portPair             string
		fs                   filesystem.Filesystem
	}
	tests := []struct {
		name             string
		args             args
		alreadyExistFile bool
		wantDebugInfo    Info
		wantErr          bool
	}{
		{
			name: "case 1: normal json write to the debug file",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					componentName: "nodejs-ex",
					appName:       "app",
					projectName:   "testing-1",
				},
				portPair: "5858:9001",
				fs:       fs,
			},
			wantDebugInfo: Info{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-ex",
					Namespace: "testing-1",
				},
				Spec: InfoSpec{
					DebugProcessID: os.Getpid(),
					App:            "app",
					RemotePort:     9001,
					LocalPort:      5858,
				},
			},
			alreadyExistFile: false,
			wantErr:          false,
		},
		{
			name: "case 2: overwrite the debug file",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					componentName: "nodejs-ex",
					appName:       "app",
					projectName:   "testing-1",
				},
				portPair: "5758:9004",
				fs:       fs,
			},
			wantDebugInfo: Info{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-ex",
					Namespace: "testing-1",
				},
				Spec: InfoSpec{
					DebugProcessID: os.Getpid(),
					App:            "app",
					RemotePort:     9004,
					LocalPort:      5758,
				},
			},
			alreadyExistFile: true,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			debugFilePath := GetDebugInfoFilePath(tt.args.defaultPortForwarder.componentName, tt.args.defaultPortForwarder.appName, tt.args.defaultPortForwarder.projectName)
			// create a already existing file
			if tt.alreadyExistFile {
				_, err := testingutil.MkFileWithContent(debugFilePath, "blah", fs)
				if err != nil {
					t.Errorf("error happened while writing, cause: %v", err)
				}
			}

			if err := createDebugInfoFile(tt.args.defaultPortForwarder, tt.args.portPair, tt.args.fs); (err != nil) != tt.wantErr {
				t.Errorf("createDebugInfoFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			readBytes, err := fs.ReadFile(debugFilePath)
			if err != nil {
				t.Errorf("error while reading file, cause: %v", err)
			}
			var odoDebugFileData Info
			err = json.Unmarshal(readBytes, &odoDebugFileData)
			if err != nil {
				t.Errorf("error occured while unmarshalling json, cause: %v", err)
			}

			if !reflect.DeepEqual(tt.wantDebugInfo, odoDebugFileData) {
				t.Errorf("odo debug info on file doesn't match, got: %v, want: %v", odoDebugFileData, tt.wantDebugInfo)
			}

			// clear the odo debug info file
			_ = fs.RemoveAll(debugFilePath)
		})
	}
}

func Test_getDebugInfo(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()

	type args struct {
		defaultPortForwarder *DefaultPortForwarder
		fs                   filesystem.Filesystem
	}
	tests := []struct {
		name               string
		args               args
		fileExists         bool
		debugPortListening bool
		readDebugFile      Info
		wantDebugFile      Info
		debugRunning       bool
	}{
		{
			name: "case 1: the debug file exists",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					appName:       "app",
					componentName: "nodejs-ex",
					projectName:   "testing-1",
				},
				fs: fs,
			},
			wantDebugFile: Info{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-ex",
					Namespace: "testing-1",
				},
				Spec: InfoSpec{
					DebugProcessID: os.Getpid(),
					App:            "app",
					RemotePort:     5858,
					LocalPort:      9001,
				},
			},
			readDebugFile: Info{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-ex",
					Namespace: "testing-1",
				},
				Spec: InfoSpec{
					DebugProcessID: os.Getpid(),
					App:            "app",
					RemotePort:     5858,
					LocalPort:      9001,
				},
			},
			debugPortListening: true,
			fileExists:         true,
			debugRunning:       true,
		},
		{
			name: "case 2: the debug file doesn't exists",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					appName:       "app",
					componentName: "nodejs-ex",
					projectName:   "testing-1",
				},
				fs: fs,
			},
			debugPortListening: true,
			wantDebugFile:      Info{},
			readDebugFile:      Info{},
			fileExists:         false,
			debugRunning:       false,
		},
		{
			name: "case 3: debug port not listening",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					appName:       "app",
					componentName: "nodejs-ex",
					projectName:   "testing-1",
				},
				fs: fs,
			},
			debugPortListening: false,
			wantDebugFile:      Info{},
			readDebugFile: Info{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-ex",
					Namespace: "testing-1",
				},
				Spec: InfoSpec{
					DebugProcessID: os.Getpid(),
					App:            "app",
					RemotePort:     5858,
					LocalPort:      9001,
				},
			},
			fileExists:   true,
			debugRunning: false,
		},
		{
			name: "case 4: the process is not running",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					appName:       "app",
					componentName: "nodejs-ex",
					projectName:   "testing-1",
				},
				fs: fs,
			},
			debugPortListening: true,
			wantDebugFile:      Info{},
			readDebugFile: Info{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-ex",
					Namespace: "testing-1",
				},
				Spec: InfoSpec{
					DebugProcessID: os.Getpid() + 818177979,
					App:            "app",
					RemotePort:     5858,
					LocalPort:      9001,
				},
			},
			fileExists:   true,
			debugRunning: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			freePort, err := util.HTTPGetFreePort()
			if err != nil {
				t.Errorf("error occured while getting a free port, cause: %v", err)
			}

			if tt.readDebugFile.Spec.LocalPort != 0 {
				tt.readDebugFile.Spec.LocalPort = freePort
			}

			if tt.wantDebugFile.Spec.LocalPort != 0 {
				tt.wantDebugFile.Spec.LocalPort = freePort
			}

			odoDebugFilePath := GetDebugInfoFilePath(tt.args.defaultPortForwarder.componentName, tt.args.defaultPortForwarder.appName, tt.args.defaultPortForwarder.projectName)
			if tt.fileExists {
				fakeString, err := fakeOdoDebugFileString(tt.readDebugFile.TypeMeta,
					tt.readDebugFile.Spec.DebugProcessID,
					tt.readDebugFile.ObjectMeta.Namespace,
					tt.readDebugFile.Spec.App,
					tt.readDebugFile.ObjectMeta.Name,
					tt.readDebugFile.Spec.RemotePort,
					tt.readDebugFile.Spec.LocalPort)

				if err != nil {
					t.Errorf("error occured while getting odo debug file string, cause: %v", err)
				}

				_, err = testingutil.MkFileWithContent(odoDebugFilePath, fakeString, fs)
				if err != nil {
					t.Errorf("error occured while writing to file, cause: %v", err)
				}
			}

			stopListenerChan := make(chan bool)
			listenerStarted := false
			if tt.debugPortListening {
				startListenerChan := make(chan bool)
				go func() {
					err := testingutil.FakePortListener(startListenerChan, stopListenerChan, tt.readDebugFile.Spec.LocalPort)
					if err != nil {
						// the fake listener failed, show error and close the channel
						t.Errorf("error while starting fake port listerner, cause: %v", err)
						close(startListenerChan)
					}
				}()
				// wait for the test server to start listening
				if <-startListenerChan {
					listenerStarted = true
				}
			}

			got, resultRunning := getInfo(tt.args.defaultPortForwarder, tt.args.fs)

			if !reflect.DeepEqual(got, tt.wantDebugFile) {
				t.Errorf("getDebugInfo() got = %v, want %v", got, tt.wantDebugFile)
			}
			if resultRunning != tt.debugRunning {
				t.Errorf("getDebugInfo() got1 = %v, want %v", resultRunning, tt.debugRunning)
			}

			// clear the odo debug info file
			_ = fs.RemoveAll(odoDebugFilePath)

			// close the listener
			if listenerStarted == true {
				stopListenerChan <- true
			}
			close(stopListenerChan)
		})
	}
}
