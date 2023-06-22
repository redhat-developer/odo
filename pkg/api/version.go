package api

/*
{
  "version": "v3.11.0",
  "gitCommit": "0acf1a5af",
  "cluster": {
    "serverURL": "https://kubernetes.docker.internal:6443",
    "kubernetes": {
      "version": "v1.25.9"
    },
    "openshift": {
      "version": "4.13.0"
    }
  },
  "podman": {
    "client": {
      "version": "4.5.1"
    }
  }
}
*/

type OdoVersion struct {
	Version   string       `json:"version"`
	GitCommit string       `json:"gitCommit"`
	Cluster   *ClusterInfo `json:"cluster,omitempty"`
	Podman    *PodmanInfo  `json:"podman,omitempty"`
}

type ClusterInfo struct {
	ServerURL  string             `json:"serverURL,omitempty"`
	Kubernetes *ClusterClientInfo `json:"kubernetes,omitempty"`
	OpenShift  *ClusterClientInfo `json:"openshift,omitempty"`
}

type ClusterClientInfo struct {
	Version string `json:"version,omitempty"`
}

type PodmanInfo struct {
	Client *PodmanClientInfo `json:"client,omitempty"`
}

type PodmanClientInfo struct {
	Version string `json:"version,omitempty"`
}
