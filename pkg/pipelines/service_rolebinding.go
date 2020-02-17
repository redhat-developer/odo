package pipelines

import (
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateRoleBinding() v1rbac.RoleBinding {
	roleBinding := v1rbac.RoleBinding{
		TypeMeta:   CreateTaskTypeMeta("RoleBinding", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: CreateTaskObjectMeta("tekton-triggers-openshift-binding"),
		Subjects:   ReturnRoleBindingSubjects(),
		RoleRef:    ReturnRoleBindingRef(),
	}
	return roleBinding
}

func ReturnRoleBindingSubjects() []v1rbac.Subject {
	return []v1rbac.Subject{
		v1rbac.Subject{
			Kind: "ServiceAccount",
			Name: "demo-sa",
		},
	}
}

func ReturnRoleBindingRef() v1rbac.RoleRef {
	return v1rbac.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "tekton-triggers-openshift-demo",
	}
}

func GenerateRole() v1rbac.Role {
	role := v1rbac.Role{
		TypeMeta:   CreateTaskTypeMeta("Role", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: CreateTaskObjectMeta("tekton-triggers-openshift-demo"),
		Rules:      GenerateRoleRules(),
	}
	return role
}

func GenerateRoleRules() []v1rbac.PolicyRule {
	return []v1rbac.PolicyRule{
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
	}
}

func GenerateServiceAccount() corev1.ServiceAccount {

	serviceAccount := corev1.ServiceAccount{
		TypeMeta:   CreateTaskTypeMeta("ServiceAccount", "v1"),
		ObjectMeta: CreateTaskObjectMeta("demo-sa"),
		Secrets:    createSecretsServiceAccount(),
	}

	return serviceAccount
}

func createSecretsServiceAccount() []corev1.ObjectReference {
	return []corev1.ObjectReference{
		corev1.ObjectReference{
			Name: "regcred",
		},
	}
}

func CreateTaskTypeMeta(kind string, api string) v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       kind,
		APIVersion: api,
	}
}

func CreateTaskObjectMeta(name string) v1.ObjectMeta {
	return v1.ObjectMeta{
		Name: name,
	}
}
