package statustracker

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/deployment"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/roles"
)

func TestCreateStatusTrackerDeployment(t *testing.T) {
	deploy := createStatusTrackerDeployment("dana-cicd")

	want := &appsv1.Deployment{
		TypeMeta:   meta.TypeMeta("Deployment", "apps/v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("dana-cicd", operatorName)),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr32(1),
			Selector: labelSelector(deployment.KubernetesAppNameLabel, operatorName),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						deployment.KubernetesAppNameLabel: operatorName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: operatorName,
					Containers: []corev1.Container{
						{
							Name:            operatorName,
							Image:           containerImage,
							Command:         []string{operatorName},
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
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
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, deploy); diff != "" {
		t.Fatalf("deployment diff: %s", diff)
	}
}

func TestResource(t *testing.T) {
	defer func(f secretSealer) {
		defaultSecretSealer = f
	}(defaultSecretSealer)

	testSecret := &ssv1alpha1.SealedSecret{}
	defaultSecretSealer = func(ns types.NamespacedName, data, secretKey string) (*ssv1alpha1.SealedSecret, error) {
		return testSecret, nil
	}

	ns := "my-test-ns"
	res, err := Resources(ns, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	name := meta.NamespacedName(ns, operatorName)
	sa := roles.CreateServiceAccount(name)
	want := []interface{}{
		sa,
		testSecret,
		roles.CreateRole(name, roleRules),
		roles.CreateRoleBinding(name, sa, "Role", operatorName),
		createStatusTrackerDeployment(ns),
	}

	if diff := cmp.Diff(want, res); diff != "" {
		t.Fatalf("deployment diff: %s", diff)
	}
}
