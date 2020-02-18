package occlient

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kylelemons/godebug/pretty"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerapi "github.com/openshift/api/image/docker10"
	dockerapiv10 "github.com/openshift/api/image/docker10"
	imagev1 "github.com/openshift/api/image/v1"
	projectv1 "github.com/openshift/api/project/v1"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

// fakeDeploymentConfig creates a fake DC.
// we "dog food" our own functions by using our templates / functions to generate this fake deployment config
func fakeDeploymentConfig(name string, image string, envVars []corev1.EnvVar, envfrom []corev1.EnvFromSource, t *testing.T) *appsv1.DeploymentConfig {

	// save component type as label
	labels := componentlabels.GetLabels(name, name, true)
	labels[componentlabels.ComponentTypeLabel] = image
	labels[componentlabels.ComponentTypeVersion] = "latest"
	labels[applabels.ApplicationLabel] = name

	// save source path as annotation
	annotations := map[string]string{"app.openshift.io/vcs-uri": "./",
		"app.kubernetes.io/component-source-type": "local",
	}

	// Create CommonObjectMeta to be passed in
	commonObjectMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}

	commonImageMeta := CommonImageMeta{
		Name:      name,
		Tag:       "latest",
		Namespace: "openshift",
		Ports:     []corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
	}

	// Generate the DeploymentConfig that will be used.
	dc := generateSupervisordDeploymentConfig(
		commonObjectMeta,
		commonImageMeta,
		envVars,
		envfrom,
		fakeResourceRequirements(),
	)

	// Add the appropriate bootstrap volumes for SupervisorD
	addBootstrapVolumeCopyInitContainer(&dc, commonObjectMeta.Name)
	addBootstrapSupervisordInitContainer(&dc, commonObjectMeta.Name)
	addBootstrapVolume(&dc, commonObjectMeta.Name)
	addBootstrapVolumeMount(&dc, commonObjectMeta.Name)

	return &dc
}

func fakeDeploymentConfigGit(name string, image string, envVars []corev1.EnvVar, containerPorts []corev1.ContainerPort) *appsv1.DeploymentConfig {

	// save component type as label
	labels := componentlabels.GetLabels(name, name, true)
	labels[componentlabels.ComponentTypeLabel] = image
	labels[componentlabels.ComponentTypeVersion] = "latest"

	// save source path as annotation
	annotations := map[string]string{"app.openshift.io/vcs-uri": "github.com/foo/bar.git",
		"app.kubernetes.io/component-source-type": "git",
	}

	// Create CommonObjectMeta to be passed in
	commonObjectMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}

	commonImageMeta := CommonImageMeta{
		Name:      name,
		Tag:       "latest",
		Namespace: "openshift",
		Ports:     []corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
	}

	// Generate the DeploymentConfig that will be used.
	dc := generateGitDeploymentConfig(
		commonObjectMeta,
		commonImageMeta.Name,
		containerPorts,
		envVars,
		fakeResourceRequirements(),
	)

	return &dc
}

func fakeResourceRequirements() *corev1.ResourceRequirements {
	var resReq corev1.ResourceRequirements

	limits := make(corev1.ResourceList)
	limits[corev1.ResourceCPU], _ = parseResourceQuantity("0.5m")
	limits[corev1.ResourceMemory], _ = parseResourceQuantity("300Mi")
	resReq.Limits = limits

	requests := make(corev1.ResourceList)
	requests[corev1.ResourceCPU], _ = parseResourceQuantity("0.5m")
	requests[corev1.ResourceMemory], _ = parseResourceQuantity("300Mi")
	resReq.Requests = requests

	return &resReq
}

// fakeImageStream gets imagestream for the reactor
func fakeImageStream(imageName string, namespace string, strTags []string) *imagev1.ImageStream {
	var tags []imagev1.NamedTagEventList
	for _, tag := range strTags {
		tags = append(tags, imagev1.NamedTagEventList{
			Tag: tag,
			Items: []imagev1.TagEvent{
				{
					DockerImageReference: "example/" + imageName + ":" + tag,
					Generation:           1,
					Image:                "sha256:9579a93ee",
				},
			},
		})
	}

	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: namespace,
		},

		Status: imagev1.ImageStreamStatus{
			Tags: tags,
		},
	}
}

// fakeImageStreams lists the imagestreams for the reactor
func fakeImageStreams(imageName string, namespace string) *imagev1.ImageStreamList {
	return &imagev1.ImageStreamList{
		Items: []imagev1.ImageStream{*fakeImageStream(imageName, namespace, []string{"latest"})},
	}
}

// fakeImageStreamImages gets imagstreamimages for the reactor
func fakeImageStreamImages(imageName string) *imagev1.ImageStreamImage {
	mdata := &dockerapi.DockerImage{
		ContainerConfig: dockerapi.DockerConfig{
			Env: []string{
				"STI_SCRIPTS_URL=http://repo/git/" + imageName,
			},

			ExposedPorts: map[string]struct{}{
				"8080/tcp": {},
			},
		},
	}

	mdataRaw, _ := json.Marshal(mdata)
	return &imagev1.ImageStreamImage{
		Image: imagev1.Image{
			DockerImageReference: "example/" + imageName + ":latest",
			DockerImageMetadata:  runtime.RawExtension{Raw: mdataRaw},
		},
	}
}

// fakeBuildStatus is used to pass fake BuildStatus to watch
func fakeBuildStatus(status buildv1.BuildPhase, buildName string) *buildv1.Build {
	return &buildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      buildName,
		},
		Status: buildv1.BuildStatus{
			Phase: status,
		},
	}
}

func fakePodStatus(status corev1.PodPhase, podName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Status: corev1.PodStatus{
			Phase: status,
		},
	}
}

func fakeImageStreamImage(imageName string, ports []string, containerConfig string) *imagev1.ImageStreamImage {
	exposedPorts := make(map[string]struct{})
	var s struct{}
	for _, port := range ports {
		exposedPorts[port] = s
	}
	builderImage := &imagev1.ImageStreamImage{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s@@sha256:9579a93ee", imageName),
		},
		Image: imagev1.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name: "@sha256:9579a93ee",
			},
			DockerImageMetadata: runtime.RawExtension{
				Object: &dockerapiv10.DockerImage{
					ContainerConfig: dockerapiv10.DockerConfig{
						ExposedPorts: exposedPorts,
					},
				},
			},
			DockerImageReference: fmt.Sprintf("docker.io/centos/%s-36-centos7@s@sha256:9579a93ee", imageName),
		},
	}
	if containerConfig != "" {
		(*builderImage).Image.DockerImageMetadata.Raw = []byte(containerConfig)
	}
	return builderImage
}

func fakePlanExternalMetaDataRaw() ([][]byte, error) {
	planExternalMetaData1 := make(map[string]string)
	planExternalMetaData1["displayName"] = "plan-name-1"

	planExternalMetaData2 := make(map[string]string)
	planExternalMetaData2["displayName"] = "plan-name-2"

	planExternalMetaDataRaw1, err := json.Marshal(planExternalMetaData1)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	planExternalMetaDataRaw2, err := json.Marshal(planExternalMetaData2)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var data [][]byte
	data = append(data, planExternalMetaDataRaw1)
	data = append(data, planExternalMetaDataRaw2)

	return data, nil
}

func fakePlanInstanceCreateParameterSchemasRaw() ([][]byte, error) {
	planInstanceCreateParameterSchema1 := make(map[string][]string)
	planInstanceCreateParameterSchema1["required"] = []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD"}

	planInstanceCreateParameterSchema2 := make(map[string][]string)
	planInstanceCreateParameterSchema2["required"] = []string{"PLAN_DATABASE_USERNAME_2", "PLAN_DATABASE_PASSWORD"}

	planInstanceCreateParameterSchemaRaw1, err := json.Marshal(planInstanceCreateParameterSchema1)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	planInstanceCreateParameterSchemaRaw2, err := json.Marshal(planInstanceCreateParameterSchema2)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	var data [][]byte
	data = append(data, planInstanceCreateParameterSchemaRaw1)
	data = append(data, planInstanceCreateParameterSchemaRaw2)

	return data, nil
}

func TestGetPVCNameFromVolumeMountName(t *testing.T) {
	type args struct {
		volumeMountName string
		dc              *appsv1.DeploymentConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test case : Deployment config with given PVC",
			args: args{
				volumeMountName: "test-pvc",
				dc: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{
							"deploymentconfig": "test",
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "test",
										VolumeMounts: []corev1.VolumeMount{
											{
												MountPath: "/tmp",
												Name:      "test-pvc",
											},
										},
									},
								},
								Volumes: []corev1.Volume{
									{
										Name: "test-pvc",
										VolumeSource: corev1.VolumeSource{
											PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
												ClaimName: "test-pvc",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: "test-pvc",
		},
		{
			name: "Test case : Deployment config without given PVC",
			args: args{
				volumeMountName: "non-existent-pvc",
				dc: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{
							"deploymentconfig": "test",
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "test",
										VolumeMounts: []corev1.VolumeMount{
											{
												MountPath: "/tmp",
												Name:      "test-pvc",
											},
										},
									},
								},
								Volumes: []corev1.Volume{
									{
										Name: "test-pvc",
										VolumeSource: corev1.VolumeSource{
											PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
												ClaimName: "test-pvc",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, _ := FakeNew()

			returnValue := fakeClient.GetPVCNameFromVolumeMountName(tt.args.volumeMountName, tt.args.dc)

			// Check for validating return value
			if returnValue != tt.want {
				t.Errorf("error in return value got: %v, expected %v", returnValue, tt.want)
			}

		})
	}
}

func TestRemoveVolumeFromDeploymentConfig(t *testing.T) {
	type args struct {
		pvc    string
		dcName string
	}
	tests := []struct {
		name     string
		dcBefore *appsv1.DeploymentConfig
		args     args
		wantErr  bool
	}{
		{
			name: "Test case : 1",
			dcBefore: &appsv1.DeploymentConfig{
				Spec: appsv1.DeploymentConfigSpec{
					Selector: map[string]string{
						"deploymentconfig": "test",
					},
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									VolumeMounts: []corev1.VolumeMount{
										{
											MountPath: "/tmp",
											Name:      "test-pvc",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "test-pvc",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "test-pvc",
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				pvc:    "test-pvc",
				dcName: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.dcBefore, nil
			})
			fakeClientSet.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})
			err := fakeClient.RemoveVolumeFromDeploymentConfig(tt.args.pvc, tt.args.dcName)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.RemoveVolumeFromDeploymentConfig(pvc, dcName) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			// Check for validating number of actions performed
			if (len(fakeClientSet.AppsClientset.Actions()) != 2) && (tt.wantErr != true) {
				t.Errorf("expected 2 actions in GetPVCFromName got: %v", fakeClientSet.Kubernetes.Actions())
			}
			updatedDc := fakeClientSet.AppsClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
			//	validating volume got removed from dc
			for _, volume := range updatedDc.Spec.Template.Spec.Volumes {
				if volume.PersistentVolumeClaim.ClaimName == tt.args.pvc {
					t.Errorf("expected volume with name : %v to be removed from dc", tt.args.pvc)
				}
			}
		})
	}
}

func TestGetPVCFromName(t *testing.T) {
	tests := []struct {
		name    string
		pvcName string
		wantPVC *corev1.PersistentVolumeClaim
		wantErr bool
	}{
		{
			name:    "storage 10Gi",
			pvcName: "postgresql",
			wantPVC: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "postgresql",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.wantPVC, nil
			})

			returnPVC, err := fakeClient.GetPVCFromName(tt.pvcName)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.GetPVCFromName(name) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			// Check for validating actions performed
			if (len(fakeClientSet.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in GetPVCFromName got: %v", fakeClientSet.Kubernetes.Actions())
			}
			// Check for value with which the function has called
			PVCname := fakeClientSet.Kubernetes.Actions()[0].(ktesting.GetAction).GetName()
			if PVCname != tt.pvcName {
				t.Errorf("Get action is performed with wrong pvcName, expected: %s, got %s", tt.pvcName, PVCname)

			}
			// Check for returnPVC and tt.wantPVC is same
			if returnPVC != tt.wantPVC {
				t.Errorf("Get action has returned pvc with wrong name, expected: %s, got %s", tt.wantPVC, returnPVC)
			}
		})
	}
}

func TestCreateRoute(t *testing.T) {
	tests := []struct {
		name       string
		urlName    string
		service    string
		portNumber intstr.IntOrString
		labels     map[string]string
		wantErr    bool
		existingDC appsv1.DeploymentConfig
		secureURL  bool
	}{
		{
			name:       "Case : mailserver",
			urlName:    "mailserver",
			service:    "mailserver",
			portNumber: intstr.FromInt(8080),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "python",
			},
			wantErr:    false,
			existingDC: *fakeDeploymentConfig("mailserver", "", nil, nil, t),
		},

		{
			name:       "Case : blog (urlName is different than service)",
			urlName:    "example",
			service:    "blog",
			portNumber: intstr.FromInt(9100),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "golang",
			},
			wantErr:    false,
			existingDC: *fakeDeploymentConfig("blog", "", nil, nil, t),
		},

		{
			name:       "Case : secure url",
			urlName:    "example",
			service:    "blog",
			portNumber: intstr.FromInt(9100),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "golang",
			},
			wantErr:    false,
			existingDC: *fakeDeploymentConfig("blog", "", nil, nil, t),
			secureURL:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()

			fkclientset.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				dc := &appsv1.DeploymentConfig{}
				dc.Name = tt.service
				return true, dc, nil
			})

			createdRoute, err := fkclient.CreateRoute(tt.urlName, tt.service, tt.portNumber, tt.labels, tt.secureURL)

			if tt.secureURL {
				wantedTLSConfig := &routev1.TLSConfig{
					Termination:                   routev1.TLSTerminationEdge,
					InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
				}
				if !reflect.DeepEqual(createdRoute.Spec.TLS, wantedTLSConfig) {
					t.Errorf("tls config is different, wanted %v, got %v", wantedTLSConfig, createdRoute.Spec.TLS)
				}
			} else {
				if createdRoute.Spec.TLS != nil {
					t.Errorf("tls config is set for a non secure url")
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.CreateRoute(string, labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if len(fkclientset.RouteClientset.Actions()) != 1 {
				t.Errorf("expected 1 action in CreateRoute got: %v", fkclientset.RouteClientset.Actions())
			}
			// Checks for return values in positive cases
			if err == nil {
				createdRoute := fkclientset.RouteClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*routev1.Route)
				// created route should be labeled with labels passed to CreateRoute
				if !reflect.DeepEqual(createdRoute.Labels, tt.labels) {
					t.Errorf("labels in created route is not matching expected labels, expected: %v, got: %v", tt.labels, createdRoute.Labels)
				}
				// route name and service that route is pointg to should match
				if createdRoute.Spec.To.Name != tt.service {
					t.Errorf("route is not matching to expected service name, expected: %s, got %s", tt.service, createdRoute)
				}
				if createdRoute.Name != tt.urlName {
					t.Errorf("route name is not matching to expected route name, expected: %s, got %s", tt.urlName, createdRoute.Name)
				}
				if createdRoute.Spec.To.Name != tt.service {
					t.Errorf("service name is not matching to expected service name, expected: %s, got %s", tt.service, createdRoute.Spec.To.Name)
				}
				if createdRoute.Spec.Port.TargetPort != tt.portNumber {
					t.Errorf("port number is not matching to expected port number, expected: %v, got %v", tt.portNumber, createdRoute.Spec.Port.TargetPort)
				}
			}
		})
	}
}

func TestAddLabelsToArgs(t *testing.T) {
	tests := []struct {
		name     string
		argsIn   []string
		labels   map[string]string
		argsOut1 []string
		argsOut2 []string
	}{
		{
			name:   "one label in empty args",
			argsIn: []string{},
			labels: map[string]string{
				"label1": "value1",
			},
			argsOut1: []string{
				"--labels", "label1=value1",
			},
		},
		{
			name: "one label with existing args",
			argsIn: []string{
				"--foo", "bar",
			},
			labels: map[string]string{
				"label1": "value1",
			},
			argsOut1: []string{
				"--foo", "bar",
				"--labels", "label1=value1",
			},
		},
		{
			name: "multiple label with existing args",
			argsIn: []string{
				"--foo", "bar",
			},
			labels: map[string]string{
				"label1": "value1",
				"label2": "value2",
			},
			argsOut1: []string{
				"--foo", "bar",
				"--labels", "label1=value1,label2=value2",
			},
			argsOut2: []string{
				"--foo", "bar",
				"--labels", "label2=value2,label1=value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsGot := addLabelsToArgs(tt.labels, tt.argsIn)

			if !reflect.DeepEqual(argsGot, tt.argsOut1) && !reflect.DeepEqual(argsGot, tt.argsOut2) {
				t.Errorf("addLabelsToArgs() \ngot:  %#v \nwant: %#v or %#v", argsGot, tt.argsOut1, tt.argsOut2)
			}
		})
	}
}

func TestParseImageName(t *testing.T) {

	tests := []struct {
		arg     string
		want1   string
		want2   string
		want3   string
		want4   string
		wantErr bool
	}{
		{
			arg:     "nodejs:8",
			want1:   "",
			want2:   "nodejs",
			want3:   "8",
			want4:   "",
			wantErr: false,
		},
		{
			arg:     "nodejs@sha256:7e56ca37d1db225ebff79dd6d9fd2a9b8f646007c2afc26c67962b85dd591eb2",
			want2:   "nodejs",
			want1:   "",
			want3:   "",
			want4:   "sha256:7e56ca37d1db225ebff79dd6d9fd2a9b8f646007c2afc26c67962b85dd591eb2",
			wantErr: false,
		},
		{
			arg:     "nodejs@sha256:asdf@",
			wantErr: true,
		},
		{
			arg:     "nodejs@@",
			wantErr: true,
		},
		{
			arg:     "nodejs::",
			wantErr: true,
		},
		{
			arg:     "nodejs",
			want1:   "",
			want2:   "nodejs",
			want3:   "latest",
			want4:   "",
			wantErr: false,
		},
		{
			arg:     "",
			wantErr: true,
		},
		{
			arg:     ":",
			wantErr: true,
		},
		{
			arg:     "myproject/nodejs:8",
			want1:   "myproject",
			want2:   "nodejs",
			want3:   "8",
			want4:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("image name: '%s'", tt.arg)
		t.Run(name, func(t *testing.T) {
			got1, got2, got3, got4, err := ParseImageName(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImageName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 != tt.want1 {
				t.Errorf("ParseImageName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseImageName() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("ParseImageName() got3 = %v, want %v", got3, tt.want3)
			}
			if got4 != tt.want4 {
				t.Errorf("ParseImageName() got4 = %v, want %v", got4, tt.want4)
			}
		})
	}
}

func TestUpdateDCAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		dcName      string
		annotations map[string]string
		existingDc  appsv1.DeploymentConfig
		wantErr     bool
	}{
		{
			name:   "existing dc",
			dcName: "nodejs",
			annotations: map[string]string{
				"app.openshift.io/vcs-uri":                "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
					Annotations: map[string]string{"app.openshift.io/vcs-uri": "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "non existing dc",
			dcName: "nodejs",
			annotations: map[string]string{
				"app.openshift.io/vcs-uri":                "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{"app.openshift.io/vcs-uri": "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()
			fkclientset.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dcName := action.(ktesting.GetAction).GetName()
				if dcName != tt.dcName {
					return true, nil, fmt.Errorf("'get' called with a different dcName")
				}

				if tt.dcName != tt.existingDc.Name {
					return true, nil, fmt.Errorf("got different dc")
				}
				return true, &tt.existingDc, nil
			})

			fkclientset.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dc := action.(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
				if dc.Name != tt.existingDc.Name {
					return true, nil, fmt.Errorf("got different dc")
				}
				return true, dc, nil
			})

			err := fkclient.UpdateDCAnnotations(tt.dcName, tt.annotations)

			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.AppsClientset.Actions()) != 2) && (tt.wantErr != true) {
					t.Errorf("expected 2 action in UpdateDeploymentConfig got: %v", fkclientset.AppsClientset.Actions())
				}

				updatedDc := fkclientset.AppsClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
				if updatedDc.Name != tt.dcName {
					t.Errorf("deploymentconfig name is not matching with expected value, expected: %s, got %s", tt.dcName, updatedDc.Name)
				}

				if !reflect.DeepEqual(updatedDc.Annotations, tt.annotations) {
					t.Errorf("deployment Config annotations not matching with expected values, expected: %s, got %s", tt.annotations, updatedDc.Annotations)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestGetBuildConfigFromName(t *testing.T) {
	tests := []struct {
		name                string
		buildName           string
		returnedBuildConfig buildv1.BuildConfig
		wantErr             bool
	}{
		{
			name:      "buildConfig with existing bc",
			buildName: "nodejs",
			returnedBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.BuildClientset.PrependReactor("get", "buildconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				buildName := action.(ktesting.GetAction).GetName()
				if buildName != tt.buildName {
					return true, nil, fmt.Errorf("'get' was called with wrong buildName")
				}
				return true, &tt.returnedBuildConfig, nil
			})

			build, err := fkclient.GetBuildConfigFromName(tt.buildName)
			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.BuildClientset.Actions()) != 1) && (tt.wantErr != true) {
					t.Errorf("expected 1 action in GetBuildConfigFromName got: %v", fkclientset.AppsClientset.Actions())
				}
				if build.Name != tt.buildName {
					t.Errorf("wrong GetBuildConfigFromName got: %v, expected: %v", build.Name, tt.buildName)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestUpdateBuildConfig(t *testing.T) {
	tests := []struct {
		name                string
		buildConfigName     string
		gitURL              string
		annotations         map[string]string
		existingBuildConfig buildv1.BuildConfig
		updatedBuildConfig  buildv1.BuildConfig
		wantErr             bool
	}{
		{
			name:            "local to git with proper parameters",
			buildConfigName: "nodejs",
			gitURL:          "https://github.com/sclorg/nodejs-ex",
			annotations: map[string]string{
				"app.openshift.io/vcs-uri":                "https://github.com/sclorg/nodejs-ex",
				"app.kubernetes.io/component-source-type": "git",
			},
			existingBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{},
				},
			},
			updatedBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{
						Source: buildv1.BuildSource{
							Git: &buildv1.GitBuildSource{
								URI: "https://github.com/sclorg/nodejs-ex",
							},
							Type: buildv1.BuildSourceGit,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()
			fkclientset.BuildClientset.PrependReactor("get", "buildconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				buildConfigName := action.(ktesting.GetAction).GetName()
				if buildConfigName != tt.buildConfigName {
					return true, nil, fmt.Errorf("'update' was called with wrong buildConfig name")
				}
				return true, &tt.existingBuildConfig, nil
			})

			fkclientset.BuildClientset.PrependReactor("update", "buildconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				buildConfig := action.(ktesting.UpdateAction).GetObject().(*buildv1.BuildConfig)
				if buildConfig.Name != tt.buildConfigName {
					return true, nil, fmt.Errorf("'update' was called with wrong buildConfig name")
				}
				return true, &tt.updatedBuildConfig, nil
			})

			err := fkclient.UpdateBuildConfig(tt.buildConfigName, tt.gitURL, tt.annotations)
			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.BuildClientset.Actions()) != 2) && (tt.wantErr != true) {
					t.Errorf("expected 2 action in GetBuildConfigFromName got: %v", fkclientset.BuildClientset.Actions())
				}

				updatedDc := fkclientset.BuildClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*buildv1.BuildConfig)
				if !reflect.DeepEqual(updatedDc.Annotations, tt.annotations) {
					t.Errorf("deployment Config annotations not matching with expected values, expected: %s, got %s", tt.annotations, updatedDc.Annotations)
				}

				if !reflect.DeepEqual(updatedDc.Spec, tt.updatedBuildConfig.Spec) {
					t.Errorf("deployment Config Spec not matching with expected values: %v", pretty.Compare(tt.updatedBuildConfig.Spec, updatedDc.Spec))
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestNewAppS2I(t *testing.T) {
	type args struct {
		commonObjectMeta   metav1.ObjectMeta
		namespace          string
		builderImage       string
		gitURL             string
		inputPorts         []string
		envVars            []string
		storageToBeMounted map[string]*corev1.PersistentVolumeClaim
	}

	tests := []struct {
		name          string
		args          args
		wantedService map[int32]corev1.Protocol
		wantErr       bool
	}{
		{
			name: "case 1: with valid gitURL and two env vars and two storage to be mounted",
			args: args{
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				envVars: []string{"key=value", "key1=value1"},
				storageToBeMounted: map[string]*corev1.PersistentVolumeClaim{
					"pvc-1": testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
					"pvc-2": testingutil.FakePVC("pvc-2", "1Gi", map[string]string{}),
				},
			},
			wantedService: map[int32]corev1.Protocol{
				8080: corev1.ProtocolTCP,
			},
			wantErr: false,
		},
		{
			name: "case 2 : binary buildSource with gitURL empty and no env vars",
			args: args{
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitURL:       "",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				inputPorts: []string{"8081/tcp", "9100/udp"},
			},
			wantedService: map[int32]corev1.Protocol{
				8081: corev1.ProtocolTCP,
				9100: corev1.ProtocolUDP,
			},
			wantErr: true,
		},
		{
			name: "case 3 : with a invalid port protocol",
			args: args{
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				inputPorts: []string{"8081", "9100/blah"},
			},
			wantedService: map[int32]corev1.Protocol{
				8081: corev1.ProtocolTCP,
				9100: corev1.ProtocolUDP,
			},
			wantErr: true,
		},
		{
			name: "case 4 : with a invalid port number",
			args: args{
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				inputPorts: []string{"8ad1", "9100/Udp"},
			},
			wantedService: map[int32]corev1.Protocol{
				8081: corev1.ProtocolTCP,
				9100: corev1.ProtocolUDP,
			},
			wantErr: true,
		},

		// TODO: Currently fails. Enable this case once fixed
		// {
		// 	name: "case 3: with empty builderImage",
		// 	args: args{
		// 		name:         "ruby",
		// 		builderImage: "",
		// 		gitURL:       "https://github.com/openshift/ruby",
		// 		labels: map[string]string{
		// 			"app": "apptmp",
		// 			"app.kubernetes.io/instance": "ruby",
		// 			"app.kubernetes.io/name": "ruby",
		// 			"app.kubernetes.io/part-of":           "apptmp",
		// 		},
		// 		annotations: map[string]string{
		// 			"app.openshift.io/vcs-uri":                   "https://github.com/openshift/ruby",
		// 			"app.kubernetes.io/component-source-type": "git",
		// 		},
		// 	},
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreams(tt.args.commonObjectMeta.Name, tt.args.commonObjectMeta.Namespace), nil
			})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStream(tt.args.commonObjectMeta.Name, tt.args.commonObjectMeta.Namespace, []string{"latest"}), nil
			})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreamimages", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreamImages(tt.args.commonObjectMeta.Name), nil
			})

			fkclientset.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvcName := action.(ktesting.GetAction).GetName()
				for _, pvc := range tt.args.storageToBeMounted {
					if pvc.Name == pvcName {
						return true, pvc, nil
					}
				}
				return true, nil, nil
			})

			fkclientset.Kubernetes.PrependReactor("update", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvc := action.(ktesting.UpdateAction).GetObject().(*corev1.PersistentVolumeClaim)
				if pvc.OwnerReferences[0].Name != tt.args.commonObjectMeta.Name {
					t.Errorf("owner reference not set for dc %s", tt.args.commonObjectMeta.Name)
				}
				return true, pvc, nil
			})

			err := fkclient.NewAppS2I(
				CreateArgs{
					Name:               tt.args.commonObjectMeta.Name,
					SourcePath:         tt.args.gitURL,
					SourceType:         config.GIT,
					ImageName:          tt.args.builderImage,
					EnvVars:            tt.args.envVars,
					Ports:              tt.args.inputPorts,
					Resources:          fakeResourceRequirements(),
					StorageToBeMounted: tt.args.storageToBeMounted,
				},
				tt.args.commonObjectMeta,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewAppS2I() error = %#v, wantErr %#v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.BuildClientset.Actions()) != 1 {
					t.Errorf("expected 1 BuildClientset.Actions() in NewAppS2I, got %v: %v", len(fkclientset.BuildClientset.Actions()), fkclientset.BuildClientset.Actions())
				}

				if len(fkclientset.AppsClientset.Actions()) != 1 {
					t.Errorf("expected 1 AppsClientset.Actions() in NewAppS2I, got: %v", fkclientset.AppsClientset.Actions())
				}

				if len(tt.args.storageToBeMounted) > 0 {
					if len(fkclientset.Kubernetes.Actions()) != len(tt.args.storageToBeMounted)*2+2 {
						t.Errorf("expected %v storage action(s) in PatchCurrentDC got : %v", len(tt.args.storageToBeMounted)*2, len(fkclientset.Kubernetes.Actions()))
					}
				} else {
					if len(fkclientset.Kubernetes.Actions()) != 2 {
						t.Errorf("expected 2 Kubernetes.Actions() in NewAppS2I, got: %v", fkclientset.Kubernetes.Actions())
					}
				}

				var createdIS *imagev1.ImageStream

				if len(tt.args.inputPorts) <= 0 {
					if len(fkclientset.ImageClientset.Actions()) != 4 {
						t.Errorf("expected 4 ImageClientset.Actions() in NewAppS2I, got %v: %v", len(fkclientset.ImageClientset.Actions()), fkclientset.ImageClientset.Actions())
					}

					// Check for imagestream objects
					createdIS = fkclientset.ImageClientset.Actions()[2].(ktesting.CreateAction).GetObject().(*imagev1.ImageStream)
				} else {
					if len(fkclientset.ImageClientset.Actions()) != 1 {
						t.Errorf("expected 3 ImageClientset.Actions() in NewAppS2I, got: %v", fkclientset.ImageClientset.Actions())
					}

					// Check for imagestream objects
					createdIS = fkclientset.ImageClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*imagev1.ImageStream)
				}

				if createdIS.Name != tt.args.commonObjectMeta.Name {
					t.Errorf("imagestream name is not matching with expected name, expected: %s, got %s", tt.args.commonObjectMeta.Name, createdIS.Name)
				}

				if !reflect.DeepEqual(createdIS.Labels, tt.args.commonObjectMeta.Labels) {
					t.Errorf("imagestream labels not matching with expected values, expected: %s, got %s", tt.args.commonObjectMeta.Labels, createdIS.Labels)
				}

				if !reflect.DeepEqual(createdIS.Annotations, tt.args.commonObjectMeta.Annotations) {
					t.Errorf("imagestream annotations not matching with expected values, expected: %s, got %s", tt.args.commonObjectMeta.Annotations, createdIS.Annotations)
				}

				// Check buildconfig objects
				createdBC := fkclientset.BuildClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*buildv1.BuildConfig)

				if tt.args.gitURL != "" {
					if createdBC.Spec.CommonSpec.Source.Git.URI != tt.args.gitURL {
						t.Errorf("git url is not matching with expected value, expected: %s, got %s", tt.args.gitURL, createdBC.Spec.CommonSpec.Source.Git.URI)
					}

					if createdBC.Spec.CommonSpec.Source.Type != "Git" {
						t.Errorf("BuildSource type is not Git as expected")
					}
				}

				// TODO: Enable once Issue #594 fixed
				// } else if createdBC.Spec.CommonSpec.Source.Type != "Binary" {
				// 	t.Errorf("BuildSource type is not Binary as expected")
				// }

				// Check deploymentconfig objects
				createdDC := fkclientset.AppsClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*appsv1.DeploymentConfig)
				if createdDC.Spec.Selector["deploymentconfig"] != tt.args.commonObjectMeta.Name {
					t.Errorf("deploymentconfig name is not matching with expected value, expected: %s, got %s", tt.args.commonObjectMeta.Name, createdDC.Spec.Selector["deploymentconfig"])
				}

				var createdSvc *corev1.Service
				if len(tt.args.storageToBeMounted) > 0 {
					// if storage are needed to be mounted, service creation depends on the storage actions in the kubernetes client
					// since each storage needs 2 actions thus we multiply 2 to the number of storage to be mounted
					createdSvc = fkclientset.Kubernetes.Actions()[len(tt.args.storageToBeMounted)*2].(ktesting.CreateAction).GetObject().(*corev1.Service)
				} else {
					// no storage action needed thus service creation is the first action in the kubernetes client
					createdSvc = fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.Service)
				}

				for port, protocol := range tt.wantedService {
					found := false
					for _, servicePort := range createdSvc.Spec.Ports {
						if servicePort.Port == port {
							found = true
							if servicePort.Protocol != protocol {
								t.Errorf("port protocol not matching, expected: %v, got %v", protocol, servicePort.Protocol)
							}
						}
					}
					if !found {
						t.Errorf("%v port with %v protocol not found", port, protocol)
						break
					}
				}
			}
		})
	}
}

func TestIsTagInImageStream(t *testing.T) {
	tests := []struct {
		name        string
		imagestream imagev1.ImageStream
		imageTag    string
		wantErr     bool
		want        bool
	}{
		{
			name:        "Case: Valid image and image tag",
			imagestream: *fakeImageStream("foo", "openshift", []string{"latest", "3.5"}),
			imageTag:    "3.5",
			want:        true,
		},
		{
			name:        "Case: Invalid image tag",
			imagestream: *fakeImageStream("bar", "testing", []string{"latest"}),
			imageTag:    "0.1",
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := isTagInImageStream(tt.imagestream, tt.imageTag)

			if got != tt.want {
				t.Errorf("GetImageStream() = %#v, want %#v\n\n", got, tt)
			}
		})
	}
}

func TestGetExposedPorts(t *testing.T) {
	tests := []struct {
		name             string
		imageTag         string
		imageStreamImage *imagev1.ImageStreamImage
		wantErr          bool
		want             []corev1.ContainerPort
	}{
		{
			name:             "Case: Valid image ports",
			imageTag:         "3.5",
			imageStreamImage: fakeImageStreamImage("python", []string{"8080/tcp"}, ""),
			want: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-%s", 8080, strings.ToLower(string("tcp"))),
					ContainerPort: 8080,
					Protocol:      "TCP",
				},
			},
		},
		{
			name:             "Case: Invalid image tag",
			imageTag:         "0.1",
			imageStreamImage: fakeImageStreamImage("python", []string{"8080---tcp"}, ""),
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, _ := FakeNew()
			fkclient.Namespace = "testing"
			got, err := fkclient.GetExposedPorts(tt.imageStreamImage)

			if !tt.wantErr == (err != nil) {
				t.Errorf("client.GetExposedPorts(imagestream imageTag) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.GetExposedPorts = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetSecret(t *testing.T) {
	tests := []struct {
		name       string
		secretNS   string
		secretName string
		wantErr    bool
		want       *corev1.Secret
	}{
		{
			name:       "Case: Valid request for retrieving a secret",
			secretNS:   "",
			secretName: "foo",
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: false,
		},
		{
			name:       "Case: Invalid request for retrieving a secret",
			secretNS:   "",
			secretName: "foo2",
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake getting Secret
			fakeClientSet.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.want.Name != tt.secretName {
					return true, nil, fmt.Errorf("'get' called with a different secret name")
				}
				return true, tt.want, nil
			})

			returnValue, err := fakeClient.GetSecret(tt.secretName, tt.secretNS)

			// Check for validating return value
			if err == nil && returnValue != tt.want {
				t.Errorf("error in return value got: %v, expected %v", returnValue, tt.want)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("\nclient.GetSecret(secretNS, secretName) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateServiceBinding(t *testing.T) {
	tests := []struct {
		name        string
		bindingNS   string
		bindingName string
		wantErr     bool
	}{
		{
			name:        "Case: Valid request for creating a secret",
			bindingNS:   "",
			bindingName: "foo",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			err := fakeClient.CreateServiceBinding(tt.bindingName, tt.bindingNS)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1 {
					t.Errorf("expected 1 ServiceCatalogClientSet.Actions() in CreateServiceBinding, got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
				}
				createdBinding := fakeClientSet.ServiceCatalogClientSet.Actions()[0].(ktesting.CreateAction).GetObject().(*scv1beta1.ServiceBinding)
				if createdBinding.Name != tt.bindingName {
					t.Errorf("the name of servicebinding was not correct, expected: %s, got: %s", tt.bindingName, createdBinding.Name)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}

		})
	}
}

func TestGetServiceBinding(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		serviceName string
		wantErr     bool
		want        *scv1beta1.ServiceBinding
	}{
		{
			name:        "Case: Valid request for retrieving a service binding",
			namespace:   "",
			serviceName: "foo",
			want: &scv1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Invalid request for retrieving a service binding",
			namespace:   "",
			serviceName: "foo2",
			want: &scv1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake getting Secret
			fakeClientSet.ServiceCatalogClientSet.PrependReactor("get", "servicebindings", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.want.Name != tt.serviceName {
					return true, nil, fmt.Errorf("'get' called with a different serviebinding name")
				}
				return true, tt.want, nil
			})

			returnValue, err := fakeClient.GetServiceBinding(tt.serviceName, tt.namespace)

			// Check for validating return value
			if err == nil && returnValue != tt.want {
				t.Errorf("error in return value got: %v, expected %v", returnValue, tt.want)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("\nclient.GetServiceBinding(serviceName, namespace) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLinkSecret(t *testing.T) {
	tests := []struct {
		name              string
		secretName        string
		componentName     string
		applicationName   string
		existingDC        appsv1.DeploymentConfig
		expectedUpdatedDC appsv1.DeploymentConfig
		wantErr           bool
	}{
		{
			name:            "Case 1: Unable to locate DeploymentConfig",
			secretName:      "foo",
			componentName:   "foo",
			applicationName: "",
			wantErr:         true,
		},
		{
			name:            "Case 2: Unable to update DeploymentConfig",
			secretName:      "foo",
			componentName:   "",
			applicationName: "foo",
			existingDC:      *fakeDeploymentConfig("foo", "", nil, nil, t),
			wantErr:         true,
		},
		{
			name:            "Case 3: Valid creation of link",
			secretName:      "secret",
			componentName:   "component",
			applicationName: "app",
			existingDC:      *fakeDeploymentConfig("component-app", "", nil, nil, nil),
			expectedUpdatedDC: *fakeDeploymentConfig("component-app", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
						},
					},
				},
				t,
			),
			wantErr: false,
		},
		{
			name:            "Case 4: Creation of link on a component that already has a different link",
			secretName:      "secret",
			componentName:   "component",
			applicationName: "app",
			existingDC: *fakeDeploymentConfig("component-app", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other"},
						},
					},
				},
				t),
			expectedUpdatedDC: *fakeDeploymentConfig("component-app", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
						},
					},
				},
				t),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake getting DC
			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				if len(tt.applicationName) == 0 {
					return true, nil, fmt.Errorf("could not find dc")
				}
				return true, &tt.existingDC, nil
			})

			// Fake updating DC
			fakeClientSet.AppsClientset.PrependReactor("patch", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				if len(tt.componentName) == 0 {
					return true, nil, fmt.Errorf("could not patch dc")
				}
				return true, &tt.expectedUpdatedDC, nil
			})

			err := fakeClient.LinkSecret(tt.secretName, tt.componentName, tt.applicationName)
			if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			} else if err == nil && !tt.wantErr {
				if len(fakeClientSet.AppsClientset.Actions()) != 2 {
					t.Errorf("expected 1 AppsClientset.Actions() in LinkSecret, got: %v", fakeClientSet.AppsClientset.Actions())
				}

				dcPatched := fakeClientSet.AppsClientset.Actions()[1].(ktesting.PatchAction).GetName()
				if dcPatched != tt.existingDC.Name {
					t.Errorf("Expected patch to be performed on dc named: %s but instead got: %s", tt.expectedUpdatedDC.Name, dcPatched)
				}
			}
		})
	}
}

func TestUnlinkSecret(t *testing.T) {
	tests := []struct {
		name              string
		secretName        string
		componentName     string
		applicationName   string
		existingDC        appsv1.DeploymentConfig
		expectedUpdatedDC appsv1.DeploymentConfig
		wantErr           bool
	}{
		{
			name:            "Case 1: Remove link from dc that has none",
			secretName:      "secret",
			componentName:   "component",
			applicationName: "app",
			existingDC:      *fakeDeploymentConfig("foo", "", nil, nil, t),
			wantErr:         true,
		},
		{
			name:            "Case 2: Remove link from dc that has no matching link",
			secretName:      "secret",
			componentName:   "component",
			applicationName: "app",
			existingDC: *fakeDeploymentConfig("foo", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other"},
						},
					},
				},
				t),
			wantErr: true,
		},
		{
			name:            "Case 3: Remove the only link",
			secretName:      "secret",
			componentName:   "component",
			applicationName: "app",
			existingDC: *fakeDeploymentConfig("component-app", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
						},
					},
				},
				t),
			expectedUpdatedDC: *fakeDeploymentConfig("component-app", "", nil, []corev1.EnvFromSource{}, t),
			wantErr:           false,
		},
		{
			name:            "Case 4: Remove a link from a dc that contains many",
			secretName:      "secret",
			componentName:   "component",
			applicationName: "app",
			existingDC: *fakeDeploymentConfig("component-app", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other1"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other2"},
						},
					},
				},
				t),
			expectedUpdatedDC: *fakeDeploymentConfig("component-app", "", nil,
				[]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other1"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "other2"},
						},
					},
				},
				t),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake getting DC
			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				if len(tt.applicationName) == 0 {
					return true, nil, fmt.Errorf("could not find dc")
				}
				return true, &tt.existingDC, nil
			})

			// Fake updating DC
			fakeClientSet.AppsClientset.PrependReactor("patch", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				if len(tt.componentName) == 0 {
					return true, nil, fmt.Errorf("could not patch dc")
				}
				return true, &tt.expectedUpdatedDC, nil
			})

			err := fakeClient.UnlinkSecret(tt.secretName, tt.componentName, tt.applicationName)
			if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			} else if err == nil && !tt.wantErr {
				if len(fakeClientSet.AppsClientset.Actions()) != 2 {
					t.Errorf("expected 1 AppsClientset.Actions() in LinkSecret, got: %v", fakeClientSet.AppsClientset.Actions())
				}

				dcPatched := fakeClientSet.AppsClientset.Actions()[1].(ktesting.PatchAction).GetName()
				if dcPatched != tt.existingDC.Name {
					t.Errorf("Expected patch to be performed on dc named: %s but instead got: %s", tt.expectedUpdatedDC.Name, dcPatched)
				}
			}
		})
	}
}

func TestListSecrets(t *testing.T) {

	tests := []struct {
		name       string
		secretList corev1.SecretList
		output     []corev1.Secret
		wantErr    bool
	}{
		{
			name: "Case 1: Ensure secrets are properly listed",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "secret1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "secret2",
						},
					},
				},
			},
			output: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secret1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secret2",
					},
				},
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := FakeNew()

		fakeClientSet.Kubernetes.PrependReactor("list", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.secretList, nil
		})

		secretsList, err := client.ListSecrets("")

		if !reflect.DeepEqual(tt.output, secretsList) {
			t.Errorf("expected output: %#v,got: %#v", tt.secretList, secretsList)
		}

		if err == nil && !tt.wantErr {
			if len(fakeClientSet.Kubernetes.Actions()) != 1 {
				t.Errorf("expected 1 action in ListSecrets got: %v", fakeClientSet.Kubernetes.Actions())
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}
}

func TestGetImageStream(t *testing.T) {
	tests := []struct {
		name           string
		imageNS        string
		imageName      string
		imageTag       string
		wantErr        bool
		want           *imagev1.ImageStream
		wantActionsCnt int
	}{
		{
			name:           "Case: Valid request for imagestream of latest version and not namespace qualified",
			imageNS:        "",
			imageName:      "foo",
			imageTag:       "latest",
			want:           fakeImageStream("foo", "testing", []string{"latest"}),
			wantActionsCnt: 1,
		},
		{
			name:           "Case: Valid explicit request for specific namespace qualified imagestream of specific version",
			imageNS:        "openshift",
			imageName:      "foo",
			imageTag:       "latest",
			want:           fakeImageStream("foo", "openshift", []string{"latest", "3.5"}),
			wantActionsCnt: 1,
		},
		{
			name:           "Case: Valid request for specific imagestream of specific version not in current namespace",
			imageNS:        "",
			imageName:      "foo",
			imageTag:       "3.5",
			want:           fakeImageStream("foo", "openshift", []string{"latest", "3.5"}),
			wantActionsCnt: 1, // Ideally supposed to be 2 but bcoz prependreactor is not parameter sensitive, the way it is mocked makes it 1
		},
		{
			name:           "Case: Invalid request for non-current and non-openshift namespace imagestream/Non-existant imagestream",
			imageNS:        "foo",
			imageName:      "bar",
			imageTag:       "3.5",
			wantErr:        true,
			wantActionsCnt: 1,
		},
		{
			name:           "Case: Request for non-existant tag",
			imageNS:        "",
			imageName:      "foo",
			imageTag:       "3.6",
			wantErr:        true,
			wantActionsCnt: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "testing"
			openshiftIS := fakeImageStream(tt.imageName, "openshift", []string{"latest", "3.5"})
			currentNSIS := fakeImageStream(tt.imageName, "testing", []string{"latest"})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.imageNS == "" {
					if isTagInImageStream(*fakeImageStream("foo", "testing", []string{"latest"}), tt.imageTag) {
						return true, currentNSIS, nil
					} else if isTagInImageStream(*fakeImageStream("foo", "openshift", []string{"latest", "3.5"}), tt.imageTag) {
						return true, openshiftIS, nil
					}
					return true, nil, fmt.Errorf("Requested imagestream %s with tag %s not found", tt.imageName, tt.imageTag)
				}
				if tt.imageNS == "testing" {
					return true, currentNSIS, nil
				}
				if tt.imageNS == "openshift" {
					return true, openshiftIS, nil
				}
				return true, nil, fmt.Errorf("Requested imagestream %s with tag %s not found", tt.imageName, tt.imageTag)
			})

			got, err := fkclient.GetImageStream(tt.imageNS, tt.imageName, tt.imageTag)
			if len(fkclientset.ImageClientset.Actions()) != tt.wantActionsCnt {
				t.Errorf("expected %d ImageClientset.Actions() in GetImageStream, got %v", tt.wantActionsCnt, fkclientset.ImageClientset.Actions())
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("\nclient.GetImageStream(imageNS, imageName, imageTag) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetImageStream() = %#v, want %#v and the current project name is %s\n\n", got, tt, fkclient.GetCurrentProjectName())
			}
		})
	}
}

func TestGetImageStreams(t *testing.T) {

	type args struct {
		name      string
		namespace string
	}

	tests := []struct {
		name    string
		args    args
		want    []imagev1.ImageStream
		wantErr bool
	}{
		{
			name: "case 1: testing a valid imagestream",
			args: args{
				name:      "ruby",
				namespace: "testing",
			},
			want: []imagev1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ruby",
						Namespace: "testing",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imagev1.TagEvent{
									{
										DockerImageReference: "example/ruby:latest",
										Generation:           1,
										Image:                "sha256:9579a93ee",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},

		// TODO: Currently fails. Enable once fixed
		// {
		//         name: "case 2: empty namespace",
		//         args: args{
		//                 name:      "ruby",
		//                 namespace: "",
		//         },
		//         wantErr: true,
		// },

		// {
		// 	name: "case 3: empty name",
		// 	args: args{
		// 		name:      "",
		// 		namespace: "testing",
		// 	},
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client, fkclientset := FakeNew()

			fkclientset.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreams(tt.args.name, tt.args.namespace), nil
			})

			got, err := client.GetImageStreams(tt.args.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageStreams() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}

			if len(fkclientset.ImageClientset.Actions()) != 1 {
				t.Errorf("expected 1 action in GetImageStreams got: %v", fkclientset.ImageClientset.Actions())
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetImageStreams() = %#v, want %#v", got, tt.want)
			}

		})
	}
}

func TestStartBuild(t *testing.T) {
	tests := []struct {
		name    string
		bcName  string
		wantErr bool
	}{
		{
			name:    "Case 1: Testing valid name",
			bcName:  "ruby",
			wantErr: false,
		},

		// TODO: Currently fails. Enable once fixed.
		// {
		// 	name:    "Case 2: Testing empty name",
		// 	bcName:  "",
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()

			fkclientset.BuildClientset.PrependReactor("create", "buildconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				build := buildv1.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.bcName,
					},
				}

				return true, &build, nil
			})

			_, err := fkclient.StartBuild(tt.bcName)
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.StartBuild(string) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.BuildClientset.Actions()) != 1 {
					t.Errorf("expected 1 action in StartBuild got: %v", fkclientset.BuildClientset.Actions())
				}

				startedBuild := fkclientset.BuildClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*buildv1.BuildRequest)

				if startedBuild.Name != tt.bcName {
					t.Errorf("buildconfig name is not matching to expected name, expected: %s, got %s", tt.bcName, startedBuild.Name)
				}
			}
		})
	}

}

func TestWaitForBuildToFinish(t *testing.T) {

	tests := []struct {
		name      string
		buildName string
		status    buildv1.BuildPhase
		wantErr   bool
	}{
		{
			name:      "phase: complete",
			buildName: "ruby",
			status:    buildv1.BuildPhaseComplete,
			wantErr:   false,
		},

		{
			name:      "phase: failed",
			buildName: "ruby",
			status:    buildv1.BuildPhaseFailed,
			wantErr:   true,
		},

		{
			name:      "phase: cancelled",
			buildName: "ruby",
			status:    buildv1.BuildPhaseCancelled,
			wantErr:   true,
		},

		{
			name:      "phase: error",
			buildName: "ruby",
			status:    buildv1.BuildPhaseError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()
			fkWatch := watch.NewFake()

			go func() {
				fkWatch.Modify(fakeBuildStatus(tt.status, tt.buildName))
			}()

			fkclientset.BuildClientset.PrependWatchReactor("builds", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			err := fkclient.WaitForBuildToFinish(tt.buildName, os.Stdout)
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.WaitForBuildToFinish(string) unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(fkclientset.BuildClientset.Actions()) != 1 {
				t.Errorf("expected 1 action in WaitForBuildToFinish got: %v", fkclientset.BuildClientset.Actions())
			}

			if err == nil {
				expectedFields := fields.OneTermEqualSelector("metadata.name", tt.buildName)
				gotFields := fkclientset.BuildClientset.Actions()[0].(ktesting.WatchAction).GetWatchRestrictions().Fields

				if !reflect.DeepEqual(expectedFields, gotFields) {
					t.Errorf("Fields not matching: expected: %s, got %s", expectedFields, gotFields)
				}
			}
		})
	}

}

func TestWaitAndGetPod(t *testing.T) {

	tests := []struct {
		name    string
		podName string
		status  corev1.PodPhase
		wantErr bool
	}{
		{
			name:    "phase: running",
			podName: "ruby",
			status:  corev1.PodRunning,
			wantErr: false,
		},

		{
			name:    "phase: failed",
			podName: "ruby",
			status:  corev1.PodFailed,
			wantErr: true,
		},

		{
			name: "phase:	unknown",
			podName: "ruby",
			status:  corev1.PodUnknown,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				fkWatch.Modify(fakePodStatus(tt.status, tt.podName))
			}()

			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			podSelector := fmt.Sprintf("deploymentconfig=%s", tt.podName)
			pod, err := fkclient.WaitAndGetPod(podSelector, corev1.PodRunning, "Waiting for component to start")

			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.WaitAndGetPod(string) unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(fkclientset.Kubernetes.Actions()) != 1 {
				t.Errorf("expected 1 action in WaitAndGetPod got: %v", fkclientset.Kubernetes.Actions())
			}

			if err == nil {
				if pod.Name != tt.podName {
					t.Errorf("pod name is not matching to expected name, expected: %s, got %s", tt.podName, pod.Name)
				}
			}

		})
	}
}

func TestWaitAndGetSecret(t *testing.T) {

	tests := []struct {
		name       string
		secretName string
		namespace  string
		wantErr    bool
	}{
		{
			name:       "Case 1: no error expected",
			secretName: "ruby",
			namespace:  "dummy",
			wantErr:    false,
		},

		{
			name:       "Case 2: error expected",
			secretName: "",
			namespace:  "dummy",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				fkWatch.Modify(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.secretName,
					},
				})
			}()

			fkclientset.Kubernetes.PrependWatchReactor("secrets", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				if len(tt.secretName) == 0 {
					return true, nil, fmt.Errorf("error watching secret")
				}
				return true, fkWatch, nil
			})

			pod, err := fkclient.WaitAndGetSecret(tt.secretName, tt.namespace)

			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.WaitAndGetSecret(string, string) unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(fkclientset.Kubernetes.Actions()) != 1 {
				t.Errorf("expected 1 action in WaitAndGetSecret got: %v", fkclientset.Kubernetes.Actions())
			}

			if err == nil {
				if pod.Name != tt.secretName {
					t.Errorf("secret name is not matching to expected name, expected: %s, got %s", tt.secretName, pod.Name)
				}
			}
		})
	}
}

func TestCreateNewProject(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		wait     bool
		wantErr  bool
	}{
		{
			name:     "Case 1: valid project name, not waiting",
			projName: "testing",
			wait:     false,
			wantErr:  false,
		},
		{
			name:     "Case 2: valid project name, waiting",
			projName: "testing2",
			wait:     true,
			wantErr:  false,
		},
		// {
		// 	name:     "Case 2: empty project name",
		// 	projName: "",
		// 	wantErr:  true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.ProjClientset.PrependReactor("create", "projectrequests", func(action ktesting.Action) (bool, runtime.Object, error) {
				proj := projectv1.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.projName,
					},
				}
				return true, &proj, nil
			})

			if tt.wait {
				fkWatch := watch.NewFake()
				// Change the status
				go func() {
					fkWatch.Modify(&projectv1.Project{
						ObjectMeta: metav1.ObjectMeta{
							Name: tt.projName,
						},
					})
				}()
				fkclientset.ProjClientset.PrependWatchReactor("projects", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					if len(tt.projName) == 0 {
						return true, nil, fmt.Errorf("error watching project")
					}
					return true, fkWatch, nil
				})
			}

			err := fkclient.CreateNewProject(tt.projName, tt.wait)
			if !tt.wantErr == (err != nil) {
				t.Errorf("client.CreateNewProject(string) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			actions := fkclientset.ProjClientset.Actions()
			actionsLen := len(actions)
			if !tt.wait && actionsLen != 1 {
				t.Errorf("expected 1 action in CreateNewProject got: %v", actions)
			}
			if tt.wait && actionsLen != 2 {
				t.Errorf("expected 2 actions in CreateNewProject when waiting for project creation got: %v", actions)
			}

			if err == nil {
				createdProj := actions[actionsLen-1].(ktesting.CreateAction).GetObject().(*projectv1.ProjectRequest)

				if createdProj.Name != tt.projName {
					t.Errorf("project name does not match the expected name, expected: %s, got: %s", tt.projName, createdProj.Name)
				}

				if tt.wait {
					expectedFields := fields.OneTermEqualSelector("metadata.name", tt.projName)
					gotFields := actions[0].(ktesting.WatchAction).GetWatchRestrictions().Fields

					if !reflect.DeepEqual(expectedFields, gotFields) {
						t.Errorf("Fields not matching: expected: %s, got %s", expectedFields, gotFields)
					}
				}
			}

		})
	}
}

func TestCreateService(t *testing.T) {
	tests := []struct {
		name             string
		commonObjectMeta metav1.ObjectMeta
		containerPorts   []corev1.ContainerPort
		wantErr          bool
		existingDC       appsv1.DeploymentConfig
	}{
		{
			name:       "Test case: with valid commonObjectName and containerPorts",
			existingDC: *fakeDeploymentConfig("foo", "", nil, nil, t),
			commonObjectMeta: metav1.ObjectMeta{
				Name: "nodejs",
				Labels: map[string]string{
					"app":                        "apptmp",
					"app.kubernetes.io/instance": "ruby",
					"app.kubernetes.io/name":     "ruby",
					"app.kubernetes.io/part-of":  "apptmp",
				},
				Annotations: map[string]string{
					"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
					"app.kubernetes.io/component-source-type": "git",
				},
			},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9100-udp",
					ContainerPort: 9100,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			ownerReference := metav1.OwnerReference{
				APIVersion: "apps.openshift.io/v1",
				Kind:       "DeploymentConfig",
				Name:       tt.existingDC.Name,
				UID:        tt.existingDC.UID,
			}

			_, err := fkclient.CreateService(tt.commonObjectMeta, tt.containerPorts, ownerReference)

			tt.commonObjectMeta.SetOwnerReferences(append(tt.commonObjectMeta.GetOwnerReferences(), ownerReference))

			if err == nil && !tt.wantErr {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 Kubernetes.Actions() in CreateService, got: %v", fkclientset.ImageClientset.Actions())
				}
				createdSvc := fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.Service)
				if !reflect.DeepEqual(tt.commonObjectMeta, createdSvc.ObjectMeta) {
					t.Errorf("ObjectMeta does not match the expected name, expected: %v, got: %v", tt.commonObjectMeta, createdSvc.ObjectMeta)
				}
				if !reflect.DeepEqual(tt.commonObjectMeta.Name, createdSvc.Spec.Selector["deploymentconfig"]) {
					t.Errorf("selector value does not match the expected name, expected: %s, got: %s", tt.commonObjectMeta.Name, createdSvc.Spec.Selector["deploymentconfig"])
				}
				for _, port := range tt.containerPorts {
					found := false
					for _, servicePort := range createdSvc.Spec.Ports {
						if servicePort.Port == port.ContainerPort {
							found = true
							if servicePort.Protocol != port.Protocol {
								t.Errorf("service protocol does not match the expected name, expected: %s, got: %s", port.Protocol, servicePort.Protocol)
							}
							if servicePort.Name != port.Name {
								t.Errorf("service name does not match the expected name, expected: %s, got: %s", port.Name, servicePort.Name)
							}
							if servicePort.TargetPort != intstr.FromInt(int(port.ContainerPort)) {
								t.Errorf("target port does not match the expected name, expected: %v, got: %v", intstr.FromInt(int(port.ContainerPort)), servicePort.TargetPort)
							}
						}
					}
					if found == false {
						t.Errorf("expected service port %s not found in the created Service", tt.name)
						break
					}
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestGetDeploymentConfigsFromSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		label    map[string]string
		wantErr  bool
	}{
		{
			name:     "true case",
			selector: "app.kubernetes.io/part-of=app",
			label: map[string]string{
				"app.kubernetes.io/part-of": "app",
			},
			wantErr: false,
		},
		{
			name:     "true case",
			selector: "app.kubernetes.io/part-of=app1",
			label: map[string]string{
				"app.kubernetes.io/part-of": "app",
			},
			wantErr: false,
		},
	}

	listOfDC := appsv1.DeploymentConfigList{
		Items: []appsv1.DeploymentConfig{
			{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/part-of": "app",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.selector) {
					return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.selector, action.(ktesting.ListAction).GetListRestrictions())
				}
				return true, &listOfDC, nil
			})
			dc, err := fakeClient.GetDeploymentConfigsFromSelector(tt.selector)

			if len(fakeClientSet.AppsClientset.Actions()) != 1 {
				t.Errorf("expected 1 AppsClientset.Actions() in GetDeploymentConfigsFromSelector, got: %v", fakeClientSet.AppsClientset.Actions())
			}

			if tt.wantErr == false && err != nil {
				t.Errorf("test failed, %#v", dc[0].Labels)
			}

			for _, dc1 := range dc {
				if !reflect.DeepEqual(dc1.Labels, tt.label) {
					t.Errorf("labels are not matching with expected labels, expected: %s, got %s", tt.label, dc1.Labels)
				}
			}

		})
	}
}

func TestCreateServiceInstance(t *testing.T) {
	type args struct {
		serviceName string
		serviceType string
		labels      map[string]string
		plan        string
		parameters  map[string]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create service instance",
			args: args{
				serviceName: "jenkins",
				serviceType: "jenkins",
				labels: map[string]string{
					"name":      "mongodb",
					"namespace": "blog",
				},
				plan:       "dev",
				parameters: map[string]string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			err := fkclient.CreateServiceInstance(tt.args.serviceName, tt.args.serviceType, tt.args.plan, tt.args.parameters, tt.args.labels)
			// Checks for error in positive cases
			if tt.wantErr == false && (err != nil) {
				t.Errorf(" client.CreateServiceInstance(serviceName,serviceType, labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			// creating a service instance also means creating a serviceBinding
			// which is why we expect 2 actions
			if len(fkclientset.ServiceCatalogClientSet.Actions()) != 2 && !tt.wantErr {
				t.Errorf("expected 1 action in CreateServiceInstace got: %v", fkclientset.ServiceCatalogClientSet.Actions())
			}

			createdServiceInstance := fkclientset.ServiceCatalogClientSet.Actions()[0].(ktesting.CreateAction).GetObject().(*scv1beta1.ServiceInstance)
			if !reflect.DeepEqual(createdServiceInstance.Labels, tt.args.labels) {
				t.Errorf("labels in created serviceInstance is not matching expected labels, expected: %v, got: %v", tt.args.labels, createdServiceInstance.Labels)
			}
			if createdServiceInstance.Name != tt.args.serviceName {
				t.Errorf("labels in created serviceInstance is not matching expected labels, expected: %v, got: %v", tt.args.serviceName, createdServiceInstance.Name)
			}
			if !reflect.DeepEqual(createdServiceInstance.Spec.ClusterServiceClassExternalName, tt.args.serviceType) {
				t.Errorf("labels in created serviceInstance is not matching expected labels, expected: %v, got: %v", tt.args.serviceType, createdServiceInstance.Spec.ClusterServiceClassExternalName)
			}
		})
	}
}

func TestGetClusterServiceClass(t *testing.T) {
	classExternalMetaData := make(map[string]interface{})
	classExternalMetaData["longDescription"] = "example long description"
	classExternalMetaData["dependencies"] = []string{"docker.io/centos/7", "docker.io/centos/8"}

	classExternalMetaDataRaw, err := json.Marshal(classExternalMetaData)
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	type args struct {
		serviceName string
	}
	tests := []struct {
		name                    string
		args                    args
		returnedServicesClasses *scv1beta1.ClusterServiceClassList
		wantedServiceClass      *scv1beta1.ClusterServiceClass
		wantErr                 bool
	}{
		{
			name: "test case 1: with one valid service class returned",
			args: args{
				serviceName: "class name",
			},
			returnedServicesClasses: &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "1dda1477cace09730bd8ed7a6505607e"},
						Spec: scv1beta1.ClusterServiceClassSpec{
							CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
								ExternalName:     "class name",
								Bindable:         false,
								Description:      "example description",
								Tags:             []string{"php", "java"},
								ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
							},
							ClusterServiceBrokerName: "broker name",
						},
					},
				},
			},
			wantedServiceClass: &scv1beta1.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{Name: "1dda1477cace09730bd8ed7a6505607e"},
				Spec: scv1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
						ExternalName:     "class name",
						Bindable:         false,
						Description:      "example description",
						Tags:             []string{"php", "java"},
						ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
					},
					ClusterServiceBrokerName: "broker name",
				},
			},
			wantErr: false,
		},
		{
			name: "test case 2: with two service classes returned",
			args: args{
				serviceName: "class name",
			},
			returnedServicesClasses: &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "1dda1477cace09730bd8ed7a6505607e"},
						Spec: scv1beta1.ClusterServiceClassSpec{
							CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
								ExternalName:     "class name",
								Bindable:         false,
								Description:      "example description",
								Tags:             []string{"php", "java"},
								ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
							},
							ClusterServiceBrokerName: "broker name",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "1dda1477cace09730bd8ed7a6505607e"},
						Spec: scv1beta1.ClusterServiceClassSpec{
							CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
								ExternalName:     "class name",
								Bindable:         false,
								Description:      "example description",
								Tags:             []string{"java"},
								ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
							},
							ClusterServiceBrokerName: "broker name 1",
						},
					},
				},
			},
			wantedServiceClass: &scv1beta1.ClusterServiceClass{},
			wantErr:            true,
		},
		{
			name: "test case 3: with no service classes returned",
			args: args{
				serviceName: "class name",
			},
			returnedServicesClasses: &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{},
			},
			wantedServiceClass: &scv1beta1.ClusterServiceClass{},
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := FakeNew()

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			if action.(ktesting.ListAction).GetListRestrictions().Fields.String() != fmt.Sprintf("spec.externalName=%v", tt.args.serviceName) {
				t.Errorf("got a different service name got: %v , expected: %v", action.(ktesting.ListAction).GetListRestrictions().Fields.String(), fmt.Sprintf("spec.externalName=%v", tt.args.serviceName))
			}
			return true, tt.returnedServicesClasses, nil
		})

		gotServiceClass, err := client.GetClusterServiceClass(tt.args.serviceName)
		if err == nil && !tt.wantErr {
			if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1 {
				t.Errorf("expected 1 action in GetServiceClassAndPlans got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
			}

			if !reflect.DeepEqual(gotServiceClass.Spec, tt.wantedServiceClass.Spec) {
				t.Errorf("different service class spec value expected: %v", pretty.Compare(gotServiceClass.Spec, tt.wantedServiceClass.Spec))
			}
			if !reflect.DeepEqual(gotServiceClass.Name, tt.wantedServiceClass.Name) {
				t.Errorf("different service class name value expected got: %v , expected: %v", gotServiceClass.Name, tt.wantedServiceClass.Name)
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}

}

func TestGetClusterPlansFromServiceName(t *testing.T) {
	planExternalMetaDataRaw, err := fakePlanExternalMetaDataRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	planInstanceCreateParameterSchemasRaw, err := fakePlanInstanceCreateParameterSchemasRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	type args struct {
		serviceClassName string
	}
	tests := []struct {
		name    string
		args    args
		want    []scv1beta1.ClusterServicePlan
		wantErr bool
	}{
		{
			name:    "test case 1 : plans found for the service class",
			args:    args{serviceClassName: "1dda1477cace09730bd8ed7a6505607e"},
			wantErr: false,
			want: []scv1beta1.ClusterServicePlan{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "67042296c7c95e84142f21f58da2ebfe",
					},
					Spec: scv1beta1.ClusterServicePlanSpec{
						ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
							Name: "1dda1477cace09730bd8ed7a6505607e",
						},
						CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
							ExternalName:                  "dev",
							Description:                   "this is a example description 1",
							ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[0]},
							InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[0]},
						},
					},
				},

				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "7f88be6129622f72554c20af879a8ce0",
					},
					Spec: scv1beta1.ClusterServicePlanSpec{
						ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
							Name: "1dda1477cace09730bd8ed7a6505607e",
						},
						CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
							ExternalName:                  "prod",
							Description:                   "this is a example description 2",
							ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[1]},
							InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[1]},
						},
					},
				},
			},
		},
		{
			name:    "test case 2 : no plans found for the service class",
			args:    args{serviceClassName: "1dda1477cace09730bd8"},
			wantErr: false,
			want:    []scv1beta1.ClusterServicePlan{},
		},
	}

	planList := scv1beta1.ClusterServicePlanList{
		Items: []scv1beta1.ClusterServicePlan{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "67042296c7c95e84142f21f58da2ebfe",
				},
				Spec: scv1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
						Name: "1dda1477cace09730bd8ed7a6505607e",
					},
					CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
						ExternalName:                  "dev",
						Description:                   "this is a example description 1",
						ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[0]},
						InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[0]},
					},
				},
			},

			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "7f88be6129622f72554c20af879a8ce0",
				},
				Spec: scv1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
						Name: "1dda1477cace09730bd8ed7a6505607e",
					},
					CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
						ExternalName:                  "prod",
						Description:                   "this is a example description 2",
						ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[1]},
						InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[1]},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := FakeNew()

			fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (bool, runtime.Object, error) {
				var pList []scv1beta1.ClusterServicePlan
				for _, plan := range planList.Items {
					if plan.Spec.ClusterServiceClassRef.Name == strings.Split(action.(ktesting.ListAction).GetListRestrictions().Fields.String(), "=")[1] {
						pList = append(pList, plan)
					}
				}

				return true, &scv1beta1.ClusterServicePlanList{Items: pList}, nil
			})

			gotPlans, err := client.GetClusterPlansFromServiceName(tt.args.serviceClassName)
			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1 {
					t.Errorf("expected 2 actions in GetServiceClassAndPlans got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
				}

				for _, wantedServicePlan := range tt.want {
					found := false
					for _, gotServicePlan := range gotPlans {
						if reflect.DeepEqual(wantedServicePlan.Spec.ExternalName, gotServicePlan.Spec.ExternalName) {
							found = true
						} else {
							continue
						}

						if !reflect.DeepEqual(wantedServicePlan.Name, gotServicePlan.Name) {
							t.Errorf("different plan name expected got: %v , expected: %v", wantedServicePlan.Name, gotServicePlan.Name)
						}

						if !reflect.DeepEqual(wantedServicePlan.Spec, gotServicePlan.Spec) {
							t.Errorf("different plan spec value expected: %v", pretty.Compare(wantedServicePlan.Spec, gotServicePlan.Spec))
						}
					}

					if !found {
						t.Errorf("service plan %v not found", wantedServicePlan.Spec.ExternalName)
					}
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}

func TestGetDeploymentConfigLabelValues(t *testing.T) {
	type args struct {
		deploymentConfigList appsv1.DeploymentConfigList
		expectedOutput       []string
	}
	tests := []struct {
		applicationName string
		name            string
		args            args
		wantErr         bool
		actions         int
	}{
		{
			name:            "Case 1 - Retrieve list",
			applicationName: "app",
			args: args{
				expectedOutput: []string{"app", "app2"},
				deploymentConfigList: appsv1.DeploymentConfigList{
					Items: []appsv1.DeploymentConfig{
						{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app.kubernetes.io/part-of": "app",
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app.kubernetes.io/part-of": "app2",
								},
							},
						},
					},
				},
			},
			wantErr: false,
			actions: 1,
		},
		{
			name:            "Case 1 - Retrieve list, different order",
			applicationName: "app",
			args: args{
				expectedOutput: []string{"app", "app2"},
				deploymentConfigList: appsv1.DeploymentConfigList{
					Items: []appsv1.DeploymentConfig{
						{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app.kubernetes.io/part-of": "app2",
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app.kubernetes.io/part-of": "app",
								},
							},
						},
					},
				},
			},
			wantErr: false,
			actions: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.args.deploymentConfigList, nil
			})

			// Run function GetServiceInstanceLabelValues
			list, err := fakeClient.GetDeploymentConfigLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)

			if err == nil && !tt.wantErr {

				// Compare arrays
				if !reflect.DeepEqual(list, tt.args.expectedOutput) {
					t.Errorf("expected %s output, got %s", tt.args.expectedOutput, list)
				}

				if (len(fakeClientSet.AppsClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in GetServiceInstanceLabelValues got %v: %v", tt.actions, len(fakeClientSet.AppsClientset.Actions()), fakeClientSet.AppsClientset.Actions())
				}

			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestGetServiceInstanceLabelValues(t *testing.T) {
	type args struct {
		serviceList    scv1beta1.ServiceInstanceList
		expectedOutput []string
		// dcBefore appsv1.DeploymentConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1 - Retrieve list",
			args: args{
				expectedOutput: []string{"app", "app2"},
				serviceList: scv1beta1.ServiceInstanceList{
					Items: []scv1beta1.ServiceInstance{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "mysql-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app",
									componentlabels.ComponentLabel:     "mysql-persistent",
									componentlabels.ComponentTypeLabel: "mysql-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "mysql-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "jenkins-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app2",
									componentlabels.ComponentLabel:     "jenkins-persistent",
									componentlabels.ComponentTypeLabel: "jenkins-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "jenkins-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			actions: 0,
		},
		{
			name: "Case 2 - Retrieve list, different order",
			args: args{
				expectedOutput: []string{"app", "app2"},
				serviceList: scv1beta1.ServiceInstanceList{
					Items: []scv1beta1.ServiceInstance{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "mysql-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app2",
									componentlabels.ComponentLabel:     "mysql-persistent",
									componentlabels.ComponentTypeLabel: "mysql-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "mysql-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "jenkins-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app",
									componentlabels.ComponentLabel:     "jenkins-persistent",
									componentlabels.ComponentTypeLabel: "jenkins-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "jenkins-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			actions: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.args.serviceList, nil
			})

			// Run function GetServiceInstanceLabelValues
			list, err := fakeClient.GetServiceInstanceLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)

			if err == nil && !tt.wantErr {

				// Compare arrays
				if !reflect.DeepEqual(list, tt.args.expectedOutput) {
					t.Errorf("expected %s output, got %s", tt.args.expectedOutput, list)
				}

				if (len(fakeClientSet.AppsClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in GetServiceInstanceLabelValues got %v: %v", tt.actions, len(fakeClientSet.AppsClientset.Actions()), fakeClientSet.AppsClientset.Actions())
				}

			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestGetServiceInstanceList(t *testing.T) {

	type args struct {
		Project  string
		Selector string
	}

	tests := []struct {
		name        string
		args        args
		serviceList scv1beta1.ServiceInstanceList
		output      []scv1beta1.ServiceInstance
		wantErr     bool
	}{
		{
			name: "test case 1",
			args: args{
				Project:  "myproject",
				Selector: "app.kubernetes.io/instance=mysql-persistent,app.kubernetes.io/part-of=app",
			},
			serviceList: scv1beta1.ServiceInstanceList{
				Items: []scv1beta1.ServiceInstance{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "mysql-persistent",
							Finalizers: []string{"kubernetes-incubator/service-catalog"},
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "mysql-persistent",
								componentlabels.ComponentTypeLabel: "mysql-persistent",
							},
							Namespace: "myproject",
						},
						Spec: scv1beta1.ServiceInstanceSpec{
							PlanReference: scv1beta1.PlanReference{
								ClusterServiceClassExternalName: "mysql-persistent",
								ClusterServicePlanExternalName:  "default",
							},
						},
						Status: scv1beta1.ServiceInstanceStatus{
							Conditions: []scv1beta1.ServiceInstanceCondition{
								{
									Reason: "ProvisionedSuccessfully",
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "jenkins-persistent",
							Finalizers: []string{"kubernetes-incubator/service-catalog"},
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "jenkins-persistent",
								componentlabels.ComponentTypeLabel: "jenkins-persistent",
							},
							Namespace: "myproject",
						},
						Spec: scv1beta1.ServiceInstanceSpec{
							PlanReference: scv1beta1.PlanReference{
								ClusterServiceClassExternalName: "jenkins-persistent",
								ClusterServicePlanExternalName:  "default",
							},
						},
						Status: scv1beta1.ServiceInstanceStatus{
							Conditions: []scv1beta1.ServiceInstanceCondition{
								{
									Reason: "ProvisionedSuccessfully",
								},
							},
						},
					},
				},
			},
			output: []scv1beta1.ServiceInstance{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "mysql-persistent",
						Finalizers: []string{"kubernetes-incubator/service-catalog"},
						Labels: map[string]string{
							applabels.ApplicationLabel:         "app",
							componentlabels.ComponentLabel:     "mysql-persistent",
							componentlabels.ComponentTypeLabel: "mysql-persistent",
						},
						Namespace: "myproject",
					},
					Spec: scv1beta1.ServiceInstanceSpec{
						PlanReference: scv1beta1.PlanReference{
							ClusterServiceClassExternalName: "mysql-persistent",
							ClusterServicePlanExternalName:  "default",
						},
					},
					Status: scv1beta1.ServiceInstanceStatus{
						Conditions: []scv1beta1.ServiceInstanceCondition{
							{
								Reason: "ProvisionedSuccessfully",
							},
						},
					},
				},
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := FakeNew()

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.args.Selector) {
				return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.args.Selector, action.(ktesting.ListAction).GetListRestrictions())
			}
			return true, &tt.serviceList, nil
		})

		svcInstanceList, err := client.GetServiceInstanceList(tt.args.Selector)

		if !reflect.DeepEqual(tt.output, svcInstanceList) {
			t.Errorf("expected output: %#v,got: %#v", tt.serviceList, svcInstanceList)
		}

		if err == nil && !tt.wantErr {
			if (len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in ListServicecatalog got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}
}

func TestDeleteServiceInstance(t *testing.T) {

	tests := []struct {
		name        string
		serviceName string
		labels      map[string]string
		serviceList scv1beta1.ServiceInstanceList
		wantErr     bool
	}{
		{
			name:        "Delete service instance",
			serviceName: "mongodb",
			labels: map[string]string{
				applabels.ApplicationLabel:     "app",
				componentlabels.ComponentLabel: "mongodb",
			},
			serviceList: scv1beta1.ServiceInstanceList{
				Items: []scv1beta1.ServiceInstance{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mongodb",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "mongodb",
								componentlabels.ComponentTypeLabel: "mongodb-persistent",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			//fake the services listing
			fkclientset.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.serviceList, nil
			})

			// Fake the servicebinding delete
			fkclientset.ServiceCatalogClientSet.PrependReactor("delete", "servicebindings", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			// Fake the serviceinstance delete
			fkclientset.ServiceCatalogClientSet.PrependReactor("delete", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fkclient.DeleteServiceInstance(tt.labels)
			// Checks for error in positive cases
			if !tt.wantErr && (err != nil) {
				t.Errorf(" client.DeleteServiceInstance(labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			// deleting based on the labels means listing the services and then delete the instance and binding for each
			// thus we have 1 list action that always takes place, plus another 2 (delete instance, delete binding)
			// for each service
			expectedNumberOfServiceCatalogActions := 1 + (2 * len(tt.serviceList.Items))
			if len(fkclientset.ServiceCatalogClientSet.Actions()) != expectedNumberOfServiceCatalogActions && !tt.wantErr {
				t.Errorf("expected %d action in CreateServiceInstace got: %v",
					expectedNumberOfServiceCatalogActions, fkclientset.ServiceCatalogClientSet.Actions())
			}

			// Check that the correct service binding was deleted
			DeletedServiceBinding := fkclientset.ServiceCatalogClientSet.Actions()[1].(ktesting.DeleteAction).GetName()
			if DeletedServiceBinding != tt.serviceName {
				t.Errorf("Delete action is performed with wrong ServiceBinding, expected: %s, got %s", tt.serviceName, DeletedServiceBinding)
			}

			// Check that the correct service instance was deleted
			DeletedServiceInstance := fkclientset.ServiceCatalogClientSet.Actions()[2].(ktesting.DeleteAction).GetName()
			if DeletedServiceInstance != tt.serviceName {
				t.Errorf("Delete action is performed with wrong ServiceInstance, expected: %s, got %s", tt.serviceName, DeletedServiceInstance)
			}
		})
	}
}

func TestPatchCurrentDC(t *testing.T) {
	dcRollOutWait := func(*appsv1.DeploymentConfig, int64) bool {
		return true
	}

	type args struct {
		ucp               UpdateComponentParams
		dcPatch           appsv1.DeploymentConfig
		prePatchDCHandler dcStructUpdater
		isGit             bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1: Test patching with nil prePatchDCHandler (local/binary to git)",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfig("foo", "foo", []corev1.EnvVar{{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"}}, []corev1.EnvFromSource{}, t),
					DcRollOutWaitCond: dcRollOutWait,
				},
				dcPatch: generateGitDeploymentConfig(metav1.ObjectMeta{Name: "foo", Annotations: map[string]string{"app.kubernetes.io/component-source-type": "git"}}, "bar",
					[]corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					fakeResourceRequirements()),
				isGit: true,
			},
			wantErr: false,
			actions: 3,
		},
		{
			name: "Case 2: Test patching with non-nil prePatchDCHandler (local/binary to git)",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfig("foo", "foo", []corev1.EnvVar{{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"}}, []corev1.EnvFromSource{}, t),
					DcRollOutWaitCond: dcRollOutWait,
				},
				dcPatch: generateGitDeploymentConfig(metav1.ObjectMeta{Name: "foo", Annotations: map[string]string{"app.kubernetes.io/component-source-type": "git"}}, "bar",
					[]corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					fakeResourceRequirements()),
				prePatchDCHandler: removeTracesOfSupervisordFromDC,
				isGit:             true,
			},
			wantErr: false,
			actions: 3,
		},
		{
			name: "Case 3: Test patching with different dc configuration (local/binary to local/binary)",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfig("foo", "foo", []corev1.EnvVar{{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"}}, []corev1.EnvFromSource{}, t),
					DcRollOutWaitCond: dcRollOutWait,
				},
				dcPatch:           *fakeDeploymentConfig("foo", "foo", []corev1.EnvVar{{Name: "key1", Value: "value1"}}, []corev1.EnvFromSource{}, t),
				prePatchDCHandler: removeTracesOfSupervisordFromDC,
				isGit:             false,
			},
			wantErr: false,
			actions: 2,
		},
		{
			name: "Case 4: Test patching with the wrong name",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfig("foo", "foo",
						[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
						[]corev1.EnvFromSource{}, t),
					DcRollOutWaitCond: dcRollOutWait,
				},
				dcPatch: generateGitDeploymentConfig(metav1.ObjectMeta{Name: "foo2"}, "bar",
					[]corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					fakeResourceRequirements(),
				),
				isGit: false,
			},
			wantErr: true,
			actions: 3,
		},
		{
			name: "Case 5: Test patching with the dc with same requirements (local/binary to local/binary)",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfig("foo", "foo",
						[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
						[]corev1.EnvFromSource{}, t,
					),
					DcRollOutWaitCond: dcRollOutWait,
				},
				dcPatch: *fakeDeploymentConfig("foo", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t,
				),
				isGit: false,
			},
			wantErr: false,
			actions: 1,
		},
		{
			name: "Case 6: Test patching (git to git) with two storage to mount",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfigGit("foo", "foo",
						[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
						[]corev1.ContainerPort{{Name: "port-1", ContainerPort: 8080}},
					),
					DcRollOutWaitCond: dcRollOutWait,
					StorageToBeMounted: map[string]*corev1.PersistentVolumeClaim{
						"pvc-1": testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
						"pvc-2": testingutil.FakePVC("pvc-2", "1Gi", map[string]string{}),
					},
				},
				dcPatch: generateGitDeploymentConfig(metav1.ObjectMeta{Name: "foo", Annotations: map[string]string{"app.kubernetes.io/component-source-type": "git"}}, "bar",
					[]corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					fakeResourceRequirements(),
				),
				isGit: true,
			},
			wantErr: false,
			actions: 3,
		},
		{
			name: "Case 7: Test patching (git to local/binary) with two storage to mount",
			args: args{
				ucp: UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					ExistingDC: fakeDeploymentConfig("foo", "foo", []corev1.EnvVar{{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"}}, []corev1.EnvFromSource{}, t),
					DcRollOutWaitCond: dcRollOutWait,
					StorageToBeMounted: map[string]*corev1.PersistentVolumeClaim{
						"pvc-1": testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
						"pvc-2": testingutil.FakePVC("pvc-2", "1Gi", map[string]string{}),
					},
				},
				dcPatch: generateGitDeploymentConfig(metav1.ObjectMeta{Name: "foo", Annotations: map[string]string{"app.kubernetes.io/component-source-type": "git"}}, "bar",
					[]corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					fakeResourceRequirements(),
				),
				isGit: false,
			},
			wantErr: false,
			actions: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake "watch"
			fkWatch := watch.NewFake()
			go func() {
				fkWatch.Modify(&tt.args.dcPatch)
			}()
			fakeClientSet.AppsClientset.PrependWatchReactor("deploymentconfigs", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// Fake getting DC
			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.args.ucp.ExistingDC, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvcName := action.(ktesting.GetAction).GetName()
				for _, pvc := range tt.args.ucp.StorageToBeMounted {
					if pvc.Name == pvcName {
						return true, pvc, nil
					}
				}
				return true, nil, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("update", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvc := action.(ktesting.UpdateAction).GetObject().(*corev1.PersistentVolumeClaim)
				if pvc.OwnerReferences[0].Name != tt.args.ucp.ExistingDC.Name {
					t.Errorf("owner reference not set for dc %s", tt.args.ucp.ExistingDC.Name)
				}
				return true, pvc, nil
			})

			// Fake the "update"
			fakeClientSet.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				dc := action.(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
				if dc.Name != tt.args.dcPatch.Name {
					return true, nil, fmt.Errorf("got different dc")
				}
				return true, dc, nil
			})

			fakeClientSet.AppsClientset.PrependReactor("create", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dc := action.(ktesting.CreateAction).GetObject().(*appsv1.DeploymentRequest)
				if dc.Name != tt.args.dcPatch.Name {
					return true, nil, fmt.Errorf("got request for different dc")
				}
				return true, &tt.args.dcPatch, nil
			})

			// Run function PatchCurrentDC
			existingContainer, err := FindContainer(tt.args.ucp.ExistingDC.Spec.Template.Spec.Containers, tt.args.ucp.ExistingDC.Name)
			if err != nil {
				t.Errorf("client.PatchCurrentDC() unexpected error attempting to fetch component container. error %v", err)
			}

			err = fakeClient.PatchCurrentDC(tt.args.dcPatch, tt.args.prePatchDCHandler, existingContainer, tt.args.ucp, tt.args.isGit)

			// Error checking PatchCurrentDC
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.PatchCurrentDC() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if len(tt.args.ucp.StorageToBeMounted) > 0 {
				if len(fakeClientSet.Kubernetes.Actions()) != len(tt.args.ucp.StorageToBeMounted)*2 {
					t.Errorf("expected %v storage action(s) in PatchCurrentDC got : %v", len(tt.args.ucp.StorageToBeMounted)*2, len(fakeClientSet.Kubernetes.Actions()))
				}
			}

			if err == nil && !tt.wantErr {
				// Check to see how many actions are being ran
				if (len(fakeClientSet.AppsClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in PatchCurrentDC got %v: %v", tt.actions, len(fakeClientSet.AppsClientset.Actions()), fakeClientSet.AppsClientset.Actions())
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestUpdateDCToGit(t *testing.T) {
	type args struct {
		name                       string
		newImage                   string
		dc                         appsv1.DeploymentConfig
		ports                      []corev1.ContainerPort
		componentSettings          config.LocalConfigInfo
		resourceLimits             corev1.ResourceRequirements
		envVars                    []corev1.EnvVar
		isDeleteSupervisordVolumes bool
		dcRollOutWaitCond          dcRollOutWait
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1: Check the function works",
			args: args{
				name:     "foo",
				newImage: "bar",

				dc: *fakeDeploymentConfig("foo", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t),
				ports:                      []corev1.ContainerPort{},
				componentSettings:          fakeComponentSettings("foo", "foo", "foo", config.GIT, "nodejs", t),
				resourceLimits:             corev1.ResourceRequirements{},
				envVars:                    []corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
				isDeleteSupervisordVolumes: false,
				dcRollOutWaitCond: func(*appsv1.DeploymentConfig, int64) bool {
					return true
				},
			},
			wantErr: false,
			actions: 3,
		},
		{
			name: "Case 2: Fail if the variable passed in is blank",
			args: args{
				name:     "foo",
				newImage: "",
				dc: *fakeDeploymentConfig("foo", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t),
				ports:                      []corev1.ContainerPort{},
				componentSettings:          fakeComponentSettings("foo", "foo", "foo", config.GIT, "foo", t),
				resourceLimits:             corev1.ResourceRequirements{},
				envVars:                    []corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
				isDeleteSupervisordVolumes: false,
				dcRollOutWaitCond: func(*appsv1.DeploymentConfig, int64) bool {
					return true
				},
			},
			wantErr: true,
			actions: 4,
		},
		{
			name: "Case 3: Fail if image retrieved doesn't match the one we want to patch",
			args: args{
				name:     "foo",
				newImage: "",
				dc: *fakeDeploymentConfig("foo2", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t),
				ports:                      []corev1.ContainerPort{},
				componentSettings:          fakeComponentSettings("foo2", "foo", "foo", config.GIT, "foo2", t),
				resourceLimits:             corev1.ResourceRequirements{},
				envVars:                    []corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
				isDeleteSupervisordVolumes: false,
				dcRollOutWaitCond: func(*appsv1.DeploymentConfig, int64) bool {
					return true
				},
			},
			wantErr: true,
			actions: 3,
		},
		{
			name: "Case 4: Check we can patch with a tag",
			args: args{
				name:     "foo",
				newImage: "bar:latest",
				dc: *fakeDeploymentConfig("foo", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t),
				ports:                      []corev1.ContainerPort{},
				componentSettings:          fakeComponentSettings("foo", "foo", "foo", config.GIT, "foo", t),
				resourceLimits:             corev1.ResourceRequirements{},
				envVars:                    []corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
				isDeleteSupervisordVolumes: false,
				dcRollOutWaitCond: func(*appsv1.DeploymentConfig, int64) bool {
					return true
				},
			},
			wantErr: false,
			actions: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake "watch"
			fkWatch := watch.NewFake()
			go func() {
				fkWatch.Modify(&tt.args.dc)
			}()
			fakeClientSet.AppsClientset.PrependWatchReactor("deploymentconfigs", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// Fake getting DC
			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.args.dc, nil
			})

			// Fake the "update"
			fakeClientSet.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				dc := action.(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)

				// Check name
				if dc.Name != tt.args.dc.Name {
					return true, nil, fmt.Errorf("got different dc")
				}

				// Check that the new patch actually has the new "image"
				if !tt.wantErr == (dc.Spec.Template.Spec.Containers[0].Image != "") {
					return true, nil, fmt.Errorf("got %s image, suppose to get %s", dc.Spec.Template.Spec.Containers[0].Image, "")
				}

				return true, dc, nil
			})

			fakeClientSet.AppsClientset.PrependReactor("create", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dc := action.(ktesting.CreateAction).GetObject().(*appsv1.DeploymentRequest)
				if dc.Name != tt.args.dc.Name {
					return true, nil, fmt.Errorf("got request for different dc")
				}
				return true, &tt.args.dc, nil
			})

			// Fake the pvc delete
			fakeClientSet.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			// Run function UpdateDCToGit
			err := fakeClient.UpdateDCToGit(UpdateComponentParams{
				CommonObjectMeta: metav1.ObjectMeta{Name: tt.args.name},
				ImageMeta: CommonImageMeta{
					Name:  tt.args.newImage,
					Ports: tt.args.ports,
				},
				ResourceLimits:    tt.args.resourceLimits,
				EnvVars:           tt.args.envVars,
				ExistingDC:        &(tt.args.dc),
				DcRollOutWaitCond: tt.args.dcRollOutWaitCond,
			},
				tt.args.isDeleteSupervisordVolumes,
			)

			// Error checking UpdateDCToGit
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.UpdateDCToGit() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !tt.wantErr {
				// Check to see how many actions are being ran
				if (len(fakeClientSet.AppsClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in UpdateDCToGit got %v: %v", tt.actions, len(fakeClientSet.AppsClientset.Actions()), fakeClientSet.AppsClientset.Actions())
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestUniqueAppendOrOverwriteEnvVars(t *testing.T) {
	tests := []struct {
		name            string
		existingEnvVars []corev1.EnvVar
		envVars         []corev1.EnvVar
		want            []corev1.EnvVar
	}{
		{
			name: "Case: Overlapping env vars appends",
			existingEnvVars: []corev1.EnvVar{
				{
					Name:  "key1",
					Value: "value1",
				},
				{
					Name:  "key2",
					Value: "value2",
				},
			},
			envVars: []corev1.EnvVar{
				{
					Name:  "key1",
					Value: "value3",
				},
				{
					Name:  "key2",
					Value: "value4",
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "key1",
					Value: "value3",
				},
				{
					Name:  "key2",
					Value: "value4",
				},
			},
		},
		{
			name: "New env vars append",
			existingEnvVars: []corev1.EnvVar{
				{
					Name:  "key1",
					Value: "value1",
				},
				{
					Name:  "key2",
					Value: "value2",
				},
			},
			envVars: []corev1.EnvVar{
				{
					Name:  "key3",
					Value: "value3",
				},
				{
					Name:  "key4",
					Value: "value4",
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "key1",
					Value: "value1",
				},
				{
					Name:  "key2",
					Value: "value2",
				},
				{
					Name:  "key3",
					Value: "value3",
				},
				{
					Name:  "key4",
					Value: "value4",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnvVars := uniqueAppendOrOverwriteEnvVars(tt.existingEnvVars, tt.envVars...)
			if len(tt.want) != len(gotEnvVars) {
				t.Errorf("Tc: %s, expected %+v, got %+v", tt.name, tt.want, gotEnvVars)
			}
			matchFound := false
			for _, wantEnv := range tt.want {
				for _, gotEnv := range gotEnvVars {
					if reflect.DeepEqual(wantEnv, gotEnv) {
						matchFound = true
					}
				}
				if !matchFound {
					t.Errorf("Tc: %s, expected %+v, got %+v", tt.name, tt.want, gotEnvVars)
				}
			}
		})
	}
}

func TestInjectS2IPaths(t *testing.T) {
	tests := []struct {
		name            string
		existingEnvVars []corev1.EnvVar
		envVars         []corev1.EnvVar
		wantLength      int
	}{
		{
			name: "Case: Overlapping env vars appends",
			existingEnvVars: []corev1.EnvVar{
				{
					Name:  EnvS2IScriptsProtocol,
					Value: "value1",
				},
				{
					Name:  EnvS2IBuilderImageName,
					Value: "value2",
				},
			},
			wantLength: 6,
		},
		{
			name: "New env vars append",
			existingEnvVars: []corev1.EnvVar{
				{
					Name:  "key1",
					Value: "value1",
				},
				{
					Name:  "key2",
					Value: "value2",
				},
			},
			wantLength: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotEnvVars := injectS2IPaths(tt.existingEnvVars, S2IPaths{
				"test", "test", "test", "test", "test", "test", "test",
			})
			if tt.wantLength != len(gotEnvVars) {
				t.Errorf("Tc: %s, expected %+v, got %+v", tt.name, tt.wantLength, len(gotEnvVars))
			}
		})
	}
}

func TestGetS2IMetaInfoFromBuilderImg(t *testing.T) {
	tests := []struct {
		name             string
		imageStreamImage *imagev1.ImageStreamImage
		want             S2IPaths
		wantErr          bool
	}{
		{
			name: "Case 1: Valid nodejs test case with image protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{"kind":"DockerImage","apiVersion":"1.0","Id":"sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234","Created":"2018-10-19T15:43:13Z","ContainerConfig":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=image:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"],"Image":"sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"DockerVersion":"18.06.0-ce","Config":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=image:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","${STI_SCRIPTS_PATH}/usage"],"Image":"57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"Architecture":"amd64","Size":221580439}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "image://",
				ScriptsPath:         "/usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 2: Valid nodejs test case with file protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{"kind":"DockerImage","apiVersion":"1.0","Id":"sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234","Created":"2018-10-19T15:43:13Z","ContainerConfig":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=file:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"],"Image":"sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"file:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"DockerVersion":"18.06.0-ce","Config":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=image:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","${STI_SCRIPTS_PATH}/usage"],"Image":"57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"Architecture":"amd64","Size":221580439}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "file://",
				ScriptsPath:         "/usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 3: Valid nodejs test case with http(s) protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{"kind":"DockerImage","apiVersion":"1.0","Id":"sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234","Created":"2018-10-19T15:43:13Z","ContainerConfig":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=http(s):///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"],"Image":"sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"http(s):///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"DockerVersion":"18.06.0-ce","Config":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=image:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","${STI_SCRIPTS_PATH}/usage"],"Image":"57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"Architecture":"amd64","Size":221580439}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "http(s)://",
				ScriptsPath:         "http(s):///usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 4: Valid openjdk test case with image(s) protocol access",
			imageStreamImage: fakeImageStreamImage(
				"redhat-openjdk18-openshift",
				[]string{"8080/tcp"},
				`{"kind":"DockerImage","apiVersion":"1.0","Id":"sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234","Created":"2018-10-19T15:43:13Z","ContainerConfig":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=http(s):///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"],"Image":"sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","org.jboss.deployments-dir": "/deployments","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/local/s2i","org.jboss.deployments-dir": "/deployments","io.openshift.s2i.destination": "/tmp","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"redhat-openjdk-18/openjdk18-openshift","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"DockerVersion":"18.06.0-ce","Config":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=image:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","${STI_SCRIPTS_PATH}/usage"],"Image":"57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"Architecture":"amd64","Size":221580439}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "image://",
				ScriptsPath:         "/usr/local/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/deployments",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "redhat-openjdk-18/openjdk18-openshift",
			},
			wantErr: false,
		},
		{
			name: "Case 5: Inalid nodejs test case with invalid protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{"kind":"DockerImage","apiVersion":"1.0","Id":"sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234","Created":"2018-10-19T15:43:13Z","ContainerConfig":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=something:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"],"Image":"sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"something:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"DockerVersion":"18.06.0-ce","Config":{"Hostname":"8911994b686d","User":"1001","ExposedPorts":{"8080/tcp":{}},"Env":["PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SUMMARY=Platform for building and running Node.js 10.12.0 applications","DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","STI_SCRIPTS_URL=image:///usr/libexec/s2i","STI_SCRIPTS_PATH=/usr/libexec/s2i","APP_ROOT=/opt/app-root","HOME=/opt/app-root/src","BASH_ENV=/opt/app-root/etc/scl_enable","ENV=/opt/app-root/etc/scl_enable","PROMPT_COMMAND=. /opt/app-root/etc/scl_enable","NODEJS_SCL=rh-nodejs8","NPM_RUN=start","NODE_VERSION=10.12.0","NPM_VERSION=6.4.1","NODE_LTS=false","NPM_CONFIG_LOGLEVEL=info","NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global","NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz","DEBUG_PORT=5858"],"Cmd":["/bin/sh","-c","${STI_SCRIPTS_PATH}/usage"],"Image":"57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34","WorkingDir":"/opt/app-root/src","Entrypoint":["container-entrypoint"],"Labels":{"com.redhat.component":"s2i-base-container","com.redhat.deployments-dir":"/opt/app-root/src","com.redhat.dev-mode":"DEV_MODE:false","com.rehdat.dev-mode.port":"DEBUG_PORT:5858","description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.description":"Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.","io.k8s.display-name":"Node.js 10.12.0","io.openshift.builder-version":"\"190ef14\"","io.openshift.expose-services":"8080:http","io.openshift.s2i.scripts-url":"image:///usr/libexec/s2i","io.openshift.tags":"builder,nodejs,nodejs-10.12.0","io.s2i.scripts-url":"image:///usr/libexec/s2i","maintainer":"Lance Ball \u003clball@redhat.com\u003e","name":"bucharestgold/centos7-s2i-nodejs","org.label-schema.build-date":"20180804","org.label-schema.license":"GPLv2","org.label-schema.name":"CentOS Base Image","org.label-schema.schema-version":"1.0","org.label-schema.vendor":"CentOS","release":"1","summary":"Platform for building and running Node.js 10.12.0 applications","version":"10.12.0"}},"Architecture":"amd64","Size":221580439}`,
			),
			want:    S2IPaths{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s2iPaths, err := GetS2IMetaInfoFromBuilderImg(tt.imageStreamImage)
			if !reflect.DeepEqual(tt.want, s2iPaths) {
				t.Errorf("s2i paths are not matching with expected values, expected: %v, got %v", tt.want, s2iPaths)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf(" GetS2IScriptsPathFromBuilderImg() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteEnvVars(t *testing.T) {
	tests := []struct {
		name           string
		existingEnvs   []corev1.EnvVar
		envTobeDeleted string
		want           []corev1.EnvVar
	}{
		{
			name: "Case 1: valid case of delete",
			existingEnvs: []corev1.EnvVar{
				{
					Name:  "abc",
					Value: "123",
				},
				{
					Name:  "def",
					Value: "456",
				},
			},
			envTobeDeleted: "def",
			want: []corev1.EnvVar{
				{
					Name:  "abc",
					Value: "123",
				},
			},
		},
		{
			name: "Case 2: valid case of delete non-existant env",
			existingEnvs: []corev1.EnvVar{
				{
					Name:  "abc",
					Value: "123",
				},
				{
					Name:  "def",
					Value: "456",
				},
			},
			envTobeDeleted: "ghi",
			want: []corev1.EnvVar{
				{
					Name:  "abc",
					Value: "123",
				},
				{
					Name:  "def",
					Value: "456",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deleteEnvVars(tt.existingEnvs, tt.envTobeDeleted)
			// Verify the passed param is not changed after call to function
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got: %+v, want: %+v", got, tt.want)
			}
		})
	}
}

func fakeComponentSettings(cmpName string, appName string, projectName string, srcType config.SrcType, cmpType string, t *testing.T) config.LocalConfigInfo {
	lci, err := config.NewLocalConfigInfo("")
	if err != nil {
		t.Errorf("failed to init fake component configuration")
		return *lci
	}
	defer os.Remove(lci.Filename)
	err = lci.SetComponentSettings(config.ComponentSettings{
		Name:        &cmpName,
		Application: &appName,
		Project:     &projectName,
		SourceType:  &srcType,
		Type:        &cmpType,
	})
	if err != nil {
		t.Errorf("failed to set component settings. Error %+v", err)
	}
	return *lci
}

func TestUpdateDCToSupervisor(t *testing.T) {
	type args struct {
		name           string
		imageName      string
		expectedImage  string
		imageNamespace string
		isToLocal      bool
		dc             appsv1.DeploymentConfig
		cmpSettings    config.LocalConfigInfo
		envVars        []corev1.EnvVar
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1: Check the function works",
			args: args{
				name:           "foo",
				imageName:      "nodejs",
				expectedImage:  "nodejs",
				imageNamespace: "openshift",
				cmpSettings:    fakeComponentSettings("foo", "foo", "foo", config.LOCAL, "nodejs", t),
				envVars:        []corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
				isToLocal:      true,
				dc: *fakeDeploymentConfig("foo", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t),
			},
			wantErr: false,
			actions: 2,
		},
		{
			name: "Case 2: Fail if unable to find container",
			args: args{
				name:           "testfoo",
				imageName:      "foo",
				expectedImage:  "foobar",
				imageNamespace: "testing",
				isToLocal:      false,
				cmpSettings:    fakeComponentSettings("foo", "foo", "foo", config.LOCAL, "nodejs", t),
				envVars:        []corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
				dc: *fakeDeploymentConfig("foo", "foo",
					[]corev1.EnvVar{{Name: "key1", Value: "value1"}, {Name: "key2", Value: "value2"}},
					[]corev1.EnvFromSource{}, t),
			},
			wantErr: true,
			actions: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake "watch"
			fkWatch := watch.NewFake()
			go func() {
				fkWatch.Modify(&tt.args.dc)
			}()
			fakeClientSet.AppsClientset.PrependWatchReactor("deploymentconfigs", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// Fake getting DC
			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.args.dc, nil
			})

			// Fake the "update"
			fakeClientSet.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				dc := action.(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)

				// Check name
				if dc.Name != tt.args.dc.Name {
					return true, nil, fmt.Errorf("got different dc")
				}

				// Check that the new patch actually has parts of supervisord in it when it's used..

				// Check that addBootstrapVolumeCopyInitContainer is the 1st initContainer and it exists
				if !tt.wantErr == (dc.Spec.Template.Spec.InitContainers[0].Name != "copy-files-to-volume") {
					return true, nil, fmt.Errorf("client.UpdateDCSupervisor() does not contain the copy-files-to-volume container within Spec.Template.Spec.InitContainers, found: %v", dc.Spec.Template.Spec.InitContainers[0].Name)
				}

				// Check that addBootstrapVolumeCopyInitContainer is the 2nd initContainer and it exists
				if !tt.wantErr == (dc.Spec.Template.Spec.InitContainers[1].Name != "copy-supervisord") {
					return true, nil, fmt.Errorf("client.UpdateDCSupervisor() does not contain the copy-supervisord container within Spec.Template.Spec.InitContainers, found: %v", dc.Spec.Template.Spec.InitContainers[1].Name)
				}

				return true, dc, nil
			})

			// Fake getting image stream
			fakeClientSet.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStream(tt.args.expectedImage, tt.args.imageNamespace, []string{"latest"}), nil
			})

			fakeClientSet.ImageClientset.PrependReactor("get", "imagestreamimages", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreamImages(tt.args.imageName), nil
			})

			// Run function UpdateDCToSupervisor
			err := fakeClient.UpdateDCToSupervisor(
				UpdateComponentParams{
					CommonObjectMeta: metav1.ObjectMeta{Name: tt.args.name},
					ImageMeta: CommonImageMeta{
						Name:      tt.args.imageName,
						Tag:       "latest",
						Namespace: "openshift",
					},
					ResourceLimits: corev1.ResourceRequirements{},
					EnvVars:        tt.args.envVars,
					ExistingDC:     &(tt.args.dc),
					DcRollOutWaitCond: func(e *appsv1.DeploymentConfig, i int64) bool {
						return true
					},
				},
				tt.args.isToLocal,
				false,
			)

			// Error checking UpdateDCToSupervisor
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.UpdateDCToSupervisor() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check to see how many actions are being ran
			if err == nil && !tt.wantErr {
				if (len(fakeClientSet.AppsClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in UpdateDCToSupervisor got %v: %v", tt.actions, len(fakeClientSet.AppsClientset.Actions()), fakeClientSet.AppsClientset.Actions())
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestIsVolumeAnEmptyDir(t *testing.T) {
	type args struct {
		VolumeName string
		dc         appsv1.DeploymentConfig
	}
	tests := []struct {
		name         string
		args         args
		wantEmptyDir bool
	}{
		{
			name: "Case 1 - Check that it is an emptyDir",
			args: args{
				VolumeName: supervisordVolumeName,
				dc:         *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantEmptyDir: true,
		},
		{
			name: "Case 2 - Check a non-existent volume",
			args: args{
				VolumeName: "foobar",
				dc:         *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantEmptyDir: false,
		},
		{
			name: "Case 3 - Check a volume that exists but is not emptyDir",
			args: args{
				VolumeName: "foo-s2idata",
				dc:         *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantEmptyDir: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, _ := FakeNew()

			// Run function IsVolumeAnEmptyDir
			isVolumeEmpty := fakeClient.IsVolumeAnEmptyDir(tt.args.VolumeName, &tt.args.dc)

			// Error checking IsVolumeAnEmptyDir
			if tt.wantEmptyDir != isVolumeEmpty {
				t.Errorf(" client.IsVolumeAnEmptyDir() unexpected %v, wantEmptyDir %v", isVolumeEmpty, tt.wantEmptyDir)
			}

		})
	}
}

func Test_GetInputEnvVarsFromStrings(t *testing.T) {
	tests := []struct {
		name          string
		envVars       []string
		wantedEnvVars []corev1.EnvVar
		wantErr       bool
	}{
		{
			name:    "Test case 1: with valid two key value pairs",
			envVars: []string{"key=value", "key1=value1"},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "value1",
				},
			},
			wantErr: false,
		},
		{
			name:    "Test case 2: one env var with missing value",
			envVars: []string{"key=value", "key1="},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "",
				},
			},
			wantErr: false,
		},
		{
			name:    "Test case 3: one env var with no value and no =",
			envVars: []string{"key=value", "key1"},
			wantErr: true,
		},
		{
			name:    "Test case 4: one env var with multiple values",
			envVars: []string{"key=value", "key1=value1=value2"},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "value1=value2",
				},
			},
			wantErr: false,
		},
		{
			name:    "Test case 5: two env var with same key",
			envVars: []string{"key=value", "key=value1"},
			wantErr: true,
		},
		{
			name:    "Test case 6: one env var with base64 encoded value",
			envVars: []string{"key=value", "key1=SSd2ZSBnb3QgYSBsb3ZlbHkgYnVuY2ggb2YgY29jb251dHMhCg=="},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "SSd2ZSBnb3QgYSBsb3ZlbHkgYnVuY2ggb2YgY29jb251dHMhCg==",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars, err := GetInputEnvVarsFromStrings(tt.envVars)

			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(tt.wantedEnvVars, envVars) {
					t.Errorf("corev1.Env values are not matching with expected values, expected: %v, got %v", tt.wantedEnvVars, envVars)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestGetEnvVarsFromDC(t *testing.T) {
	tests := []struct {
		name            string
		dcName          string
		projectName     string
		returnedDC      appsv1.DeploymentConfig
		returnedEnvVars []corev1.EnvVar
		wantErr         bool
	}{
		{
			name:        "case 1: with valid existing dc and one valid env var pair",
			dcName:      "nodejs-app",
			projectName: "project",
			returnedDC: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs-app",
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{
										{
											Name:  "key",
											Value: "value",
										},
									},
								},
							},
						},
					},
				},
			},
			returnedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
			},
			wantErr: false,
		},
		{
			name:        "case 2: with valid existing dc and two valid env var pairs",
			dcName:      "nodejs-app",
			projectName: "project",
			returnedDC: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs-app",
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{
										{
											Name:  "key",
											Value: "value",
										},
										{
											Name:  "key-1",
											Value: "value-1",
										},
									},
								},
							},
						},
					},
				},
			},
			returnedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key-1",
					Value: "value-1",
				},
			},
			wantErr: false,
		},
		{
			name:        "case 3: with non valid existing dc",
			dcName:      "nodejs-app",
			projectName: "project",
			returnedDC: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly-app",
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{},
								},
							},
						},
					},
				},
			},
			returnedEnvVars: []corev1.EnvVar{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dcName := action.(ktesting.GetAction).GetName()
				if dcName != tt.dcName {
					return true, nil, fmt.Errorf("get dc called with different name, expected: %s, got %s", tt.dcName, dcName)
				}
				return true, &tt.returnedDC, nil
			})

			envVars, err := fakeClient.GetEnvVarsFromDC(tt.dcName)

			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if len(fakeClientSet.AppsClientset.Actions()) != 1 {
					t.Errorf("expected 1 action in GetBuildConfigFromName got: %v", fakeClientSet.AppsClientset.Actions())
				}

				if !reflect.DeepEqual(tt.returnedEnvVars, envVars) {
					t.Errorf("env vars are not matching with expected values, expected: %s, got %s", tt.returnedEnvVars, envVars)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func Test_updateEnvVar(t *testing.T) {
	type args struct {
		dc           *appsv1.DeploymentConfig
		inputEnvVars []corev1.EnvVar
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test case 1: tests with single container in dc and no existing env vars",
			args: args{
				dc: fakeDeploymentConfig("foo", "foo", nil, nil, t),
				inputEnvVars: []corev1.EnvVar{
					{
						Name:  "key",
						Value: "value",
					},
					{
						Name:  "key-1",
						Value: "value-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test case 2: tests with single container in dc and existing env vars",
			args: args{
				dc: fakeDeploymentConfig("foo", "foo", []corev1.EnvVar{{Name: "key-1", Value: "key-1"}},
					[]corev1.EnvFromSource{}, t),
				inputEnvVars: []corev1.EnvVar{
					{
						Name:  "key-2",
						Value: "value-2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test case 3: tests with double container in dc",
			args: args{
				dc: &appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "wildfly-app",
					},
					Spec: appsv1.DeploymentConfigSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{},
									},
									{
										Env: []corev1.EnvVar{},
									},
								},
							},
						},
					},
				},
				inputEnvVars: []corev1.EnvVar{
					{
						Name:  "key",
						Value: "value",
					},
					{
						Name:  "key-1",
						Value: "value-1",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateEnvVar(tt.args.dc, tt.args.inputEnvVars)

			if err == nil && !tt.wantErr {
				found := false
				for _, inputEnv := range tt.args.inputEnvVars {
					for _, foundEnv := range tt.args.dc.Spec.Template.Spec.Containers[0].Env {
						if reflect.DeepEqual(inputEnv, foundEnv) {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("update env vars are not matching, expected dc to contain: %v", tt.args.inputEnvVars)
				}

			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func Test_findContainer(t *testing.T) {
	type args struct {
		name       string
		containers []corev1.Container
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case 1 - Find the container",
			args: args{
				name: "foo",
				containers: []corev1.Container{
					{
						Name: "foo",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tmp",
								Name:      "test-pvc",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2 - Error if container not found",
			args: args{
				name: "foo2",
				containers: []corev1.Container{
					{
						Name: "foo",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tmp",
								Name:      "test-pvc",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3 - Error when passing in blank container name",
			args: args{
				name: "",
				containers: []corev1.Container{
					{
						Name: "foo",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tmp",
								Name:      "test-pvc",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 4 - Check against multiple containers (rather than one)",
			args: args{
				name: "foo",
				containers: []corev1.Container{
					{
						Name: "bar",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tmp",
								Name:      "test-pvc",
							},
						},
					},
					{
						Name: "foo",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tmp",
								Name:      "test-pvc",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Run function findContainer
			container, err := FindContainer(tt.args.containers, tt.args.name)

			// Check that the container matches the name
			if err == nil && container.Name != tt.args.name {
				t.Errorf("Wrong container returned, wanted container %v, got %v", tt.args.name, container.Name)
			}

			if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestWaitAndGetDC(t *testing.T) {
	type args struct {
		name       string
		annotation string
		value      string
		dc         appsv1.DeploymentConfig
		timeout    time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1 - Check that the function actually works",
			args: args{
				name:       "foo",
				annotation: "app.kubernetes.io/component-source-type",
				value:      "git",
				dc: *fakeDeploymentConfig("foo", "bar",
					[]corev1.EnvVar{}, []corev1.EnvFromSource{}, t),
				timeout: 3 * time.Second,
			},
			wantErr: false,
			actions: 1,
		},
		{
			name: "Case 2 - Purposefully timeout / error",
			args: args{
				name:       "foo",
				annotation: "app.kubernetes.io/component-source-type",
				value:      "foobar",
				dc: *fakeDeploymentConfig("foo", "bar",
					[]corev1.EnvVar{}, []corev1.EnvFromSource{}, t),
				timeout: 3 * time.Second,
			},
			wantErr: true,
			actions: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()
			fkWatch := watch.NewFake()
			go func() {
				fkWatch.Modify(&tt.args.dc)
			}()
			fakeClientSet.AppsClientset.PrependWatchReactor("deploymentconfigs", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})
			// Run function WaitAndGetDC
			_, err := fakeClient.WaitAndGetDC(tt.args.name, 0, tt.args.timeout, func(*appsv1.DeploymentConfig, int64) bool {
				return !tt.wantErr
			})
			// Error checking WaitAndGetDC
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.WaitAndGetDC() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && !tt.wantErr {
				// Check to see how many actions are being ran
				if (len(fakeClientSet.AppsClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in WaitAndGetDC got %v: %v", tt.actions, len(fakeClientSet.AppsClientset.Actions()), fakeClientSet.AppsClientset.Actions())
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}

func TestCreateBuildConfig(t *testing.T) {
	type args struct {
		commonObjectMeta metav1.ObjectMeta
		namespace        string
		builderImage     string
		gitURL           string
		gitRef           string
		envVars          []corev1.EnvVar
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1 - Generate and create the BuildConfig",
			args: args{
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				gitRef:       "master",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				envVars: []corev1.EnvVar{
					{
						Name:  "key",
						Value: "value",
					},
					{
						Name:  "key1",
						Value: "value1",
					},
				},
			},
			wantErr: false,
			actions: 1,
		},
		{
			name: "Case 2 - Generate and create the BuildConfig but fail with unable to find image name",
			args: args{
				builderImage: "fakeimagename:notlatest",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				envVars: []corev1.EnvVar{
					{
						Name:  "key",
						Value: "value",
					},
					{
						Name:  "key1",
						Value: "value1",
					},
				},
			},
			wantErr: true,
			actions: 1,
		},
		{
			name: "Case 3 - Generate and create the BuildConfig but fail with unable to parse image name",
			args: args{
				builderImage: "::",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				envVars: []corev1.EnvVar{
					{
						Name:  "key",
						Value: "value",
					},
					{
						Name:  "key1",
						Value: "value1",
					},
				},
			},
			wantErr: true,
			actions: 1,
		},
		{
			name: "Case 4 - Generate and create the BuildConfig and pass in no envVars",
			args: args{
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitURL:       "https://github.com/openshift/ruby",
				gitRef:       "develop",
				commonObjectMeta: metav1.ObjectMeta{
					Name: "ruby",
					Labels: map[string]string{
						"app":                        "apptmp",
						"app.kubernetes.io/instance": "ruby",
						"app.kubernetes.io/name":     "ruby",
						"app.kubernetes.io/part-of":  "apptmp",
					},
					Annotations: map[string]string{
						"app.openshift.io/vcs-uri":                "https://github.com/openshift/ruby",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				envVars: []corev1.EnvVar{},
			},
			wantErr: false,
			actions: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStream(tt.args.commonObjectMeta.Name, tt.args.commonObjectMeta.Namespace, []string{"latest"}), nil
			})

			// Run function CreateBuildConfig
			bc, err := fakeClient.CreateBuildConfig(tt.args.commonObjectMeta, tt.args.builderImage, tt.args.gitURL, tt.args.gitRef, tt.args.envVars)

			if err == nil && !tt.wantErr {
				// Check to see how many actions are being ran
				if (len(fakeClientSet.ImageClientset.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in CreateBuildConfig got %v: %v", tt.actions, len(fakeClientSet.ImageClientset.Actions()), fakeClientSet.ImageClientset.Actions())
				}

				// Check to see that names match
				if bc.ObjectMeta.Name != tt.args.commonObjectMeta.Name {
					t.Errorf("Expected buildConfig name %s, got '%s'", tt.args.commonObjectMeta.Name, bc.ObjectMeta.Name)
				}

				// Check to see that labels match
				if !reflect.DeepEqual(tt.args.commonObjectMeta.Labels, bc.ObjectMeta.Labels) {
					t.Errorf("Expected equal labels, got %+v, expected %+v", tt.args.commonObjectMeta.Labels, bc.ObjectMeta.Labels)
				}

				// Check to see that annotations match
				if !reflect.DeepEqual(tt.args.commonObjectMeta.Annotations, bc.ObjectMeta.Annotations) {
					t.Errorf("Expected equal annotations, got %+v, expected %+v", tt.args.commonObjectMeta.Annotations, bc.ObjectMeta.Annotations)
				}

			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}

func TestGetServiceClassesByCategory(t *testing.T) {
	t.Run("GetServiceClassesByCategory should work", func(t *testing.T) {
		client, fakeClientSet := FakeNew()
		foo := testingutil.FakeClusterServiceClass("foo", "footag", "footag2")
		bar := testingutil.FakeClusterServiceClass("bar", "")
		boo := testingutil.FakeClusterServiceClass("boo")
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{
					foo,
					bar,
					boo,
				},
			}, nil
		})

		expected := map[string][]scv1beta1.ClusterServiceClass{"footag": {foo}, "other": {bar, boo}}
		categories, err := client.GetServiceClassesByCategory()

		if err != nil {
			t.Errorf("test failed due to %s", err.Error())
		}

		if !reflect.DeepEqual(expected, categories) {
			t.Errorf("test failed, expected %v, got %v", expected, categories)
		}
	})
}

func TestGetMatchingPlans(t *testing.T) {
	t.Run("GetMatchingPlans should work", func(t *testing.T) {
		client, fakeClientSet := FakeNew()
		foo := testingutil.FakeClusterServiceClass("foo", "footag", "footag2")
		dev := testingutil.FakeClusterServicePlan("dev", 1)
		classId := foo.Spec.ExternalID
		dev.Spec.ClusterServiceClassRef.Name = classId
		prod := testingutil.FakeClusterServicePlan("prod", 2)
		prod.Spec.ClusterServiceClassRef.Name = classId

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			value, _ := action.(ktesting.ListAction).GetListRestrictions().Fields.RequiresExactMatch("spec.clusterServiceClassRef.name")
			if value != classId {
				t.Errorf("cluster service plans list should have been filtered on 'spec.clusterServiceClassRef.name==%s'", classId)
			}

			return true, &scv1beta1.ClusterServicePlanList{
				Items: []scv1beta1.ClusterServicePlan{
					dev,
					prod,
				},
			}, nil
		})

		expected := map[string]scv1beta1.ClusterServicePlan{"dev": dev, "prod": prod}
		plans, err := client.GetMatchingPlans(foo)

		if err != nil {
			t.Errorf("test failed due to %s", err.Error())
		}

		if !reflect.DeepEqual(expected, plans) {
			t.Errorf("test failed, expected %v, got %v", expected, plans)
		}
	})
}

func TestStartDeployment(t *testing.T) {
	tests := []struct {
		name           string
		deploymentName string
		wantErr        bool
	}{
		{
			name:           "Case 1: Testing valid name",
			deploymentName: "ruby",
			wantErr:        false,
		},
		{
			name:           "Case 2: Testing invalid name",
			deploymentName: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()

			fkclientset.AppsClientset.PrependReactor("create", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				deploymentConfig := appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.deploymentName,
					},
				}
				return true, &deploymentConfig, nil
			})

			_, err := fkclient.StartDeployment(tt.deploymentName)
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.StartDeployment(string) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.AppsClientset.Actions()) != 1 {
					t.Errorf("expected 1 action in StartDeployment got: %v", fkclientset.AppsClientset.Actions())
				} else {
					startedDeployment := fkclientset.AppsClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*appsv1.DeploymentRequest)

					if startedDeployment.Name != tt.deploymentName {
						t.Errorf("deployment name is not matching to expected name, expected: %s, got %s", tt.deploymentName, startedDeployment.Name)
					}

					if startedDeployment.Latest == false {
						t.Errorf("deployment is not set to latest")
					}
				}
			}
		})
	}
}

func TestGetPortsFromBuilderImage(t *testing.T) {

	type args struct {
		componentType string
	}
	tests := []struct {
		name           string
		imageNamespace string
		args           args
		want           []string
		wantErr        bool
	}{
		{
			name:           "component type: nodejs",
			imageNamespace: "openshift",
			args:           args{componentType: "nodejs"},
			want:           []string{"8080/TCP"},
			wantErr:        false,
		},
		{
			name:           "component type: php",
			imageNamespace: "openshift",
			args:           args{componentType: "php"},
			want:           []string{"8080/TCP", "8443/TCP"},
			wantErr:        false,
		},
		{
			name:           "component type: is empty",
			imageNamespace: "openshift",
			args:           args{componentType: ""},
			want:           []string{},
			wantErr:        true,
		},
		{
			name:           "component type: is invalid",
			imageNamespace: "openshift",
			args:           args{componentType: "abc"},
			want:           []string{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "testing"
			// Fake getting image stream
			fkclientset.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStream(tt.args.componentType, tt.imageNamespace, []string{"latest"}), nil
			})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreamimages", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreamImage(tt.args.componentType, tt.want, ""), nil
			})
			got, err := fkclient.GetPortsFromBuilderImage(tt.args.componentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetPortsFromBuilderImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !sliceEqual(got, tt.want) {
				t.Errorf("Client.GetPortsFromBuilderImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// sliceEqual checks equality of two slices irrespective of the element ordering
func sliceEqual(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	xc := make([]string, len(x))
	yc := make([]string, len(y))

	copy(xc, x)
	copy(yc, y)

	sort.Strings(xc)
	sort.Strings(yc)

	return reflect.DeepEqual(xc, yc)
}

func TestIsSubDir(t *testing.T) {

	tests := []struct {
		name     string
		baseDir  string
		otherDir string
		matches  bool
	}{
		{
			name:     "Case 1: same dirs with slashes",
			baseDir:  "/abcd/",
			otherDir: "/abcd",
			matches:  true,
		},
		{
			name:     "Case 2: same dirs with slashes order reverse",
			baseDir:  "/abcd",
			otherDir: "/abcd/",
			matches:  true,
		},
		{
			name:     "Case 3: other dir same prefix",
			baseDir:  "/abcd",
			otherDir: "/abcde/",
			matches:  false,
		},
		{
			name:     "Case 4: other dir same prefix more complex",
			baseDir:  "/abcde/fg",
			otherDir: "/abcde/fgh",
			matches:  false,
		},
		{
			name:     "Case 5: other dir same prefix more complex matching",
			baseDir:  "/abcde/fg",
			otherDir: "/abcde/fg/h",
			matches:  true,
		},
		{
			name:     "Case 6: dirs with ..",
			baseDir:  "/abcde/fg/../h",
			otherDir: "/abcde/h/ij",
			matches:  true,
		},
		{
			name:     "Case 7: dirs with .. not matching",
			baseDir:  "/abcde/fg/../h",
			otherDir: "/abcde/fg/h",
			matches:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isSubDir(tt.baseDir, tt.otherDir) != tt.matches {
				t.Errorf("the outcome for %s and %s is not expected", tt.baseDir, tt.otherDir)
			}
		})
	}

}
