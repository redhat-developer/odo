package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func TestCreateDevCDPipelineRun(t *testing.T) {
	validDevCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("dev-cd-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("dev-cd-pipeline"),
			Resources:          createDevResource(),
		},
	}
	template := createDevCDPipelineRun()
	if diff := cmp.Diff(validDevCDPipeline, template); diff != "" {
		t.Fatalf("createDevCDPipelineRun failed:\n%s", diff)
	}

}

func TestCreateDevCIPipelineRun(t *testing.T) {
	validDevCIPipelineRun := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("dev-ci-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("dev-ci-pipeline"),
			Resources:          createDevResource(),
		},
	}
	template := createDevCIPipelineRun()
	if diff := cmp.Diff(validDevCIPipelineRun, template); diff != "" {
		t.Fatalf("createDevCIPipelineRun failed:\n%s", diff)
	}
}

func TestCreateStageCDPipelineRUn(t *testing.T) {
	validStageCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("stage-cd-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("stage-ci-pipeline"),
			Resources:          createStageResources(),
		},
	}
	template := createStageCDPipelineRun()
	if diff := cmp.Diff(validStageCDPipeline, template); diff != "" {
		t.Fatalf("createStageCDPipelineRun failed:\n%s", diff)
	}
}

func TestCreateStageCIPipelineRun(t *testing.T) {
	validStageCIPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("stage-ci-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("stage-ci-pipeline"),
			Resources:          createStageResources(),
		},
	}
	template := createStageCIPipelineRun()
	if diff := cmp.Diff(validStageCIPipeline, template); diff != "" {
		t.Fatalf("createStageCIPipelineRun failed:\n%s", diff)
	}
}
