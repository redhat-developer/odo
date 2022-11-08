package api

type Project struct {
	Name   string `json:"name,omitempty"`
	Active bool   `json:"active"`
}
