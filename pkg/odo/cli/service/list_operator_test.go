package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetOrderedServicesNames(t *testing.T) {
	tests := []struct {
		name     string
		services map[string]*serviceItem
		want     []string
	}{
		{
			name: "Unordered names",
			services: map[string]*serviceItem{
				"name3": nil,
				"name1": nil,
				"name2": nil,
			},
			want: []string{"name1", "name2", "name3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOrderedServicesNames(tt.services)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("Failed %s: got: %q, want: %q", t.Name(), result, tt.want)
			}
		})
	}
}

type getTabularInfoResult struct {
	managedByOdo    string
	state           string
	durationContent bool
}

func TestGetTabularInfo(t *testing.T) {

	tests := []struct {
		name             string
		service          *serviceItem
		devfileComponent string
		want             getTabularInfoResult
	}{
		{
			name: "case 1: service in cluster managed by current devfile",
			service: &serviceItem{
				ClusterInfo: &clusterInfo{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "odo",
						"app.kubernetes.io/instance":   "component1",
					},
				},
				InDevfile: true,
			},
			devfileComponent: "component1",
			want: getTabularInfoResult{
				managedByOdo:    "Yes (component1)",
				state:           "Pushed",
				durationContent: true,
			},
		},
		{
			name: "case 2: service in cluster not managed by Odo",
			service: &serviceItem{
				ClusterInfo: &clusterInfo{
					Labels: map[string]string{},
				},
				InDevfile: false,
			},
			devfileComponent: "component1",
			want: getTabularInfoResult{
				managedByOdo:    "No",
				state:           "",
				durationContent: true,
			},
		},
		{
			name: "case 3: service in cluster absent from current devfile",
			service: &serviceItem{
				ClusterInfo: &clusterInfo{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "odo",
						"app.kubernetes.io/instance":   "component1",
					},
				},
				InDevfile: false,
			},
			devfileComponent: "component1",
			want: getTabularInfoResult{
				managedByOdo:    "Yes (component1)",
				state:           "Deleted locally",
				durationContent: true,
			},
		},
		{
			name: "case 4: service in cluster maaged by another devfile",
			service: &serviceItem{
				ClusterInfo: &clusterInfo{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "odo",
						"app.kubernetes.io/instance":   "component2",
					},
				},
				InDevfile: false,
			},
			devfileComponent: "component1",
			want: getTabularInfoResult{
				managedByOdo:    "Yes (component2)",
				state:           "",
				durationContent: true,
			},
		},
		{
			name: "case 5: service defined in devfile, not in cluster",
			service: &serviceItem{
				ClusterInfo: nil,
				InDevfile:   true,
			},
			devfileComponent: "component1",
			want: getTabularInfoResult{
				managedByOdo:    "Yes (component1)",
				state:           "Not pushed",
				durationContent: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			managedByOdo, state, duration := getTabularInfo(tt.service, tt.devfileComponent)
			if managedByOdo != tt.want.managedByOdo {
				t.Errorf("Failed %s: managedByOdo got: %q, want: %q", t.Name(), managedByOdo, tt.want.managedByOdo)
			}
			if state != tt.want.state {
				t.Errorf("Failed %s: state got: %q, want: %q", t.Name(), state, tt.want.state)
			}
			if len(duration) > 0 != tt.want.durationContent {
				t.Errorf("Failed %s: duration content got: %v, want: %v", t.Name(), len(duration) > 0, tt.want.durationContent)
			}
		})
	}
}

func TestMixServices(t *testing.T) {
	atime, _ := time.Parse(time.RFC3339, "2021-06-02T08:39:20Z00:00")
	tests := []struct {
		name               string
		clusterListInlined []string
		devfileList        []string
		want               []serviceItem
	}{
		{
			name: "two in cluster and two in devfile, including one in common",
			clusterListInlined: []string{`
kind: kind1
metadata:
  name: name1
  labels:
    app.kubernetes.io/managed-by: odo
    app.kubernetes.io/instance: component1
  creationTimestamp: 2021-06-02T08:39:20Z00:00
spec:
  field1: value1`,
				`
kind: kind2
metadata:
  name: name2
  labels:
    app.kubernetes.io/managed-by: odo
    app.kubernetes.io/instance: component2
  creationTimestamp: 2021-06-02T08:39:20Z00:00
spec:
  field2: value2`},
			devfileList: []string{`
kind: kind1
metadata:
    name: name1
spec:
    field1: value1`, `
kind: kind3
metadata:
    name: name3
spec:
    field3: value3`},
			want: []serviceItem{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "kind1/name1",
					},
					ClusterInfo: &clusterInfo{
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "odo",
							"app.kubernetes.io/instance":   "component1",
						},
						CreationTimestamp: atime,
					},
					InDevfile: true,
					Deployed:  true,
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "kind2/name2",
					},
					ClusterInfo: &clusterInfo{
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "odo",
							"app.kubernetes.io/instance":   "component2",
						},
						CreationTimestamp: atime,
					},
					InDevfile: false,
					Deployed:  true,
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "kind3/name3",
					},
					ClusterInfo: nil,
					InDevfile:   true,
					Deployed:    false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usCluster := make([]unstructured.Unstructured, len(tt.clusterListInlined))
			for i, clusterInlined := range tt.clusterListInlined {
				err := yaml.Unmarshal([]byte(clusterInlined), &usCluster[i])
				if err != nil {
					t.Errorf("Failed to unmarshal spec manifest %q: %u", clusterInlined, err)
				}
			}
			usDevfiles := make(map[string]unstructured.Unstructured)
			for _, devfile := range tt.devfileList {
				usDevfile := unstructured.Unstructured{}
				err := yaml.Unmarshal([]byte(devfile), &usDevfile)
				if err != nil {
					t.Errorf("Failed to unmarshal spec manifest %q, %u", devfile, err)
				}
				usDevfiles[usDevfile.GetKind()+"/"+usDevfile.GetName()] = usDevfile
			}
			result := mixServices(usCluster, usDevfiles)
			for i := range result.Items {
				if reflect.DeepEqual(result.Items[i].Manifest, unstructured.Unstructured{}) {
					t.Errorf("Manifest is empty")
				}
				// do not check manifest content
				result.Items[i].Manifest = unstructured.Unstructured{}
			}
			if !reflect.DeepEqual(result.Items, tt.want) {
				t.Errorf("Failed %s\n\ngot: %+v\n\nwant: %+v\n", t.Name(), result.Items, tt.want)
			}
		})
	}
}
