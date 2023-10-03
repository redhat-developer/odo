package remotecmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
)

const (
	_podName       = "my-pod"
	_containerName = "my-container"
	statFile       = `1 (tail) S 0 1 1 0 -1 1077952768 943 0 0 0 1 1 0 0 20 0 1 0 171838 5050368 338 18446744073709551615 94133334573056 94133335487553 140737112090992 0 0 0 0 0 0 0 0 0 17 1 0 0 0 0 0 94133335803888 94133335849476 94133343424512 140737112095206 140737112095282 140737112095282 140737112096746 0
118 (bash) S 0 118 118 34816 128 4210944 1144 454 0 0 0 1 0 0 20 0 1 0 185395 4554752 926 18446744073709551615 93924054794240 93924055688405 140724979077904 0 0 0 65536 3686404 1266761467 0 0 0 17 1 0 0 0 0 0 93924055927824 93924055975568 93924085239808 140724979079714 140724979079719 140724979079719 140724979081194 0
81 (sh) (param) S 0 81 81 0 -1 4210944 693 0 0 0 0 0 0 0 20 0 1 0 172021 4284416 760 18446744073709551615 94666717065216 94666717959381 140728008896192 0 0 0 65536 4 65538 0 0 0 17 0 0 0 0 0 0 94666718198800 94666718246544 94666730864640 140728008903100 140728008903254 140728008903254 140728008904688 0
87 (main) S 81 81 81 0 -1 4210688 541 0 0 0 0 0 0 0 20 0 5 0 172022 1032048640 1892 18446744073709551615 4194304 6405776 140730311069152 0 0 0 0 0 2143420159 0 0 0 17 1 0 0 0 0 0 8507392 8757920 34906112 140730311072280 140730311072287 140730311072287 140730311073777 0
128 (cat) R 118 128 118 34816 128 4210688 152 0 0 0 0 0 0 0 20 0 1 0 193754 5185536 625 18446744073709551615 94301628837888 94301629752385 140721174235312 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0 94301630068720 94301630114308 94301634404352 140721174243694 140721174243865 140721174243865 140721174245355 0
128 (cat) R 118 128 118 34816 128 4210688 152 0 0 0 0 0 0 0 20 0 1 0 193754 5185536 625 18446744073709551615 94301628837888 94301629752385 140721174235312 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0 94301630068720 94301630114308 94301634404352 140721174243694 140721174243865 140721174243865 140721174245355 0
222 (my-cmd) S 87 81 81 0 -1 4210688 541 0 0 0 0 0 0 0 20 0 5 0 172022 1032048640 1892 18446744073709551615 4194304 6405776 140730311069152 0 0 0 0 0 2143420159 0 0 0 17 1 0 0 0 0 0 8507392 8757920 34906112 140730311072280 140730311072287 140730311072287 140730311073777 0
223 (my-cmd) S 87 81 81 0 -1 4210688 541 0 0 0 0 0 0 0 20 0 5 0 172022 1032048640 1892 18446744073709551615 4194304 6405776 140730311069152 0 0 0 0 0 2143420159 0 0 0 17 1 0 0 0 0 0 8507392 8757920 34906112 140730311072280 140730311072287 140730311072287 140730311073777 0
333 (my-cmd) S 222 81 81 0 -1 4210688 541 0 0 0 0 0 0 0 20 0 5 0 172022 1032048640 1892 18446744073709551615 4194304 6405776 140730311069152 0 0 0 0 0 2143420159 0 0 0 17 1 0 0 0 0 0 8507392 8757920 34906112 140730311072280 140730311072287 140730311072287 140730311073777 0
334 (my-cmd) S 222 81 81 0 -1 4210688 541 0 0 0 0 0 0 0 20 0 5 0 172022 1032048640 1892 18446744073709551615 4194304 6405776 140730311069152 0 0 0 0 0 2143420159 0 0 0 17 1 0 0 0 0 0 8507392 8757920 34906112 140730311072280 140730311072287 140730311072287 140730311073777 0`
)

func Test_kubeExecProcessHandler_GetProcessInfoForCommand(t *testing.T) {
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "stopped status if PID file missing",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
			name: "error status if kill -0 command exit status is non-zero and process exit code recorded as failing",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123\n1"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			want: RemoteProcessInfo{
				Pid:    123,
				Status: Errored,
			},
		},
		{
			name: "running status if kill -0 command exit status is zero",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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

			execClient := exec.NewExecClient(kubeClient)
			k := NewKubeExecProcessHandler(execClient)
			got, err := k.GetProcessInfoForCommand(context.Background(), cmdDef, _podName, _containerName)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("kubeExecProcessHandler.GetProcessInfoForCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_kubeExecProcessHandler_StartProcessForCommand(t *testing.T) {
	kill0CmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", p)}
	}

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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c",
						fmt.Sprintf("echo $$ > %[1]s &&   (%s) 1>>/proc/1/fd/1 2>>/proc/1/fd/2; echo $? >> %[1]s",
							getPidFileForCommand(execCmdWithoutWorkingDir), execCmdWithoutWorkingDir.CmdLine)}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("Hello"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(execCmdWithoutWorkingDir))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			isCmdExpectedToRun: true,
			expectedStatuses:   []RemoteProcessStatus{Starting, Running, Stopped},
		},
		{
			name:   "command with all fields returned an error",
			cmdDef: fullExecCmd,
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c",
						fmt.Sprintf("echo $$ > %[1]s && cd %s && export ENV_VAR1='value1' ENV_VAR2='value2' && (%s) 1>>/proc/1/fd/1 2>>/proc/1/fd/2; echo $? >> %[1]s",
							getPidFileForCommand(fullExecCmd), fullExecCmd.WorkingDir, fullExecCmd.CmdLine)}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("error while running command"))
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(fullExecCmd))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123\n1"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			isCmdExpectedToRun: true,
			expectedStatuses:   []RemoteProcessStatus{Starting, Running, Errored},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			execClient := exec.NewExecClient(kubeClient)
			k := NewKubeExecProcessHandler(execClient)

			var wg sync.WaitGroup
			wg.Add(len(tt.expectedStatuses)) //number of invocations of outputHandler
			var statusesReported []RemoteProcessStatus
			err := k.StartProcessForCommand(context.Background(), tt.cmdDef, _podName, _containerName, func(status RemoteProcessStatus, stdout []string, stderr []string, err error) {
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

			if diff := cmp.Diff(tt.expectedStatuses, statusesReported); diff != "" {
				t.Errorf("kubeExecProcessHandler.StartProcessForCommand() expectedStatuses mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_kubeExecProcessHandler_StopProcessForCommand(t *testing.T) {
	cmdDef := CommandDefinition{Id: "my-run"}
	retrieveChildrenCmdProvider := func() []string {
		return []string{ShellExecutable, "-c", "cat /proc/*/stat || true"}
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "nothing to do if PID file missing",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
		},
		{
			name: "error while determining process children",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider()),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
				// parent process should still be killed.
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stderr.Write([]byte("no such process"))
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			wantErr: true,
		},
		{
			name: "no process children killed if no children file found",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("rm -f %s", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error which should be ignored"))
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider()),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stderr.Write([]byte("no such process"))
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
		},
		{
			name: "process children should get killed",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("rm -f %s", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error which should be ignored"))
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("81"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider()),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(statFile))
						return err
					})
				for _, p := range []int{333, 334, 222, 223, 87, 81} {
					kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(p)),
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
					kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(p)),
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
							_, _ = stderr.Write([]byte("no such process"))
							_, err := stdout.Write([]byte("1"))
							return err
						})
				}
			},
		},
		{
			name: "error if any child process could not be killed",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("rm -f %s", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error which should be ignored"))
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq([]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("81"))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(retrieveChildrenCmdProvider()),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(statFile))
						return err
					})
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(333)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("error killing process 333"))
				// parent process should be stopped in all cases
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(killCmdProvider(81)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(kill0CmdProvider(81)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, _ = stderr.Write([]byte("no such process"))
						_, err := stdout.Write([]byte("1"))
						return err
					})
			},
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			execClient := exec.NewExecClient(kubeClient)
			k := NewKubeExecProcessHandler(execClient)
			err := k.StopProcessForCommand(context.Background(), cmdDef, _podName, _containerName)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_kubeExecProcessHandler_getProcessInfoFromPid(t *testing.T) {
	cmdProvider := func(p int) []string {
		return []string{ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", p)}
	}
	for _, tt := range []struct {
		name                 string
		kubeClientCustomizer func(*kclient.MockClientInterface)
		pid                  int
		lastKnownExitStatus  int
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmdProvider(123)),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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

			execClient := exec.NewExecClient(kubeClient)
			k := NewKubeExecProcessHandler(execClient)
			got, err := k.getProcessInfoFromPid(context.Background(), tt.pid, tt.lastKnownExitStatus, _podName, _containerName)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("kubeExecProcessHandler.getProcessInfoFromPid() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_kubeExecProcessHandler_getRemoteProcessPID(t *testing.T) {
	cmdDef := CommandDefinition{Id: "my-run"}
	cmd := []string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", getPidFileForCommand(cmdDef))}
	for _, tt := range []struct {
		name                  string
		kubeClientCustomizer  func(*kclient.MockClientInterface)
		wantPid               int
		wantLastKnownExitCode int
		wantErr               bool
	}{
		{
			name: "error returned at command execution",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			wantErr: true,
		},
		{
			name: "missing pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
		},
		{
			name: "unexpected number of lines in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123\n234\n345"))
						return err
					})
			},
			wantErr: true,
		},
		{
			name: "invalid content in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("invalid-pid"))
						return err
					})
			},
			wantErr: true,
		},
		{
			name: "valid content in pid file with trailing spaces",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(" 123 "))
						return err
					})
			},
			wantPid: 123,
		},
		{
			name: "valid content in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123"))
						return err
					})
			},
			wantPid: 123,
		},
		{
			name: "negative value in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("-1"))
						return err
					})
			},
			wantPid: -1,
		},
		{
			name: "valid content with zero exit status code in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123\n0"))
						return err
					})
			},
			wantPid:               123,
			wantLastKnownExitCode: 0,
		},
		{
			name: "valid content with non-zero exit status code in pid file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123\n1"))
						return err
					})
			},
			wantPid:               123,
			wantLastKnownExitCode: 1,
		},
		{
			name: "error returned content if non-number recorded in pid file as process last-known exit code",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte("123\nNAN"))
						return err
					})
			},
			wantErr: true,
			wantPid: 123,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			execClient := exec.NewExecClient(kubeClient)
			kubeExecClient := NewKubeExecProcessHandler(execClient)
			got, lastKnownExitStatus, err := kubeExecClient.getRemoteProcessPID(context.Background(), cmdDef, _podName, _containerName)
			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.wantPid, got); diff != "" {
				t.Errorf("kubeExecProcessHandler.getRemoteProcessPID() wantPid mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantLastKnownExitCode, lastKnownExitStatus); diff != "" {
				t.Errorf("kubeExecProcessHandler.getRemoteProcessPID() wantLastKnownExitCode mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_kubeExecProcessHandler_getProcessChildren(t *testing.T) {
	const ppid = 123
	cmdProvider := func() []string {
		return []string{ShellExecutable, "-c", "cat /proc/*/stat || true"}
	}

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
				cmd := cmdProvider()
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("an error"))
			},
			ppid:    ppid,
			wantErr: true,
		},
		{
			name: "missing stat file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				cmd := cmdProvider()
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stderr.Write([]byte("no such file or directory"))
						return err
					})
			},
			ppid: ppid,
			want: nil,
		},
		{
			name: "empty stat file",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				cmd := cmdProvider()
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName), gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(""))
						return err
					})
			},
			ppid: ppid,
			want: nil,
		},
		{
			name: "stat file with children at several levels",
			kubeClientCustomizer: func(kclient *kclient.MockClientInterface) {
				cmd := cmdProvider()
				kclient.EXPECT().ExecCMDInContainer(gomock.Any(), gomock.Eq(_containerName), gomock.Eq(_podName),
					gomock.Eq(cmd),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
						_, err := stdout.Write([]byte(statFile))
						return err
					})
			},
			ppid: 81,
			want: []int{333, 334, 222, 223, 87},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.kubeClientCustomizer != nil {
				tt.kubeClientCustomizer(kubeClient)
			}

			execClient := exec.NewExecClient(kubeClient)
			kubeExecClient := NewKubeExecProcessHandler(execClient)
			got, err := kubeExecClient.getProcessChildren(context.Background(), tt.ppid, _podName, _containerName)
			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("kubeExecProcessHandler.getProcessChildren() mismatch (-want +got):\n%s", diff)
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
