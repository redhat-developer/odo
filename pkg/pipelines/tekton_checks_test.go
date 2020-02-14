package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
)

func TestTektonChecker(t *testing.T) {
	var tests = []struct {
		desscription       string
		existingCRDs       []runtime.Object
		wantedResult       bool
		wantedErrorMessage string
	}{
		{"All required CRDs are present.",
			[]runtime.Object{
				crd("pipelines.tekton.dev"),
				crd("pipelineresources.tekton.dev"),
				crd("pipelineruns.tekton.dev"),
				crd("triggerbindings.tekton.dev"),
				crd("triggertemplates.tekton.dev"),
				crd("clustertasks.tekton.dev"),
				crd("taskruns.tekton.dev"),
				crd("tasks.tekton.dev"),
			},
			true,
			"",
		},
		{"All required CRDs are present with some extra CRDs",
			[]runtime.Object{
				crd("pipelines.tekton.dev"),
				crd("pipelineresources.tekton.dev"),
				crd("pipelineruns.tekton.dev"),
				crd("triggerbindings.tekton.dev"),
				crd("triggertemplates.tekton.dev"),
				crd("clustertasks.tekton.dev"),
				crd("taskruns.tekton.dev"),
				crd("tasks.tekton.dev"),
				crd("something else.tekton.dev"),
			},
			true,
			"",
		},
		{"Missed one required pipeline CRD",
			[]runtime.Object{
				crd("pipelineresources.tekton.dev"),
				crd("pipelineruns.tekton.dev"),
				crd("triggerbindings.tekton.dev"),
				crd("triggertemplates.tekton.dev"),
				crd("clustertasks.tekton.dev"),
				crd("taskruns.tekton.dev"),
				crd("tasks.tekton.dev"),
			},
			false,
			"",
		},
		{"Missed one required trigger CRD",
			[]runtime.Object{
				crd("pipelines.tekton.dev"),
				crd("pipelineresources.tekton.dev"),
				crd("pipelineruns.tekton.dev"),
				crd("triggertemplates.tekton.dev"),
				crd("clustertasks.tekton.dev"),
				crd("taskruns.tekton.dev"),
				crd("tasks.tekton.dev"),
			},
			false,
			"",
		},
		{"Missed more than one required CRDs",
			[]runtime.Object{
				crd("pipelineresources.tekton.dev"),
				crd("pipelineruns.tekton.dev"),
				crd("triggertemplates.tekton.dev"),
				crd("clustertasks.tekton.dev"),
				crd("taskruns.tekton.dev"),
				crd("tasks.tekton.dev"),
			},
			false,
			"",
		},
		{"Zero required CRDs",
			[]runtime.Object{},
			false,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.desscription, func(t *testing.T) {

			tektonChecker, err := newFakeChecker(test.existingCRDs)
			if err != nil {
				t.Fatal(err)
			}
			result, err := tektonChecker.checkInstall()
			if result != test.wantedResult {
				t.Fatalf("Want check result '%v', but got '%v'", test.wantedResult, result)
			}
			var errMessage string = ""
			if err != nil {
				errMessage = err.Error()
			}

			if diff := cmp.Diff(test.wantedErrorMessage, errMessage); diff != "" {
				t.Fatalf("Unexpected error \n%s", diff)
			}
		})
	}
}

func newFakeChecker(objs []runtime.Object) (*tektonChecker, error) {

	// get client set from client config
	cs := fake.NewSimpleClientset(objs...)

	return &tektonChecker{
		strategy: &checkStrategy{
			requiredCRDs: requiredCRDNames,
			client:       cs.ApiextensionsV1beta1().CustomResourceDefinitions(),
		},
	}, nil
}

func crd(name string) *v1beta1.CustomResourceDefinition {
	return &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}
