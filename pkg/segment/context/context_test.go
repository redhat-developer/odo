package context

import (
	"context"
	"reflect"
	"testing"

	odoFake "github.com/openshift/odo/pkg/kclient/fake"

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
		want   string
		groups []string
	}{
		{
			want:   "openshift3",
			groups: []string{"project.openshift.io/v1"},
		},
		{
			want:   "openshift4",
			groups: []string{"project.openshift.io/v1", "operators.coreos.com/v1alpha1"},
		},
		{
			want: "kubernetes",
		},
		{
			want: "not-found",
		},
	}

	for _, tt := range tests {
		var fakeClient *occlient.Client
		if tt.want != "not-found" {
			fakeClient, _ = occlient.FakeNew()
		}
		if tt.groups != nil {
			setupCluster(fakeClient, tt.groups)
		}

		ctx := NewContext(context.Background())
		SetClusterType(ctx, fakeClient)

		got := GetContextProperties(ctx)[ClusterType]
		if got != tt.want {
			t.Errorf("got: %q, want: %q", got, tt.want)
		}
	}
}

var apiResourceList = map[string]*metav1.APIResourceList{
	"operators.coreos.com/v1alpha1": {
		GroupVersion: "operators.coreos.com/v1alpha1",
		APIResources: []metav1.APIResource{{
			Name:         "clusterserviceversions",
			SingularName: "clusterserviceversion",
			Namespaced:   false,
			Kind:         "ClusterServiceVersion",
			ShortNames:   []string{"csv", "csvs"},
		}},
	},
	"project.openshift.io/v1": {
		GroupVersion: "project.openshift.io/v1",
		APIResources: []metav1.APIResource{{
			Name:         "projects",
			SingularName: "project",
			Namespaced:   false,
			Kind:         "Project",
			ShortNames:   []string{"proj"},
		}},
	},
}

// setupCluster adds resource groups to the client
func setupCluster(fakeClient *occlient.Client, groupVersion []string) {
	fd := odoFake.NewFakeDiscovery()
	for _, group := range groupVersion {
		fd.AddResourceList(group, apiResourceList[group])
	}
	fakeClient.GetKubeClient().SetDiscoveryInterface(fd)
}
