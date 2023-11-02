//
//
// Copyright Red Hat
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

package v1alpha2

// ImageType describes the type of image.
// Only one of the following image type may be specified.
// +kubebuilder:validation:Enum=Dockerfile
type ImageType string

const (
	DockerfileImageType ImageType = "Dockerfile"
)

type BaseImage struct {
}

// Component that allows the developer to build a runtime image for outerloop
type ImageComponent struct {
	BaseComponent `json:",inline"`
	Image         `json:",inline"`
}

type Image struct {
	// Name of the image for the resulting outerloop build
	ImageName  string `json:"imageName"`
	ImageUnion `json:",inline"`
}

// +union
// +devfile:getter:generate
type ImageUnion struct {
	// Type of image
	//
	// +unionDiscriminator
	// +optional
	ImageType ImageType `json:"imageType,omitempty"`

	// Allows specifying dockerfile type build
	// +optional
	Dockerfile *DockerfileImage `json:"dockerfile,omitempty"`

	// Defines if the image should be built during startup.
	//
	// Default value is `false`
	// +optional
	// +devfile:default:value=false
	AutoBuild *bool `json:"autoBuild,omitempty"`
}
