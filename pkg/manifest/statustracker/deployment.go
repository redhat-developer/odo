package statustracker

import (
	"fmt"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openshift/odo/pkg/manifest/deployment"
	"github.com/openshift/odo/pkg/manifest/meta"
	"github.com/openshift/odo/pkg/manifest/roles"
	"github.com/openshift/odo/pkg/manifest/secrets"
)

const (
	operatorName   = "commit-status-tracker"
	containerImage = "quay.io/redhat-developer/commit-status-tracker:v0.0.1"
)

type secretSealer = func(types.NamespacedName, string, string) (*ssv1alpha1.SealedSecret, error)

var defaultSecretSealer secretSealer = secrets.CreateSealedSecret

var (
	roleRules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "services", "services/finalizers", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"monitoring.coreos.com"},
			Resources: []string{"servicemonitors"},
			Verbs:     []string{"get", "create"},
		},
		{
			APIGroups:     []string{"apps"},
			Resources:     []string{"deployments/finalizers"},
			ResourceNames: []string{operatorName},
			Verbs:         []string{"update"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"replicasets", "deployments"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"tekton.dev"},
			Resources: []string{"pipelineruns"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}

	statusTrackerEnv = []corev1.EnvVar{
		{
			Name: "WATCH_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "OPERATOR_NAME",
			Value: operatorName,
		},
	}
)

func createStatusTrackerDeployment(ns string) *appsv1.Deployment {
	return deployment.Create(ns, operatorName, containerImage,
		deployment.ServiceAccount(operatorName),
		deployment.Env(statusTrackerEnv),
		deployment.Command([]string{operatorName}))
}

func createRoleBinding(ns string, roleKind, roleName string, subjects []rbacv1.Subject) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta:   meta.TypeMeta("RoleBinding", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, operatorName)),
		Subjects:   subjects,
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
	}
}

// Resources returns a list of newly created resources that are required start
// the status-tracker service.
func Resources(ns, token string) ([]interface{}, error) {
	name := meta.NamespacedName(ns, operatorName)
	sa := roles.CreateServiceAccount(name)

	githubAuth, err := defaultSecretSealer(meta.NamespacedName(ns, "commit-status-tracker-git-secret"), token, "token")
	if err != nil {
		return nil, fmt.Errorf("failed to generate Status Tracker Secret: %w", err)
	}
	return []interface{}{
		sa,
		githubAuth,
		roles.CreateRole(name, roleRules),
		roles.CreateRoleBinding(name, sa, "Role", operatorName),
		createStatusTrackerDeployment(ns),
	}, nil
}

func ptr32(i int32) *int32 {
	return &i
}

func labelSelector(name, value string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			name: value,
		},
	}
}
