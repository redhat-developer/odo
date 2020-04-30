package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var (
	sName = "pipeline"
)

func TestCreateDevCDPipelineRun(t *testing.T) {
	validDevCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-cd-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("app-cd-pipeline"),
			Resources:          createDevResource("example.com:5000/testing/testing:$(params.gitsha)", "$(params.gitsha)"),
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
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-ci-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("app-ci-pipeline"),
			Params: []pipelinev1.Param{
				createBindingParam("REPO", "$(params.fullname)"),
				createBindingParam("COMMIT_SHA", "$(params.gitsha)"),
			},
			Resources: createDevResource("example.com:5000/testing/testing:$(params.gitref)-$(params.gitsha)", "$(params.gitref)"),
		},
	}
	template := createDevCIPipelineRun(sName, "example.com:5000/testing/testing")
	if diff := cmp.Diff(validDevCIPipelineRun, template); diff != "" {
		t.Fatalf("createDevCIPipelineRun failed:\n%s", diff)
	}
}

func TestCreateCDPipelineRun(t *testing.T) {
	validStageCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "cd-deploy-from-push-pipeline-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("cd-deploy-from-push-pipeline"),
			Resources:          createResources(),
		},
	}
	template := createCDPipelineRun(sName)
	if diff := cmp.Diff(validStageCDPipeline, template); diff != "" {
		t.Fatalf("createCDPipelineRun failed:\n%s", diff)
	}
}

func TestCreateStageCIPipelineRun(t *testing.T) {
	validStageCIPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "ci-dryrun-from-pr-pipeline-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("ci-dryrun-from-pr-pipeline"),
			Resources:          createResources(),
		},
	}
	template := createCIPipelineRun(sName)
	if diff := cmp.Diff(validStageCIPipeline, template); diff != "" {
		t.Fatalf("createCIPipelineRun failed:\n%s", diff)
	}
}
