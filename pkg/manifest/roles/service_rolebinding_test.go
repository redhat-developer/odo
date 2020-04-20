package roles

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/manifest/meta"
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	roleBindingName = "my-test-role-binding"
	roleName        = "test-role"
)

var testRules = []v1rbac.PolicyRule{
	{
		APIGroups: []string{"tekton.dev"},
		Resources: []string{"eventlisteners", "triggerbindings", "triggertemplates", "tasks", "taskruns"},
		Verbs:     []string{"get"},
	},
}

func TestClusterRoleBinding(t *testing.T) {
	want := &v1rbac.ClusterRoleBinding{
		TypeMeta: clusterRoleBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name: "test-clusterbinding",
		},
		Subjects: []v1rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "pipeline",
				Namespace: "cicd",
			},
		},
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "edit",
		},
	}
	sa := &corev1.ServiceAccount{
		TypeMeta:   serviceAccountTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("cicd", "pipeline")),
	}
	got := CreateClusterRoleBinding(meta.NamespacedName("", "test-clusterbinding"), sa, "ClusterRole", "edit")
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("CreateClusterRoleBinding() failed:%v\n", diff)
	}
}

func TestRoleBinding(t *testing.T) {
	want := &v1rbac.RoleBinding{
		TypeMeta: roleBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name: roleBindingName,
		},
		Subjects: []v1rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "pipeline",
				Namespace: "testing",
			},
		},
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}
	sa := &corev1.ServiceAccount{
		TypeMeta:   serviceAccountTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("testing", "pipeline")),
	}
	roleBindingTask := CreateRoleBinding(
		meta.NamespacedName("", roleBindingName),
		sa, "Role", roleName)
	if diff := cmp.Diff(want, roleBindingTask); diff != "" {
		t.Errorf("TestRoleBinding() failed:\n%s", diff)
	}

}

func TestRoleBindingForSubjects(t *testing.T) {
	want := &v1rbac.RoleBinding{
		TypeMeta: roleBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: "testns",
		},
		Subjects: []v1rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "pipeline",
				Namespace: "testing",
			},
			{
				Kind:      "ServiceAccount",
				Name:      "pipeline",
				Namespace: "testing2",
			},
		},
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}

	roleBinding := CreateRoleBindingForSubjects(meta.NamespacedName("testns", roleBindingName), "Role", roleName,
		[]v1rbac.Subject{{Kind: "ServiceAccount", Name: "pipeline", Namespace: "testing"},
			{Kind: "ServiceAccount", Name: "pipeline", Namespace: "testing2"},
		})

	if diff := cmp.Diff(want, roleBinding); diff != "" {
		t.Errorf("TestRoleBindingForSubjects() failed:\n%s", diff)
	}

}

func TestCreateRole(t *testing.T) {
	want := &v1rbac.Role{
		TypeMeta: roleTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name: roleName,
		},
		Rules: testRules,
	}
	roleTask := CreateRole(meta.NamespacedName("", roleName), testRules)
	if diff := cmp.Diff(roleTask, want); diff != "" {
		t.Errorf("TestCreateRole() failed:\n%s", diff)
	}
}

func TestServiceAccount(t *testing.T) {
	want := &corev1.ServiceAccount{
		TypeMeta: serviceAccountTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline",
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "regcred",
			},
		},
	}
	servicetask := CreateServiceAccount(meta.NamespacedName("", "pipeline"))
	servicetask = AddSecretToSA(servicetask, "regcred")
	if diff := cmp.Diff(servicetask, want); diff != "" {
		t.Errorf("TestServiceAccount() failed:\n%s", diff)
	}
}

func TestAddSecretToSA(t *testing.T) {
	validSA := &corev1.ServiceAccount{
		TypeMeta: serviceAccountTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name: "pipeline",
		},
	}
	validSecrets := []corev1.ObjectReference{
		{
			Name: "regcred",
		},
	}
	sa := AddSecretToSA(validSA, "regcred")
	if diff := cmp.Diff(sa.Secrets, validSecrets); diff != "" {
		t.Errorf("addSecretToSA() failed:\n%s", diff)
	}
}

func TestCreateClusterRole(t *testing.T) {
	validClusterRole := &v1rbac.ClusterRole{
		TypeMeta: clusterRoleTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name: ClusterRoleName,
		},
		Rules: testRules,
	}
	clusterRole := CreateClusterRole(meta.NamespacedName("", ClusterRoleName), testRules)
	if diff := cmp.Diff(validClusterRole, clusterRole); diff != "" {
		t.Fatalf("createClusterRole() failed:\n%v", diff)
	}
}
