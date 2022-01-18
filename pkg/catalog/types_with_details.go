package catalog

import (
	"errors"
	"fmt"
	"sort"
)

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
