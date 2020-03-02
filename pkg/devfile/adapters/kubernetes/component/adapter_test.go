package component

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func TestStart(t *testing.T) {

	testComponentName := "test"

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Invalid devfile",
			componentType: "",
			wantErr:       true,
		},
		{
			name:          "Case: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				podStatus := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: testComponentName,
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}
				fkWatch.Modify(podStatus)
			}()
			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			componentAdapter := New(adapterCtx, *fkclient)
			err := componentAdapter.Start()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter start unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
