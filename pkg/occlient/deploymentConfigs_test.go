package occlient

import (
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
	"reflect"
	"testing"
	"time"
)

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

func TestListDeploymentConfigs(t *testing.T) {
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
			dc, err := fakeClient.ListDeploymentConfigs(tt.selector)

			if len(fakeClientSet.AppsClientset.Actions()) != 1 {
				t.Errorf("expected 1 AppsClientset.Actions() in ListDeploymentConfigs, got: %v", fakeClientSet.AppsClientset.Actions())
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

func TestGetDeploymentConfigFromSelector(t *testing.T) {
	type args struct {
		selector string
	}
	tests := []struct {
		name           string
		args           args
		returnedDCList *appsv1.DeploymentConfigList
		want           *appsv1.DeploymentConfig
		wantErr        bool
	}{
		{
			name: "case 1: only one dc returned",
			args: args{
				"app=app",
			},
			returnedDCList: &appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{
					*testingutil.OneFakeDeploymentConfigWithMounts("comp-0", "nodejs", "app", nil),
				},
			},
			want: testingutil.OneFakeDeploymentConfigWithMounts("comp-0", "nodejs", "app", nil),
		},
		{
			name: "case 2: no dc returned",
			args: args{
				"app=app",
			},
			returnedDCList: &appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{},
			},
			wantErr: true,
		},
		{
			name: "case 3: two dc returned",
			args: args{
				"app=app",
			},
			returnedDCList: &appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{
					*testingutil.OneFakeDeploymentConfigWithMounts("comp-0", "nodejs", "app", nil),
					*testingutil.OneFakeDeploymentConfigWithMounts("comp-1", "nodejs", "app", nil),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.args.selector) {
					return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.args.selector, action.(ktesting.ListAction).GetListRestrictions())
				}
				return true, tt.returnedDCList, nil
			})

			got, err := fakeClient.GetDeploymentConfigFromSelector(tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeploymentConfigFromSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeploymentConfigFromSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}
