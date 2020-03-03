package pipelines

import (
	"github.com/openshift/odo/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createServiceAccount(name types.NamespacedName) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta:   meta.TypeMeta("ServiceAccount", "v1"),
		ObjectMeta: meta.ObjectMeta(name),
		Secrets:    []corev1.ObjectReference{},
	}
}

func addSecretToSA(sa *corev1.ServiceAccount, secretName string) *corev1.ServiceAccount {
	sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secretName})
	return sa
}

// createRoleBinding creates a RoleBinding given name, sa, roleKind, and roleName
func createRoleBinding(name types.NamespacedName, sa *corev1.ServiceAccount, roleKind, roleName string) *v1rbac.RoleBinding {
	return &v1rbac.RoleBinding{
		TypeMeta:   meta.TypeMeta("RoleBinding", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: meta.ObjectMeta(name),
		Subjects: []v1rbac.Subject{
			v1rbac.Subject{
				Kind: sa.Kind,
				Name: sa.Name,
			},
		},
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
	}
}

// createRole creates a Role given a name and policyRules
func createRole(name types.NamespacedName, policyRules []v1rbac.PolicyRule) *v1rbac.Role {
	return &v1rbac.Role{
		TypeMeta:   meta.TypeMeta("Role", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: meta.ObjectMeta(name),
		Rules:      policyRules,
	}
}
