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

package validation

import (
	"net/url"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// attribute keys for imported and overridden elements
// the value of those keys is the resource information
const (
	// attribute key of the imported element resource information
	ImportSourceAttribute = "api.devfile.io/imported-from"
	// attribute key of the parent overridden element resource information
	ParentOverrideAttribute = "api.devfile.io/parent-override-from"
	// attribute key of the plugin overridden element resource information
	PluginOverrideAttribute = "api.devfile.io/plugin-override-from"
)

// getCommandsMap iterates through the commands and returns a map of command
func getCommandsMap(commands []v1alpha2.Command) map[string]v1alpha2.Command {
	commandMap := make(map[string]v1alpha2.Command, len(commands))

	for _, command := range commands {
		command.Id = strings.ToLower(command.Id)
		commandMap[command.Id] = command
	}

	return commandMap
}

// ValidateURI checks if the string is with valid uri format, return error if not valid
func ValidateURI(uri string) error {
	if strings.HasPrefix(uri, "http") {
		if _, err := url.ParseRequestURI(uri); err != nil {
			return err
		}
	} else if _, err := url.Parse(uri); err != nil {
		return err
	}

	return nil
}
