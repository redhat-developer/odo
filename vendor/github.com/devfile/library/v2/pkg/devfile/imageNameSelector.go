//
// Copyright 2023 Red Hat, Inc.
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

package devfile

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/distribution/distribution/v3/reference"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

var k8sSerializer = json.NewSerializerWithOptions(
	json.DefaultMetaFactory,
	scheme.Scheme,
	scheme.Scheme,
	json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
	})

// replaceImageNames parses all Image components in the specified Devfile object and,
// for each relative image name, replaces the value in all matching Image, Container and Kubernetes/Openshift components.
//
// An image is said to be relative if it has a canonical name different from its actual name.
// For example, image names like 'nodejs-devtools', 'nodejs-devtools:some-tag', 'nodejs-devtools@digest', or even 'some_name_different_from_localhost/nodejs-devtools'  are all relative because
// their canonical form (as returned by the Distribution library) will be prefixed with 'docker.io/library/'.
// On the other hand, image names like 'docker.io/library/nodejs-devtools', 'localhost/nodejs-devtools@digest'  or 'quay.io/nodejs-devtools:some-tag' are absolute.
//
// A component is said to be matching if the base name of the image used in this component is the same as the base name of the image component, regardless of its tag, digest or registry.
// For example, if the Devfile has an Image component with an image named 'nodejs-devtools' and 2 Container components using an image named 'nodejs-devtools:some-tag' and another absolute image named
// 'quay.io/nodejs-devtools@digest', both image names in the two Container components will be replaced by a value described below (because the base names of those images are 'nodejs-devtools', which
// match the base name of the relative image name of the Image Component).
// But `nodejs-devtools2` or 'ghcr.io/some-user/nodejs-devtools3' do not match the 'nodejs-devtools' image name and won't be replaced.
//
// For Kubernetes and OpenShift components, this function assumes that the actual resource manifests are inlined in the components,
// in order to perform any replacements for matching image names.
// At the moment, this function only supports replacements in Kubernetes native resource types (Pod, CronJob, Job, DaemonSet; Deployment, ReplicaSet, ReplicationController, StatefulSet).
//
// Absolute images and non-matching image references are left unchanged.
//
// And the replacement is done by using the following format: "<registry>/<devfileName>-<baseImageName>:<imageTag>",
// where both <registry>  and <imageTag>  are set by the tool itself (either via auto-detection or via user input).
func replaceImageNames(d *parser.DevfileObj, registry string, imageTag string) (err error) {
	var imageComponents []v1.Component
	imageComponents, err = d.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1.ImageComponentType},
	})
	if err != nil {
		return err
	}

	var isAbs bool
	var imageRef reference.Named
	for _, comp := range imageComponents {
		imageName := comp.Image.ImageName
		isAbs, imageRef, err = parseImageReference(imageName)
		if err != nil {
			return err
		}
		if isAbs {
			continue
		}
		baseImageName := getImageSimpleName(imageRef)

		replacement := baseImageName
		if d.GetMetadataName() != "" {
			replacement = fmt.Sprintf("%s-%s", d.GetMetadataName(), replacement)
		}
		if registry != "" {
			replacement = fmt.Sprintf("%s/%s", strings.TrimSuffix(registry, "/"), replacement)
		}
		if imageTag != "" {
			replacement += fmt.Sprintf(":%s", imageTag)
		}

		// Replace so that the image can be built and pushed to the registry specified by the tool.
		comp.Image.ImageName = replacement

		// Replace in matching container components
		err = handleContainerComponents(d, baseImageName, replacement)
		if err != nil {
			return err
		}

		// Replace in matching Kubernetes and OpenShift components
		err = handleKubernetesLikeComponents(d, baseImageName, replacement)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseImageReference uses the Docker reference library to detect if the image name is absolute or not
// and returns a struct from which we can extract the domain, tag and digest if needed.
func parseImageReference(imageName string) (isAbsolute bool, imageRef reference.Named, err error) {
	imageRef, err = reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return false, nil, err
	}

	// Non-canonical image references are not absolute.
	// For example, "nodejs-devtools" will be parsed as "docker.io/library/nodejs-devtools"
	isAbsolute = imageRef.String() == imageName

	return isAbsolute, imageRef, nil
}

func getImageSimpleName(img reference.Named) string {
	p := reference.Path(img)
	i := strings.LastIndex(p, "/")
	result := p
	if i >= 0 {
		result = strings.TrimPrefix(p[i:], "/")
	}
	return result
}

func hasMatch(baseImageName, compImage string) (bool, error) {
	_, imageRef, err := parseImageReference(compImage)
	if err != nil {
		return false, err
	}
	return getImageSimpleName(imageRef) == baseImageName, nil
}

func handleContainerComponents(d *parser.DevfileObj, baseImageName, replacement string) (err error) {
	var containerComponents []v1.Component
	containerComponents, err = d.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1.ContainerComponentType},
	})
	if err != nil {
		return err
	}

	for _, comp := range containerComponents {
		var match bool
		match, err = hasMatch(baseImageName, comp.Container.Image)
		if err != nil {
			return err
		}
		if !match {
			continue
		}
		comp.Container.Image = replacement
	}
	return nil
}

func handleKubernetesLikeComponents(d *parser.DevfileObj, baseImageName, replacement string) error {
	var allK8sOcComponents []v1.Component

	k8sComponents, err := d.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1.KubernetesComponentType},
	})
	if err != nil {
		return err
	}
	allK8sOcComponents = append(allK8sOcComponents, k8sComponents...)

	ocComponents, err := d.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1.OpenshiftComponentType},
	})
	if err != nil {
		return err
	}
	allK8sOcComponents = append(allK8sOcComponents, ocComponents...)

	updateImageInPodSpecIfNeeded := func(obj runtime.Object, ps *corev1.PodSpec) (string, error) {
		handleContainer := func(c *corev1.Container) (match bool, err error) {
			match, err = hasMatch(baseImageName, c.Image)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil
			}
			c.Image = replacement
			return true, nil
		}
		for i := range ps.Containers {
			if _, err = handleContainer(&ps.Containers[i]); err != nil {
				return "", err
			}
		}
		for i := range ps.InitContainers {
			if _, err = handleContainer(&ps.InitContainers[i]); err != nil {
				return "", err
			}
		}
		for i := range ps.EphemeralContainers {
			if _, err = handleContainer((*corev1.Container)(&ps.EphemeralContainers[i].EphemeralContainerCommon)); err != nil {
				return "", err
			}
		}

		//Encode obj back into a YAML string
		var s strings.Builder
		err = k8sSerializer.Encode(obj, &s)
		if err != nil {
			return "", err
		}

		return s.String(), nil
	}

	handleK8sContent := func(content string) (newContent string, err error) {
		multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewBufferString(content)))
		var yamlAsStringList []string
		var buf []byte
		var obj runtime.Object
		for {
			buf, err = multidocReader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", err
			}

			obj, _, err = k8sSerializer.Decode(buf, nil, nil)
			if err != nil {
				// Use raw string as it is, as it might be a Custom Resource with a Kind that is not known
				// by the K8s decoder.
				yamlAsStringList = append(yamlAsStringList, strings.TrimSpace(string(buf)))
				continue
			}

			newYaml := string(buf)
			switch r := obj.(type) {
			case *batchv1.CronJob:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.JobTemplate.Spec.Template.Spec)
			case *appsv1.DaemonSet:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.Template.Spec)
			case *appsv1.Deployment:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.Template.Spec)
			case *batchv1.Job:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.Template.Spec)
			case *corev1.Pod:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec)
			case *appsv1.ReplicaSet:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.Template.Spec)
			case *corev1.ReplicationController:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.Template.Spec)
			case *appsv1.StatefulSet:
				newYaml, err = updateImageInPodSpecIfNeeded(r, &r.Spec.Template.Spec)
			}

			if err != nil {
				return "", err
			}

			yamlAsStringList = append(yamlAsStringList, strings.TrimSpace(newYaml))
		}

		return strings.Join(yamlAsStringList, "\n---\n"), nil
	}

	var newContent string
	for _, comp := range allK8sOcComponents {
		if comp.Kubernetes != nil {
			newContent, err = handleK8sContent(comp.Kubernetes.Inlined)
			if err != nil {
				return err
			}
			comp.Kubernetes.Inlined = newContent
		} else {
			newContent, err = handleK8sContent(comp.Openshift.Inlined)
			if err != nil {
				return err
			}
			comp.Openshift.Inlined = newContent
		}
	}

	return nil
}
