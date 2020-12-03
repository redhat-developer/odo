package occlient

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kylelemons/godebug/pretty"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerapi "github.com/openshift/api/image/docker10"
	imagev1 "github.com/openshift/api/image/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/pkg/errors"
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

func TestCreateServiceBinding(t *testing.T) {
	tests := []struct {
		name        string
		bindingNS   string
		bindingName string
		labels      map[string]string
		wantErr     bool
	}{
		{
			name:        "Case: Valid request for creating a secret",
			bindingNS:   "",
			bindingName: "foo",
			labels:      map[string]string{"app": "app"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			err := fakeClient.CreateServiceBinding(tt.bindingName, tt.bindingNS, tt.labels)

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
