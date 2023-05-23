package port

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/remotecmd"
)

const (
	podName       = "my-pod"
	containerName = "my-container"
)

var cmd = []string{
	remotecmd.ShellExecutable, "-c",
	"cat /proc/net/tcp /proc/net/udp /proc/net/tcp6 /proc/net/udp6 || true",
}

const aggregatedContentFromProcNetFiles = `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
0: 0100007F:4E21 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 2798686 1 0000000000000000 100 0 0 10 0
1: 690A0A0A:192B 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 24399 1 0000000000000000 100 0 0 10 0
2: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 28227 1 0000000000000000 100 0 0 10 0
6: 00000000:14EB 00000000:0000 0A 00000000:00000000 00:00000000 00000000   193        0 18937 1 0000000000000000 100 0 0 10 5
10: 690A0A0A:A0B6 85B2FA8E:01BB 06 00000000:00000000 03:00000C93 00000000     0        0 0 3 0000000000000000
11: 690A0A0A:DC86 6E882E34:01BB 01 00000000:00000000 02:00000418 00000000  1000        0 5992580 2 0000000000000000 29 4 30 10 -1
invalid_state: 690A0A0A:DC86 6E882E34:01BB ZZZ 00000000:00000000 02:00000418 00000000  1000        0 5992580 2 0000000000000000 29 4 30 10 -1
invalid_local_port: 690A0A0A:WXYZ 6E882E34:01BB 0A 00000000:00000000 02:00000418 00000000  1000        0 5992580 2 0000000000000000 29 4 30 10 -1

sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
299: 690A0A0A:87BE EE4BFA8E:01BB 01 00000000:00000000 00:00000000 00000000  1000        0 5879134 2 0000000000000000 0
3670: FB0000E0:14E9 00000000:0000 07 00000000:00000000 00:00000000 00000000  1000        0 5422041 2 0000000000000000 0
invalid_local_addr: 000ZZ000:14E9 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 5422041 2 0000000000000000 0

sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
0: 00000000000000000000000001000000:0277 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 30021 1 0000000000000000 100 0 0 10 0
1: 0000000000000000FFFF00000100007F:F76E 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 383053 1 0000000000000000 100 0 0 10 0
3: 00000000000000000000000000000000:1388 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 29332 1 0000000000000000 100 0 0 10 0
too_long_local_addr 00000000000000000000000001000000123456789010:0277 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 30021 1 0000000000000000 100 0 0 10 0

sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
623: 00000000000000000000000000000000:8902 00000000000000000000000000000000:0000 07 00000000:00000000 00:00000000 00000000  1000        0 5976457 2 0000000000000000 0
32077: 00000000000000000000000000000000:83E0 00000000000000000000000000000000:0000 07 00000000:00000000 00:00000000 00000000  1000        0 5915479 2 0000000000000000 0
invalid_local_addr_port 0000000000000000000000000100000:0277:123 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 30021 1 0000000000000000 100 0 0 10 0

`

func TestDetectRemotePortsBoundOnLoopback(t *testing.T) {
	inputPorts := []api.ForwardedPort{
		{ContainerPort: 20001},
		{ContainerPort: 6443},
		{ContainerPort: 22},
		{ContainerPort: 5355},
		{ContainerPort: 631},
		{ContainerPort: 63342},
		{ContainerPort: 5000},
		{ContainerPort: 8080},
		{ContainerPort: 5858},
	}
	type args struct {
		execClientCustomizer func(client *exec.MockClient)
		podName              string
		containerName        string
		ports                []api.ForwardedPort
	}
	tests := []struct {
		name    string
		args    args
		want    []api.ForwardedPort
		wantErr bool
	}{
		{
			name: "error while executing command",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(nil, nil, errors.New("some-err"))
				},
				podName:       podName,
				containerName: containerName,
				ports:         inputPorts,
			},
			wantErr: true,
		},
		{
			name: "no active connections",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(`
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
`, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
				ports:         inputPorts,
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "no input ports",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Times(0)
				},
				podName:       podName,
				containerName: containerName,
				ports:         nil,
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "with different connections",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
				ports:         inputPorts,
			},
			wantErr: false,
			want: []api.ForwardedPort{
				{ContainerPort: 20001},
				{ContainerPort: 631},
				{ContainerPort: 63342},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			execClient := exec.NewMockClient(ctrl)
			tt.args.execClientCustomizer(execClient)

			got, err := DetectRemotePortsBoundOnLoopback(context.Background(), execClient, tt.args.podName, tt.args.containerName, tt.args.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectRemotePortsBoundOnLoopback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("detectRemotePortsBoundOnLoopback() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetListeningConnections(t *testing.T) {
	type args struct {
		execClientCustomizer func(client *exec.MockClient)
		podName              string
		containerName        string
	}
	tests := []struct {
		name    string
		args    args
		want    []Connection
		wantErr bool
	}{
		{
			name: "error while executing command",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(nil, nil, errors.New("some-err"))
				},
				podName:       podName,
				containerName: containerName,
			},
			wantErr: true,
		},
		{
			name: "no active connections",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(`
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
`, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "with different connections",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
			},
			wantErr: false,
			want: []Connection{
				{LocalAddress: "127.0.0.1", LocalPort: 20001, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "10.10.10.105", LocalPort: 6443, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "0.0.0.0", LocalPort: 22, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "0.0.0.0", LocalPort: 5355, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:0000:0000:0001",
					LocalPort:     631,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "LISTEN",
				},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:FFFF:7F00:0001",
					LocalPort:     63342,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "LISTEN",
				},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:0000:0000:0000",
					LocalPort:     5000,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "LISTEN",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			execClient := exec.NewMockClient(ctrl)
			tt.args.execClientCustomizer(execClient)

			got, err := GetListeningConnections(context.Background(), execClient, tt.args.podName, tt.args.containerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getListeningConnections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getListeningConnections() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetConnections(t *testing.T) {
	type args struct {
		execClientCustomizer func(client *exec.MockClient)
		podName              string
		containerName        string
		statePredicate       func(state int) bool
	}
	tests := []struct {
		name    string
		args    args
		want    []Connection
		wantErr bool
	}{
		{
			name: "error while executing command",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(nil, nil, errors.New("some-err"))
				},
				podName:       podName,
				containerName: containerName,
			},
			wantErr: true,
		},
		{
			name: "no active connections",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(`
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
`, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "non-matching filter on state",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
				statePredicate: func(state int) bool {
					return stateToString(state) == "some unknown state"
				},
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "filter on state: ESTABLISHED",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
				statePredicate: func(state int) bool {
					return stateToString(state) == "ESTABLISHED"
				},
			},
			wantErr: false,
			want: []Connection{
				{LocalAddress: "10.10.10.105", LocalPort: 56454, RemoteAddress: "52.46.136.110", RemotePort: 443, State: "ESTABLISHED"},
				{LocalAddress: "10.10.10.105", LocalPort: 34750, RemoteAddress: "142.250.75.238", RemotePort: 443, State: "ESTABLISHED"},
			},
		},
		{
			name: "all connections",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil)
				},
				podName:       podName,
				containerName: containerName,
			},
			wantErr: false,
			want: []Connection{
				{LocalAddress: "127.0.0.1", LocalPort: 20001, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "10.10.10.105", LocalPort: 6443, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "0.0.0.0", LocalPort: 22, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "0.0.0.0", LocalPort: 5355, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "LISTEN"},
				{LocalAddress: "10.10.10.105", LocalPort: 41142, RemoteAddress: "142.250.178.133", RemotePort: 443, State: "TIME_WAIT"},
				{LocalAddress: "10.10.10.105", LocalPort: 56454, RemoteAddress: "52.46.136.110", RemotePort: 443, State: "ESTABLISHED"},
				{LocalAddress: "10.10.10.105", LocalPort: 34750, RemoteAddress: "142.250.75.238", RemotePort: 443, State: "ESTABLISHED"},
				{LocalAddress: "224.0.0.251", LocalPort: 5353, RemoteAddress: "0.0.0.0", RemotePort: 0, State: "CLOSE"},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:0000:0000:0001",
					LocalPort:     631,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "LISTEN",
				},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:FFFF:7F00:0001",
					LocalPort:     63342,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "LISTEN",
				},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:0000:0000:0000",
					LocalPort:     5000,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "LISTEN",
				},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:0000:0000:0000",
					LocalPort:     35074,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "CLOSE",
				},
				{
					LocalAddress:  "0000:0000:0000:0000:0000:0000:0000:0000",
					LocalPort:     33760,
					RemoteAddress: "0000:0000:0000:0000:0000:0000:0000:0000",
					RemotePort:    0,
					State:         "CLOSE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			execClient := exec.NewMockClient(ctrl)
			tt.args.execClientCustomizer(execClient)

			got, err := GetConnections(context.Background(), execClient, tt.args.podName, tt.args.containerName, tt.args.statePredicate)
			if (err != nil) != tt.wantErr {
				t.Errorf("getListeningConnections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getListeningConnections() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckAppPortsListening(t *testing.T) {
	type args struct {
		execClientCustomizer func(client *exec.MockClient)
		containerPortMapping map[string][]int
		timeout              time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "no container port mapping",
			wantErr: false,
		},
		{
			name:    "container with no ports",
			wantErr: false,
			args: args{
				containerPortMapping: map[string][]int{
					containerName:   {},
					"my-other-cont": {},
				},
				timeout: 5 * time.Second,
			},
		},
		{
			name: "error while checking for ports",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Any(), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(nil, []string{"an error"}, errors.New("some error")).AnyTimes()
				},
				containerPortMapping: map[string][]int{
					// all ports are opened, as decoded from aggregatedContentFromProcNetFiles
					containerName:   {20001, 6443, 22},
					"my-other-cont": {5355},
				},
				timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "too small timeout reached while checking for ports, even if they are all opened",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Any(), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(nil, []string{"an error"}, errors.New("some error")).AnyTimes()
				},
				containerPortMapping: map[string][]int{
					// all ports are opened, as decoded from aggregatedContentFromProcNetFiles
					containerName: {20001, 6443, 22},
					// all ports are opened, as decoded from aggregatedContentFromProcNetFiles
					"my-other-cont": {5355},
				},
				timeout: 1 * time.Millisecond,
			},
			wantErr: true,
		},
		{
			name: "at least one of the ports are not opened",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil).AnyTimes()
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq("my-other-cont"), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil).AnyTimes()
				},
				containerPortMapping: map[string][]int{
					// all ports are coming from aggregatedContentFromProcNetFiles
					containerName: {20001, 6443, 22},
					// port 5355 is coming from aggregatedContentFromProcNetFiles, but 12345 is intentionally not there
					"my-other-cont": {5355, 12345},
				},
				timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "all ports are opened in the container",
			args: args{
				execClientCustomizer: func(client *exec.MockClient) {
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq(containerName), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil).AnyTimes()
					client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Eq(cmd), gomock.Eq(podName), gomock.Eq("my-other-cont"), gomock.Eq(false), gomock.Nil(), gomock.Nil()).
						Return(strings.Split(aggregatedContentFromProcNetFiles, "\n"), nil, nil).AnyTimes()
				},
				containerPortMapping: map[string][]int{
					// all ports are opened, as decoded from aggregatedContentFromProcNetFiles
					containerName: {20001, 6443, 22},
					// all ports are opened, as decoded from aggregatedContentFromProcNetFiles
					"my-other-cont": {5355},
				},
				timeout: 6 * time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			execClient := exec.NewMockClient(ctrl)
			if tt.args.execClientCustomizer != nil {
				tt.args.execClientCustomizer(execClient)
			}

			gotErr := CheckAppPortsListening(context.Background(), execClient, podName, tt.args.containerPortMapping, tt.args.timeout)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CheckAppPortsListening() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}
