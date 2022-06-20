package remotecmd

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/kclient"
)

const (
	_podName       = "my-pod"
	_containerName = "my-container"
)

func TestKubeExecProcessHandler_GetProcessInfoForCommand(t *testing.T) {
	cmdDef := CommandDefinition{Id: "my-run"}
	kill0CmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", p)}
	}
	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		pid                  int
		want                 RemoteProcessInfo
		wantErr              bool
	}{
		{
			name: "error returned when checking pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "stopped status if PID file missing",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    0,
				Status: Stopped,
			},
		},
		{
			name: "unknown status if negative value stored in PID file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("-1"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    -1,
				Status: Unknown,
			},
			wantErr: true,
		},
		{
			name: "stopped status if kill -0 command exit status is non-zero",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Stopped,
			},
		},
		{
			name: "running status if kill -0 command exit status is non-zero",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("0"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Running,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			k := NewKubeExecProcessHandler()
			got, err := k.GetProcessInfoForCommand(cmdDef, kubeClient, _podName, _containerName)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestKubeExecProcessHandler_StartProcessForCommand(t *testing.T) {
	execCmdWithoutWorkingDir := CommandDefinition{
		Id:      "my-exec-cmd",
		CmdLine: "echo Hello; sleep 300",
	}
	fullExecCmd := CommandDefinition{
		Id:         "my-exec-cmd",
		CmdLine:    "tail -f /path/to/a/file",
		WorkingDir: "/path/to/working/dir",
		EnvVars: []CommandEnvVar{
			{
				Key:   "ENV_VAR1",
				Value: "value1",
			},
			{
				Key:   "ENV_VAR2",
				Value: "value2",
			},
		},
	}
	for _, tt := range []struct {
		name                 string
		cmdDef               CommandDefinition
		kubeClientCustomizer func(*kclient.MockClientInterface)
		isCmdExpectedToRun   bool
		wantErr              bool
		expectedStatuses     []RemoteProcessStatus
	}{
		{
			name:   "command execution returned no error",
			cmdDef: execCmdWithoutWorkingDir,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c",
						fmt.Sprintf("echo $$ > %s &&   (%s) 1>>/proc/1/fd/1 2>>/proc/1/fd/2",
							getPidFileForCommand(execCmdWithoutWorkingDir), execCmdWithoutWorkingDir.CmdLine)}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("Hello"))
						return err
					})
			},
			isCmdExpectedToRun: true,
			expectedStatuses:   []RemoteProcessStatus{Starting, Stopped},
		},
		{
			name:   "command with all fields returned no error",
			cmdDef: fullExecCmd,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c",
						fmt.Sprintf("echo $$ > %s && cd %s && export ENV_VAR1='value1' ENV_VAR2='value2' && (%s) 1>>/proc/1/fd/1 2>>/proc/1/fd/2",
							getPidFileForCommand(fullExecCmd), fullExecCmd.WorkingDir, fullExecCmd.CmdLine)}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("error while running command"))
			},
			isCmdExpectedToRun: true,
			expectedStatuses:   []RemoteProcessStatus{Starting, Stopped},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			k := NewKubeExecProcessHandler()

			var wg sync.WaitGroup
			wg.Add(2) //number of invocations of outputHandler
			var statusesReported []RemoteProcessStatus
			err := k.StartProcessForCommand(tt.cmdDef, kubeClient, _podName, _containerName,
				func(status RemoteProcessStatus, stdout []string, stderr []string, err error) {
					defer wg.Done()
					statusesReported = append(statusesReported, status)
				})

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if tt.isCmdExpectedToRun && waitTimeout(&wg, 10*time.Second) {
				t.Errorf("timeout waiting for output handler to get called")
				return
			}

			if !reflect.DeepEqual(tt.expectedStatuses, statusesReported) {
				t.Errorf("expected %v, got %v", tt.expectedStatuses, statusesReported)
			}
		})
	}
}

func TestKubeExecProcessHandler_StopProcessForCommand(t *testing.T) {
	cmdDef := CommandDefinition{Id: "my-run"}
	retrieveChildrenCmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("cat /proc/%[1]d/task/%[1]d/children || true", p)}
	}
	killCmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("kill %d || true", p)}
	}
	kill0CmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", p)}
	}

	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		pid                  int
		wantErr              bool
	}{
		{
			name: "error returned when checking pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "nothing to do if PID file missing",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
		},
		{
			name: "error while determining process children",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "no process children killed if missing children file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stderr.Write([]byte("no such process"))
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
		},
		{
			name: "process children should get killed",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("987 765"))
						return err
					})
				for _, p := range []int{987, 765} {
					kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(p)),
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
					kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(p)),
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
							_, _ = stderr.Write([]byte("no such process"))
							_, err := stdout.Write([]byte("1"))
							return err
						})
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			kubeClient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
				gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("rm -f %s", getPidFileForCommand(cmdDef))}),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("an error which should be ignored"))
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			k := NewKubeExecProcessHandler()
			err := k.StopProcessForCommand(cmdDef, kubeClient, _podName, _containerName)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getProcessInfoFromPid(t *testing.T) {
	cmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", p)}
	}
	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		pid                  int
		want                 RemoteProcessInfo
		wantErr              bool
	}{
		{
			name:    "pid < 0",
			pid:     -1,
			wantErr: true,
			want: RemoteProcessInfo{
				Pid:    -1,
				Status: Unknown,
			},
		},
		{
			name: "pid == 0",
			want: RemoteProcessInfo{
				Status: Stopped,
			},
		},
		{
			name: "error when checking process status",
			pid:  123,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Unknown,
			},
		},
		{
			name: "non-integer content returned by kill command output",
			pid:  123,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("should-not-happen"))
						return err
					})
			},
			wantErr: true,
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Unknown,
			},
		},
		{
			name: "kill command returned non-zero exit status code",
			pid:  123,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Stopped,
			},
		},
		{
			name: "kill command returned 0 as exit status code",
			pid:  123,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("0"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Running,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			k := NewKubeExecProcessHandler()
			got, err := k.getProcessInfoFromPid(tt.pid, kubeClient, _podName, _containerName)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func Test_getRemoteProcessPID(t *testing.T) {
	cmdDef := CommandDefinition{Id: "my-run"}
	cmd := []string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}
	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		want                 int
		wantErr              bool
	}{
		{
			name: "error returned at command execution",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "missing pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
			},
			want: 123,
		},
		{
			name: "negative value in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("-1"))
						return err
					})
			},
			want: -1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			got, err := getRemoteProcessPID(kubeClient, cmdDef, _podName, _containerName)
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

	cmd := []string{ShellExecutable, "-c", fmt.Sprintf("cat /proc/%[1]d/task/%[1]d/children || true", ppid)}
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			ppid:    ppid,
			wantErr: true,
		},
		{
			name: "missing children file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Eq(_containerName), gomock.Eq(_podName),
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

			got, err := getProcessChildren(tt.ppid, kubeClient, _podName, _containerName)
			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}

		})
	}
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
