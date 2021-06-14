package context

import (
	"context"
	"reflect"
	"sync"
	"testing"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/discovery/fake"

	"github.com/openshift/odo/pkg/occlient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetContextProperties(t *testing.T) {
	ckey, value := "preferenceKey", "consenttelemetry"
	ctx := NewContext(context.Background())
	setContextProperty(ctx, ckey, value)

	got := GetContextProperties(ctx)
	want := map[string]interface{}{ckey: value}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: %q got: %q", want, got)
	}
}

func TestSetComponentType(t *testing.T) {
	want := "java"
	for _, value := range []string{"java", "java:8", "myproject/java:8"} {
		ctx := NewContext(context.Background())
		SetComponentType(ctx, value)

		if got, contains := GetContextProperties(ctx)[ComponentType]; !contains || got != want {
			t.Errorf("component type was not set. Got: %q, Want: %q", got, want)
		}
	}
}

func TestSetClusterType(t *testing.T) {
	tests := []struct {
		want  string
		setup func(client *occlient.Client)
	}{
		{
			want:  "openshift3",
			setup: fakeProjects,
		},
		{
			want:  "openshift4",
			setup: setupOCP4,
		},
		{
			want:  "kubernetes",
			setup: nil,
		},
		{
			want:  "not-found",
			setup: nil,
		},
	}

	for _, tt := range tests {
		var fakeClient *occlient.Client
		if tt.want != "not-found" {
			fakeClient, _ = occlient.FakeNew()
		}
		if tt.setup != nil {
			tt.setup(fakeClient)
		}

		ctx := NewContext(context.Background())
		SetClusterType(ctx, fakeClient)

		got := GetContextProperties(ctx)[ClusterType]
		if got != tt.want {
			t.Errorf("got: %q, want: %q", got, tt.want)
		}
	}
}

type resourceMapEntry struct {
	list *metav1.APIResourceList
	err  error
}

type fakeDiscovery struct {
	*fake.FakeDiscovery

	lock        sync.Mutex
	resourceMap map[string]*resourceMapEntry
}

var fakeDiscoveryWithProject = &fakeDiscovery{
	resourceMap: map[string]*resourceMapEntry{
		"project.openshift.io/v1": {
			list: &metav1.APIResourceList{
				GroupVersion: "project.openshift.io/v1",
				APIResources: []metav1.APIResource{{
					Name:         "projects",
					SingularName: "project",
					Namespaced:   false,
					Kind:         "Project",
					ShortNames:   []string{"proj"},
				}},
			},
		},
	},
}

var fakeDiscoveryOCP4 = &fakeDiscovery{
	resourceMap: map[string]*resourceMapEntry{
		"operators.coreos.com/v1alpha1": {
			list: &metav1.APIResourceList{
				GroupVersion: "operators.coreos.com/v1alpha1",
				APIResources: []metav1.APIResource{{
					Name:         "clusterserviceversions",
					SingularName: "clusterserviceversion",
					Namespaced:   false,
					Kind:         "ClusterServiceVersion",
					ShortNames:   []string{"csv", "csvs"},
				}},
			},
		},
		"project.openshift.io/v1": {
			list: &metav1.APIResourceList{
				GroupVersion: "project.openshift.io/v1",
				APIResources: []metav1.APIResource{{
					Name:         "projects",
					SingularName: "project",
					Namespaced:   false,
					Kind:         "Project",
					ShortNames:   []string{"proj"},
				}},
			},
		},
	},
}

func fakeProjects(fakeClient *occlient.Client) {
	fakeClient.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithProject)
}

// setupOCP4 adds fakeDiscovery with clusterserviceversion and project
func setupOCP4(fakeClient *occlient.Client) {
	fakeClient.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryOCP4)
}

func (c *fakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if rl, ok := c.resourceMap[groupVersion]; ok {
		return rl.list, rl.err
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, "")
}
