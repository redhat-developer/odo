package remotecmd

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/kclient"
)

func TestExecuteCommand(t *testing.T) {
	cmd := []string{"echo", "Hello"}

	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		wantErr              bool
	}{
		{
			name: "command returning an error",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stderr.Write([]byte("error running command"))
						return errors.New("some error")
					})
			},
			wantErr: true,
		},
		{
			name: "command not returning an error",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("Hello"))
						return err
					})
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			err := ExecuteCommand(cmd, kubeClient, _podName, _containerName, false, nil, nil)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteCommandAndGetOutput(t *testing.T) {
	for _, tt := range []struct {
		name                 string
		cmd                  []string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		wantErr              bool
		wantStdout           []string
		wantStderr           []string
	}{
		{
			name: "command returning an error",
			cmd:  []string{"unknown-cmd"},
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq([]string{"unknown-cmd"}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stderr.Write([]byte("error running command\nanother message"))
						return errors.New("some error")
					})
			},
			wantErr:    true,
			wantStderr: []string{"error running command", "another message"},
		},
		{
			name: "command not returning an error",
			cmd:  []string{"cat", "/path/to/my/file"},
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq([]string{"cat", "/path/to/my/file"}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stdout.Write([]byte("Hello World\n\n\n"))
						_, _ = stdout.Write([]byte("Lorem Ipsum Dolor Sit Amet\n"))
						_, _ = stderr.Write([]byte("some message written to stderr"))
						return nil
					})
			},
			wantStdout: []string{"Hello World", "", "", "Lorem Ipsum Dolor Sit Amet"},
			wantStderr: []string{"some message written to stderr"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			stdout, stderr, err := ExecuteCommandAndGetOutput(kubeClient, _podName, _containerName, false, tt.cmd...)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.wantStdout, stdout) {
				t.Errorf("expected %+q for stdout, got %+q", tt.wantStdout, stdout)
			}
			if !reflect.DeepEqual(tt.wantStderr, stderr) {
				t.Errorf("expected %+q for stderr, got %+q", tt.wantStderr, stderr)
			}
		})
	}
}
