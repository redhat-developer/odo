package debug

import (
	"encoding/json"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"testing"
)

// fakeOdoDebugFileString creates a json string of a fake OdoDebugFile
func fakeOdoDebugFileString(typeMeta v1.TypeMeta, processId int, projectName, appName, componentName string, remotePort, localPort int) (string, error) {
	odoDebugFile := OdoDebugFile{
		TypeMeta:       typeMeta,
		DebugProcessId: processId,
		ProjectName:    projectName,
		AppName:        appName,
		ComponentName:  componentName,
		RemotePort:     remotePort,
		LocalPort:      localPort,
	}

	data, err := json.Marshal(odoDebugFile)
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
		wantDebugInfo    OdoDebugFile
		wantErr          bool
	}{
		{
			name: "case 1: normal json write to the debug file",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					componentName: "nodejs-ex",
					appName:       "app",
				},
				portPair: "5858:9001",
				fs:       fs,
			},
			wantDebugInfo: OdoDebugFile{
				TypeMeta: v1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				DebugProcessId: os.Getpid(),
				ProjectName:    "testing-1",
				AppName:        "app",
				ComponentName:  "nodejs-ex",
				RemotePort:     9001,
				LocalPort:      5858,
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
				},
				portPair: "5758:9004",
				fs:       fs,
			},
			wantDebugInfo: OdoDebugFile{
				TypeMeta: v1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				DebugProcessId: os.Getpid(),
				ProjectName:    "testing-1",
				AppName:        "app",
				ComponentName:  "nodejs-ex",
				RemotePort:     9004,
				LocalPort:      5758,
			},
			alreadyExistFile: true,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, _ := occlient.FakeNew()
			client.Namespace = "testing-1"
			tt.args.defaultPortForwarder.client = client

			debugFilePath := GetDebugInfoFilePath(client, tt.args.defaultPortForwarder.componentName, tt.args.defaultPortForwarder.appName)
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
			var odoDebugFileData OdoDebugFile
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
		readDebugFile      OdoDebugFile
		wantDebugFile      OdoDebugFile
		debugRunning       bool
	}{
		{
			name: "case 1: the debug file exists",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					appName:       "app",
					componentName: "nodejs-ex",
				},
				fs: fs,
			},
			wantDebugFile: OdoDebugFile{
				TypeMeta: v1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				DebugProcessId: os.Getpid(),
				ProjectName:    "testing-1",
				AppName:        "app",
				ComponentName:  "nodejs-ex",
				RemotePort:     5858,
				LocalPort:      9001,
			},
			readDebugFile: OdoDebugFile{
				TypeMeta: v1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				DebugProcessId: os.Getpid(),
				ProjectName:    "testing-1",
				AppName:        "app",
				ComponentName:  "nodejs-ex",
				RemotePort:     5858,
				LocalPort:      9001,
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
				},
				fs: fs,
			},
			debugPortListening: true,
			wantDebugFile:      OdoDebugFile{},
			readDebugFile:      OdoDebugFile{},
			fileExists:         false,
			debugRunning:       false,
		},
		{
			name: "case 3: debug port not listening",
			args: args{
				defaultPortForwarder: &DefaultPortForwarder{
					appName:       "app",
					componentName: "nodejs-ex",
				},
				fs: fs,
			},
			debugPortListening: false,
			wantDebugFile:      OdoDebugFile{},
			readDebugFile: OdoDebugFile{
				TypeMeta: v1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				DebugProcessId: os.Getpid(),
				ProjectName:    "testing-1",
				AppName:        "app",
				ComponentName:  "nodejs-ex",
				RemotePort:     5858,
				LocalPort:      9001,
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
				},
				fs: fs,
			},
			debugPortListening: true,
			wantDebugFile:      OdoDebugFile{},
			readDebugFile: OdoDebugFile{
				TypeMeta: v1.TypeMeta{
					Kind:       "OdoDebugInfo",
					APIVersion: "v1",
				},
				DebugProcessId: os.Getpid() + 818177979,
				ProjectName:    "testing-1",
				AppName:        "app",
				ComponentName:  "nodejs-ex",
				RemotePort:     5858,
				LocalPort:      9001,
			},
			fileExists:   true,
			debugRunning: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, _ := occlient.FakeNew()
			client.Namespace = "testing-1"
			tt.args.defaultPortForwarder.client = client

			odoDebugFilePath := GetDebugInfoFilePath(tt.args.defaultPortForwarder.client, tt.args.defaultPortForwarder.componentName, tt.args.defaultPortForwarder.appName)
			if tt.fileExists {
				fakeString, err := fakeOdoDebugFileString(tt.readDebugFile.TypeMeta,
					tt.readDebugFile.DebugProcessId,
					tt.readDebugFile.ProjectName,
					tt.readDebugFile.AppName,
					tt.readDebugFile.ComponentName,
					tt.readDebugFile.RemotePort,
					tt.readDebugFile.LocalPort)

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
					err := testingutil.FakePortListener(startListenerChan, stopListenerChan, tt.readDebugFile.LocalPort)
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

			got, resultRunning := getDebugInfo(tt.args.defaultPortForwarder, tt.args.fs)

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
