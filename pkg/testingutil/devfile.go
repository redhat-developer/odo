package testingutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
)

// GetFakeContainerComponent returns a fake container component for testing
func GetFakeContainerComponent(name string, ports ...int) v1.Component {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"
	mountSources := true

	var endpoints []v1.Endpoint
	for _, p := range ports {
		endpoints = append(endpoints, v1.Endpoint{
			Name:       fmt.Sprintf("port-%d", p),
			TargetPort: p,
		})
	}

	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Container: &v1.ContainerComponent{
				Container: v1.Container{
					Image:       image,
					Env:         []v1.EnvVar{},
					MemoryLimit: memoryLimit,
					VolumeMounts: []v1.VolumeMount{{
						Name: volumeName,
						Path: volumePath,
					}},
					MountSources: &mountSources,
				},
				Endpoints: endpoints,
			}}}

}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name, size string) v1.Component {
	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Volume: &v1.VolumeComponent{
				Volume: v1.Volume{
					Size: size,
				}}}}

}

// GetTestDevfileObj returns a devfile object for testing
func GetTestDevfileObj(fs devfilefs.Filesystem) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	devfileData.SetMetadata(devfilepkg.DevfileMetadata{Name: "my-nodejs-app"})

	_ = devfileData.AddCommands([]v1.Command{
		{
			Id: "devbuild",
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{
					WorkingDir: "/projects/nodejs-starter",
				},
			},
		},
	})
	_ = devfileData.AddComponents([]v1.Component{
		{
			Name: "runtime",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image: "quay.io/nodejs-12",
					},
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-3030",
							TargetPort: 3000,
						},
					},
				},
			},
		},
		{
			Name: "loadbalancer",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image: "quay.io/nginx",
					},
				},
			},
		},
	})

	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}

// GetTestDevfileObjWithPreStopEvents returns a devfile object with preStop event.
// This function can further be extended to accept other type of events.
func GetTestDevfileObjWithPreStopEvents(fs devfilefs.Filesystem, preStopId, preStopCMD string) parser.DevfileObj {
	obj := GetTestDevfileObj(fs)
	_ = obj.Data.AddCommands([]v1.Command{
		{
			Id: preStopId,
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{
					CommandLine: preStopCMD,
					Component:   "runtime",
					WorkingDir:  "/projects/nodejs-starter",
				},
			},
		},
	})
	_ = obj.Data.AddEvents(v1.Events{
		DevWorkspaceEvents: v1.DevWorkspaceEvents{
			PreStop: []string{strings.ToLower(preStopId)},
		}})
	return obj
}

// GetTestDevfileObjFromFile takes the filename of devfile from tests/examples/source/devfiles/nodejs and returns a parser.DevfileObj
func GetTestDevfileObjFromFile(fileName string) parser.DevfileObj {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the devfile
	devfilePath := filepath.Join(filepath.Dir(filename), "..", "..", "tests", "examples", filepath.Join("source", "devfiles", "nodejs", fileName))

	devfileObj, _, err := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath})
	if err != nil {
		return parser.DevfileObj{}
	}
	return devfileObj
}
