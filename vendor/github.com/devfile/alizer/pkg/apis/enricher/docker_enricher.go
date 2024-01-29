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

package enricher

import (
	"context"
	"github.com/devfile/alizer/pkg/apis/model"
)

type DockerEnricher struct{}

type DockerFrameworkDetector interface {
	DoPortsDetection(component *model.Component, ctx *context.Context)
}

func (d DockerEnricher) GetSupportedLanguages() []string {
	return []string{"dockerfile"}
}

func (d DockerEnricher) DoEnrichLanguage(language *model.Language, _ *[]string) {
	// The Dockerfile language does not contain frameworks
}

func (d DockerEnricher) DoEnrichComponent(component *model.Component, _ model.DetectionSettings, _ *context.Context) {
	projectName := GetDefaultProjectName(component.Path)
	component.Name = projectName

	ports := GetPortsFromDockerFile(component.Path)
	if len(ports) > 0 {
		component.Ports = ports
	}
}

func (d DockerEnricher) IsConfigValidForComponentDetection(language string, config string) bool {
	return IsConfigurationValidForLanguage(language, config)
}
