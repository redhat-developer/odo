package binding

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	"github.com/redhat-developer/odo/pkg/kclient"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
)

const deploymentKind = "Deployment"

var deploymentGVK = appsv1.SchemeGroupVersion.WithKind(deploymentKind)
var deploymentApiVersion, _ = deploymentGVK.ToAPIVersionAndKind()

const clusterKind = "Cluster"

var clusterGV = schema.GroupVersion{
	Group:   "postgresql.k8s.enterprisedb.io",
	Version: "v1",
}
var clusterGVK = clusterGV.WithKind(clusterKind)
var clusterGVR = clusterGV.WithResource("clusters")
var ClusterAPIVersion, _ = clusterGVK.ToAPIVersionAndKind()

func TestBindingClient_GetFlags(t *testing.T) {
	type args struct {
		flags map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "service and name flags are present",
			args: args{flags: map[string]string{"service": "redisService", "name": "mybinding", "v": "9"}},
			want: map[string]string{"service": "redisService", "name": "mybinding"},
		},
		{
			name: "only one flag is present",
			args: args{map[string]string{"service": "redisService", "v": "9"}},
			want: map[string]string{"service": "redisService"},
		},
		{
			name: "no relevant flags are present",
			args: args{map[string]string{"v": "9"}},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &BindingClient{}
			if got := o.GetFlags(tt.args.flags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindingClient_GetServiceInstances(t *testing.T) {
	ns := "my-ns"

	var clusterUnstructured unstructured.Unstructured

	clusterUnstructured.SetGroupVersionKind(clusterGVK)
	clusterUnstructured.SetName("postgres-cluster")

	serviceBindingInstance := servicebinding.BindableKinds{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BindableKinds",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "bindable-kinds",
		},
		Status: []servicebinding.BindableKindsStatus{
			{
				Group:   "redis.redis.opstreelabs.in",
				Kind:    "Redis",
				Version: "v1beta1",
			},
			{
				Group:   "postgresql.k8s.enterprisedb.io",
				Kind:    "Cluster",
				Version: "v1",
			},
		},
	}
	type fields struct {
		kubernetesClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "obtained service instances",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().IsServiceBindingSupported().Return(true, nil)
					client.EXPECT().GetBindableKinds().Return(serviceBindingInstance, nil)
					client.EXPECT().GetBindableKindStatusRestMapping(serviceBindingInstance.Status).Return([]*meta.RESTMapping{
						{Resource: clusterGVR, GroupVersionKind: clusterGVK},
					}, nil)

					client.EXPECT().ListDynamicResources("", clusterGVR).Return(&unstructured.UnstructuredList{Items: []unstructured.Unstructured{clusterUnstructured}}, nil)
					return client
				},
			},
			want: map[string]unstructured.Unstructured{
				"postgres-cluster (Cluster.postgresql.k8s.enterprisedb.io)": clusterUnstructured,
			},
			wantErr: false,
		},
		{
			name: "obtained service instances from specific namespace",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().IsServiceBindingSupported().Return(true, nil)
					client.EXPECT().GetBindableKinds().Return(serviceBindingInstance, nil)
					client.EXPECT().GetBindableKindStatusRestMapping(serviceBindingInstance.Status).Return([]*meta.RESTMapping{
						{Resource: clusterGVR, GroupVersionKind: clusterGVK},
					}, nil)

					client.EXPECT().ListDynamicResources(ns, clusterGVR).Return(&unstructured.UnstructuredList{Items: []unstructured.Unstructured{clusterUnstructured}}, nil)
					return client
				},
			},
			args: args{
				namespace: ns,
			},
			want: map[string]unstructured.Unstructured{
				"postgres-cluster (Cluster.postgresql.k8s.enterprisedb.io)": clusterUnstructured,
			},
			wantErr: false,
		},
		{
			name: "do not fail if no bindable kind service was found",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().IsServiceBindingSupported().Return(true, nil)
					client.EXPECT().GetBindableKinds().Return(serviceBindingInstance, nil)
					client.EXPECT().GetBindableKindStatusRestMapping(serviceBindingInstance.Status).Return(nil, nil)
					return client
				},
			},
			want:    map[string]unstructured.Unstructured{},
			wantErr: false,
		},
		{
			name: "do not fail if no instances of the bindable kind services was found",
			fields: fields{kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
				client := kclient.NewMockClientInterface(ctrl)
				client.EXPECT().IsServiceBindingSupported().Return(true, nil)
				client.EXPECT().GetBindableKinds().Return(serviceBindingInstance, nil)
				client.EXPECT().GetBindableKindStatusRestMapping(serviceBindingInstance.Status).Return([]*meta.RESTMapping{
					{Resource: clusterGVR, GroupVersionKind: clusterGVK},
				}, nil)

				client.EXPECT().ListDynamicResources("", clusterGVR).Return(&unstructured.UnstructuredList{Items: nil}, nil)
				return client
			}},
			want:    map[string]unstructured.Unstructured{},
			wantErr: false,
		},
		{
			name: "error out if the servicebinding CRD is not found",
			fields: fields{kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
				client := kclient.NewMockClientInterface(ctrl)
				client.EXPECT().IsServiceBindingSupported().Return(false, nil)
				return client
			}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &BindingClient{
				kubernetesClient: tt.fields.kubernetesClient(ctrl),
			}
			got, err := o.GetServiceInstances(tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceInstances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServiceInstances() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindingClient_AddBindingToDevfile(t *testing.T) {
	bindingName := "my-nodejs-app-cluster-sample"
	ns := "my-ns"

	var clusterUnstructured unstructured.Unstructured
	clusterUnstructured.SetGroupVersionKind(clusterGVK)
	clusterUnstructured.SetName("cluster-sample")

	serviceBindingRef := servicebinding.Service{
		Id: &bindingName,
		NamespacedRef: servicebinding.NamespacedRef{
			Ref: servicebinding.Ref{
				Group:    clusterGVK.Group,
				Version:  clusterGVK.Version,
				Kind:     clusterGVK.Kind,
				Name:     clusterUnstructured.GetName(),
				Resource: "clusters",
			},
		},
	}

	serviceBindingRefWithNs := servicebinding.Service{
		Id: &bindingName,
		NamespacedRef: servicebinding.NamespacedRef{
			Ref: servicebinding.Ref{
				Group:    clusterGVK.Group,
				Version:  clusterGVK.Version,
				Kind:     clusterGVK.Kind,
				Name:     clusterUnstructured.GetName(),
				Resource: "clusters",
			},
			Namespace: &ns,
		},
	}

	type fields struct {
		kubernetesClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		bindingName         string
		bindAsFiles         bool
		namingStrategy      string
		namespace           string
		unstructuredService unstructured.Unstructured
		obj                 parser.DevfileObj
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    parser.DevfileObj
		wantErr bool
	}{
		{
			name: "successfully add binding",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().NewServiceBindingServiceObject("", clusterUnstructured, bindingName).Return(serviceBindingRef, nil)
					client.EXPECT().GetDeploymentAPIVersion().Return(deploymentGVK, nil)
					return client
				},
			},
			args: args{
				bindingName:         bindingName,
				bindAsFiles:         false,
				unstructuredService: clusterUnstructured,
				obj:                 odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs()),
			},
			want:    getDevfileObjWithServiceBinding(bindingName, "", false, ""),
			wantErr: false,
		},
		{
			name: "successfully add binding with namespace",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().NewServiceBindingServiceObject(ns, clusterUnstructured, bindingName).Return(serviceBindingRefWithNs, nil)
					client.EXPECT().GetDeploymentAPIVersion().Return(deploymentGVK, nil)
					return client
				},
			},
			args: args{
				bindingName:         bindingName,
				bindAsFiles:         false,
				namespace:           ns,
				unstructuredService: clusterUnstructured,
				obj:                 odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs()),
			},
			want:    getDevfileObjWithServiceBinding(bindingName, ns, false, ""),
			wantErr: false,
		},
		{
			name: "successfully added binding for a Service Binding bound as files",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().NewServiceBindingServiceObject("", clusterUnstructured, bindingName).Return(serviceBindingRef, nil)
					client.EXPECT().GetDeploymentAPIVersion().Return(deploymentGVK, nil)
					return client
				},
			},
			args: args{
				bindingName:         bindingName,
				bindAsFiles:         true,
				unstructuredService: clusterUnstructured,
				obj:                 odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs()),
			},
			want:    getDevfileObjWithServiceBinding(bindingName, "", true, ""),
			wantErr: false,
		},
		{
			name: "successfully added binding for a Service Binding bound as files and namespace",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().NewServiceBindingServiceObject(ns, clusterUnstructured, bindingName).Return(serviceBindingRefWithNs, nil)
					client.EXPECT().GetDeploymentAPIVersion().Return(deploymentGVK, nil)
					return client
				},
			},
			args: args{
				bindingName:         bindingName,
				bindAsFiles:         true,
				namespace:           ns,
				unstructuredService: clusterUnstructured,
				obj:                 odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs()),
			},
			want:    getDevfileObjWithServiceBinding(bindingName, ns, true, ""),
			wantErr: false,
		},
		{
			name: "successfully added binding for a Service Binding with naming strategy",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().NewServiceBindingServiceObject("", clusterUnstructured, bindingName).Return(serviceBindingRef, nil)
					client.EXPECT().GetDeploymentAPIVersion().Return(deploymentGVK, nil)
					return client
				},
			},
			args: args{
				bindingName:         bindingName,
				bindAsFiles:         true,
				namingStrategy:      "uppercase",
				unstructuredService: clusterUnstructured,
				obj:                 odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs()),
			},
			want:    getDevfileObjWithServiceBinding(bindingName, "", true, "uppercase"),
			wantErr: false,
		},
		{
			name: "successfully added binding for a Service Binding with naming strategy and namespace",
			fields: fields{
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().NewServiceBindingServiceObject(ns, clusterUnstructured, bindingName).Return(serviceBindingRefWithNs, nil)
					client.EXPECT().GetDeploymentAPIVersion().Return(deploymentGVK, nil)
					return client
				},
			},
			args: args{
				bindingName:         bindingName,
				bindAsFiles:         true,
				namespace:           ns,
				namingStrategy:      "uppercase",
				unstructuredService: clusterUnstructured,
				obj:                 odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs()),
			},
			want:    getDevfileObjWithServiceBinding(bindingName, ns, true, "uppercase"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &BindingClient{
				kubernetesClient: tt.fields.kubernetesClient(ctrl),
			}
			got, err := o.AddBindingToDevfile(tt.args.bindingName, tt.args.bindAsFiles, tt.args.namespace, tt.args.namingStrategy, tt.args.unstructuredService, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddBindingToDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddBindingToDevfile(): %v", pretty.Compare(got, tt.want))
			}
		})
	}
}

func getDevfileObjWithServiceBinding(bindingName string, ns string, bindAsFiles bool, namingStrategy string) parser.DevfileObj {
	obj := odoTestingUtil.GetTestDevfileObj(filesystem.NewFakeFs())
	_ = obj.Data.AddComponents([]v1alpha2.Component{{
		Name: bindingName,
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: &v1alpha2.KubernetesComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					BaseComponent: v1alpha2.BaseComponent{},
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Inlined: getServiceBindingInlinedContent(ns, bindAsFiles, namingStrategy),
					},
				},
			},
		},
	}})
	return obj
}

func getServiceBindingInlinedContent(ns string, bindAsFiles bool, namingStrategy string) string {
	h := fmt.Sprintf(`apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  creationTimestamp: null
  name: my-nodejs-app-cluster-sample
spec:
  application:
    group: apps
    kind: Deployment
    name: my-nodejs-app-app
    version: v1
  bindAsFiles: %v
  detectBindingResources: true`, bindAsFiles)

	fNoNamespace := `
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: my-nodejs-app-cluster-sample
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1
status:
  secret: ""
`

	fNamespace := fmt.Sprintf(`
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: my-nodejs-app-cluster-sample
    kind: Cluster
    name: cluster-sample
    namespace: %s
    resource: clusters
    version: v1
status:
  secret: ""
`, ns)

	if namingStrategy != "" {
		if ns != "" {
			return fmt.Sprintf(`%s
  namingStrategy: %s%s`, h, namingStrategy, fNamespace)
		}
		return fmt.Sprintf(`%s
  namingStrategy: %s%s`, h, namingStrategy, fNoNamespace)
	}

	if ns != "" {
		return h + fNamespace
	}

	return h + fNoNamespace
}
