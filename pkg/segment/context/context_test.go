package context

import (
	"context"
	"reflect"
	"testing"

	projectv1 "github.com/openshift/api/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"

	"github.com/openshift/odo/pkg/occlient"
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
	ckey, value := "componentType", "maven"
	ctx := NewContext(context.Background())
	SetComponentType(ctx, value)

	if _, contains := GetContextProperties(ctx)[ckey]; !contains {
		t.Errorf("component type was not set.")
	}
}

func TestSetClusterType(t *testing.T) {
	ckey := "clusterType"
	fakeClient, fakeClientset := occlient.FakeNew()
	ctx := NewContext(context.Background())
	cases := []struct {
		want  string
		setup func(*occlient.FakeClientset)
	}{
		{
			"openshift3",
			nil,
		},
		{
			"openshift4",
			fakeClusterVersion,
		},
		{
			"vanilla-kubernetes",
			fakeProjects,
		},
	}
	for _, c := range cases {
		if c.setup != nil {
			c.setup(fakeClientset)
		}
		SetClusterType(ctx, fakeClient)
		got, _ := GetContextProperties(ctx)[ckey]
		if got != c.want {
			t.Errorf("got: %q, want%q", got, c.want)
		}
	}
}

func fakeClusterVersion(fakeClientset *occlient.FakeClientset) {

}

func fakeProjects(fakeClientset *occlient.FakeClientset) {
	fakeClientset.ProjClientset.PrependReactor("list", "projectrequests", func(action ktesting.Action) (bool, runtime.Object, error) {
		proj := projectv1.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
		return true, &proj, nil
	})
}
