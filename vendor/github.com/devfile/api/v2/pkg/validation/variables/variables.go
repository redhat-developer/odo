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

package variables

import (
	"regexp"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// example of the regex: {{variable}} / {{ variable }}
var globalVariableRegex = regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

// VariableWarning stores the invalid variable references for each devfile object
type VariableWarning struct {
	// Commands stores a map of command ids to the invalid variable references
	Commands map[string][]string

	// Components stores a map of component names to the invalid variable references
	Components map[string][]string

	// Projects stores a map of project names to the invalid variable references
	Projects map[string][]string

	// StarterProjects stores a map of starter project names to the invalid variable references
	StarterProjects map[string][]string

	// DependentProjects stores a map of dependent project names to invalid variable references
	DependentProjects map[string][]string
}

// ValidateAndReplaceGlobalVariable validates the workspace template spec data for global variable references and replaces them with the variable value
func ValidateAndReplaceGlobalVariable(workspaceTemplateSpec *v1alpha2.DevWorkspaceTemplateSpec) VariableWarning {

	var variableWarning VariableWarning

	if workspaceTemplateSpec != nil {
		// Validate the components and replace for global variable
		variableWarning.Components = ValidateAndReplaceForComponents(workspaceTemplateSpec.Variables, workspaceTemplateSpec.Components)

		// Validate the commands and replace for global variable
		variableWarning.Commands = ValidateAndReplaceForCommands(workspaceTemplateSpec.Variables, workspaceTemplateSpec.Commands)

		// Validate the projects and replace for global variable
		variableWarning.Projects = ValidateAndReplaceForProjects(workspaceTemplateSpec.Variables, workspaceTemplateSpec.Projects)

		// Validate the starter projects and replace for global variable
		variableWarning.StarterProjects = ValidateAndReplaceForStarterProjects(workspaceTemplateSpec.Variables, workspaceTemplateSpec.StarterProjects)

		// Validate the starter projects and replace for global variable
		variableWarning.DependentProjects = ValidateAndReplaceForProjects(workspaceTemplateSpec.Variables, workspaceTemplateSpec.DependentProjects)
	}

	return variableWarning
}

// validateAndReplaceDataWithVariable validates the string for a global variable and replaces it. An error
// is returned if the string references an invalid global variable key
func validateAndReplaceDataWithVariable(val string, variables map[string]string) (string, error) {
	matches := globalVariableRegex.FindAllStringSubmatch(val, -1)
	var invalidKeys []string
	for _, match := range matches {
		varValue, ok := variables[match[1]]
		if !ok {
			invalidKeys = append(invalidKeys, match[1])
		} else {
			val = strings.Replace(val, match[0], varValue, -1)
		}
	}

	if len(invalidKeys) > 0 {
		return val, &InvalidKeysError{Keys: invalidKeys}
	}

	return val, nil
}
