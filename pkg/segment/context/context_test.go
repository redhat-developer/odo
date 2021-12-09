package context

import (
	"context"
	"reflect"
	"testing"
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

// TODO(feloy) test with fake kclient implementation
//func TestSetClusterType(t *testing.T) {
//	tests := []struct {
//		want   string
//		groups []string
//	}{
//		{
//			want:   "openshift3",
//			groups: []string{"project.openshift.io/v1"},
//		},
//		{
//			want:   "openshift4",
//			groups: []string{"project.openshift.io/v1", "operators.coreos.com/v1alpha1"},
//		},
//		{
//			want: "kubernetes",
//		},
//		{
//			want: NOTFOUND,
//		},
//	}
//
//	for _, tt := range tests {
//		var fakeClient *kclient.Client
//		if tt.want != NOTFOUND {
//			fakeClient, _ = kclient.FakeNew()
//		}
//		if tt.groups != nil {
//			setupCluster(fakeClient, tt.groups)
//		}
//
//		ctx := NewContext(context.Background())
//		SetClusterType(ctx, fakeClient)
//
//		got := GetContextProperties(ctx)[ClusterType]
//		if got != tt.want {
//			t.Errorf("got: %q, want: %q", got, tt.want)
//		}
//	}
//}

func TestGetTelemetryStatus(t *testing.T) {
	want := true
	ctx := NewContext(context.Background())
	setContextProperty(ctx, TelemetryStatus, want)
	got := GetTelemetryStatus(ctx)
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestSetTelemetryStatus(t *testing.T) {
	want := false
	ctx := NewContext(context.Background())
	SetTelemetryStatus(ctx, want)
	got := GetContextProperties(ctx)[TelemetryStatus]
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

//var apiResourceList = map[string]*metav1.APIResourceList{
//	"operators.coreos.com/v1alpha1": {
//		GroupVersion: "operators.coreos.com/v1alpha1",
//		APIResources: []metav1.APIResource{{
//			Name:         "clusterserviceversions",
//			SingularName: "clusterserviceversion",
//			Namespaced:   false,
//			Kind:         "ClusterServiceVersion",
//			ShortNames:   []string{"csv", "csvs"},
//		}},
//	},
//	"project.openshift.io/v1": {
//		GroupVersion: "project.openshift.io/v1",
//		APIResources: []metav1.APIResource{{
//			Name:         "projects",
//			SingularName: "project",
//			Namespaced:   false,
//			Kind:         "Project",
//			ShortNames:   []string{"proj"},
//		}},
//	},
//}

// setupCluster adds resource groups to the client
//func setupCluster(fakeClient kclient.ClientInterface, groupVersion []string) {
//	fd := odoFake.NewFakeDiscovery()
//	for _, group := range groupVersion {
//		fd.AddResourceList(group, apiResourceList[group])
//	}
//	fakeClient.SetDiscoveryInterface(fd)
//}
