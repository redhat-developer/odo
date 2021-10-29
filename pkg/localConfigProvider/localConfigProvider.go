package localConfigProvider

// URLKind is an enum to indicate the type of the URL i.e ingress/route
type URLKind string

const (
	INGRESS URLKind = "ingress"
	ROUTE   URLKind = "route"
)

// LocalURL holds URL related information
type LocalURL struct {
	// Name of the URL
	Name string `yaml:"Name,omitempty" json:"name,omitempty"`
	// Port number for the url of the component, required in case of components which expose more than one service port
	Port int `yaml:"Port,omitempty" json:"port,omitempty"`
	// Indicates if the URL should be a secure https one
	Secure bool `yaml:"Secure,omitempty" json:"secure,omitempty"`
	// Cluster host
	Host string `yaml:"Host,omitempty" json:"host,omitempty"`
	// TLS secret name to create ingress to provide a secure URL
	TLSSecret string `yaml:"TLSSecret,omitempty" json:"tlsSecret,omitempty"`
	// Exposed port number for docker container, required for local scenarios
	ExposedPort int `yaml:"ExposedPort,omitempty" json:"exposedPort,omitempty"`
	// Kind is the kind of the URL
	Kind URLKind `yaml:"Kind,omitempty" json:"kind,omitempty"`
	// Path is the path of the URL
	Path string `yaml:"-" json:"-"`
	// Container is the container of the URL
	Container string `yaml:"-" json:"-"`
	// Protocol is the protocol of the URL
	Protocol string `yaml:"-" json:"-"`
}

// LocalStorage holds storage related information
type LocalStorage struct {
	// Name of the storage
	Name string `yaml:"Name,omitempty"`
	// Size of the storage
	Size string `yaml:"Size,omitempty"`
	// Path of the storage to which it will be mounted on the container
	Path string `yaml:"Path,omitempty"`
	// Container is the container name on which this storage is mounted
	Container string `yaml:"-" json:"-"`
}

// LocalContainer holds the container related information
type LocalContainer struct {
	Name string `yaml:"Name" json:"Name"`
}

// LocalConfigProvider is an interface which all local config providers need to implement
// currently implemented by EnvInfo
type LocalConfigProvider interface {
	GetApplication() string
	GetName() string
	GetNamespace() string
	GetDebugPort() int
	GetContainers() ([]LocalContainer, error)

	GetURL(name string) (*LocalURL, error)
	CompleteURL(url *LocalURL) error
	ValidateURL(url LocalURL) error
	CreateURL(url LocalURL) error
	DeleteURL(name string) error
	GetContainerPorts(container string) ([]string, error)
	GetComponentPorts() ([]string, error)
	ListURLs() ([]LocalURL, error)

	GetStorage(name string) (*LocalStorage, error)
	CompleteStorage(storage *LocalStorage)
	ValidateStorage(storage LocalStorage) error
	CreateStorage(storage LocalStorage) error
	DeleteStorage(name string) error
	ListStorage() ([]LocalStorage, error)
	GetStorageMountPath(storageName string) (string, error)

	Exists() bool
}
