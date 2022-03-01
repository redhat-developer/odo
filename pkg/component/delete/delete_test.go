package delete

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/kclient"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktesting "k8s.io/client-go/testing"
	"reflect"
	"testing"
)

// TODO : Add tests

func TestDeleteComponentClient_DeleteComponent(t *testing.T) {
	type fields struct {
		kubeClient kclient.ClientInterface
	}
	type args struct {
		devfileObj    parser.DevfileObj
		componentName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do := &DeleteComponentClient{
				kubeClient: tt.fields.kubeClient,
			}
			if err := do.DeleteComponent(tt.args.devfileObj, tt.args.componentName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteComponentClient_UnDeploy(t *testing.T) {
	type fields struct {
		kubeClient kclient.ClientInterface
	}
	type args struct {
		devfileObj parser.DevfileObj
		path       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DeleteComponentClient{
				kubeClient: tt.fields.kubeClient,
			}
			if err := o.UnDeploy(tt.args.devfileObj, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("UnDeploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteComponentClient_getPod(t *testing.T) {
	type fields struct {
		kubeClient kclient.ClientInterface
	}
	type args struct {
		componentName string
		appName       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantPod *corev1.Pod
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := DeleteComponentClient{
				kubeClient: tt.fields.kubeClient,
			}
			gotPod, err := o.getPod(tt.args.componentName, tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPod, tt.wantPod) {
				t.Errorf("getPod() gotPod = %v, want %v", gotPod, tt.wantPod)
			}
		})
	}
}

func TestAdapterDelete(t *testing.T) {
	type args struct {
		labels map[string]string
	}

	emptyPods := &corev1.PodList{
		Items: []corev1.Pod{},
	}

	tests := []struct {
		name            string
		args            args
		existingPod     *corev1.PodList
		componentName   string
		componentExists bool
		wantErr         bool
	}{
		{
			name: "case 1: component exists and given labels are valid",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*odoTestingUtil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "component",
			componentExists: true,
			wantErr:         false,
		},
		{
			name: "case 2: component exists and given labels are not valid",
			args: args{labels: nil},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*odoTestingUtil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "component",
			componentExists: true,
			wantErr:         true,
		},
		{
			name: "case 3: component doesn't exists",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*odoTestingUtil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "nocomponent",
			componentExists: false,
			wantErr:         false,
		},
		{
			name: "case 4: resource forbidden",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*odoTestingUtil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "resourceforbidden",
			componentExists: false,
			wantErr:         false,
		},
		{
			name: "case 5: component check error",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*odoTestingUtil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "componenterror",
			componentExists: true,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("delete-collection", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				if util.ConvertLabelsToSelector(tt.args.labels) != action.(ktesting.DeleteCollectionAction).GetListRestrictions().Labels.String() {
					return true, nil, errors.Errorf("collection labels are not matching, wanted: %v, got: %v", util.ConvertLabelsToSelector(tt.args.labels), action.(ktesting.DeleteCollectionAction).GetListRestrictions().Labels.String())
				}
				return true, nil, nil
			})

			fkclientset.Kubernetes.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if tt.componentName == "nocomponent" {
					return true, emptyPods, &kclient.PodNotFoundError{Selector: "somegarbage"}
				} else if tt.componentName == "resourceforbidden" {
					return true, emptyPods, kerrors.NewForbidden(schema.GroupResource{}, "", nil)
				} else if tt.componentName == "componenterror" {
					return true, emptyPods, errors.Errorf("pod check error")
				}
				return true, tt.existingPod, nil
			})

			if err := component.Delete(fkclient, devObj, tt.componentName, "app", tt.args.labels, false, false); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteComponentClient_ListKubernetesComponents(t *testing.T) {
	type fields struct {
		kubeClient kclient.ClientInterface
	}
	type args struct {
		devfileObj parser.DevfileObj
		path       string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantList []unstructured.Unstructured
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DeleteComponentClient{
				kubeClient: tt.fields.kubeClient,
			}
			gotList, err := o.ListKubernetesComponents(tt.args.devfileObj, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListKubernetesComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotList, tt.wantList) {
				t.Errorf("ListKubernetesComponents() gotList = %v, want %v", gotList, tt.wantList)
			}
		})
	}
}
