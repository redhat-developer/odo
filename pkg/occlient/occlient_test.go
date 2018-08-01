package occlient

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerapi "github.com/openshift/api/image/docker10"
	imagev1 "github.com/openshift/api/image/v1"
	projectv1 "github.com/openshift/api/project/v1"
	routev1 "github.com/openshift/api/route/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	urlLabels "github.com/redhat-developer/odo/pkg/url/labels"
)

// fakeImageStream gets imagestream for the reactor
func fakeImageStream(imageName string, namespace string) *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: namespace,
		},

		Status: imagev1.ImageStreamStatus{
			Tags: []imagev1.NamedTagEventList{
				{
					Tag: "latest",
					Items: []imagev1.TagEvent{
						{DockerImageReference: "example/" + imageName + ":latest"},
						{Generation: 1},
						{Image: imageName + "@sha256:9579a93ee"},
					},
				},
			},
		},
	}
}

// fakeImageStreams lists the imagestreams for the reactor
func fakeImageStreams(imageName string, namespace string) *imagev1.ImageStreamList {
	return &imagev1.ImageStreamList{
		Items: []imagev1.ImageStream{*fakeImageStream(imageName, namespace)},
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
		wantErr bool
	}{
		{
			name:    "storage 10Gi",
			pvcName: "postgresql",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			_, err := fakeClient.GetPVCFromName(tt.pvcName)

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
		})
	}
}

func TestCreatePVC(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		labels  map[string]string
		wantErr bool
	}{
		{
			name: "storage 10Gi",
			size: "10Gi",
			labels: map[string]string{
				"name":      "mongodb",
				"namespace": "blog",
			},
			wantErr: false,
		},
		{
			name: "storage 1024",
			size: "1024",
			labels: map[string]string{
				"name":      "PostgreSQL",
				"namespace": "backend",
			},
			wantErr: false,
		},
		{
			name: "storage invalid size",
			size: "4#0",
			labels: map[string]string{
				"name":      "MySQL",
				"namespace": "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			_, err := fkclient.CreatePVC(tt.name, tt.size, tt.labels)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.CreatePVC(name, size, labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if (len(fkclientset.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in CreatePVC got: %v", fkclientset.RouteClientset.Actions())
			}
			// Checks for return values in positive cases
			if err == nil {
				createdPVC := fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.PersistentVolumeClaim)
				quantity, err := resource.ParseQuantity(tt.size)
				if err != nil {
					t.Errorf("failed to create quantity by calling resource.ParseQuantity(%v)", tt.size)
				}

				// created PVC should be labeled with labels passed to CreatePVC
				if !reflect.DeepEqual(createdPVC.Labels, tt.labels) {
					t.Errorf("labels in created route is not matching expected labels, expected: %v, got: %v", tt.labels, createdPVC.Labels)
				}
				// name, size of createdPVC should be matching to size, name passed to CreatePVC
				if !reflect.DeepEqual(createdPVC.Spec.Resources.Requests["storage"], quantity) {
					t.Errorf("size of PVC is not matching to expected size, expected: %v, got %v", quantity, createdPVC.Spec.Resources.Requests["storage"])
				}
				if createdPVC.Name != tt.name {
					t.Errorf("PVC name is not matching to expected name, expected: %s, got %s", tt.name, createdPVC.Name)
				}
			}
		})
	}
}

func TestDeletePVC(t *testing.T) {
	tests := []struct {
		name    string
		pvcName string
		wantErr bool
	}{
		{
			name:    "storage 10Gi",
			pvcName: "postgresql",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fakeClient.DeletePVC(tt.pvcName)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.DeletePVC(name) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if (len(fakeClientSet.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in DeletePVC got: %v", fakeClientSet.Kubernetes.Actions())
			}

			// Check for value with which the function has called
			DeletedPVC := fakeClientSet.Kubernetes.Actions()[0].(ktesting.DeleteAction).GetName()
			if DeletedPVC != tt.pvcName {
				t.Errorf("Delete action is performed with wrong pvcName, expected: %s, got %s", tt.pvcName, DeletedPVC)

			}
		})
	}
}

func TestCreateRoute(t *testing.T) {
	tests := []struct {
		name    string
		urlName string
		service string
		labels  map[string]string
		wantErr bool
	}{
		{
			name:    "Case : mailserver",
			urlName: "mailserver",
			service: "mailserver",
			labels: map[string]string{
				"SLA": "High",
				"app.kubernetes.io/component-name": "backend",
				"app.kubernetes.io/component-type": "python",
			},
			wantErr: false,
		},

		{
			name:    "Case : blog (urlName is different than service)",
			urlName: "example",
			service: "blog",
			labels: map[string]string{
				"SLA": "High",
				"app.kubernetes.io/component-name": "backend",
				"app.kubernetes.io/component-type": "golang",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()

			_, err := fkclient.CreateRoute(tt.urlName, tt.service, tt.labels)

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

func Test_parseImageName(t *testing.T) {

	tests := []struct {
		arg     string
		want1   string
		want2   string
		want3   string
		wantErr bool
	}{
		{
			arg:     "nodejs:8",
			want1:   "nodejs",
			want2:   "8",
			want3:   "",
			wantErr: false,
		},
		{
			arg:     "nodejs@sha256:7e56ca37d1db225ebff79dd6d9fd2a9b8f646007c2afc26c67962b85dd591eb2",
			want1:   "nodejs",
			want2:   "",
			want3:   "sha256:7e56ca37d1db225ebff79dd6d9fd2a9b8f646007c2afc26c67962b85dd591eb2",
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
			want1:   "nodejs",
			want2:   "latest",
			want3:   "",
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
	}
	for _, tt := range tests {
		name := fmt.Sprintf("image name: '%s'", tt.arg)
		t.Run(name, func(t *testing.T) {
			got1, got2, got3, err := parseImageName(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseImageName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got1 != tt.want1 {
				t.Errorf("parseImageName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("parseImageName() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("parseImageName() got3 = %v, want %v", got3, tt.want3)
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
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
					Annotations: map[string]string{"app.kubernetes.io/url": "https://github.com/sclorg/nodejs-ex",
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
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{"app.kubernetes.io/url": "https://github.com/sclorg/nodejs-ex",
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

func TestSetupForSupervisor(t *testing.T) {
	quantity, _ := resource.ParseQuantity("1Gi")
	errQuantity, _ := resource.ParseQuantity("2Gi")

	tests := []struct {
		name        string
		dcName      string
		projectName string
		annotations map[string]string
		labels      map[string]string
		existingDc  appsv1.DeploymentConfig
		createdPVC  corev1.PersistentVolumeClaim
		wantErr     bool
	}{
		{
			name:        "setup with normal correct values",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			labels: map[string]string{
				"app": "apptmp",
				"app.kubernetes.io/component-name": "ruby",
				"app.kubernetes.io/component-type": "ruby",
				"app.kubernetes.io/name":           "apptmp",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{"app.kubernetes.io/url": "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"deploymentconfig": "wildfly",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "wildfly:latest",
									Name:  "wildfly",
								},
							},
						},
					},
				},
			},
			createdPVC: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-s2idata", "wildfly"),
					Labels: map[string]string{
						"app": "apptmp",
						"app.kubernetes.io/component-name": "wildfly",
						"app.kubernetes.io/component-type": "wildfly",
						"app.kubernetes.io/name":           "apptmp",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: quantity,
						},
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
				},
			},

			wantErr: false,
		},
		{
			name:        "setup with wrong pvc name",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			labels: map[string]string{
				"app": "apptmp",
				"app.kubernetes.io/component-name": "ruby",
				"app.kubernetes.io/component-type": "ruby",
				"app.kubernetes.io/name":           "apptmp",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{
						"app.kubernetes.io/url":                   "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"deploymentconfig": "wildfly",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "wildfly:latest",
									Name:  "wildfly",
								},
							},
						},
					},
				},
			},
			createdPVC: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Labels: map[string]string{
						"app": "apptmp",
						"app.kubernetes.io/component-name": "wildfly",
						"app.kubernetes.io/component-type": "wildfly",
						"app.kubernetes.io/name":           "apptmp",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: quantity,
						},
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
				},
			},

			wantErr: true,
		},
		{
			name:        "setup with wrong pvc specs",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			labels: map[string]string{
				"app": "apptmp",
				"app.kubernetes.io/component-name": "ruby",
				"app.kubernetes.io/component-type": "ruby",
				"app.kubernetes.io/name":           "apptmp",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{
						"app.kubernetes.io/url":                   "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"deploymentconfig": "wildfly",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "wildfly:latest",
									Name:  "wildfly",
								},
							},
						},
					},
				},
			},
			createdPVC: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-s2idata", "wildfly"),
					Labels: map[string]string{
						"app": "apptmp",
						"app.kubernetes.io/component-name": "wildfly",
						"app.kubernetes.io/component-type": "wildfly",
						"app.kubernetes.io/name":           "apptmp",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: errQuantity,
						},
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
				},
			},
			wantErr: true,
		},
		{
			name:        "setup with non existing dc",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			labels: map[string]string{
				"app": "apptmp",
				"app.kubernetes.io/component-name": "ruby",
				"app.kubernetes.io/component-type": "ruby",
				"app.kubernetes.io/name":           "apptmp",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
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
					return true, nil, fmt.Errorf("'get' called with different dcName")
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

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				createdPvc := action.(ktesting.CreateAction).GetObject().(*corev1.PersistentVolumeClaim)
				if createdPvc.ObjectMeta.Name != tt.createdPVC.ObjectMeta.Name {
					return true, nil, fmt.Errorf("got a different pvc name")
				}
				if !reflect.DeepEqual(createdPvc.Spec, tt.createdPVC.Spec) {
					return true, nil, fmt.Errorf("got a different pvc spec")
				}
				return true, &tt.createdPVC, nil
			})

			err := fkclient.SetupForSupervisor(tt.dcName, tt.projectName, tt.annotations, tt.labels)

			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.AppsClientset.Actions()) != 2) && (tt.wantErr != true) {
					t.Errorf("expected 2 action in UpdateDeploymentConfig got: %v", fkclientset.AppsClientset.Actions())
				}

				// Check for validating actions performed
				if (len(fkclientset.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
					t.Errorf("expected 1 action in CreatePVC got: %v", fkclientset.AppsClientset.Actions())
				}

				updatedDc := fkclientset.AppsClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
				if updatedDc.Name != tt.dcName {
					t.Errorf("deploymentconfig name is not matching with expected value, expected: %s, got %s", tt.dcName, updatedDc.Name)
				}

				createdPVC := fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.PersistentVolumeClaim)
				if createdPVC.ObjectMeta.Name != fmt.Sprintf("%s-s2idata", tt.dcName) {
					t.Errorf("pvc name is not matching with expected value, expected: %s, got %s", tt.createdPVC.ObjectMeta.Name, createdPVC.ObjectMeta.Name)
				}

				if !reflect.DeepEqual(updatedDc.Annotations, tt.annotations) {
					t.Errorf("deployment Config annotations not matching with expected values, expected: %s, got %s", tt.annotations, updatedDc.Annotations)
				}

				if !reflect.DeepEqual(createdPVC.Name, tt.createdPVC.Name) {
					t.Errorf("created PVC not matching with expected values, expected: %v, got %v", tt.createdPVC, createdPVC)
				}

				if !reflect.DeepEqual(createdPVC.Spec, tt.createdPVC.Spec) {
					t.Errorf("created PVC spec not matching with expected values, expected: %v, got %v", createdPVC.Spec, tt.createdPVC.Spec)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestCleanupAfterSupervisor(t *testing.T) {
	tests := []struct {
		name        string
		dcName      string
		projectName string
		annotations map[string]string
		existingDc  appsv1.DeploymentConfig
		pvcName     string
		wantErr     bool
	}{
		{
			name:        "proper parameters and one volume,volumeMount and initContainer",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{"app.kubernetes.io/url": "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "wildfly:latest",
									Name:  "wildfly",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name: getAppRootVolumeName("wildfly"),
										},
									},
								},
							},
							InitContainers: []corev1.Container{
								{
									Name:  "copy-files-to-volume",
									Image: "wildfly:latest",
									Command: []string{
										"copy-files-to-volume",
										"/opt/app-root",
										"/mnt/app-root"},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      getAppRootVolumeName("wildfly"),
											MountPath: "/mnt",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: getAppRootVolumeName("wildfly"),
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: getAppRootVolumeName("wildfly"),
										},
									},
								},
							},
						},
					},
				},
			},
			pvcName: fmt.Sprintf("%s-s2idata", "wildfly"),
			wantErr: false,
		},
		{
			name:        "proper parameters and two volume,volumeMount and initContainer",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wildfly",
					Annotations: map[string]string{"app.kubernetes.io/url": "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "wildfly:latest",
									Name:  "wildfly",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name: getAppRootVolumeName("wildfly"),
										},
										{
											Name: "backend",
										},
									},
								},
							},
							InitContainers: []corev1.Container{
								{
									Name:  "copy-files-to-volume",
									Image: "wildfly:latest",
									Command: []string{
										"copy-files-to-volume",
										"/opt/app-root",
										"/mnt/app-root"},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      getAppRootVolumeName("wildfly"),
											MountPath: "/mnt",
										},
									},
								},
								{
									Name:  "xyz",
									Image: "xyz:latest",
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: getAppRootVolumeName("wildfly"),
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: getAppRootVolumeName("wildfly"),
										},
									},
								},
								{
									Name: "backend",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "backend",
										},
									},
								},
							},
						},
					},
				},
			},
			pvcName: fmt.Sprintf("%s-s2idata", "wildfly"),
			wantErr: false,
		},
		{
			name:        "proper parameters and one volume,volumeMount and initContainer and non existing dc",
			dcName:      "wildfly",
			projectName: "project-testing",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingDc: appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
					Annotations: map[string]string{"app.kubernetes.io/url": "https://github.com/sclorg/nodejs-ex",
						"app.kubernetes.io/component-source-type": "git",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "wildfly:latest",
									Name:  "wildfly",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name: getAppRootVolumeName("wildfly"),
										},
									},
								},
							},
							InitContainers: []corev1.Container{
								{
									Name:  "copy-files-to-volume",
									Image: "wildfly:latest",
									Command: []string{
										"copy-files-to-volume",
										"/opt/app-root",
										"/mnt/app-root"},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      getAppRootVolumeName("wildfly"),
											MountPath: "/mnt",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: getAppRootVolumeName("wildfly"),
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: getAppRootVolumeName("wildfly"),
										},
									},
								},
							},
						},
					},
				},
			},
			pvcName: fmt.Sprintf("%s-s2idata", "wildfly"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lenVolumes := len(tt.existingDc.Spec.Template.Spec.Volumes)
			lenVolumeMounts := len(tt.existingDc.Spec.Template.Spec.Containers[0].VolumeMounts)
			lenInitContainer := len(tt.existingDc.Spec.Template.Spec.InitContainers)

			fkclient, fkclientset := FakeNew()
			fkclientset.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dcName := action.(ktesting.GetAction).GetName()
				if tt.dcName != dcName {
					return true, nil, fmt.Errorf("got different dc")
				}
				return true, &tt.existingDc, nil
			})

			fkclientset.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				dc := action.(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
				if dc.Name != tt.dcName {
					return true, nil, fmt.Errorf("got different dc")
				}
				return true, dc, nil
			})

			fkclientset.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvcName := action.(ktesting.DeleteAction).GetName()
				if pvcName != getAppRootVolumeName(tt.dcName) {
					return true, nil, fmt.Errorf("the pvc name is not matching the required pvc name")
				}
				return true, nil, nil
			})

			err := fkclient.CleanupAfterSupervisor(tt.dcName, tt.projectName, tt.annotations)
			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.AppsClientset.Actions()) != 2) && (tt.wantErr != true) {
					t.Errorf("expected 2 action in UpdateDeploymentConfig got: %v", fkclientset.AppsClientset.Actions())
				}

				// Check for validating actions performed
				if (len(fkclientset.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
					t.Errorf("expected 1 action in CreatePVC got: %v", fkclientset.Kubernetes.Actions())
				}

				updatedDc := fkclientset.AppsClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
				if updatedDc.Name != tt.dcName {
					t.Errorf("deploymentconfig name is not matching with expected value, expected: %s, got %s", tt.dcName, updatedDc.Name)
				}

				if !reflect.DeepEqual(updatedDc.Annotations, tt.annotations) {
					t.Errorf("deployment Config annotations not matching with expected values, expected: %s, got %s", tt.annotations, updatedDc.Annotations)
				}

				if lenVolumes == len(updatedDc.Spec.Template.Spec.Volumes) {
					t.Errorf("could not remove any volumes, expected: %s, got %s", "0 volume", "1 volume")
				}

				if lenVolumeMounts == len(updatedDc.Spec.Template.Spec.Containers[0].VolumeMounts) {
					t.Errorf("could not remove any volumeMounts, expected: %s, got %s", "0 volume", "1 volume")
				}

				if lenInitContainer == len(updatedDc.Spec.Template.Spec.InitContainers) {
					t.Errorf("could not remove any volumeMounts, expected: %s, got %s", "0 volume", "1 volume")
				}

				if lenVolumes-1 != len(updatedDc.Spec.Template.Spec.Volumes) {
					t.Errorf("wrong number of Volumes deleted, expected: %s, removed %d", "1 volume", lenVolumes-len(updatedDc.Spec.Template.Spec.Volumes))
				}

				if lenInitContainer-1 != len(updatedDc.Spec.Template.Spec.InitContainers) {
					t.Errorf("wrong number of InitContainer deleted, expected: %s, removed %d", "1 initContainer", lenInitContainer-len(updatedDc.Spec.Template.Spec.InitContainers))
				}

				if lenVolumeMounts-1 != len(updatedDc.Spec.Template.Spec.Containers[0].VolumeMounts) {
					t.Errorf("wrong number of VolumeMounts deleted, expected: %s, removed %d", "1 volumeMount", lenInitContainer-len(updatedDc.Spec.Template.Spec.Containers[0].VolumeMounts))
				}

				for _, initContainer := range updatedDc.Spec.Template.Spec.InitContainers {
					if initContainer.Name == "copy-files-to-volume" {
						t.Errorf("could not remove 'copy-volume-to-volume' InitContainer and instead some other InitContainer was removed")
					}
				}

				for _, volume := range updatedDc.Spec.Template.Spec.Volumes {
					if volume.Name == getAppRootVolumeName(tt.dcName) {
						t.Errorf("could not remove %s volume", getAppRootVolumeName(tt.dcName))
					}
				}

				for _, volumeMount := range updatedDc.Spec.Template.Spec.Containers[0].VolumeMounts {
					if volumeMount.Name == getAppRootVolumeName(tt.dcName) {
						t.Errorf("could not remove %s volumeMount", getAppRootVolumeName(tt.dcName))
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

func TestGetBuildConfig(t *testing.T) {
	tests := []struct {
		name                string
		buildName           string
		projectName         string
		returnedBuildConfig buildv1.BuildConfig
		wantErr             bool
	}{
		{
			name:        "buildConfig with existing bc",
			buildName:   "nodejs",
			projectName: "project-app",
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

			build, err := fkclient.GetBuildConfig(tt.buildName, tt.projectName)
			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.BuildClientset.Actions()) != 1) && (tt.wantErr != true) {
					t.Errorf("expected 1 action in GetBuildConfig got: %v", fkclientset.AppsClientset.Actions())
				}
				if build.Name != tt.buildName {
					t.Errorf("wrong GetBuildConfig got: %v, expected: %v", build.Name, tt.buildName)
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
		projectName         string
		gitUrl              string
		annotations         map[string]string
		existingBuildConfig buildv1.BuildConfig
		updatedBuildConfig  buildv1.BuildConfig
		wantErr             bool
	}{
		{
			name:            "git to local with proper parameters",
			buildConfigName: "nodejs",
			projectName:     "app",
			gitUrl:          "",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
				"app.kubernetes.io/component-source-type": "local",
			},
			existingBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
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
			updatedBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
					Annotations: map[string]string{
						"app.kubernetes.io/url":                   "file:///temp/nodejs-ex",
						"app.kubernetes.io/component-source-type": "local",
					},
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{
						Source: buildv1.BuildSource{
							Git: &buildv1.GitBuildSource{
								URI: bootstrapperURI,
								Ref: bootstrapperRef,
							},
							Type: buildv1.BuildSourceGit,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:            "local to git with proper parameters",
			buildConfigName: "nodejs",
			projectName:     "app",
			gitUrl:          "https://github.com/sclorg/nodejs-ex",
			annotations: map[string]string{
				"app.kubernetes.io/url":                   "https://github.com/sclorg/nodejs-ex",
				"app.kubernetes.io/component-source-type": "git",
			},
			existingBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{
						Source: buildv1.BuildSource{
							Git: &buildv1.GitBuildSource{
								URI: bootstrapperURI,
								Ref: bootstrapperRef,
							},
							Type: buildv1.BuildSourceGit,
						},
					},
				},
			},
			updatedBuildConfig: buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs",
					Annotations: map[string]string{
						"app.kubernetes.io/url":                   "https://github.com/sclorg/nodejs-ex",
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

			err := fkclient.UpdateBuildConfig(tt.buildConfigName, tt.projectName, tt.gitUrl, tt.annotations)
			if err == nil && !tt.wantErr {
				// Check for validating actions performed
				if (len(fkclientset.BuildClientset.Actions()) != 2) && (tt.wantErr != true) {
					t.Errorf("expected 2 action in GetBuildConfig got: %v", fkclientset.BuildClientset.Actions())
				}

				updatedDc := fkclientset.BuildClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*buildv1.BuildConfig)
				if !reflect.DeepEqual(updatedDc.Annotations, tt.annotations) {
					t.Errorf("deployment Config annotations not matching with expected values, expected: %s, got %s", tt.annotations, updatedDc.Annotations)
				}

				if !reflect.DeepEqual(updatedDc.Spec, tt.updatedBuildConfig.Spec) {
					t.Errorf("deployment Config Spec not matching with expected values, expected: %v, got %v", tt.updatedBuildConfig.Spec, updatedDc.Spec)
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
		name         string
		namespace    string
		builderImage string
		gitUrl       string
		labels       map[string]string
		annotations  map[string]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: with valid gitUrl",
			args: args{
				name:         "ruby",
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitUrl:       "https://github.com/openshift/ruby",
				labels: map[string]string{
					"app": "apptmp",
					"app.kubernetes.io/component-name": "ruby",
					"app.kubernetes.io/component-type": "ruby",
					"app.kubernetes.io/name":           "apptmp",
				},
				annotations: map[string]string{
					"app.kubernetes.io/url":                   "https://github.com/openshift/ruby",
					"app.kubernetes.io/component-source-type": "git",
				},
			},
			wantErr: false,
		},

		{
			name: "case 2 : binary buildSource with gitUrl empty",
			args: args{
				name:         "ruby",
				builderImage: "ruby:latest",
				namespace:    "testing",
				gitUrl:       "",
				labels: map[string]string{
					"app": "apptmp",
					"app.kubernetes.io/component-name": "ruby",
					"app.kubernetes.io/component-type": "ruby",
					"app.kubernetes.io/name":           "apptmp",
				},
				annotations: map[string]string{
					"app.kubernetes.io/url":                   "https://github.com/openshift/ruby",
					"app.kubernetes.io/component-source-type": "git",
				},
			},
			wantErr: false,
		},

		// TODO: Currently fails. Enable this case once fixed
		// {
		// 	name: "case 3: with empty builderImage",
		// 	args: args{
		// 		name:         "ruby",
		// 		builderImage: "",
		// 		gitUrl:       "https://github.com/openshift/ruby",
		// 		labels: map[string]string{
		// 			"app": "apptmp",
		// 			"app.kubernetes.io/component-name": "ruby",
		// 			"app.kubernetes.io/component-type": "ruby",
		// 			"app.kubernetes.io/name":           "apptmp",
		// 		},
		// 		annotations: map[string]string{
		// 			"app.kubernetes.io/url":                   "https://github.com/openshift/ruby",
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
				return true, fakeImageStreams(tt.args.name, tt.args.namespace), nil
			})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStream(tt.args.name, tt.args.namespace), nil
			})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreamimages", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreamImages(tt.args.name), nil
			})

			err := fkclient.NewAppS2I(tt.args.name,
				tt.args.builderImage,
				tt.args.gitUrl,
				tt.args.labels,
				tt.args.annotations)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewAppS2I() error = %#v, wantErr %#v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.ImageClientset.Actions()) != 3 {
					t.Errorf("expected 3 ImageClientset.Actions() in NewAppS2I, got: %v", fkclientset.ImageClientset.Actions())
				}

				if len(fkclientset.BuildClientset.Actions()) != 1 {
					t.Errorf("expected 1 BuildClientset.Actions() in NewAppS2I, got: %v", fkclientset.BuildClientset.Actions())
				}

				if len(fkclientset.AppsClientset.Actions()) != 1 {
					t.Errorf("expected 1 AppsClientset.Actions() in NewAppS2I, go: %v", fkclientset.AppsClientset.Actions())
				}

				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 Kubernetes.Actions() in NewAppS2I, go: %v", fkclientset.Kubernetes.Actions())
				}

				// Check for imagestream objects
				createdIS := fkclientset.ImageClientset.Actions()[2].(ktesting.CreateAction).GetObject().(*imagev1.ImageStream)

				if createdIS.Name != tt.args.name {
					t.Errorf("imagestream name is not matching with expected name, expected: %s, got %s", tt.args.name, createdIS.Name)
				}

				if !reflect.DeepEqual(createdIS.Labels, tt.args.labels) {
					t.Errorf("imagestream labels not matching with expected values, expected: %s, got %s", tt.args.labels, createdIS.Labels)
				}

				if !reflect.DeepEqual(createdIS.Annotations, tt.args.annotations) {
					t.Errorf("imagestream annotations not matching with expected values, expected: %s, got %s", tt.args.annotations, createdIS.Annotations)
				}

				// Check buildconfig objects
				createdBC := fkclientset.BuildClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*buildv1.BuildConfig)

				if tt.args.gitUrl != "" {
					if createdBC.Spec.CommonSpec.Source.Git.URI != tt.args.gitUrl {
						t.Errorf("git url is not matching with expected value, expected: %s, got %s", tt.args.gitUrl, createdBC.Spec.CommonSpec.Source.Git.URI)
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
				if createdDC.Spec.Selector["deploymentconfig"] != tt.args.name {
					t.Errorf("deploymentconfig name is not matching with expected value, expected: %s, got %s", tt.args.name, createdDC.Spec.Selector["deploymentconfig"])
				}

				createdSvc := fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.Service)

				// ExposedPorts 8080 in fakeImageStreamImages()
				if createdSvc.Spec.Ports[0].Port != 8080 {
					t.Errorf("Svc port not matching, expected: 8080, got %v", createdSvc.Spec.Ports[0].Port)
				}

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
									{DockerImageReference: "example/ruby:latest"},
									{Generation: 1},
									{Image: "ruby@sha256:9579a93ee"},
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

			err := fkclient.WaitForBuildToFinish(tt.buildName)
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
			pod, err := fkclient.WaitAndGetPod(podSelector)

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

func TestCreateNewProject(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		wantErr  bool
	}{
		{
			name:     "Case 1: valid project name",
			projName: "testing",
			wantErr:  false,
		},

		// TODO: Currently fails. Enable once fixed.
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

			err := fkclient.CreateNewProject(tt.projName)
			if !tt.wantErr == (err != nil) {
				t.Errorf("client.CreateNewProject(string) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if len(fkclientset.ProjClientset.Actions()) != 1 {
				t.Errorf("expected 1 action in CreateNewProject got: %v", fkclientset.ProjClientset.Actions())
			}

			if err == nil {
				createdProj := fkclientset.ProjClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*projectv1.ProjectRequest)

				if createdProj.Name != tt.projName {
					t.Errorf("project name does not match the expected name, expected: %s, got: %s", tt.projName, createdProj.Name)
				}
			}
		})
	}
}

func TestListRoutes(t *testing.T) {
	tests := []struct {
		name          string
		labelSelector string
		wantLabels    map[string]string
		routesList    routev1.RouteList
		wantErr       bool
	}{
		{
			name:          "existing url",
			labelSelector: "app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=app",
			wantLabels: map[string]string{
				applabels.ApplicationLabel:     "app",
				componentlabels.ComponentLabel: "nodejs",
			},
			routesList: routev1.RouteList{
				Items: []routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "nodejs",
							Labels: map[string]string{
								applabels.ApplicationLabel:     "app",
								componentlabels.ComponentLabel: "nodejs",
							},
						},
						Spec: routev1.RouteSpec{
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "nodejs-app",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "wildfly",
							Labels: map[string]string{
								applabels.ApplicationLabel:     "app",
								componentlabels.ComponentLabel: "wildfly",
								urlLabels.UrlLabel:             "wildfly",
							},
						},
						Spec: routev1.RouteSpec{
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "wildfly-app",
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

		fakeClientSet.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
			if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.labelSelector) {
				return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.labelSelector, action.(ktesting.ListAction).GetListRestrictions())
			}
			return true, &tt.routesList, nil
		})

		_, err := client.ListRoutes(tt.labelSelector)
		if err == nil && !tt.wantErr {
			if (len(fakeClientSet.RouteClientset.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in ListRoutes got: %v", fakeClientSet.RouteClientset.Actions())
			}
		} else if err == nil && tt.wantErr {
			t.Error("error was expected, but no error was returned")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
		}
	}
}
