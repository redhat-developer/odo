package v1alpha2

import "github.com/devfile/api/v2/pkg/attributes"

// EndpointProtocol defines the application and transport protocols of the traffic that will go through this endpoint.
// Only one of the following protocols may be specified: http, ws, tcp, udp.
// +kubebuilder:validation:Enum=http;https;ws;wss;tcp;udp
type EndpointProtocol string

const (
	// Endpoint will have `http` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `https` when the `secure` field is set to `true`
	HTTPEndpointProtocol EndpointProtocol = "http"
	// Endpoint will have `https` traffic, typically on a TCP connection
	HTTPSEndpointProtocol EndpointProtocol = "https"
	// Endpoint will have `ws` traffic, typically on a TCP connection
	// It will be automaticaly promoted to `wss` when the `secure` field is set to `true`
	WSEndpointProtocol EndpointProtocol = "ws"
	// Endpoint will have `wss` traffic, typically on a TCP connection
	WSSEndpointProtocol EndpointProtocol = "wss"
	// Endpoint will have traffic on a TCP connection,
	// without specifying an application protocol
	TCPEndpointProtocol EndpointProtocol = "tcp"
	// Endpoint will have traffic on an UDP connection,
	// without specifying an application protocol
	UDPEndpointProtocol EndpointProtocol = "udp"
)

// EndpointExposure describes the way an endpoint is exposed on the network.
// Only one of the following exposures may be specified: public, internal, none.
// +kubebuilder:validation:Enum=public;internal;none
type EndpointExposure string

const (
	// Endpoint will be exposed on the public network, typically through
	// a K8S ingress or an OpenShift route
	PublicEndpointExposure EndpointExposure = "public"
	// Endpoint will be exposed internally outside of the main devworkspace POD,
	// typically by K8S services, to be consumed by other elements running
	// on the same cloud internal network.
	InternalEndpointExposure EndpointExposure = "internal"
	// Endpoint will not be exposed and will only be accessible
	// inside the main devworkspace POD, on a local address.
	NoneEndpointExposure EndpointExposure = "none"
)

// +devfile:getter:generate
type Endpoint struct {
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// The port number should be unique.
	TargetPort int `json:"targetPort"`

	// Describes how the endpoint should be exposed on the network.
	//
	// - `public` means that the endpoint will be exposed on the public network, typically through
	// a K8S ingress or an OpenShift route.
	//
	// - `internal` means that the endpoint will be exposed internally outside of the main devworkspace POD,
	// typically by K8S services, to be consumed by other elements running
	// on the same cloud internal network.
	//
	// - `none` means that the endpoint will not be exposed and will only be accessible
	// inside the main devworkspace POD, on a local address.
	//
	// Default value is `public`
	// +optional
	// +kubebuilder:default=public
	Exposure EndpointExposure `json:"exposure,omitempty"`

	// Describes the application and transport protocols of the traffic that will go through this endpoint.
	//
	// - `http`: Endpoint will have `http` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `https` when the `secure` field is set to `true`.
	//
	// - `https`: Endpoint will have `https` traffic, typically on a TCP connection.
	//
	// - `ws`: Endpoint will have `ws` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `wss` when the `secure` field is set to `true`.
	//
	// - `wss`: Endpoint will have `wss` traffic, typically on a TCP connection.
	//
	// - `tcp`: Endpoint will have traffic on a TCP connection, without specifying an application protocol.
	//
	// - `udp`: Endpoint will have traffic on an UDP connection, without specifying an application protocol.
	//
	// Default value is `http`
	// +optional
	// +kubebuilder:default=http
	Protocol EndpointProtocol `json:"protocol,omitempty"`

	// Describes whether the endpoint should be secured and protected by some
	// authentication process. This requires a protocol of `https` or `wss`.
	// +optional
	// +devfile:default:value=false
	Secure *bool `json:"secure,omitempty"`

	// Path of the endpoint URL
	// +optional
	Path string `json:"path,omitempty"`

	// Map of implementation-dependant string-based free-form attributes.
	//
	// Examples of Che-specific attributes:
	//
	// - cookiesAuthEnabled: "true" / "false",
	//
	// - type: "terminal" / "ide" / "ide-dev",
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes attributes.Attributes `json:"attributes,omitempty"`

	// +optional
	// Annotations to be added to Kubernetes Ingress or Openshift Route
	Annotations map[string]string `json:"annotation,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}
