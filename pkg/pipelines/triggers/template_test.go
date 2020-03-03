package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

var (
	serviceAccName = "pipeline"
)

func TestCreateDevCDDeployTemplate(t *testing.T) {
	validDevCDTemplate := triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta("testns", "dev-cd-deploy-from-master-template"),
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

func TestCreatedevCIBuildPRTemplate(t *testing.T) {
	validdevCIPRTemplate := triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta("testns", "dev-ci-build-from-pr-template"),

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
		ObjectMeta: meta.CreateObjectMeta("testns", "stage-cd-deploy-from-push-template"),
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

func TestCreateStageCIdryrunptemplate(t *testing.T) {
	validStageCIdryrunTemplate := triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta("testns", "stage-ci-dryrun-from-pr-template"),

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
	template := createStageCIdryrunptemplate("testns", serviceAccName)
	if diff := cmp.Diff(validStageCIdryrunTemplate, template); diff != "" {
		t.Fatalf("createStageCIdryrunptemplate failed:\n%s", diff)
	}

}
