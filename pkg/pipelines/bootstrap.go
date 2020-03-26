package pipelines

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/routes"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/tasks"
	"github.com/openshift/odo/pkg/pipelines/triggers"
	"sigs.k8s.io/yaml"
)

const (
	dockerSecretName     = "regcred"
	saName               = "pipeline"
	roleName             = "pipelines-service-role"
	roleBindingName      = "pipelines-service-role-binding"
	devRoleBindingName   = "pipeline-edit-dev"
	stageRoleBindingName = "pipeline-edit-stage"
)

// BootstrapParameters is a struct that provides the optional flags
type BootstrapParameters struct {
	DeploymentPath           string
	GithubHookSecret         string
	GithubToken              string
	GitRepo                  string
	InternalRegistryHostname string
	ImageRepo                string
	Prefix                   string
	DockerConfigJSONFileName string
	SkipChecks               bool
}

// Bootstrap is the main driver for getting OpenShift pipelines for GitOps
// configured with a basic configuration.
func Bootstrap(o *BootstrapParameters) error {

	if !o.SkipChecks {
		installed, err := checkTektonInstall()
		if err != nil {
			return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
		}
		if !installed {
			return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
		}
	}

	isInternalRegistry, imageRepo, err := validateImageRepo(o)
	if err != nil {
		return err
	}

	outputs := make([]interface{}, 0)
	namespaces := namespaceNames(o.Prefix)
	for _, n := range createNamespaces(values(namespaces)) {
		outputs = append(outputs, n)
	}

	if o.GithubToken != "" {
		githubAuth, err := secrets.CreateSealedSecret(meta.NamespacedName(namespaces["cicd"], "github-auth"), o.GithubToken, "token")
		if err != nil {
			return fmt.Errorf("failed to generate Status Tracker Secret: %w", err)
		}

		outputs = append(outputs, githubAuth)
	}

	if o.GithubHookSecret != "" {
		githubSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(namespaces["cicd"], eventlisteners.GitOpsWebhookSecret), o.GithubHookSecret, eventlisteners.WebhookSecretKey)
		if err != nil {
			return fmt.Errorf("failed to generate GitHub Webhook Secret: %w", err)
		}

		outputs = append(outputs, githubSecret)
	}

	// Create Tasks
	tasks := tasks.Generate(namespaces["cicd"])
	for _, task := range tasks {
		outputs = append(outputs, task)
	}

	// Create trigger templates
	templates := triggers.GenerateTemplates(namespaces["cicd"], saName, imageRepo)
	for _, template := range templates {
		outputs = append(outputs, template)
	}

	// Create trigger bindings
	bindings := triggers.GenerateBindings(namespaces["cicd"])
	for _, binding := range bindings {
		outputs = append(outputs, binding)
	}

	// Create Pipelines
	outputs = append(outputs, createPipelines(namespaces, isInternalRegistry, o.DeploymentPath)...)

	// Create Event Listener
	eventListener := eventlisteners.Generate(o.GitRepo, namespaces["cicd"], saName)
	outputs = append(outputs, eventListener)

	// Create route
	route := routes.Generate(namespaces["cicd"])
	outputs = append(outputs, route)

	// Don't add this service account to outputs as this is the default service account created by Pipeline Operator
	sa := createServiceAccount(meta.NamespacedName(namespaces["cicd"], saName))

	manifests, err := createManifestsForImageRepo(sa, isInternalRegistry, imageRepo, o, namespaces)
	if err != nil {
		return err
	}
	outputs = append(outputs, manifests...)

	//  Create Role, Role Bindings, and ClusterRole Bindings
	outputs = append(outputs, createRoleBindings(namespaces, sa)...)

	return marshalOutputs(os.Stdout, outputs)
}

func createRoleBindings(ns map[string]string, sa *corev1.ServiceAccount) []interface{} {
	out := make([]interface{}, 0)

	role := createRole(meta.NamespacedName(ns["cicd"], roleName), rules)
	out = append(out, role)
	out = append(out, createRoleBinding(meta.NamespacedName(ns["cicd"], roleBindingName), sa, role.Kind, role.Name))
	out = append(out, createRoleBinding(meta.NamespacedName(ns["dev"], devRoleBindingName), sa, "ClusterRole", "edit"))
	out = append(out, createRoleBinding(meta.NamespacedName(ns["stage"], stageRoleBindingName), sa, "ClusterRole", "edit"))

	return out
}

// createManifestsForImageRepo creates manifests like namespaces, secret, and role bindng for using image repo
func createManifestsForImageRepo(sa *corev1.ServiceAccount, isInternalRegistry bool, imageRepo string, o *BootstrapParameters, namespaces map[string]string) ([]interface{}, error) {
	out := make([]interface{}, 0)

	if isInternalRegistry {
		// Provide access to service account for using internal registry
		internalRegistryNamespace := strings.Split(imageRepo, "/")[1]

		clientSet, err := getClientSet()
		if err != nil {
			return nil, err
		}
		namespaceExists, err := checkNamespace(clientSet, internalRegistryNamespace)
		if err != nil {
			return nil, err
		}
		if !namespaceExists {
			out = append(out, createNamespace(internalRegistryNamespace))
		}

		// pipelines sa should have access to internal registry
		out = append(out, createRoleBinding(meta.NamespacedName(internalRegistryNamespace, "internal-registry-binding"), sa, "ClusterRole", "edit"))

		// add image puller role to allow pulling app images across namespaces from dev and stage envirnments
		out = append(out, createRoleBindingForSubjects(meta.NamespacedName(internalRegistryNamespace, "image-puller-binding"), "ClusterRole", "system:image-puller",
			[]v1rbac.Subject{v1rbac.Subject{Kind: "ServiceAccount", Name: "default", Namespace: namespaces["dev"]},
				v1rbac.Subject{Kind: "ServiceAccount", Name: "default", Namespace: namespaces["stage"]},
			}))

	} else {
		// Add secret to service account if external registry is used
		dockerSecret, err := createDockerSecret(o.DockerConfigJSONFileName, namespaces["cicd"])
		if err != nil {
			return nil, err
		}
		out = append(out, dockerSecret)
		// add secret and sa to outputs
		out = append(out, addSecretToSA(sa, dockerSecretName))
	}

	return out, nil
}

func createPipelines(ns map[string]string, isInternalRegistry bool, deploymentPath string) []interface{} {
	out := make([]interface{}, 0)
	out = append(out, createDevCIPipeline(meta.NamespacedName(ns["cicd"], "dev-ci-pipeline"), isInternalRegistry))
	out = append(out, createCIPipeline(meta.NamespacedName(ns["cicd"], "stage-ci-pipeline"), ns["stage"]))
	out = append(out, createDevCDPipeline(meta.NamespacedName(ns["cicd"], "dev-cd-pipeline"), deploymentPath, ns["dev"], isInternalRegistry))
	out = append(out, createCDPipeline(meta.NamespacedName(ns["cicd"], "stage-cd-pipeline"), ns["stage"]))
	return out

}

// createDockerSecret creates Docker secret
func createDockerSecret(dockerConfigJSONFileName, ns string) (*ssv1alpha1.SealedSecret, error) {
	if dockerConfigJSONFileName == "" {
		return nil, errors.New("failed to generate path to file: --dockerconfigjson flag is not provided")
	}

	authJSONPath, err := homedir.Expand(dockerConfigJSONFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %w", err)
	}

	f, err := os.Open(authJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docker file '%s' : %w", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := secrets.CreateSealedDockerConfigSecret(meta.NamespacedName(ns, dockerSecretName), f)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil

}

func values(m map[string]string) []string {
	values := []string{}
	for _, v := range m {
		values = append(values, v)

	}
	return values
}

// marshalOutputs marshal outputs to given writer
func marshalOutputs(out io.Writer, outputs []interface{}) error {
	for _, r := range outputs {
		data, err := yaml.Marshal(r)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		_, err = fmt.Fprintf(out, "%s---\n", data)
		if err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}
	return nil
}

// validateImageRepo validates the input image repo.  It determines if it is
// for internal registry and prepend internal registry hostname if neccessary.
func validateImageRepo(o *BootstrapParameters) (bool, string, error) {
	components := strings.Split(o.ImageRepo, "/")

	// repo url has minimum of 2 components
	if len(components) < 2 {
		return false, "", imageRepoValidationError(o.ImageRepo)
	}

	for _, v := range components {
		// check for empty components
		if strings.TrimSpace(v) == "" {
			return false, "", imageRepoValidationError(o.ImageRepo)
		}
		// check for white spaces
		if len(v) > len(strings.TrimSpace(v)) {
			return false, "", imageRepoValidationError(o.ImageRepo)
		}
	}

	if len(components) == 2 {
		if components[0] == "docker.io" || components[0] == "quay.io" {
			// we recognize docker.io and quay.io.  It is missing one component
			return false, "", imageRepoValidationError(o.ImageRepo)
		}
		// We have format like <project>/<app> which is an internal registry.
		// We prepend the internal registry hostname.
		return true, o.InternalRegistryHostname + "/" + o.ImageRepo, nil
	}

	// Check the first component to see if it is an internal registry
	if len(components) == 3 {
		return components[0] == o.InternalRegistryHostname, o.ImageRepo, nil
	}

	// > 3 components.  invalid repo
	return false, "", imageRepoValidationError(o.ImageRepo)
}

func imageRepoValidationError(imageRepo string) error {
	return fmt.Errorf("failed to parse image repo:%s, expected image repository in the form <registry>/<username>/<repository> or <project>/<app> for internal registry", imageRepo)
}
