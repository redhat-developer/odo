package url

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/openshift/odo/v2/pkg/localConfigProvider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertEnvinfoURL(t *testing.T) {
	serviceName := "testService"
	urlName := "testURL"
	host := "com"
	secretName := "test-tls-secret"
	tests := []struct {
		name       string
		envInfoURL localConfigProvider.LocalURL
		wantURL    URL
	}{
		{
			name: "Case 1: insecure URL",
			envInfoURL: localConfigProvider.LocalURL{
				Name:   urlName,
				Host:   host,
				Port:   8080,
				Secure: false,
				Kind:   localConfigProvider.INGRESS,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Host: fmt.Sprintf("%s.%s", urlName, host), Port: 8080, Secure: false, Kind: localConfigProvider.INGRESS},
			},
		},
		{
			name: "Case 2: secure Ingress URL without tls secret defined",
			envInfoURL: localConfigProvider.LocalURL{
				Name:   urlName,
				Host:   host,
				Port:   8080,
				Secure: true,
				Kind:   localConfigProvider.INGRESS,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Host: fmt.Sprintf("%s.%s", urlName, host), Port: 8080, Secure: true, TLSSecret: fmt.Sprintf("%s-tlssecret", serviceName), Kind: localConfigProvider.INGRESS},
			},
		},
		{
			name: "Case 3: secure Ingress URL with tls secret defined",
			envInfoURL: localConfigProvider.LocalURL{
				Name:      urlName,
				Host:      host,
				Port:      8080,
				Secure:    true,
				TLSSecret: secretName,
				Kind:      localConfigProvider.INGRESS,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Host: fmt.Sprintf("%s.%s", urlName, host), Port: 8080, Secure: true, TLSSecret: secretName, Kind: localConfigProvider.INGRESS},
			},
		},
		{
			name: "Case 4: Insecure route URL",
			envInfoURL: localConfigProvider.LocalURL{
				Name: urlName,
				Port: 8080,
				Kind: localConfigProvider.ROUTE,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Port: 8080, Secure: false, Kind: localConfigProvider.ROUTE},
			},
		},
		{
			name: "Case 4: Secure route URL",
			envInfoURL: localConfigProvider.LocalURL{
				Name:   urlName,
				Port:   8080,
				Secure: true,
				Kind:   localConfigProvider.ROUTE,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Port: 8080, Secure: true, Kind: localConfigProvider.ROUTE},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := ConvertEnvInfoURL(tt.envInfoURL, serviceName)
			if !reflect.DeepEqual(url, tt.wantURL) {
				t.Errorf("Expected %v, got %v", tt.wantURL, url)
			}
		})
	}
}

func TestGetURLString(t *testing.T) {
	cases := []struct {
		name          string
		protocol      string
		URL           string
		ingressDomain string
		expected      string
	}{
		{
			name:          "all blank without s2i",
			protocol:      "",
			URL:           "",
			ingressDomain: "",
			expected:      "",
		},
		{
			name:          "devfile case",
			protocol:      "http",
			URL:           "",
			ingressDomain: "spring-8080.192.168.39.247.nip.io",
			expected:      "http://spring-8080.192.168.39.247.nip.io",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			output := GetURLString(testCase.protocol, testCase.URL, testCase.ingressDomain)
			if output != testCase.expected {
				t.Errorf("Expected: %v, got %v", testCase.expected, output)
			}
		})
	}
}
