package remotecmd

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/kclient"
)

const (
	podName       = "my-pod"
	containerName = "my-container"
)

func Test_getRemoteProcessPID(t *testing.T) {
	devfileCmd := v1alpha2.Command{Id: "my-run"}
	cmd := []string{common.ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(devfileCmd))}
	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		want                 int
		wantErr              bool
	}{
		{
			name: "error returned at command execution",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "missing pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
		},
		{
			name: "unexpected number of lines in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stdout.Write([]byte("123\n"))
						_, err := stdout.Write([]byte("234"))
						return err
					})
			},
			wantErr: true,
		},
		{
			name: "invalid content in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("invalid-pid"))
						return err
					})
			},
			wantErr: true,
		},
		{
			name: "valid content in pid file with trailing spaces",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(" 123 "))
						return err
					})
			},
			want: 123,
		},
		{
			name: "valid content in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
			},
			want: 123,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			got, err := getRemoteProcessPID(kubeClient, devfileCmd, podName, containerName)
			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func Test_getProcessChildren(t *testing.T) {
	const ppid = 123

	cmd := []string{common.ShellExecutable, "-c", fmt.Sprintf("cat /proc/%[1]d/task/%[1]d/children || true", ppid)}
	for _, tt := range []struct {
		name                 string
		ppid                 int
		kubeClientCustomizer func(*kclient.MockClientInterface)
		want                 []int
		wantErr              bool
	}{
		{
			name:    "pid < 0",
			ppid:    -1,
			wantErr: true,
		},
		{
			name:    "pid = 0",
			wantErr: true,
		},
		{
			name: "error returned at command execution",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			ppid:    ppid,
			wantErr: true,
		},
		{
			name: "missing children file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
			ppid: ppid,
		},
		{
			name: "one child in children file without trailing space",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName),
					gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("987"))
						return err
					})
			},
			ppid: ppid,
			want: []int{987},
		},
		{
			name: "one child in children file with trailing space",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName),
					gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("987 "))
						return err
					})
			},
			ppid: ppid,
			want: []int{987},
		},
		{
			name: "multiple children in children file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName),
					gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(" 987 765 432 "))
						return err
					})
			},
			ppid: ppid,
			want: []int{987, 765, 432},
		},
		{
			name: "multiple children in children file (on many lines)",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName),
					gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stdout.Write([]byte(" 987 765 \n"))
						_, err := stdout.Write([]byte("432"))
						return err
					})
			},
			ppid: ppid,
			want: []int{987, 765, 432},
		},
		{
			name: "multiple children in children file, with non-integer pid",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(containerName), gomock.Eq(podName),
					gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("987 765 an-invalid-pid 432 321"))
						return err
					})
			},
			ppid:    ppid,
			wantErr: true,
			want:    []int{987, 765},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			got, err := getProcessChildren(tt.ppid, kubeClient, podName, containerName)
			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}

		})
	}
}
