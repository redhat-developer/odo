package webhook

import (
	"errors"
	"fmt"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/git"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/openshift/odo/pkg/pipelines/routes"
	"github.com/openshift/odo/pkg/pipelines/secrets"
)

type webhookInfo struct {
	clusterResource *resources
	repository      *git.Repository
	gitRepoURL      string
	cicdNamepace    string
	listenerURL     string
	accessToken     string
	serviceName     *QualifiedServiceName
	isCICD          bool
}

// QualifiedServiceName represents three part name of a service (Environment, Application, and Service)
type QualifiedServiceName struct {
	EnvironmentName string
	ApplicationName string
	ServiceName     string
}

// Create creates a new webhook on the target Git Repository
// It returns the ID of created webhook.
func Create(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) (string, error) {

	webhook, err := newWebhookInfo(accessToken, pipelinesFile, serviceName, isCICD)
	if err != nil {
		return "", err
	}

	exists, err := webhook.exists()
	if err != nil {
		return "", err
	}

	if exists {
		return "", errors.New("webhook already exists")
	}

	return webhook.create()
}

// Delete deletes webhooks on the target Git Repository that match the listener address
// It returns the IDs of deleted webhooks.
func Delete(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) ([]string, error) {

	webhook, err := newWebhookInfo(accessToken, pipelinesFile, serviceName, isCICD)
	if err != nil {
		return nil, err
	}

	ids, err := webhook.list()
	if err != nil {
		return nil, err
	}

	return webhook.delete(ids)
}

// List returns an array of webhook IDs for the target Git repository/listeners
func List(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) ([]string, error) {

	webhook, err := newWebhookInfo(accessToken, pipelinesFile, serviceName, isCICD)
	if err != nil {
		return nil, err
	}

	return webhook.list()
}

func newWebhookInfo(accessToken, pipelinesFile string, serviceName *QualifiedServiceName, isCICD bool) (*webhookInfo, error) {

	manifest, err := config.ParseFile(ioutils.NewFilesystem(), pipelinesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pipelines: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, err
	}

	gitRepoURL := getRepoURL(manifest, isCICD, serviceName)
	if gitRepoURL == "" {
		return nil, errors.New("failed to find Git repostory URL in manifest")
	}

	cicdEnv, err := manifest.GetCICDEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to get CICD environment: %w", err)
	}
	cicdNamepace := cicdEnv.Name

	clusterResources, err := newResources()
	if err != nil {
		return nil, err
	}

	repository, err := git.NewRepository(gitRepoURL, accessToken)
	if err != nil {
		return nil, err
	}

	listenerURL, err := getListenerURL(clusterResources, cicdNamepace)
	if err != nil {
		return nil, fmt.Errorf("failed to get event listener URL: %w", err)
	}

	return &webhookInfo{clusterResources, repository, gitRepoURL, cicdNamepace, listenerURL, accessToken, serviceName, isCICD}, nil
}

func (w *webhookInfo) exists() (bool, error) {

	ids, err := w.repository.ListWebhooks(w.listenerURL)
	if err != nil {
		return false, err
	}

	return len(ids) > 0, nil
}

func (w *webhookInfo) list() ([]string, error) {

	return w.repository.ListWebhooks(w.listenerURL)
}

func (w *webhookInfo) delete(ids []string) ([]string, error) {

	return w.repository.DeleteWebhooks(ids)
}

func (w *webhookInfo) create() (string, error) {

	secret, err := getWebhookSecret(w.clusterResource, w.cicdNamepace, w.isCICD, w.serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to get webhook secret: %w", err)
	}

	return w.repository.CreateWebhook(w.listenerURL, secret)
}

// Get Git repository URL whether it is CICD configuration or service source repository
// Return "" if not found
func getRepoURL(manifest *config.Manifest, isCICD bool, serviceName *QualifiedServiceName) string {

	if isCICD {
		return manifest.GitOpsURL
	}

	return getSourceRepoURL(manifest, serviceName)
}

// Get service source repository URL.  Return "" if not found
func getSourceRepoURL(manifest *config.Manifest, service *QualifiedServiceName) string {

	for _, env := range manifest.Environments {
		if env.Name == service.EnvironmentName {
			for _, app := range env.Apps {
				if app.Name == service.ApplicationName {
					for _, svc := range app.Services {
						if svc.Name == service.ServiceName {
							return svc.SourceURL
						}
					}
					break
				}
			}
			break
		}
	}

	return ""
}

func getListenerURL(r *resources, cicdNamespace string) (string, error) {

	hasTLS, host, err := r.getListenerAddress(cicdNamespace, routes.GitOpsWebhookEventListenerRouteName)
	if err != nil {
		return "", err
	}

	return buildURL(host, hasTLS), nil
}

func buildURL(host string, hasTLS bool) string {

	scheme := "http"
	if hasTLS {
		scheme = scheme + "s"
	}

	return scheme + "://" + host
}

func getWebhookSecret(r *resources, namespace string, isCICD bool, service *QualifiedServiceName) (string, error) {

	var secretName string
	if isCICD {
		secretName = eventlisteners.GitOpsWebhookSecret
	} else {
		// currently, use the app name to create webhook secret name.
		// also currently, service webhook secret are in CICI namespace
		secretName = secrets.MakeServiceWebhookSecretName(service.ApplicationName)
	}

	return r.getWebhookSecret(namespace, secretName, eventlisteners.WebhookSecretKey)
}
