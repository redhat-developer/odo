package servicebindingrequest

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/envvars"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
}

// createServiceIndexPath returns a string slice with fields representing a path to a resource in the
// environment variable context. This function cleans fields that might contain invalid characters to
// be used in Go template; for example, a Group might contain the "." character, which makes it
// harder to refer using Go template direct accessors and is substituted by an underbar "_".
func createServiceIndexPath(name string, gvk schema.GroupVersionKind) []string {
	return []string{
		gvk.Version,
		strings.ReplaceAll(gvk.Group, ".", "_"),
		gvk.Kind,
		strings.ReplaceAll(name, "-", "_"),
	}

}

func buildServiceEnvVars(svcCtx *ServiceContext, globalEnvVarPrefix string) (map[string]string, error) {
	prefixes := []string{}
	if len(globalEnvVarPrefix) > 0 {
		prefixes = append(prefixes, globalEnvVarPrefix)
	}
	if svcCtx.EnvVarPrefix != nil && len(*svcCtx.EnvVarPrefix) > 0 {
		prefixes = append(prefixes, *svcCtx.EnvVarPrefix)
	}
	if svcCtx.EnvVarPrefix == nil {
		prefixes = append(prefixes, svcCtx.Service.GroupVersionKind().Kind)
	}

	return envvars.Build(svcCtx.EnvVars, prefixes...)
}

func (r *Retriever) processServiceContext(
	svcCtx *ServiceContext,
	customEnvVarCtx map[string]interface{},
	globalEnvVarPrefix string,
) (map[string][]byte, []string, error) {
	svcEnvVars, err := buildServiceEnvVars(svcCtx, globalEnvVarPrefix)
	if err != nil {
		return nil, nil, err
	}

	// contribute the entire resource to the context shared with the custom env parser
	gvk := svcCtx.Service.GetObjectKind().GroupVersionKind()

	// add an entry in the custom environment variable context, allowing the user to use the
	// following expression:
	//
	// `{{ index "v1alpha1" "postgresql.baiju.dev" "Database", "db-testing", "status", "connectionUrl" }}`
	err = unstructured.SetNestedField(
		customEnvVarCtx, svcCtx.Service.Object, gvk.Version, gvk.Group, gvk.Kind,
		svcCtx.Service.GetName())
	if err != nil {
		return nil, nil, err
	}

	// add an entry in the custom environment variable context with modified key names (group
	// names have the "." separator changed to underbar and "-" in the resource name is changed
	// to underbar "_" as well).
	//
	// `{{ .v1alpha1.postgresql_baiju_dev.Database.db_testing.status.connectionUrl }}`
	err = unstructured.SetNestedField(
		customEnvVarCtx,
		svcCtx.Service.Object,
		createServiceIndexPath(svcCtx.Service.GetName(), svcCtx.Service.GroupVersionKind())...,
	)
	if err != nil {
		return nil, nil, err
	}

	envVars := make(map[string][]byte, len(svcEnvVars))
	for k, v := range svcEnvVars {
		envVars[k] = []byte(v)
	}

	var volumeKeys []string
	volumeKeys = append(volumeKeys, svcCtx.VolumeKeys...)

	return envVars, volumeKeys, nil
}

// ProcessServiceContexts returns environment variables and volume keys from a ServiceContext slice.
func (r *Retriever) ProcessServiceContexts(
	globalEnvVarPrefix string,
	svcCtxs ServiceContextList,
	envVarTemplates []corev1.EnvVar,
) (map[string][]byte, []string, error) {
	customEnvVarCtx := make(map[string]interface{})
	volumeKeys := make([]string, 0)
	envVars := make(map[string][]byte)

	for _, svcCtx := range svcCtxs {
		s, v, err := r.processServiceContext(svcCtx, customEnvVarCtx, globalEnvVarPrefix)
		if err != nil {
			return nil, nil, err
		}
		for k, v := range s {
			envVars[k] = []byte(v)
		}
		volumeKeys = append(volumeKeys, v...)
	}

	envParser := NewCustomEnvParser(envVarTemplates, customEnvVarCtx)
	customEnvVars, err := envParser.Parse()
	if err != nil {
		r.logger.Error(
			err, "Creating envVars", "Templates", envVarTemplates, "TemplateContext", customEnvVarCtx)
		return nil, nil, err
	}

	for k, v := range customEnvVars {
		prefix := []string{}
		if len(globalEnvVarPrefix) > 0 {
			prefix = append(prefix, globalEnvVarPrefix)
		}
		prefix = append(prefix, k)
		k = strings.Join(prefix, "_")
		envVars[k] = []byte(v.(string))
	}

	return envVars, volumeKeys, nil
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(
	client dynamic.Interface,
) *Retriever {
	return &Retriever{
		logger: log.NewLog("retriever"),
		client: client,
	}
}
