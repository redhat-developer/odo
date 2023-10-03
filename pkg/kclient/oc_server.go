package kclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	dfutil "github.com/devfile/library/v2/pkg/util"
	configv1 "github.com/openshift/api/config/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/klog"
)

// isServerUp returns true if server is up and running
// server parameter has to be a valid url
func isServerUp(server string, timeout time.Duration) bool {
	address, err := dfutil.GetHostWithPort(server)
	if err != nil {
		klog.V(3).Infof("Unable to parse url %s (%s)", server, err)
	}
	klog.V(3).Infof("Trying to connect to server %s", address)
	_, connectionError := net.DialTimeout("tcp", address, timeout)
	if connectionError != nil {
		klog.V(3).Info(fmt.Errorf("unable to connect to server: %w", connectionError))
		return false
	}

	klog.V(3).Infof("Server %v is up", server)
	return true
}

// ServerInfo contains the fields that contain the server's information like
// address, OpenShift and Kubernetes versions
type ServerInfo struct {
	Address           string
	OpenShiftVersion  string
	KubernetesVersion string
}

// GetServerVersion will fetch the Server Host, OpenShift and Kubernetes Version
// It will be shown on the execution of odo version command
func (c *Client) GetServerVersion(timeout time.Duration) (*ServerInfo, error) {
	var info ServerInfo

	// This will fetch the information about Server Address
	config, err := c.KubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get server's address: %w", err)
	}
	info.Address = config.Host

	// checking if the server is reachable
	if !isServerUp(config.Host, timeout) {
		return nil, errors.New("unable to connect to OpenShift cluster, it may be down")
	}

	// This will fetch the information about OpenShift Version
	coreGet := c.GetClient().CoreV1().RESTClient().Get()
	rawOpenShiftVersion, err := coreGet.AbsPath("/apis/config.openshift.io/v1/clusterversions/version").Do(context.TODO()).Raw()
	if err != nil {
		klog.V(3).Info("Unable to get OpenShift Version: ", err)
	} else {
		var openShiftVersion configv1.ClusterVersion
		if e := json.Unmarshal(rawOpenShiftVersion, &openShiftVersion); e != nil {
			return nil, fmt.Errorf("unable to unmarshal OpenShift version %v: %w", string(rawOpenShiftVersion), e)
		}
		info.OpenShiftVersion = openShiftVersion.Status.Desired.Version
	}

	// This will fetch the information about Kubernetes Version
	rawKubernetesVersion, err := coreGet.AbsPath("/version").Do(context.TODO()).Raw()
	if err != nil {
		return nil, fmt.Errorf("unable to get Kubernetes Version: %w", err)
	}
	var kubernetesVersion version.Info
	if err := json.Unmarshal(rawKubernetesVersion, &kubernetesVersion); err != nil {
		return nil, fmt.Errorf("unable to unmarshal Kubernetes Version: %v: %w", string(rawKubernetesVersion), err)
	}
	info.KubernetesVersion = kubernetesVersion.GitVersion

	return &info, nil
}

func (c *Client) GetOCVersion() (string, error) {
	clusterVersion, err := c.configClient.ClusterVersions().Get(context.TODO(), "version", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	switch {
	case kerrors.IsForbidden(err), kerrors.IsNotFound(err):
		return "", err
	}
	if clusterVersion != nil {
		if len(clusterVersion.Status.History) == 1 {
			return clusterVersion.Status.History[0].Version, nil
		}
		for _, update := range clusterVersion.Status.History {
			if update.State == configv1.CompletedUpdate {
				// obtain the version from the last completed update
				return update.Version, nil
			}
		}
	}
	return "", errors.New("unable to get OC version")
}
