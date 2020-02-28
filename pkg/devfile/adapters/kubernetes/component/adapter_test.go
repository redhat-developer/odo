package component

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func TestComponentAdapter(t *testing.T) {

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

			componentAdapter := New(adapterCtx, *fkclient)
			podTemplateSpec, err := componentAdapter.Initialize()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter initialize unexpected error %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr && (err != nil) {
				// if we want an error, return since the remaining test is not valid
				return
			}

			// Change the status
			go func() {
				fkWatch.Modify(kclient.FakePodStatus(corev1.PodRunning, testComponentName))
			}()
			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			err = componentAdapter.Start(podTemplateSpec)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter start unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
