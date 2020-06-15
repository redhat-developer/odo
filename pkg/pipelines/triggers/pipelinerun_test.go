package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

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
			Resources:          createDevResource("$(params.gitsha)"),
		},
	}
	template := createDevCDPipelineRun(sName)
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
				createPipelineBindingParam("REPO", "$(params.fullname)"),
				createPipelineBindingParam("COMMIT_SHA", "$(params.gitsha)"),
				createPipelineBindingParam("TLSVERIFY", "$(params.tlsVerify)"),
			},
			Resources: createDevResource("$(params.gitref)"),
		},
	}
	template := createDevCIPipelineRun(sName)
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

func TestCreateDevResource(t *testing.T) {
	want := []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelinev1alpha1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", "test"),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
		{
			Name: "runtime-image",
			ResourceSpec: &pipelinev1alpha1.PipelineResourceSpec{
				Type: "image",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("url", "$(params.imageRepo):$(params.gitref)-$(params.gitsha)"),
				},
			},
		},
	}
	got := createDevResource("test")
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("createDevResource() failed: \n%s", diff)
	}
}
