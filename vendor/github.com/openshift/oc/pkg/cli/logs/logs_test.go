package logs

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/logs"

	buildv1 "github.com/openshift/api/build/v1"
	buildfake "github.com/openshift/client-go/build/clientset/versioned/fake"
)

// TestLogsFlagParity makes sure that our copied flags don't slip during rebases
func TestLogsFlagParity(t *testing.T) {
	streams := genericclioptions.NewTestIOStreamsDiscard()
	kubeCmd := logs.NewCmdLogs(nil, streams)
	originCmd := NewCmdLogs("oc", nil, streams)

	kubeCmd.LocalFlags().VisitAll(func(kubeFlag *pflag.Flag) {
		originFlag := originCmd.LocalFlags().Lookup(kubeFlag.Name)
		if originFlag == nil {
			t.Errorf("missing %v flag", kubeFlag.Name)
			return
		}

		if !reflect.DeepEqual(originFlag, kubeFlag) {
			t.Errorf("flag %v %v does not match %v", kubeFlag.Name, kubeFlag, originFlag)
		}
	})
}

func TestRunLogForPipelineStrategy(t *testing.T) {
	bld := buildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "foo-0",
			Namespace:   "foo",
			Annotations: map[string]string{buildv1.BuildJenkinsBlueOceanLogURLAnnotation: "https://foo"},
		},
		Spec: buildv1.BuildSpec{
			CommonSpec: buildv1.CommonSpec{
				Strategy: buildv1.BuildStrategy{
					JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{},
				},
			},
		},
	}

	fakebc := buildfake.NewSimpleClientset(&bld)
	streams, _, out, _ := genericclioptions.NewTestIOStreams()

	testCases := []struct {
		o runtime.Object
	}{
		{
			o: &bld,
		},
		{
			o: &buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "foo",
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{
						Strategy: buildv1.BuildStrategy{
							JenkinsPipelineStrategy: &buildv1.JenkinsPipelineBuildStrategy{},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		o := &LogsOptions{
			LogsOptions: &logs.LogsOptions{
				IOStreams: streams,
				Object:    tc.o,
				Namespace: "foo",
				Options:   &corev1.PodLogOptions{},
			},
			Client: fakebc.BuildV1(),
		}
		if err := o.RunLog(); err != nil {
			t.Errorf("%#v: RunLog error %v", tc.o, err)
		}
		if !strings.Contains(out.String(), "https://foo") {
			t.Errorf("%#v: RunLog did not have https://foo, but rather had: %s", tc.o, out.String())
		}
	}

}
