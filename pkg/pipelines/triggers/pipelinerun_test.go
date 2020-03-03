package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

var (
	sName = "pipeline"
)

func TestCreateDevCDPipelineRun(t *testing.T) {
	validDevCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("dev-cd-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("dev-cd-pipeline"),
			Resources:          createDevResource("example.com:5000/testing/testing:$(params.gitref)"),
		},
	}
	template := createDevCDPipelineRun(sName, "example.com:5000/testing/testing")
	if diff := cmp.Diff(validDevCDPipeline, template); diff != "" {
		t.Fatalf("createDevCDPipelineRun failed:\n%s", diff)
	}

}

func TestCreateDevCIPipelineRun(t *testing.T) {
	validDevCIPipelineRun := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("dev-ci-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("dev-ci-pipeline"),
			Params: []pipelinev1.Param{
				createBindingParam("REPO", "$(params.fullname)"),
				createBindingParam("COMMIT_SHA", "$(params.gitsha)"),
			},
			Resources: createDevResource("example.com:5000/testing/testing:$(params.gitref)-$(params.gitsha)"),
		},
	}
	template := createDevCIPipelineRun(sName, "example.com:5000/testing/testing")
	if diff := cmp.Diff(validDevCIPipelineRun, template); diff != "" {
		t.Fatalf("createDevCIPipelineRun failed:\n%s", diff)
	}
}

func TestCreateStageCDPipelineRun(t *testing.T) {
	validStageCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("stage-cd-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("stage-cd-pipeline"),
			Resources:          createStageResources(),
		},
	}
	template := createStageCDPipelineRun(sName)
	if diff := cmp.Diff(validStageCDPipeline, template); diff != "" {
		t.Fatalf("createStageCDPipelineRun failed:\n%s", diff)
	}
}

func TestCreateStageCIPipelineRun(t *testing.T) {
	validStageCIPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("stage-ci-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("stage-ci-pipeline"),
			Resources:          createStageResources(),
		},
	}
	template := createStageCIPipelineRun(sName)
	if diff := cmp.Diff(validStageCIPipeline, template); diff != "" {
		t.Fatalf("createStageCIPipelineRun failed:\n%s", diff)
	}
}
