package generator

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var fakeResources corev1.ResourceRequirements

func init() {
	fakeResources = *fakeResourceRequirements()
}

func TestGenerateContainer(t *testing.T) {

	tests := []struct {
		name         string
		image        string
		isPrivileged bool
		command      []string
		args         []string
		envVars      []corev1.EnvVar
		resourceReqs corev1.ResourceRequirements
		ports        []corev1.ContainerPort
	}{
		{
			name:         "",
			image:        "",
			isPrivileged: false,
			command:      []string{},
			args:         []string{},
			envVars:      []corev1.EnvVar{},
			resourceReqs: corev1.ResourceRequirements{},
			ports:        []corev1.ContainerPort{},
		},
		{
			name:         "container1",
			image:        "quay.io/eclipse/che-java8-maven:nightly",
			isPrivileged: true,
			command:      []string{"tail"},
			args:         []string{"-f", "/dev/null"},
			envVars: []corev1.EnvVar{
				{
					Name:  "test",
					Value: "123",
				},
			},
			resourceReqs: fakeResources,
			ports: []corev1.ContainerPort{
				{
					Name:          "port-9090",
					ContainerPort: 9090,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerParams := ContainerParams{
				Name:         tt.name,
				Image:        tt.image,
				IsPrivileged: tt.isPrivileged,
				Command:      tt.command,
				Args:         tt.args,
				EnvVars:      tt.envVars,
				ResourceReqs: tt.resourceReqs,
				Ports:        tt.ports,
			}
			container := GenerateContainer(containerParams)

			if container.Name != tt.name {
				t.Errorf("expected %s, actual %s", tt.name, container.Name)
			}

			if container.Image != tt.image {
				t.Errorf("expected %s, actual %s", tt.image, container.Image)
			}

			if tt.isPrivileged {
				if *container.SecurityContext.Privileged != tt.isPrivileged {
					t.Errorf("expected %t, actual %t", tt.isPrivileged, *container.SecurityContext.Privileged)
				}
			} else if tt.isPrivileged == false && container.SecurityContext != nil {
				t.Errorf("expected security context to be nil but it was defined")
			}

			if len(container.Command) != len(tt.command) {
				t.Errorf("expected %d, actual %d", len(tt.command), len(container.Command))
			} else {
				for i := range container.Command {
					if container.Command[i] != tt.command[i] {
						t.Errorf("expected %s, actual %s", tt.command[i], container.Command[i])
					}
				}
			}

			if len(container.Args) != len(tt.args) {
				t.Errorf("expected %d, actual %d", len(tt.args), len(container.Args))
			} else {
				for i := range container.Args {
					if container.Args[i] != tt.args[i] {
						t.Errorf("expected %s, actual %s", tt.args[i], container.Args[i])
					}
				}
			}

			if len(container.Env) != len(tt.envVars) {
				t.Errorf("expected %d, actual %d", len(tt.envVars), len(container.Env))
			} else {
				for i := range container.Env {
					if container.Env[i].Name != tt.envVars[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.envVars[i].Name, container.Env[i].Name)
					}
					if container.Env[i].Value != tt.envVars[i].Value {
						t.Errorf("expected value %s, actual value %s", tt.envVars[i].Value, container.Env[i].Value)
					}
				}
			}

			if len(container.Ports) != len(tt.ports) {
				t.Errorf("expected %d, actual %d", len(tt.ports), len(container.Ports))
			} else {
				for i := range container.Ports {
					if container.Ports[i].Name != tt.ports[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.ports[i].Name, container.Ports[i].Name)
					}
					if container.Ports[i].ContainerPort != tt.ports[i].ContainerPort {
						t.Errorf("expected port number is %v, actual %v", tt.ports[i].ContainerPort, container.Ports[i].ContainerPort)
					}
				}
			}

		})
	}
}

func TestGeneratePodTemplateSpec(t *testing.T) {

	container := &corev1.Container{
		Name:            "container1",
		Image:           "image1",
		ImagePullPolicy: corev1.PullAlways,

		Command: []string{"tail"},
		Args:    []string{"-f", "/dev/null"},
		Env:     []corev1.EnvVar{},
	}

	tests := []struct {
		podName        string
		namespace      string
		serviceAccount string
		labels         map[string]string
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {

			objectMeta := CreateObjectMeta(tt.podName, tt.namespace, tt.labels, nil)
			podTemplateSpecParams := PodTemplateSpecParams{
				ObjectMeta: objectMeta,
				Containers: []corev1.Container{*container},
				// Volumes:    utils.GetOdoContainerVolumes(),
			}
			podTemplateSpec := GeneratePodTemplateSpec(podTemplateSpecParams)

			if podTemplateSpec.Name != tt.podName {
				t.Errorf("expected %s, actual %s", tt.podName, podTemplateSpec.Name)
			}

			if podTemplateSpec.Namespace != tt.namespace {
				t.Errorf("expected %s, actual %s", tt.namespace, podTemplateSpec.Namespace)
			}

			// if !hasVolumeWithName(OdoSourceVolume, podTemplateSpec.Spec.Volumes) {
			// 	t.Errorf("volume with name: %s not found", OdoSourceVolume)
			// }
			if len(podTemplateSpec.Labels) != len(tt.labels) {
				t.Errorf("expected %d, actual %d", len(tt.labels), len(podTemplateSpec.Labels))
			} else {
				for i := range podTemplateSpec.Labels {
					if podTemplateSpec.Labels[i] != tt.labels[i] {
						t.Errorf("expected %s, actual %s", tt.labels[i], podTemplateSpec.Labels[i])
					}
				}
			}

		})
	}
}

func TestGeneratePVCSpec(t *testing.T) {

	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		{
			name:    "1",
			size:    "1Gi",
			wantErr: false,
		},
		{
			name:    "2",
			size:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			quantity, err := resource.ParseQuantity(tt.size)
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("resource.ParseQuantity unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			pvcSpec := GeneratePVCSpec(quantity)
			if pvcSpec.AccessModes[0] != corev1.ReadWriteOnce {
				t.Errorf("AccessMode Error: expected %s, actual %s", corev1.ReadWriteMany, pvcSpec.AccessModes[0])
			}

			pvcSpecQuantity := pvcSpec.Resources.Requests["storage"]
			if pvcSpecQuantity.String() != quantity.String() {
				t.Errorf("pvcSpec.Resources.Requests Error: expected %v, actual %v", pvcSpecQuantity.String(), quantity.String())
			}
		})
	}
}

func TestGenerateIngressSpec(t *testing.T) {

	tests := []struct {
		name      string
		parameter IngressParams
	}{
		{
			name: "1",
			parameter: IngressParams{
				ServiceName:   "service1",
				IngressDomain: "test.1.2.3.4.nip.io",
				PortNumber: intstr.IntOrString{
					IntVal: 8080,
				},
				TLSSecretName: "testTLSSecret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ingressSpec := GenerateIngressSpec(tt.parameter)

			if ingressSpec.Rules[0].Host != tt.parameter.IngressDomain {
				t.Errorf("expected %s, actual %s", tt.parameter.IngressDomain, ingressSpec.Rules[0].Host)
			}

			if ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServicePort != tt.parameter.PortNumber {
				t.Errorf("expected %v, actual %v", tt.parameter.PortNumber, ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServicePort)
			}

			if ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServiceName != tt.parameter.ServiceName {
				t.Errorf("expected %s, actual %s", tt.parameter.ServiceName, ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServiceName)
			}

			if ingressSpec.TLS[0].SecretName != tt.parameter.TLSSecretName {
				t.Errorf("expected %s, actual %s", tt.parameter.TLSSecretName, ingressSpec.TLS[0].SecretName)
			}

		})
	}
}

func TestGenerateServiceSpec(t *testing.T) {
	port1 := corev1.ContainerPort{
		Name:          "port-9090",
		ContainerPort: 9090,
	}
	port2 := corev1.ContainerPort{
		Name:          "port-8080",
		ContainerPort: 8080,
	}

	tests := []struct {
		name  string
		ports []corev1.ContainerPort
	}{
		{
			name:  "singlePort",
			ports: []corev1.ContainerPort{port1},
		},
		{
			name:  "multiplePorts",
			ports: []corev1.ContainerPort{port1, port2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceSpecParams := ServiceSpecParams{
				ContainerPorts: tt.ports,
				SelectorLabels: map[string]string{
					"component": tt.name,
				},
			}
			serviceSpec := GenerateServiceSpec(serviceSpecParams)

			if len(serviceSpec.Ports) != len(tt.ports) {
				t.Errorf("expected service ports length is %v, actual %v", len(tt.ports), len(serviceSpec.Ports))
			} else {
				for i := range serviceSpec.Ports {
					if serviceSpec.Ports[i].Name != tt.ports[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.ports[i].Name, serviceSpec.Ports[i].Name)
					}
					if serviceSpec.Ports[i].Port != tt.ports[i].ContainerPort {
						t.Errorf("expected port number is %v, actual %v", tt.ports[i].ContainerPort, serviceSpec.Ports[i].Port)
					}
				}
			}

		})
	}
}

func fakeResourceRequirements() *corev1.ResourceRequirements {
	var resReq corev1.ResourceRequirements

	limits := make(corev1.ResourceList)
	limits[corev1.ResourceCPU], _ = resource.ParseQuantity("0.5m")
	limits[corev1.ResourceMemory], _ = resource.ParseQuantity("300Mi")
	resReq.Limits = limits

	requests := make(corev1.ResourceList)
	requests[corev1.ResourceCPU], _ = resource.ParseQuantity("0.5m")
	requests[corev1.ResourceMemory], _ = resource.ParseQuantity("300Mi")
	resReq.Requests = requests

	return &resReq
}

// func hasVolumeWithName(name string, volMounts []corev1.Volume) bool {
// 	for _, vm := range volMounts {
// 		if vm.Name == name {
// 			return true
// 		}
// 	}
// 	return false
// }
