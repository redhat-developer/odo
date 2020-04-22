package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/odo/pkg/manifest/meta"
)

// ServiceAccount is an option that configures the deployment's pods to execute
// with the provided service account name.
func ServiceAccount(sa string) podSpecFunc {
	return func(c *corev1.PodSpec) {
		c.ServiceAccountName = sa
	}
}

// Env adds an environment to the first container in the PodSpec.
func Env(env []corev1.EnvVar) podSpecFunc {
	return func(c *corev1.PodSpec) {
		c.Containers[0].Env = env
	}
}

// Command configures the command for the first container in the PodSpec.
func Command(s []string) podSpecFunc {
	return func(c *corev1.PodSpec) {
		c.Containers[0].Command = s
	}
}

// ContainerPort configures a port for the first container as a ContainerPort
// with the specified port number.
func ContainerPort(p int32) podSpecFunc {
	return func(c *corev1.PodSpec) {
		c.Containers[0].Ports = []corev1.ContainerPort{
			{ContainerPort: p},
		}
	}
}

// Create creates and returns a Deployment with the specified configuration.
func Create(ns, name, image string, opts ...podSpecFunc) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta:   meta.TypeMeta("Deployment", "apps/v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, name)),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr32(1),
			Selector: labelSelector("name", name),
			Template: podTemplate(name, image, opts...),
		},
	}
}

type podSpecFunc func(t *corev1.PodSpec)

func podTemplate(name, image string, opts ...podSpecFunc) corev1.PodTemplateSpec {
	podSpec := &corev1.PodSpec{
		ServiceAccountName: "default",
		Containers: []corev1.Container{
			{
				Name:            name,
				Image:           image,
				ImagePullPolicy: corev1.PullAlways,
			},
		},
	}

	for _, o := range opts {
		o(podSpec)
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"name": name,
			},
		},
		Spec: *podSpec,
	}
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
