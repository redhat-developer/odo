//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"bytes"
	"fmt"
	"io"

	"github.com/devfile/library/v2/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	k8yaml "sigs.k8s.io/yaml"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
)

// YamlSrc specifies the src of the yaml in either Path, URL or Data format
type YamlSrc struct {
	// Path to the yaml file
	Path string
	// URL of the yaml file
	URL string
	// Data is the yaml content in []byte format
	Data []byte
}

// KubernetesResources struct contains the Deployments, Services,
// Routes and Ingresses resources
type KubernetesResources struct {
	Deployments []appsv1.Deployment
	Services    []corev1.Service
	Routes      []routev1.Route
	Ingresses   []extensionsv1.Ingress
	Others      []interface{}
}

// ReadKubernetesYaml reads a yaml Kubernetes file from either the Path, URL or Data provided.
// It returns all the parsed Kubernetes objects as an array of interface.
// Consumers interested in the Kubernetes resources are expected to Unmarshal
// it to the struct of the respective Kubernetes resource. If a Path is being passed,
// provide a filesystem, otherwise nil can be passed in
func ReadKubernetesYaml(src YamlSrc, fs *afero.Afero) ([]interface{}, error) {

	var data []byte
	var err error

	if src.URL != "" {
		params := util.HTTPRequestParams{URL: src.URL}
		data, err = util.DownloadInMemory(params)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to download file %q", src.URL)
		}
	} else if src.Path != "" {
		if fs == nil {
			return nil, fmt.Errorf("cannot read from %s because fs passed in was nil", src.Path)
		}
		data, err = fs.ReadFile(src.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read yaml from path %q", src.Path)
		}
	} else if len(src.Data) > 0 {
		data = src.Data
	}

	var values []interface{}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var value interface{}
		err = dec.Decode(&value)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}

// ParseKubernetesYaml Unmarshals the interface array of the Kubernetes resources
// and returns it as a KubernetesResources struct. Only Deployment, Service, Route
// and Ingress are processed. Consumers interested in other Kubernetes resources
// are expected to parse the values interface array an Unmarshal it to their
// desired Kuberenetes struct
func ParseKubernetesYaml(values []interface{}) (KubernetesResources, error) {
	var deployments []appsv1.Deployment
	var services []corev1.Service
	var routes []routev1.Route
	var ingresses []extensionsv1.Ingress
	var otherResources []interface{}

	for _, value := range values {
		var deployment appsv1.Deployment
		var service corev1.Service
		var route routev1.Route
		var ingress extensionsv1.Ingress
		var otherResource interface{}

		byteData, err := k8yaml.Marshal(value)
		if err != nil {
			return KubernetesResources{}, err
		}

		kubernetesMap := value.(map[string]interface{})
		kind := kubernetesMap["kind"]

		switch kind {
		case "Deployment":
			err = k8yaml.Unmarshal(byteData, &deployment)
			deployments = append(deployments, deployment)
		case "Service":
			err = k8yaml.Unmarshal(byteData, &service)
			services = append(services, service)
		case "Route":
			err = k8yaml.Unmarshal(byteData, &route)
			routes = append(routes, route)
		case "Ingress":
			err = k8yaml.Unmarshal(byteData, &ingress)
			ingresses = append(ingresses, ingress)
		default:
			err = k8yaml.Unmarshal(byteData, &otherResource)
			otherResources = append(otherResources, otherResource)
		}

		if err != nil {
			return KubernetesResources{}, err
		}
	}

	return KubernetesResources{
		Deployments: deployments,
		Services:    services,
		Routes:      routes,
		Ingresses:   ingresses,
		Others:      otherResources,
	}, nil
}
