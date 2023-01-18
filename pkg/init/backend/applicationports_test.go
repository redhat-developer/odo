package backend

import (
	"bytes"
	"testing"
	
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	devfilefs "github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	
	"github.com/redhat-developer/odo/pkg/testingutil"
)

var fs = devfilefs.NewFakeFs()

func buildDevfileObjWithComponents(components ...v1.Component) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
	devfileData.SetMetadata(devfilepkg.DevfileMetadata{Name: "my-nodejs-app"})
	_ = devfileData.AddComponents(components)
	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}

func Test_handleApplicationPorts(t *testing.T) {
	type devfileProvider func() parser.DevfileObj
	type args struct {
		devfileObjProvider devfileProvider
		ports              []int
	}

	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantProvider devfileProvider
	}{
		{
			name: "no component, no ports to set",
			args: args{
				devfileObjProvider: func() parser.DevfileObj { return buildDevfileObjWithComponents() },
			},
			wantProvider: func() parser.DevfileObj { return buildDevfileObjWithComponents() },
		},
		{
			name: "multiple container components, no ports to set",
			args: args{
				devfileObjProvider: func() parser.DevfileObj {
					return buildDevfileObjWithComponents(
						testingutil.GetFakeContainerComponent("cont1", 8080, 8081, 8082),
						testingutil.GetFakeContainerComponent("cont2", 9080, 9081, 9082),
					)
				},
			},
			wantProvider: func() parser.DevfileObj {
				return buildDevfileObjWithComponents(
					testingutil.GetFakeContainerComponent("cont1", 8080, 8081, 8082),
					testingutil.GetFakeContainerComponent("cont2", 9080, 9081, 9082),
				)
			},
		},
		{
			name: "no container components",
			args: args{
				devfileObjProvider: func() parser.DevfileObj {
					return buildDevfileObjWithComponents(testingutil.GetFakeVolumeComponent("vol1", "1Gi"))
				},
				ports: []int{8888, 8889, 8890},
			},
			wantProvider: func() parser.DevfileObj {
				return buildDevfileObjWithComponents(testingutil.GetFakeVolumeComponent("vol1", "1Gi"))
			},
		},
		{
			name: "more than one container components",
			args: args{
				devfileObjProvider: func() parser.DevfileObj {
					return buildDevfileObjWithComponents(
						testingutil.GetFakeContainerComponent("cont1", 8080, 8081, 8082),
						testingutil.GetFakeContainerComponent("cont2", 9080, 9081, 9082),
						testingutil.GetFakeVolumeComponent("vol1", "1Gi"),
					)
				},
				ports: []int{8888, 8889, 8890},
			},
			wantProvider: func() parser.DevfileObj {
				return buildDevfileObjWithComponents(
					testingutil.GetFakeContainerComponent("cont1", 8080, 8081, 8082),
					testingutil.GetFakeContainerComponent("cont2", 9080, 9081, 9082),
					testingutil.GetFakeVolumeComponent("vol1", "1Gi"),
				)
			},
		},
		{
			name: "single container component with both application and debug ports",
			args: args{
				devfileObjProvider: func() parser.DevfileObj {
					contWithDebug := testingutil.GetFakeContainerComponent("cont1", 18080, 18081, 18082)
					contWithDebug.ComponentUnion.Container.Endpoints = append(contWithDebug.ComponentUnion.Container.Endpoints,
						v1.Endpoint{Name: "debug", TargetPort: 5005},
						v1.Endpoint{Name: "debug-another", TargetPort: 5858},
					)
					return buildDevfileObjWithComponents(
						contWithDebug,
						testingutil.GetFakeVolumeComponent("vol1", "1Gi"))
				},
				ports: []int{3000, 9000},
			},
			wantProvider: func() parser.DevfileObj {
				newCont := testingutil.GetFakeContainerComponent("cont1")
				newCont.ComponentUnion.Container.Endpoints = append(newCont.ComponentUnion.Container.Endpoints,
					v1.Endpoint{Name: "port-3000-tcp", TargetPort: 3000, Protocol: v1.TCPEndpointProtocol},
					v1.Endpoint{Name: "port-9000-tcp", TargetPort: 9000, Protocol: v1.TCPEndpointProtocol},
					v1.Endpoint{Name: "debug", TargetPort: 5005},
					v1.Endpoint{Name: "debug-another", TargetPort: 5858},
				)
				return buildDevfileObjWithComponents(
					newCont,
					testingutil.GetFakeVolumeComponent("vol1", "1Gi"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			got, err := handleApplicationPorts(&output, tt.args.devfileObjProvider(), tt.args.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleApplicationPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantProvider(), got,
				cmp.AllowUnexported(devfileCtx.DevfileCtx{}),
				cmpopts.IgnoreInterfaces(struct{ devfilefs.Filesystem }{})); diff != "" {
				t.Errorf("handleApplicationPorts() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
