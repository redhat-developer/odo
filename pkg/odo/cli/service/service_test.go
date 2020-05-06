package service

import (
	"github.com/openshift/odo/pkg/testingutil"
	"sort"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestCompletions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler completion.ContextualizedPredictor
		cmd     *cobra.Command
		last    string
		want    []string
	}{
		{
			name:    "Completing service create without input returns all available service class external names",
			handler: completion.ServiceClassCompletionHandler,
			cmd:     NewCmdServiceCreate(createRecommendedCommandName, createRecommendedCommandName),
			want:    []string{"foo", "bar", "boo"},
		},
		{
			name:    "Completing service delete without input returns all available service instances",
			handler: completion.ServiceCompletionHandler,
			cmd:     NewCmdServiceDelete(deleteRecommendedCommandName, deleteRecommendedCommandName),
			want:    []string{"foo"},
		},
	}

	client, fakeClientSet := occlient.FakeNew()
	fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.ClusterServiceClassList{
			Items: []v1beta1.ClusterServiceClass{
				testingutil.FakeClusterServiceClass("foo"),
				testingutil.FakeClusterServiceClass("bar"),
				testingutil.FakeClusterServiceClass("boo"),
			},
		}, nil
	})
	fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1beta1.ServiceInstanceList{
			Items: []v1beta1.ServiceInstance{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app.kubernetes.io/part-of": "foo", componentlabels.ComponentLabel: "foo", componentlabels.ComponentTypeLabel: "service"},
					},
					Status: v1beta1.ServiceInstanceStatus{
						Conditions: []v1beta1.ServiceInstanceCondition{
							{
								Reason: "some reason",
							},
						},
					},
				},
			},
		}, nil
	})
	context := genericclioptions.NewFakeContext("", "", "", client)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := complete.Args{Last: tt.last}

			got := tt.handler(tt.cmd, completion.NewParsedArgs(a, tt.cmd), context)

			if !equal(got, tt.want) {
				t.Errorf("Failed %s: got: %q, want: %q", t.Name(), got, tt.want)
			}
		})
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
