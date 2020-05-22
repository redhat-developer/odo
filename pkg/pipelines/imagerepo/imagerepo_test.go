package imagerepo

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	v1rbac "k8s.io/api/rbac/v1"
)

func TestCreateInternalRegistryRoleBinding(t *testing.T) {

	cicd := &config.Environment{
		Name: "test-cicd",
	}
	sa := roles.CreateServiceAccount(meta.NamespacedName("test-cicd", "pipeline"))
	gotFilename, got := createInternalRegistryRoleBinding(cicd, "new-proj", sa)

	want := res.Resources{"environments/test-cicd/base/pipelines/02-rolebindings/internal-registry-new-proj-binding.yaml": &v1rbac.RoleBinding{
		TypeMeta:   meta.TypeMeta("RoleBinding", "rbac.authorization.k8s.io/v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("new-proj", "internal-registry-new-proj-binding")),
		Subjects:   []v1rbac.Subject{{Kind: sa.Kind, Name: sa.Name, Namespace: sa.Namespace}},
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "edit",
		},
	}}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("resources do not match:\n%s", diff)
	}

	if diff := cmp.Diff(gotFilename, "02-rolebindings/internal-registry-new-proj-binding.yaml"); diff != "" {
		t.Errorf("filename do not match:\n%s", diff)
	}
}

func TestValidateImageRepo(t *testing.T) {

	errorMsg := "failed to parse image repo:%s, expected image repository in the form <registry>/<username>/<repository> or <project>/<app> for internal registry"

	tests := []struct {
		description                string
		internalRegistryHostname   string
		imageRepo                  string
		expectedError              string
		expectedIsInternalRegistry bool
		expectedImageRepo          string
	}{
		{
			"Valid image regsitry URL",
			"image-registry.openshift-image-registry.svc:5000",
			"quay.io/sample-user/sample-repo",
			"",
			false,
			"quay.io/sample-user/sample-repo",
		},
		{
			"Valid image regsitry URL random registry",
			"image-registry.openshift-image-registry.svc:5000",
			"random.io/sample-user/sample-repo",
			"",
			false,
			"random.io/sample-user/sample-repo",
		},
		{
			"Valid image regsitry URL docker.io",
			"image-registry.openshift-image-registry.svc:5000",
			"docker.io/sample-user/sample-repo",
			"",
			false,
			"docker.io/sample-user/sample-repo",
		},
		{
			"Invalid image registry URL with missing repo name",
			"image-registry.openshift-image-registry.svc:5000",
			"quay.io/sample-user",
			fmt.Sprintf(errorMsg, "quay.io/sample-user"),
			false,
			"",
		},
		{
			"Invalid image registry URL with missing repo name docker.io",
			"image-registry.openshift-image-registry.svc:5000",
			"docker.io/sample-user",
			fmt.Sprintf(errorMsg, "docker.io/sample-user"),
			false,
			"",
		},
		{
			"Invalid image registry URL with whitespaces",
			"image-registry.openshift-image-registry.svc:5000",
			"quay.io/sample-user/ ",
			fmt.Sprintf(errorMsg, "quay.io/sample-user/ "),
			false,
			"",
		},
		{
			"Invalid image registry URL with whitespaces in between",
			"image-registry.openshift-image-registry.svc:5000",
			"quay.io/sam\tple-user/",
			fmt.Sprintf(errorMsg, "quay.io/sam\tple-user/"),
			false,
			"",
		},
		{
			"Invalid image registry URL with leading whitespaces",
			"image-registry.openshift-image-registry.svc:5000",
			"quay.io/ sample-user/",
			fmt.Sprintf(errorMsg, "quay.io/ sample-user/"),
			false,
			"",
		},
		{
			"Valid internal registry URL",
			"image-registry.openshift-image-registry.svc:5000",
			"image-registry.openshift-image-registry.svc:5000/project/app",
			"",
			true,
			"image-registry.openshift-image-registry.svc:5000/project/app",
		},
		{
			"Invalid internal registry URL implicit starts with '/'",
			"image-registry.openshift-image-registry.svc:5000",
			"/project/app",
			fmt.Sprintf(errorMsg, "/project/app"),
			false,
			"",
		},
		{
			"Valid internal registry URL implicit",
			"image-registry.openshift-image-registry.svc:5000",
			"project/app",
			"",
			true,
			"image-registry.openshift-image-registry.svc:5000/project/app",
		},
		{
			"Invalid too many URL components docker",
			"image-registry.openshift-image-registry.svc:5000",
			"docker.io/foo/project/app",
			fmt.Sprintf(errorMsg, "docker.io/foo/project/app"),
			false,
			"",
		},
		{
			"Invalid too many URL components internal",
			"image-registry.openshift-image-registry.svc:5000",
			"image-registry.openshift-image-registry.svc:5000/project/app/foo",
			fmt.Sprintf(errorMsg, "image-registry.openshift-image-registry.svc:5000/project/app/foo"),
			false,
			"",
		},
		{
			"Invalid not enough URL components, no slash",
			"image-registry.openshift-image-registry.svc:5000",
			"docker.io",
			fmt.Sprintf(errorMsg, "docker.io"),
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			isInternalRegistry, imageRepo, error := ValidateImageRepo(test.imageRepo,
				test.internalRegistryHostname,
			)
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
