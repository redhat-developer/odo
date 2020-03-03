package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func TestCreateDevCIPipeline(t *testing.T) {

	DevCIpipeline := createDevCIPipeline(meta.NamespacedName("cicd-environment", "dev-ci-pipeline"))

	want := &pipelinev1.Pipeline{
		TypeMeta: PipelineTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-ci-pipeline",
			Namespace: "cicd-environment",
		},
		Spec: pipelinev1.PipelineSpec{
			Params: []pipelinev1.ParamSpec{
				pipelinev1.ParamSpec{
					Name: "REPO",
					Type: "string",
				},
				pipelinev1.ParamSpec{
					Name: "COMMIT_SHA",
					Type: "string",
				},
			},
			Resources: []pipelinev1.PipelineDeclaredResource{
				pipelinev1.PipelineDeclaredResource{
					Name: "source-repo",
					Type: "git",
				},
				pipelinev1.PipelineDeclaredResource{
					Name: "runtime-image",
					Type: "image",
				},
			},
			Tasks: []pipelinev1.PipelineTask{

				pipelinev1.PipelineTask{
					Name: "build-image",
					TaskRef: &pipelinev1.TaskRef{
						Name: "buildah-task",
					},
					Resources: &pipelinev1.PipelineTaskResources{
						Inputs: []pipelinev1.PipelineTaskInputResource{
							{Name: "source",
								Resource: "source-repo"},
						},
						Outputs: []pipelinev1.PipelineTaskOutputResource{
							{Name: "image",
								Resource: "runtime-image"},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, DevCIpipeline); diff != "" {
		t.Fatalf("TestCreateDevCIPipeline() failed got\n%s", diff)
	}
}

func TestCreateStageCIPipeline(t *testing.T) {

	stageCIpipeline := createStageCIPipeline(meta.NamespacedName("cicd-environment", "stage-ci-pipeline"), "stage-environment")

	want := &pipelinev1.Pipeline{
		TypeMeta: PipelineTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stage-ci-pipeline",
			Namespace: "cicd-environment",
		},
		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				pipelinev1.PipelineDeclaredResource{
					Name: "source-repo",
					Type: "git",
				},
			},

			Tasks: []pipelinev1.PipelineTask{
				pipelinev1.PipelineTask{
					Name: "apply-source",
					TaskRef: &pipelinev1.TaskRef{
						Name: "deploy-from-source-task",
					},
					Resources: &pipelinev1.PipelineTaskResources{
						Inputs: []pipelinev1.PipelineTaskInputResource{
							{Name: "source",
								Resource: "source-repo"},
						},
					},
					Params: []pipelinev1.Param{
						{Name: "NAMESPACE", Value: pipelinev1.ArrayOrString{Type: "string", StringVal: "stage-environment"}},
						{Name: "DRYRUN", Value: pipelinev1.ArrayOrString{Type: "string", StringVal: "true"}},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, stageCIpipeline); diff != "" {
		t.Fatalf("TestcreateStageCIPipeline() failed got\n%s", diff)
	}

}

func TestCreateDevCDPipeline(t *testing.T) {
	DevCDpipeline := createDevCDPipeline(meta.NamespacedName("cicd-environment", "dev-cd-pipeline"), "usr/path/", "dev-environment")
	want := &pipelinev1.Pipeline{
		TypeMeta: PipelineTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-cd-pipeline",
			Namespace: "cicd-environment",
		},

		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				pipelinev1.PipelineDeclaredResource{
					Name: "source-repo",
					Type: "git",
				},
				pipelinev1.PipelineDeclaredResource{
					Name: "runtime-image",
					Type: "image",
				},
			},
			Tasks: []pipelinev1.PipelineTask{

				pipelinev1.PipelineTask{
					Name: "build-image",
					TaskRef: &pipelinev1.TaskRef{
						Name: "buildah-task",
					},

					Resources: &pipelinev1.PipelineTaskResources{
						Inputs: []pipelinev1.PipelineTaskInputResource{
							{Name: "source",
								Resource: "source-repo"},
						},
						Outputs: []pipelinev1.PipelineTaskOutputResource{
							{Name: "image",
								Resource: "runtime-image"},
						},
					},
				},

				pipelinev1.PipelineTask{
					Name: "deploy-image",
					TaskRef: &pipelinev1.TaskRef{
						Name: "deploy-using-kubectl-task",
					},
					RunAfter: []string{"build-image"},
					Resources: &pipelinev1.PipelineTaskResources{
						Inputs: []pipelinev1.PipelineTaskInputResource{
							{Name: "source", Resource: "source-repo"},
							{Name: "image", Resource: "runtime-image"},
						},
					},
					Params: []pipelinev1.Param{
						{Name: "PATHTODEPLOYMENT", Value: pipelinev1.ArrayOrString{Type: "string", StringVal: "usr/path/"}},
						{Name: "YAMLPATHTOIMAGE", Value: pipelinev1.ArrayOrString{Type: "string", StringVal: "spec.template.spec.containers[0].image"}},
						{Name: "NAMESPACE", Value: pipelinev1.ArrayOrString{Type: "string", StringVal: "dev-environment"}},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, DevCDpipeline); diff != "" {
		t.Fatalf("TestCreateDevCDPipeline() failed got\n%s", diff)
	}
}

func TestCreateStageCDPipeline(t *testing.T) {
	stageCDpipeline := createStageCDPipeline(meta.NamespacedName("cicd-environment", "stage-cd-pipeline"), "stage-environment")
	want := &pipelinev1.Pipeline{
		TypeMeta: PipelineTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stage-cd-pipeline",
			Namespace: "cicd-environment",
		},
		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				pipelinev1.PipelineDeclaredResource{
					Name: "source-repo",
					Type: "git",
				},
			},

			Tasks: []pipelinev1.PipelineTask{
				pipelinev1.PipelineTask{
					Name: "apply-source",
					TaskRef: &pipelinev1.TaskRef{
						Name: "deploy-from-source-task",
					},
					Resources: &pipelinev1.PipelineTaskResources{
						Inputs: []pipelinev1.PipelineTaskInputResource{
							{Name: "source",
								Resource: "source-repo"},
						},
					},
					Params: []pipelinev1.Param{
						{Name: "NAMESPACE", Value: pipelinev1.ArrayOrString{Type: "string", StringVal: "stage-environment"}},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, stageCDpipeline); diff != "" {
		t.Fatalf("TestcreateStageCSPipeline() failed got\n%s", diff)
	}

}
