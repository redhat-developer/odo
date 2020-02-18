package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoleBinding(t *testing.T) {
	roleBinding := v1rbac.RoleBinding{
		TypeMeta: v1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "tekton-triggers-openshift-binding",
		},
		Subjects: []v1rbac.Subject{
			v1rbac.Subject{
				Kind: "ServiceAccount",
				Name: "demo-sa",
			},
		},
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "tekton-triggers-openshift-demo",
		},
	}
	roleBindingTask := createRoleBinding("tekton-triggers-openshift-binding", "demo-sa", "Role", "tekton-triggers-openshift-demo")
	if diff := cmp.Diff(roleBinding, roleBindingTask); diff != "" {
		t.Errorf("GenerateGithubStatusTask() failed:\n%s", diff)
	}

}

func TestCreateRole(t *testing.T) {
	role := v1rbac.Role{
		TypeMeta: v1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "tekton-triggers-openshift-demo",
		},
		Rules: []v1rbac.PolicyRule{
			v1rbac.PolicyRule{
				APIGroups: []string{"tekton.dev"},
				Resources: []string{"eventlisteners", "triggerbindings", "triggertemplates", "tasks", "taskruns"},
				Verbs:     []string{"get"},
			},
			v1rbac.PolicyRule{
				APIGroups: []string{"tekton.dev"},
				Resources: []string{"pipelineruns", "pipelineresources", "taskruns"},
				Verbs:     []string{"create"},
			},
		},
	}
	roleTask := createRole("tekton-triggers-openshift-demo", rules)
	if diff := cmp.Diff(role, roleTask); diff != "" {
		t.Errorf("GenerateGithubStatusTask() failed:\n%s", diff)
	}

}

func ServiceAccountTest(t *testing.T) {

	serviceAccount := corev1.ServiceAccount{
		TypeMeta: v1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "demo-sa",
		},
		Secrets: []corev1.ObjectReference{
			corev1.ObjectReference{
				Name: "regcred",
			},
		},
	}
	servicetask := createServiceAccount("demo-sa", "regcred")
	if diff := cmp.Diff(servicetask, serviceAccount); diff != "" {
		t.Errorf("GenerateGithubStatusTask() failed:\n%s", diff)
	}

}
