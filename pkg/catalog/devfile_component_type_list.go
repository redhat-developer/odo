package catalog

import "sort"

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
