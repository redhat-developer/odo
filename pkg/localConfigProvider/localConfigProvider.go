package localConfigProvider

// URLKind is an enum to indicate the type of the URL i.e ingress/route
type URLKind string

const (
	DOCKER  URLKind = "docker"
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

// LocalConfigProvider is an interface which all local config providers need to implement
// currently for openshift there is localConfigInfo and for devfile its EnvInfo.
// The reason this interface is declared here instead of config package is because
// some day local config would get deprecated and hence to keep the interfaces in the new package
type LocalConfigProvider interface {
	GetApplication() string
	GetName() string
	GetNamespace() string
	GetDebugPort() int

	GetURL(name string) *LocalURL
	CompleteURL(url *LocalURL) error
	ValidateURL(url LocalURL) error
	CreateURL(url LocalURL) error
	DeleteURL(name string) error
	GetPorts() []string
	ListURLs() []LocalURL

	Exists() bool
}
