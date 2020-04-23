package argocd

import (
	"path/filepath"
	"sort"

	// This is a hack because ArgoCD doesn't support a compatible (code-wise)
	// version of k8s in common with odo.
	argoappv1 "github.com/openshift/odo/pkg/manifest/argocd/v1alpha1"

	"github.com/openshift/odo/pkg/manifest/config"
	"github.com/openshift/odo/pkg/manifest/meta"
	res "github.com/openshift/odo/pkg/manifest/resources"
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
	gitOpsApp       = "gitops-app"
)

func Build(argoNS, repoURL string, m *config.Manifest) (res.Resources, error) {
	// Without a RepositoryURL we can't do anything.
	if repoURL == "" {
		return res.Resources{}, nil
	}
	argoEnv, err := m.GetArgoCDEnvironment()
	// If there's no ArgoCD environment, then we don't need to do anything.
	if err != nil {
		return res.Resources{}, nil
	}

	files := make(res.Resources)
	eb := &argocdBuilder{repoURL: repoURL, files: files, argoEnv: argoEnv, argoNS: argoNS}
	err = m.Walk(eb)
	return eb.files, err
}

type argocdBuilder struct {
	repoURL string
	argoEnv *config.Environment
	files   res.Resources
	argoNS  string
}

func (b *argocdBuilder) Application(env *config.Environment, app *config.Application) error {
	basePath := filepath.Join(config.PathForEnvironment(b.argoEnv), "config")
	argoFiles := res.Resources{}
	filename := filepath.Join(basePath, env.Name+"-"+app.Name+"-app.yaml")
	argoFiles[filename] = makeApplication(env.Name+"-"+app.Name, b.argoNS, defaultProject, env.Name, defaultServer, makeSource(env, app, b.repoURL))
	b.files = res.Merge(argoFiles, b.files)
	return nil
}

func (b *argocdBuilder) Environment(env *config.Environment) error {
	if env.IsCICD {
		basePath := filepath.Join(config.PathForEnvironment(b.argoEnv), "config")
		filename := filepath.Join(basePath, gitOpsApp+".yaml")
		sourcePath := filepath.Join(config.PathForEnvironment(env), "base")
		b.files[filename] = makeApplication(gitOpsApp, b.argoNS, defaultProject, env.Name, defaultServer, argoappv1.ApplicationSource{RepoURL: b.repoURL, Path: sourcePath})
	}
	if !env.IsArgoCD {
		return nil
	}
	basePath := filepath.Join(config.PathForEnvironment(b.argoEnv), "config")
	filename := filepath.Join(basePath, "kustomization.yaml")
	resourceNames := []string{}
	for k, _ := range b.files {
		resourceNames = append(resourceNames, filepath.Base(k))
	}
	sort.Strings(resourceNames)
	b.files[filename] = &res.Kustomization{Resources: resourceNames}
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
