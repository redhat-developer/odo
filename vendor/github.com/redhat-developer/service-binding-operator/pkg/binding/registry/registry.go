package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

type Registry interface {
	GetAnnotations(serviceGVK schema.GroupVersionKind) (map[string]string, bool)
	Register(serviceGVK schema.GroupVersionKind, annotations map[string]string)
}

type impl struct {
	annotationMap map[schema.GroupVersionKind]map[string]string
}

var ServiceAnnotations = New()

func New() Registry {
	return &impl{
		annotationMap: map[schema.GroupVersionKind]map[string]string{
			schema.GroupVersionKind{Group: "redis.redis.opstreelabs.in", Version: "v1beta1", Kind: "Redis"}: {
				"service.binding/type":     "redis",
				"service.binding/host":     "path={.metadata.name}",
				"service.binding/password": "path={.spec.kubernetesConfig.redisSecret.name},objectType=Secret,sourceKey=password,optional=true",
			},
			schema.GroupVersionKind{Group: "postgres-operator.crunchydata.com", Version: "v1beta1", Kind: "PostgresCluster"}: {
				"service.binding/type":     "postgresql",
				"service.binding/provider": "crunchydata",
				"service.binding":          "path={.metadata.name}-pguser-{.metadata.name},objectType=Secret",
				"service.binding/database": "path={.metadata.name}-pguser-{.metadata.name},objectType=Secret,sourceKey=dbname",
				"service.binding/username": "path={.metadata.name}-pguser-{.metadata.name},objectType=Secret,sourceKey=user",
				"service.binding/cert":     "path={.metadata.name}-cluster-cert,objectType=Secret",
			},
			schema.GroupVersionKind{Group: "pxc.percona.com", Version: "v1-8-0", Kind: "PerconaXtraDBCluster"}: {
				"service.binding/type":     "mysql",
				"service.binding/provider": "percona",
				"service.binding/database": "mysql",
				"service.binding":          "path={.spec.secretsName},objectType=Secret",
				"service.binding/host":     "path={.status.host}",
				"service.binding/port":     "3306",
				"service.binding/username": "root",
				"service.binding/password": "path={.spec.secretsName},objectType=Secret,sourceKey=root",
			},
			schema.GroupVersionKind{Group: "pxc.percona.com", Version: "v1-9-0", Kind: "PerconaXtraDBCluster"}: {
				"service.binding/type":     "mysql",
				"service.binding/provider": "percona",
				"service.binding/database": "mysql",
				"service.binding":          "path={.spec.secretsName},objectType=Secret",
				"service.binding/host":     "path={.status.host}",
				"service.binding/port":     "3306",
				"service.binding/username": "root",
				"service.binding/password": "path={.spec.secretsName},objectType=Secret,sourceKey=root",
			},
			schema.GroupVersionKind{Group: "pxc.percona.com", Version: "v1-10-0", Kind: "PerconaXtraDBCluster"}: {
				"service.binding/type":     "mysql",
				"service.binding/provider": "percona",
				"service.binding/database": "mysql",
				"service.binding":          "path={.spec.secretsName},objectType=Secret",
				"service.binding/host":     "path={.status.host}",
				"service.binding/port":     "3306",
				"service.binding/username": "root",
				"service.binding/password": "path={.spec.secretsName},objectType=Secret,sourceKey=root",
			},
			schema.GroupVersionKind{Group: "psmdb.percona.com", Version: "v1-9-0", Kind: "PerconaServerMongoDB"}: {
				"service.binding/type":     "mongodb",
				"service.binding/provider": "percona",
				"service.binding":          "path={.spec.secrets.users},objectType=Secret",
				"service.binding/username": "path={.spec.secrets.users},objectType=Secret,sourceKey=MONGODB_USER_ADMIN_USER",
				"service.binding/password": "path={.spec.secrets.users},objectType=Secret,sourceKey=MONGODB_USER_ADMIN_PASSWORD",
				"service.binding/host":     "path={.status.host}",
			},
			schema.GroupVersionKind{Group: "psmdb.percona.com", Version: "v1-10-0", Kind: "PerconaServerMongoDB"}: {
				"service.binding/type":     "mongodb",
				"service.binding/provider": "percona",
				"service.binding":          "path={.spec.secrets.users},objectType=Secret",
				"service.binding/username": "path={.spec.secrets.users},objectType=Secret,sourceKey=MONGODB_USER_ADMIN_USER",
				"service.binding/password": "path={.spec.secrets.users},objectType=Secret,sourceKey=MONGODB_USER_ADMIN_PASSWORD",
				"service.binding/host":     "path={.status.host}",
			},
			schema.GroupVersionKind{Group: "psmdb.percona.com", Version: "v1", Kind: "PerconaServerMongoDB"}: {
				"service.binding/type":     "mongodb",
				"service.binding/provider": "percona",
				"service.binding":          "path={.spec.secrets.users},objectType=Secret",
				"service.binding/username": "path={.spec.secrets.users},objectType=Secret,sourceKey=MONGODB_USER_ADMIN_USER",
				"service.binding/password": "path={.spec.secrets.users},objectType=Secret,sourceKey=MONGODB_USER_ADMIN_PASSWORD",
				"service.binding/host":     "path={.status.host}",
			},
			schema.GroupVersionKind{Group: "postgresql.k8s.enterprisedb.io", Version: "v1", Kind: "Cluster"}: {
				"service.binding/type":     "postgresql",
				"service.binding/host":     "path={.status.writeService}",
				"service.binding/provider": "enterprisedb",
				"service.binding":          "path={.metadata.name}-{.spec.bootstrap.initdb.owner},objectType=Secret",
				"service.binding/database": "path={.spec.bootstrap.initdb.database}",
			},
			schema.GroupVersionKind{Group: "rabbitmq.com", Version: "v1beta1", Kind: "RabbitmqCluster"}: {
				"servicebinding.io/provisioned-service": "true",
			},
		},
	}
}

func (i *impl) GetAnnotations(serviceGVK schema.GroupVersionKind) (map[string]string, bool) {
	result, found := i.annotationMap[serviceGVK]
	return result, found
}

func (i *impl) Register(serviceGVK schema.GroupVersionKind, annotations map[string]string) {
	i.annotationMap[serviceGVK] = annotations
}
