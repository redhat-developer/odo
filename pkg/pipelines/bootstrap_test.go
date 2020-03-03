package pipelines

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateImageRepo(t *testing.T) {

	errorMsg := "failed to parse image repo:%s, expected image repository in the form <registry>/<username>/<repository> or <project>/<app> for internal registry"

	tests := []struct {
		description                string
		options                    BootstrapParameters
		expectedError              string
		expectedIsInternalRegistry bool
		expectedImageRepo          string
	}{
		{
			"Valid image regsitry URL",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sample-user/sample-repo",
			},
			"",
			false,
			"quay.io/sample-user/sample-repo",
		},
		{
			"Valid image regsitry URL random registry",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "random.io/sample-user/sample-repo",
			},
			"",
			false,
			"random.io/sample-user/sample-repo",
		},
		{
			"Valid image regsitry URL docker.io",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io/sample-user/sample-repo",
			},
			"",
			false,
			"docker.io/sample-user/sample-repo",
		},
		{
			"Invalid image registry URL with missing repo name",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sample-user",
			},
			fmt.Sprintf(errorMsg, "quay.io/sample-user"),
			false,
			"",
		},
		{
			"Invalid image registry URL with missing repo name docker.io",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io/sample-user",
			},
			fmt.Sprintf(errorMsg, "docker.io/sample-user"),
			false,
			"",
		},
		{
			"Invalid image registry URL with whitespaces",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sample-user/ ",
			},
			fmt.Sprintf(errorMsg, "quay.io/sample-user/ "),
			false,
			"",
		},
		{
			"Invalid image registry URL with whitespaces in between",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sam\tple-user/",
			},
			fmt.Sprintf(errorMsg, "quay.io/sam\tple-user/"),
			false,
			"",
		},
		{
			"Invalid image registry URL with leading whitespaces",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/ sample-user/",
			},
			fmt.Sprintf(errorMsg, "quay.io/ sample-user/"),
			false,
			"",
		},
		{
			"Valid internal registry URL",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "image-registry.openshift-image-registry.svc:5000/project/app",
			},
			"",
			true,
			"image-registry.openshift-image-registry.svc:5000/project/app",
		},
		{
			"Invalid internal registry URL implicit starts with '/'",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "/project/app",
			},
			fmt.Sprintf(errorMsg, "/project/app"),
			false,
			"",
		},
		{
			"Valid internal registry URL implicit",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "project/app",
			},
			"",
			true,
			"image-registry.openshift-image-registry.svc:5000/project/app",
		},
		{
			"Invalid too many URL components docker",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io/foo/project/app",
			},
			fmt.Sprintf(errorMsg, "docker.io/foo/project/app"),
			false,
			"",
		},
		{
			"Invalid too many URL components internal",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "image-registry.openshift-image-registry.svc:5000/project/app/foo",
			},
			fmt.Sprintf(errorMsg, "image-registry.openshift-image-registry.svc:5000/project/app/foo"),
			false,
			"",
		},
		{
			"Invalid not enough URL components, no slash",
			BootstrapParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io",
			},
			fmt.Sprintf(errorMsg, "docker.io"),
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			isInternalRegistry, imageRepo, error := validateImageRepo(&test.options)
			if diff := cmp.Diff(isInternalRegistry, test.expectedIsInternalRegistry); diff != "" {
				t.Errorf("validateImageRepo() failed:\n%s", diff)
			}
			if diff := cmp.Diff(imageRepo, test.expectedImageRepo); diff != "" {
				t.Errorf("validateImageRepo() failed:\n%s", diff)
			}
			errorString := ""
			if error != nil {
				errorString = error.Error()
			}
			if diff := cmp.Diff(errorString, test.expectedError); diff != "" {
				t.Errorf("validateImageRepo() failed:\n%s", diff)
			}
		})
	}
}
