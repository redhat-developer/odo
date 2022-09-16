package api

type BindableService struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Group     string `json:"group,omitempty"`
	Service   string `json:"service,omitempty"`
}
