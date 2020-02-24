package pipelines

import (
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
)

// createServiceAccount creates a ServiceAccount given name and secretName
func createServiceAccount(name apitypes.NamespacedName, secretName string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta:   typeMeta("ServiceAccount", "v1"),
		ObjectMeta: objectMeta(name),
		Secrets: []corev1.ObjectReference{
			corev1.ObjectReference{Name: secretName},
		},
	}
}

// createRoleBinding creates a RoleBinding given name, sa, roleKind, and roleName
func createRoleBinding(name apitypes.NamespacedName, sa *corev1.ServiceAccount, roleKind, roleName string) *v1rbac.RoleBinding {
	return &v1rbac.RoleBinding{
		TypeMeta:   typeMeta("RoleBinding", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: objectMeta(name),
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
func createRole(name apitypes.NamespacedName, policyRules []v1rbac.PolicyRule) *v1rbac.Role {
	return &v1rbac.Role{
		TypeMeta:   typeMeta("Role", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: objectMeta(name),
		Rules:      policyRules,
	}
}
