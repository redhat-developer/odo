package occlient

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerapi "github.com/openshift/api/image/docker10"
	imagev1 "github.com/openshift/api/image/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	annotations := map[string]string{
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
				Object: &dockerapi.DockerImage{
					ContainerConfig: dockerapi.DockerConfig{
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

func Test_getExposedPortsFromISI(t *testing.T) {
	tests := []struct {
		name             string
		imageTag         string
		imageStreamImage *imagev1.ImageStreamImage
		wantErr          bool
		want             []corev1.ContainerPort
	}{
		{
			name:     "Case: Valid image ports in ContainerConfig",
			imageTag: "3.5",
			imageStreamImage: &imagev1.ImageStreamImage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name/imagename@@sha256:9579a93ee",
				},
				Image: imagev1.Image{
					ObjectMeta: metav1.ObjectMeta{
						Name: "@sha256:9579a93ee",
					},
					DockerImageMetadata: runtime.RawExtension{
						Object: &dockerapi.DockerImage{
							ContainerConfig: dockerapi.DockerConfig{
								ExposedPorts: map[string]struct{}{
									"8080/tcp": {},
								},
							},
						},
					},
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      "TCP",
				},
			},
		},
		{
			name:     "Case: Valid image ports in Config",
			imageTag: "3.5",
			imageStreamImage: &imagev1.ImageStreamImage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name/imagename@@sha256:9579a93ee",
				},
				Image: imagev1.Image{
					ObjectMeta: metav1.ObjectMeta{
						Name: "@sha256:9579a93ee",
					},
					DockerImageMetadata: runtime.RawExtension{
						Object: &dockerapi.DockerImage{
							Config: &dockerapi.DockerConfig{
								ExposedPorts: map[string]struct{}{
									"8080/tcp": {},
								},
							},
						},
					},
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      "TCP",
				},
			},
		},
		{
			name:     "Case: Valid image ports in both Config and ContainerConfig",
			imageTag: "3.5",
			imageStreamImage: &imagev1.ImageStreamImage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "name/imagename@@sha256:9579a93ee",
				},
				Image: imagev1.Image{
					ObjectMeta: metav1.ObjectMeta{
						Name: "@sha256:9579a93ee",
					},
					DockerImageMetadata: runtime.RawExtension{
						Object: &dockerapi.DockerImage{
							ContainerConfig: dockerapi.DockerConfig{
								ExposedPorts: map[string]struct{}{
									"8080/tcp": {},
									"9090/tcp": {},
								},
							},
							Config: &dockerapi.DockerConfig{
								ExposedPorts: map[string]struct{}{
									"9090/tcp": {},
									"9191/tcp": {},
								},
							},
						},
					},
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      "TCP",
				},
				{
					Name:          "9090-tcp",
					ContainerPort: 9090,
					Protocol:      "TCP",
				},
				{
					Name:          "9191-tcp",
					ContainerPort: 9191,
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
			got, err := getExposedPortsFromISI(tt.imageStreamImage)

			// sort result, map is used behind the scene so the ordering might be different
			sort.Slice(got, func(i, j int) bool {
				return got[i].ContainerPort < got[j].ContainerPort
			})

			if !tt.wantErr == (err != nil) {
				t.Errorf("client.GetExposedPorts(imagestream imageTag) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.GetExposedPorts = %#v, want %#v", got, tt.want)
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
				VolumeName: common.SupervisordVolumeName,
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
