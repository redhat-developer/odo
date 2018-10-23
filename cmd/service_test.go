package cmd

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/occlient"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
	"sort"
	"testing"
)

func TestCompletions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler completionHandler
		last    string
		want    []string
	}{
		{
			name:    "Completing service create without input returns all available service class external names",
			handler: Suggesters[getCommandSuggesterNameFrom("create")],
			want:    []string{"foo", "bar", "boo"},
		},
	}

	client, fakeClientSet := occlient.FakeNew()
	fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.ClusterServiceClassList{
			Items: []v1beta1.ClusterServiceClass{
				fakeClusterServiceClass("foo"),
				fakeClusterServiceClass("bar"),
				fakeClusterServiceClass("boo"),
			},
		}, nil
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := complete.Args{Last: tt.last}
			tt.handler.client = func() *occlient.Client {
				return client
			}
			got := tt.handler.Predict(a)
			if !equal(got, tt.want) {
				t.Errorf("Failed %s: got: %q, want: %q", t.Name(), got, tt.want)
			}
		})
	}
}

func fakeClusterServiceClass(name string) v1beta1.ClusterServiceClass {
	return v1beta1.ClusterServiceClass{
		Spec: v1beta1.ClusterServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: name,
			},
		},
	}
}

func equal(s1, s2 []string) bool {
	sort.Strings(s1)
	sort.Strings(s2)
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
