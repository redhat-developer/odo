package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoleBinding(t *testing.T) {
	want := &v1rbac.RoleBinding{
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
	sa := &corev1.ServiceAccount{
		TypeMeta:   typeMeta("ServiceAccount", "v1"),
		ObjectMeta: objectMeta(namespacedName("testing", "demo-sa")),
	}
	roleBindingTask := createRoleBinding(
		namespacedName("", "tekton-triggers-openshift-binding"),
		sa, "Role", "tekton-triggers-openshift-demo")
	if diff := cmp.Diff(want, roleBindingTask); diff != "" {
		t.Errorf("TestRoleBinding() failed:\n%s", diff)
	}

}

func TestCreateRole(t *testing.T) {
	want := &v1rbac.Role{
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
	roleTask := createRole(namespacedName("", "tekton-triggers-openshift-demo"), rules)
	if diff := cmp.Diff(roleTask, want); diff != "" {
		t.Errorf("TestCreateRole() failed:\n%s", diff)
	}
}

func TestServiceAccount(t *testing.T) {
	want := &corev1.ServiceAccount{
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
	servicetask := createServiceAccount(namespacedName("", "demo-sa"), "regcred")
	if diff := cmp.Diff(servicetask, want); diff != "" {
		t.Errorf("TestServiceAccount() failed:\n%s", diff)
	}
}
