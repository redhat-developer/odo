package url

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openshift/odo/v2/pkg/kclient"
	"github.com/openshift/odo/v2/pkg/kclient/fake"
	"github.com/openshift/odo/v2/pkg/localConfigProvider"
	"github.com/openshift/odo/v2/pkg/occlient"
	"github.com/openshift/odo/v2/pkg/testingutil"
	kappsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestPush(t *testing.T) {
	type deleteParameters struct {
		string
		localConfigProvider.URLKind
	}
	type args struct {
		isRouteSupported             bool
		networkingV1IngressSupported bool
		extensionV1IngressSupported  bool
	}
	tests := []struct {
		name                string
		args                args
		componentName       string
		applicationName     string
		existingLocalURLs   []localConfigProvider.LocalURL
		existingClusterURLs URLList
		deletedItems        []deleteParameters
		createdURLs         []URL
		wantErr             bool
	}{
		{
			name: "no urls on local config and cluster",
			args: args{
				isRouteSupported:             true,
				networkingV1IngressSupported: false,
				extensionV1IngressSupported:  true,
			},
			componentName:   "nodejs",
			applicationName: "app",
		},
		{
			name:            "2 urls on local config and 0 on openshift cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args: args{
				isRouteSupported:             true,
				networkingV1IngressSupported: true,
				extensionV1IngressSupported:  false,
			},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "example-1",
					Port:   9090,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				}),
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example-1",
					Port:   9090,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				}),
			},
		},
		{
			name:            "0 url on local config and 2 on openshift cluster",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: false, extensionV1IngressSupported: true},
			existingClusterURLs: NewURLList([]URL{
				NewURL(testingutil.GetSingleRoute("example", 8080, "wildfly", "app")),
				NewURL(testingutil.GetSingleRoute("example-1", 9100, "wildfly", "app")),
			}),
			deletedItems: []deleteParameters{
				{"example", localConfigProvider.ROUTE},
				{"example-1", localConfigProvider.ROUTE},
			},
		},
		{
			name:            "2 url on local config and 2 on openshift cluster, but they are different",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "example-local-1",
					Port:   9090,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURL(testingutil.GetSingleRoute("example", 8080, "wildfly", "app")),
				NewURL(testingutil.GetSingleRoute("example-1", 9100, "wildfly", "app")),
			}),
			deletedItems: []deleteParameters{
				{"example", localConfigProvider.ROUTE},
				{"example-1", localConfigProvider.ROUTE},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				}),
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example-local-1",
					Port:   9090,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				}),
			},
		},
		{
			name:            "5 urls (both types and different configurations) on config and openshift cluster are in sync",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Path:   "/",
					Host:   "com",
					Kind:   localConfigProvider.INGRESS,
				},
				{
					Name:   "example-1",
					Port:   9100,
					Secure: false,
					Path:   "/",
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "example-default-secret",
					Port:   8080,
					Secure: true,
					Path:   "/",
					Host:   "com",
					Kind:   localConfigProvider.INGRESS,
				},
				{
					Name:      "example-user-secret",
					Port:      8080,
					Secure:    true,
					Path:      "/",
					Host:      "com",
					TLSSecret: "secret-name",
					Kind:      localConfigProvider.INGRESS,
				},
				{
					Name:   "example-11",
					Port:   9100,
					Secure: true,
					Path:   "/",
					Kind:   localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example", "nodejs", "app", true, false), false),
				NewURLFromKubernetesIngress(fake.GetSingleSecureKubernetesIngress("example-default-secret", "nodejs", "app", "", true, false), false),
				NewURLFromKubernetesIngress(fake.GetSingleSecureKubernetesIngress("example-user-secret", "nodejs", "app", "secret-name", true, false), false),
				NewURL(testingutil.GetSingleRoute("example-1", 9100, "nodejs", "app")),
				NewURL(testingutil.GetSingleSecureRoute("example-11", 9100, "nodejs", "app")),
			}),
			createdURLs: []URL{},
		},
		{
			name:              "0 urls on env file and cluster",
			componentName:     "nodejs",
			applicationName:   "app",
			args:              args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{},
		},
		{
			name:            "2 urls on env file and 0 on openshift cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
				{
					Name: "example-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				}),
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				}),
			},
		},
		{
			name:              "0 urls on env file and 2 on openshift cluster",
			componentName:     "nodejs",
			applicationName:   "app",
			args:              args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{},
			existingClusterURLs: NewURLList([]URL{
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example-0", "nodejs", "app", true, false), false),
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example-1", "nodejs", "app", true, false), false),
			}),
			deletedItems: []deleteParameters{
				{"example-0", localConfigProvider.INGRESS},
				{"example-1", localConfigProvider.INGRESS},
			},
		},
		{
			name:            "2 urls on env file and 2 on openshift cluster, but they are different",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-local-0",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
				{
					Name: "example-local-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example-0", "nodejs", "app", true, false), false),
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example-1", "nodejs", "app", true, false), false),
			}),
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example-local-0",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				}),
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example-local-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				}),
			},
			deletedItems: []deleteParameters{
				{"example-0", localConfigProvider.INGRESS},
				{"example-1", localConfigProvider.INGRESS},
			},
		},
		{
			name:            "2 urls on env file and openshift cluster are in sync",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:     "example-0",
					Host:     "com",
					Port:     8080,
					Secure:   false,
					Protocol: "http",
					Kind:     localConfigProvider.INGRESS,
					Path:     "/",
				},
				{
					Name:     "example-1",
					Host:     "com",
					Port:     9090,
					Secure:   false,
					Protocol: "http",
					Kind:     localConfigProvider.INGRESS,
					Path:     "/",
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURLFromKubernetesIngress(fake.GetKubernetesIngressListWithMultiple("wildfly", "app", true, false).Items[0], false),
				NewURLFromKubernetesIngress(fake.GetKubernetesIngressListWithMultiple("wildfly", "app", true, false).Items[1], false),
			}),
			createdURLs:  []URL{},
			deletedItems: []deleteParameters{},
		},
		{
			name:            "2 (1 ingress,1 route) urls on env file and 2 on openshift cluster (1 ingress,1 route), but they are different",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-local-0",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				},
				{
					Name: "example-local-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example-0", "nodejs", "app", true, false), false),
				NewURL(testingutil.GetSingleRoute("example-1", 9090, "nodejs", "app")),
			}),
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example-local-0",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				}),
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example-local-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				}),
			},
			deletedItems: []deleteParameters{
				{"example-0", localConfigProvider.INGRESS},
				{"example-1", localConfigProvider.ROUTE},
			},
		},
		{
			name:            "create a ingress on a kubernetes cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: false, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:      "example",
					Host:      "com",
					TLSSecret: "secret",
					Port:      8080,
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:      "example",
					Host:      "com",
					TLSSecret: "secret",
					Port:      8080,
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				}),
			},
		},
		{
			name:            "url with same name exists on env and cluster but with different specs",
			componentName:   "nodejs",
			applicationName: "app",
			args: args{
				isRouteSupported:             true,
				networkingV1IngressSupported: true,
				extensionV1IngressSupported:  false,
			},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-local-0",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURLFromKubernetesIngress(fake.GetSingleKubernetesIngress("example-local-0", "nodejs", "app", true, false), false),
			}),
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example-local-0",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				}),
			},
			deletedItems: []deleteParameters{
				{"example-local-0", localConfigProvider.INGRESS},
			},
			wantErr: false,
		},
		{
			name:            "url with same name exists on config and cluster but with different specs",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: NewURLList([]URL{
				NewURL(testingutil.GetSingleRoute("example-local-0-app", 9090, "nodejs", "app")),
			}),
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				}),
			},
			deletedItems: []deleteParameters{
				{"example-local-0-app", localConfigProvider.ROUTE},
			},
			wantErr: false,
		},
		{
			name:            "create a secure route url",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: true,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example",
					Port:   8080,
					Secure: true,
					Kind:   localConfigProvider.ROUTE,
				}),
			},
		},
		{
			name:            "create a secure ingress url with empty user given tls secret",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Host:   "com",
					Secure: true,
					Port:   8080,
					Kind:   localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example",
					Host:   "com",
					Secure: true,
					Port:   8080,
					Kind:   localConfigProvider.INGRESS,
				}),
			},
		},
		{
			name:            "create a secure ingress url with user given tls secret",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:      "example",
					Host:      "com",
					TLSSecret: "secret",
					Port:      8080,
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:      "example",
					Host:      "com",
					TLSSecret: "secret",
					Port:      8080,
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				}),
			},
		},
		{
			name:          "no host defined for ingress should not create any URL",
			componentName: "nodejs",
			args:          args{isRouteSupported: false, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				},
			},
			wantErr:     false,
			createdURLs: []URL{},
		},
		{
			name:            "should create route in openshift cluster if endpoint is defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Kind:   localConfigProvider.ROUTE,
					Secure: false,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example",
					Port:   8080,
					Kind:   localConfigProvider.ROUTE,
					Secure: false,
				}),
			},
		},
		{
			name:            "should create ingress if endpoint is defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name: "example",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				}),
			},
		},
		{
			name:            "should create route in openshift cluster with path defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Path:   "/testpath",
					Kind:   localConfigProvider.ROUTE,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Path:   "/testpath",
					Kind:   localConfigProvider.ROUTE,
				}),
			},
		},
		{
			name:            "should create ingress with path defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, networkingV1IngressSupported: true, extensionV1IngressSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Host:   "com",
					Port:   8080,
					Secure: false,
					Path:   "/testpath",
					Kind:   localConfigProvider.INGRESS,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				NewURLFromLocalURL(localConfigProvider.LocalURL{
					Name:   "example",
					Host:   "com",
					Port:   8080,
					Secure: false,
					Path:   "/testpath",
					Kind:   localConfigProvider.INGRESS,
				}),
			},
		},
	}
	for _, tt := range tests {
		//tt.name = fmt.Sprintf("case %d: ", testNum+1) + tt.name
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfigProvider := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfigProvider.EXPECT().GetName().Return(tt.componentName).AnyTimes()
			mockLocalConfigProvider.EXPECT().GetApplication().Return(tt.applicationName).AnyTimes()
			mockLocalConfigProvider.EXPECT().ListURLs().Return(tt.existingLocalURLs, nil)

			mockURLClient := NewMockClient(ctrl)
			mockURLClient.EXPECT().ListFromCluster().Return(tt.existingClusterURLs, nil)

			for i := range tt.createdURLs {
				mockURLClient.EXPECT().Create(tt.createdURLs[i]).Times(1)
			}

			for i := range tt.deletedItems {
				mockURLClient.EXPECT().Delete(gomock.Eq(tt.deletedItems[i].string), gomock.Eq(tt.deletedItems[i].URLKind)).Times(1)
			}

			fakeClient, _ := occlient.FakeNew()
			fakeKClient, fakeKClientSet := kclient.FakeNewWithIngressSupports(tt.args.networkingV1IngressSupported, tt.args.extensionV1IngressSupported)

			fakeKClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kappsv1.DeploymentList{
					Items: []kappsv1.Deployment{
						*testingutil.CreateFakeDeployment(tt.componentName),
					},
				}, nil
			})

			fakeClient.SetKubeClient(fakeKClient)

			if err := Push(PushParameters{
				LocalConfigProvider: mockLocalConfigProvider,
				URLClient:           mockURLClient,
				IsRouteSupported:    tt.args.isRouteSupported,
			}); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
