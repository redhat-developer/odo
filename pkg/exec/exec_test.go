package exec

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	_podName       = "my-pod"
	_containerName = "my-container"
)

type fakePlatform struct {
	execCMDInContainer func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error
}

func (o fakePlatform) ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return o.execCMDInContainer(containerName, podName, cmd, stdout, stderr, stdin, tty)
}

func (o fakePlatform) GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error) {
	panic("not implemented yet")
}

func (o fakePlatform) GetPodsMatchingSelector(selector string) (*corev1.PodList, error) {
	panic("not implemented yet")
}

func (o fakePlatform) GetAllResourcesFromSelector(selector string, ns string) ([]unstructured.Unstructured, error) {
	panic("not implemented yet")
}

func (o fakePlatform) GetAllPodsInNamespaceMatchingSelector(selector string, ns string) (*corev1.PodList, error) {
	panic("not implemented yet")
}

func (o fakePlatform) GetRunningPodFromSelector(selector string) (*corev1.Pod, error) {
	panic("not implemented yet")
}

func TestExecuteCommand(t *testing.T) {
	for _, tt := range []struct {
		name               string
		cmd                []string
		execCMDInContainer func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error
		wantErr            bool
		wantStdout         []string
		wantStderr         []string
	}{
		{
			name: "command returning an error",
			cmd:  []string{"unknown-cmd"},
			execCMDInContainer: func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
				_, _ = stderr.Write([]byte("error running command\nanother message"))
				return errors.New("some error")
			},
			wantErr:    true,
			wantStderr: []string{"error running command", "another message"},
		},
		{
			name: "command not returning an error",
			cmd:  []string{"cat", "/path/to/my/file"},
			execCMDInContainer: func(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
				_, _ = stdout.Write([]byte("Hello World\n\n\n"))
				_, _ = stdout.Write([]byte("Lorem Ipsum Dolor Sit Amet\n"))
				_, _ = stderr.Write([]byte("some message written to stderr"))
				return nil
			},
			wantStdout: []string{"Hello World", "", "", "Lorem Ipsum Dolor Sit Amet"},
			wantStderr: []string{"some message written to stderr"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			platformClient := fakePlatform{
				execCMDInContainer: tt.execCMDInContainer,
			}

			execClient := NewExecClient(platformClient)
			stdout, stderr, err := execClient.ExecuteCommand(tt.cmd, _podName, _containerName, false, nil, nil)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.wantStdout, stdout); diff != "" {
				t.Errorf("ExecClient.ExecuteCommand() wantStdout mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantStderr, stderr); diff != "" {
				t.Errorf("ExecClient.ExecuteCommand() wantStderr mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
