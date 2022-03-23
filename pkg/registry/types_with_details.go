package registry

import (
	"errors"
	"fmt"
	"sort"
)

// GetOrderedLabels returns a list of labels for a list of project types
func (types TypesWithDetails) GetOrderedLabels() []string {
	sortedTypes := sortTypes(types)
	stringTypes := []string{}
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
func (types TypesWithDetails) GetAtOrderedPosition(pos int) (DevfileStack, error) {
	sortedTypes := sortTypes(types)
	for _, typ := range sortedTypes {
		detailsList := types[typ]
		if pos >= len(detailsList) {
			pos -= len(detailsList)
			continue
		}
		return detailsList[pos], nil
	}
	return DevfileStack{}, errors.New("index not found")
}

func sortTypes(types TypesWithDetails) []string {
	sortedTypes := make([]string, 0, len(types))
	for typ := range types {
		sortedTypes = append(sortedTypes, typ)
	}
	sort.Strings(sortedTypes)
	return sortedTypes
}
