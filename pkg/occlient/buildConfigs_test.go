package occlient

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

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

			err := fkclient.WaitForBuildToFinish(tt.buildName, os.Stdout, time.Second*5)
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
