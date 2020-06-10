package argocd

import (
	"path/filepath"
	"sort"

	// This is a hack because ArgoCD doesn't support a compatible (code-wise)
	// version of k8s in common with odo.
	argoappv1 "github.com/openshift/odo/pkg/pipelines/argocd/v1alpha1"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"
)

var (
	applicationTypeMeta = meta.TypeMeta(
		"Application",
		"argoproj.io/v1alpha1",
	)

	syncPolicy = &argoappv1.SyncPolicy{
		Automated: &argoappv1.SyncPolicyAutomated{
			Prune:    true,
			SelfHeal: true,
		},
	}
)

const (
	defaultServer   = "https://kubernetes.default.svc"
	defaultProject  = "default"
	ArgoCDNamespace = "argocd"
)

func Build(argoNS, repoURL string, m *config.Manifest) (res.Resources, error) {
	// Without a RepositoryURL we can't do anything.
	if repoURL == "" {
		return res.Resources{}, nil
	}
	argoCDConfig := m.GetArgoCDConfig()
	if argoCDConfig == nil {
		return res.Resources{}, nil
	}

	files := make(res.Resources)
	eb := &argocdBuilder{repoURL: repoURL, files: files, argoCDConfig: argoCDConfig, argoNS: argoNS}
	err := m.Walk(eb)
	if err != nil {
		return nil, err
	}

	err = argoCDConfigResources(eb.argoCDConfig, eb.files)
	if err != nil {
		return nil, err
	}
	return eb.files, err
}

type argocdBuilder struct {
	repoURL      string
	argoCDConfig *config.ArgoCDConfig
	files        res.Resources
	argoNS       string
}

func (b *argocdBuilder) Application(env *config.Environment, app *config.Application) error {
	basePath := filepath.Join(config.PathForArgoCD(), "config")
	argoFiles := res.Resources{}
	filename := filepath.Join(basePath, env.Name+"-"+app.Name+"-app.yaml")
	argoFiles[filename] = makeApplication(env.Name+"-"+app.Name, b.argoNS, defaultProject, env.Name, defaultServer, makeSource(env, app, b.repoURL))
	b.files = res.Merge(argoFiles, b.files)
	return nil
}

func argoCDConfigResources(argoCDConfig *config.ArgoCDConfig, files res.Resources) error {
	if argoCDConfig.Namespace == "" {
		return nil
	}
	basePath := filepath.Join(config.PathForArgoCD(), "config")
	filename := filepath.Join(basePath, "kustomization.yaml")
	resourceNames := []string{}
	for k := range files {
		resourceNames = append(resourceNames, filepath.Base(k))
	}
	sort.Strings(resourceNames)
	files[filename] = &res.Kustomization{Resources: resourceNames}
	return nil
}

func makeSource(env *config.Environment, app *config.Application, repoURL string) argoappv1.ApplicationSource {
	if app.ConfigRepo == nil {
		return argoappv1.ApplicationSource{
			RepoURL: repoURL,
			Path:    filepath.Join(config.PathForApplication(env, app), "base"),
		}
	}
	return argoappv1.ApplicationSource{
		RepoURL:        app.ConfigRepo.URL,
		Path:           app.ConfigRepo.Path,
		TargetRevision: app.ConfigRepo.TargetRevision,
	}
}

func makeApplication(appName, argoNS, project, ns, server string, source argoappv1.ApplicationSource) *argoappv1.Application {
	return &argoappv1.Application{
		TypeMeta:   applicationTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(argoNS, appName)),
		Spec: argoappv1.ApplicationSpec{
			Project: project,
			Destination: argoappv1.ApplicationDestination{
				Namespace: ns,
				Server:    server,
			},
			Source:     source,
			SyncPolicy: syncPolicy,
		},
	}
}
