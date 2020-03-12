package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	serviceAccName = "pipeline"
)

func TestCreateDevCDDeployTemplate(t *testing.T) {
	validDevCDTemplate := triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("testns", "dev-cd-deploy-from-master-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				pipelinev1.ParamSpec{
					Name:        "gitsha",
					Description: "The specific commit SHA.",
				},
				pipelinev1.ParamSpec{
					Name:        "gitrepositoryurl",
					Description: "The git repository url",
				},
			},

			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createDevCDResourcetemplate(serviceAccName, "example.com:5000/testing/testing"),
				},
			},
		},
	}

	template := createDevCDDeployTemplate("testns", serviceAccName, "example.com:5000/testing/testing")
	if diff := cmp.Diff(validDevCDTemplate, template); diff != "" {
		t.Fatalf("CreateDevCDDeployTemplate failed:\n%s", diff)
	}
}

func TestCreateDevCIBuildPRTemplate(t *testing.T) {
	validdevCIPRTemplate := triggersv1.TriggerTemplate{
		TypeMeta: triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("testns", "dev-ci-build-from-pr-template"),
			statusTrackerAnnotations("dev-ci-build-from-pr", "Dev CI Build")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				pipelinev1.ParamSpec{
					Name:        "gitref",
					Description: "The git branch for this PR",
				},
				pipelinev1.ParamSpec{
					Name:        "gitsha",
					Description: "the specific commit SHA.",
				},
				pipelinev1.ParamSpec{
					Name:        "gitrepositoryurl",
					Description: "The git repository url",
				},

				pipelinev1.ParamSpec{
					Name:        "fullname",
					Description: "The GitHub repository for this PullRequest.",
				},
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createDevCIResourceTemplate(serviceAccName, "example.com:5000/testing/testing"),
				},
			},
		},
	}
	template := createDevCIBuildPRTemplate("testns", serviceAccName, "example.com:5000/testing/testing")
	if diff := cmp.Diff(validdevCIPRTemplate, template); diff != "" {
		t.Fatalf("CreatedevCIBuildPRTemplate failed:\n%s", diff)
	}
}

func TestCreateStageCDPushTemplate(t *testing.T) {
	ValidStageCDPushTemplate := triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("testns", "stage-cd-deploy-from-push-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				pipelinev1.ParamSpec{
					Name:        "gitref",
					Description: "The git revision",
					Default: &pipelinev1.ArrayOrString{
						StringVal: "master",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.ParamSpec{
					Name:        "gitrepositoryurl",
					Description: "The git repository url",
				},
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createStageCDResourceTemplate(serviceAccName),
				},
			},
		},
	}
	template := createStageCDPushTemplate("testns", serviceAccName)
	if diff := cmp.Diff(ValidStageCDPushTemplate, template); diff != "" {
		t.Fatalf("createStageCDPushTemplate failed:\n%s", diff)
	}
}

func TestCreateStageCIDryRunTemplate(t *testing.T) {
	validStageCIDryRunTemplate := triggersv1.TriggerTemplate{
		TypeMeta: triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("testns", "stage-ci-dryrun-from-pr-template"),
			statusTrackerAnnotations("stage-ci-dryrun-from-pr", "Stage CI Dry Run")),

		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				pipelinev1.ParamSpec{
					Name:        "gitref",
					Description: "The git revision",
					Default: &pipelinev1.ArrayOrString{
						StringVal: "master",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.ParamSpec{
					Name:        "gitrepositoryurl",
					Description: "The git repository url",
				},
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createStageCIResourceTemplate(serviceAccName),
				},
			},
		},
	}
	template := createStageCIDryRunTemplate("testns", serviceAccName)
	if diff := cmp.Diff(validStageCIDryRunTemplate, template); diff != "" {
		t.Fatalf("createStageCIdryrunptemplate failed:\n%s", diff)
	}

}

func TestStatusTrackerAnnotations(t *testing.T) {
	om := meta.ObjectMeta(meta.NamespacedName("test-ns", "name"))
	statusTrackerAnnotations("my-pipeline", "description")(&om)

	want := v1.ObjectMeta{
		Name:      "name",
		Namespace: "test-ns",
		Annotations: map[string]string{
			"tekton.dev/git-status":         "true",
			"tekton.dev/status-context":     "my-pipeline",
			"tekton.dev/status-description": "description",
		},
	}

	if diff := cmp.Diff(want, om); diff != "" {
		t.Fatalf("statusTrackerAnnotations failed:\n%s", diff)
	}
}
