package kclient

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/redhat-developer/odo/pkg/preference"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/watch"

	ktesting "k8s.io/client-go/testing"
)

func fakePodStatus(status corev1.PodPhase, podName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Status: corev1.PodStatus{
			Phase: status,
		},
	}
}

// NOTE: We do *not* collection the amount of actions taken in this function as there could be any number of fake
// 'event' actions that are happening in the background.
func TestWaitAndGetPodWithEvents(t *testing.T) {
	tests := []struct {
		name                string
		podName             string
		status              corev1.PodPhase
		wantEventWarning    bool
		wantErr             bool
		eventWarningMessage string
	}{
		{
			name:             "Case 1: Pod running",
			podName:          "ruby",
			status:           corev1.PodRunning,
			wantEventWarning: false,
			wantErr:          false,
		},
		{
			name:             "Case 2: Pod failed",
			podName:          "ruby",
			status:           corev1.PodFailed,
			wantEventWarning: false,
			wantErr:          true,
		},
		{
			name:             "Case 3: Pod unknown",
			podName:          "ruby",
			status:           corev1.PodUnknown,
			wantEventWarning: false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeClient, fakeClientSet := FakeNew()
			fakePodWatch := watch.NewRaceFreeFake()

			// Watch for Pods
			fakePod := fakePodStatus(tt.status, tt.podName)
			go func(pod *corev1.Pod) {
				fakePodWatch.Modify(pod)
			}(fakePod)

			// Prepend watch reactor (beginning of the chain)
			fakeClientSet.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fakePodWatch, nil
			})

			podSelector := fmt.Sprintf("deploymentconfig=%s", tt.podName)

			pod, err := fakeClient.WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, preference.DefaultPushTimeout)

			if !tt.wantErr == (err != nil) {
				t.Errorf("client.WaitAndGetPod(string) unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if pod.Name != tt.podName {
					t.Errorf("pod name is not matching to expected name, expected: %s, got %s", tt.podName, pod.Name)
				}
			}

		})
	}
}

func TestGetOnePodFromSelector(t *testing.T) {
	fakePod := FakePodStatus(corev1.PodRunning, "nodejs")
	fakePod.Labels["component"] = "nodejs"

	fakePodWithDeletionTimeStamp := FakePodStatus(corev1.PodRunning, "nodejs")
	fakePodWithDeletionTimeStamp.Labels["component"] = "nodejs"
	currentTime := metav1.NewTime(time.Now())
	fakePodWithDeletionTimeStamp.DeletionTimestamp = &currentTime

	type args struct {
		selector string
	}
	tests := []struct {
		name         string
		args         args
		returnedPods *corev1.PodList
		want         *corev1.Pod
		wantErr      bool
	}{
		{
			name: "valid number of pods",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{
					*fakePod,
				},
			},
			want:    fakePod,
			wantErr: false,
		},
		{
			name: "zero pods",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
		{
			name: "mutiple pods",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{
					*fakePod,
					*fakePod,
				},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
		{
			name: "pod is in the deletion state",
			args: args{selector: fmt.Sprintf("component=%s", "nodejs")},
			returnedPods: &corev1.PodList{
				Items: []corev1.Pod{
					*fakePodWithDeletionTimeStamp,
				},
			},
			want:    &corev1.Pod{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()

			fkclientset.Kubernetes.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if action.(ktesting.ListAction).GetListRestrictions().Labels.String() != fmt.Sprintf("component=%s", "nodejs") {
					t.Errorf("list called with different selector want:%s, got:%s", fmt.Sprintf("component=%s", "nodejs"), action.(ktesting.ListAction).GetListRestrictions().Labels.String())
				}
				return true, tt.returnedPods, nil
			})

			got, err := fkclient.GetOnePodFromSelector(tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOnePodFromSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr && err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOnePodFromSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPodUsingComponentName(t *testing.T) {
	fakePod := FakePodStatus(corev1.PodRunning, "nodejs")
	fakePod.Labels["component"] = "nodejs"

	type args struct {
		componentName string
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Pod
		wantErr bool
	}{
		{
			name: "list called with same component name",
			args: args{
				componentName: "nodejs",
			},
			want:    fakePod,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.Kubernetes.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if action.(ktesting.ListAction).GetListRestrictions().Labels.String() != fmt.Sprintf("component=%s", tt.args.componentName) {
					t.Errorf("list called with different selector want:%s, got:%s", fmt.Sprintf("component=%s", tt.args.componentName), action.(ktesting.ListAction).GetListRestrictions().Labels.String())
				}
				return true, &corev1.PodList{
					Items: []corev1.Pod{
						*fakePod,
					},
				}, nil
			})

			got, err := fkclient.GetPodUsingComponentName(tt.args.componentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodUsingComponentName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodUsingComponentName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
