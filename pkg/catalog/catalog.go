package catalog

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/redhat-developer/odo/pkg/kclient"
)

type Client interface {
	GetDevfileRegistries(registryName string) ([]Registry, error)
	ListDevfileComponents(registryName string) (DevfileComponentTypeList, error)
	GetStarterProjectsNames(details DevfileComponentType) ([]string, error)
	SearchComponent(client kclient.ClientInterface, name string) ([]string, error)
}

// GetLanguages returns the list of unique languages, ordered by name,
// from a list of registry items
func (o *DevfileComponentTypeList) GetLanguages() []string {
	languagesMap := map[string]bool{}
	for _, item := range o.Items {
		languagesMap[item.Language] = true
	}

	languages := make([]string, 0, len(languagesMap))
	for k := range languagesMap {
		languages = append(languages, k)
	}
	sort.Strings(languages)
	return languages
}

type TypesWithDetails map[string][]DevfileComponentType

// GetProjectTypes returns the list of project types and associated details
// from a list of registry items
func (o *DevfileComponentTypeList) GetProjectTypes(language string) TypesWithDetails {
	types := TypesWithDetails{}
	for _, item := range o.Items {
		if item.Language != language {
			continue
		}
		if _, found := types[item.DisplayName]; !found {
			types[item.DisplayName] = []DevfileComponentType{}
		}
		types[item.DisplayName] = append(types[item.DisplayName], item)
	}
	return types
}

// GetOrderedLabels returns a list of labels for a list of project types
func (types TypesWithDetails) GetOrderedLabels() []string {
	stringTypes := []string{}

	sortedTypes := make([]string, 0, len(types))
	for typ := range types {
		sortedTypes = append(sortedTypes, typ)
	}
	sort.Strings(sortedTypes)

	for _, typ := range sortedTypes {
		detailsList := types[typ]
		if len(detailsList) == 1 {
			stringTypes = append(stringTypes, typ)
		} else {
			for _, details := range detailsList {
				stringTypes = append(stringTypes, fmt.Sprintf("%s (%s, registry: %s)", typ, details.Name, details.Registry.Name))
			}
		}
	}
	return stringTypes
}

// GetAtOrderedPosition returns the project type at the given position,
// when the list of project types is ordered by GetOrderedLabels
func (types TypesWithDetails) GetAtOrderedPosition(pos int) (DevfileComponentType, error) {
	sortedTypes := make([]string, 0, len(types))
	for typ := range types {
		sortedTypes = append(sortedTypes, typ)
	}
	sort.Strings(sortedTypes)

	for _, typ := range sortedTypes {
		detailsList := types[typ]
		if pos >= len(detailsList) {
			pos -= len(detailsList)
			continue
		}
		return detailsList[pos], nil
	}
	return DevfileComponentType{}, errors.New("index not found")
}
