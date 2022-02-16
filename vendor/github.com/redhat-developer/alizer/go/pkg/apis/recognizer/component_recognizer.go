/*******************************************************************************
 * Copyright (c) 2022 Red Hat, Inc.
 * Distributed under license by Red Hat, Inc. All rights reserved.
 * This program is made available under the terms of the
 * Eclipse Public License v2.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 * Red Hat, Inc.
 ******************************************************************************/
package recognizer

import (
	"io/fs"
	"path/filepath"
	"strings"

	enricher "github.com/redhat-developer/alizer/go/pkg/apis/enricher"
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	"github.com/redhat-developer/alizer/go/pkg/utils/langfiles"
)

type Component struct {
	Path      string
	Languages []language.Language
}

func DetectComponents(path string) ([]Component, error) {
	files, err := getFilePaths(path)
	if err != nil {
		return []Component{}, err
	}
	components, err := detectComponents(files)
	if err != nil {
		return []Component{}, err
	}

	// it may happen that a language has no a specific configuration file (e.g opposite to JAVA -> pom.xml and Nodejs -> package.json)
	// we then rely on the language recognizer
	directoriesNotBelongingToExistingComponent := getDirectoriesWithoutConfigFile(path, components)
	components = append(components, getComponentsWithoutConfigFile(directoriesNotBelongingToExistingComponent)...)

	return components, nil
}

/*
	getComponentsWithoutConfigFile retrieves the components which are written with a language that does not require a config file
	Parameters:
		directories: list of directories to analyze
	Returns:
		components found
*/
func getComponentsWithoutConfigFile(directories []string) []Component {
	var components []Component
	for _, dir := range directories {
		component, _ := detectComponent(dir, "")
		if component.Path != "" && isLangForNoConfigComponent(component.Languages) {
			components = append(components, component)
		}
	}
	return components
}

/*
	isLangForNoConfigComponent verify if main language requires any config file
	Parameters:
		component:
	Returns:
		bool: true if language does not require any config file
*/
func isLangForNoConfigComponent(languages []language.Language) bool {
	if len(languages) == 0 {
		return false
	}

	lang, err := langfiles.Get().GetLanguageByNameOrAlias(languages[0].Name)
	if err != nil {
		return false
	}

	return len(lang.ConfigurationFiles) == 0
}

/*
	getDirectoriesPathsWithoutConfigFile retrieves all directories that do not contain any Component
	Parameters:
		root: root folder where to start the search
		components: list of components already detected
	Returns:
		list of directories path that does not contain any component
*/
func getDirectoriesWithoutConfigFile(root string, components []Component) []string {
	if len(components) == 0 {
		return []string{root}
	}
	directories := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !strings.EqualFold(root, path) && d.IsDir() && !isAnyComponentInPath(path, components) {
			directories = getParentFolders(path, directories)
		}
		return nil
	})
	if err != nil {
		return []string{}
	}
	return directories
}

/**
* Return all paths which are not sub-folders of some other path within the list
* Target will be added to the list if it is not a sub-folder of any other path within the list
* If a path in the list is sub-folder of Target, that path will be removed.
*
* @param target new path to be added
* @param directories list of all previously added paths
* @return the list containing all paths which are not sub-folders of any other
 */
func getParentFolders(target string, directories []string) []string {
	updatedDirectories := []string{}
	for _, dir := range directories {
		if isFirstPathParentOfSecond(dir, target) {
			return directories
		}

		if isFirstPathParentOfSecond(target, dir) {
			continue
		}
		updatedDirectories = append(updatedDirectories, dir)
	}

	updatedDirectories = append(updatedDirectories, target)
	return updatedDirectories
}

/*
	isAnyComponentInPath checks if a component is present in path
	Parameters:
		path: path where to search for component
		components: list of components
	Returns:
		true if a component is found starting from path
*/
func isAnyComponentInPath(path string, components []Component) bool {
	for _, component := range components {
		if strings.EqualFold(path, component.Path) || isFirstPathParentOfSecond(component.Path, path) || isFirstPathParentOfSecond(path, component.Path) {
			return true
		}
	}
	return false
}

/*
	isFirstPathParentOfSecond check if first path is parent (direct or not) of second path
	Parameters:
		firstPath: path to be used as parent
		secondPath: path to be used as child
	Returns:
		true if firstPath is part of secondPath
*/
func isFirstPathParentOfSecond(firstPath string, secondPath string) bool {
	return strings.Contains(secondPath, firstPath)
}

/*
	detectComponents detect components by analyzing all files
	Parameters:
		files: list of files to analyze
	Returns:
		list of components detected or err if any error occurs
*/
func detectComponents(files []string) ([]Component, error) {
	configurationPerLanguage := langfiles.Get().GetConfigurationPerLanguageMapping()
	var components []Component
	for _, file := range files {
		dir, fileName := filepath.Split(file)
		if language, isConfig := configurationPerLanguage[fileName]; isConfig && isConfigurationValid(language, file) {
			component, err := detectComponent(dir, language)
			if err != nil {
				return []Component{}, err
			}
			if component.Path != "" {
				components = append(components, component)
			}
		}
	}
	return components, nil
}

/*
	detectComponent returns a Component if found:
							- language must be enabled for component detection
							- there should be at least one framework detected
					, error otherwise
	Parameters:
		root: path to be used as root where to start the detection
		language: language to be used as target for detection
	Returns:
		component detected or error if any error occurs
*/
func detectComponent(root string, language string) (Component, error) {
	languages, err := Analyze(root)
	if err != nil {
		return Component{}, err
	}
	languages = getLanguagesWeightedByConfigFile(languages, language)
	if len(languages) > 0 {
		if mainLang := languages[0]; mainLang.CanBeComponent && len(mainLang.Frameworks) > 0 {
			return Component{
				Path:      root,
				Languages: languages,
			}, nil
		}
	}

	return Component{}, nil

}

/*
	getLanguagesWeightedByConfigFile returns the list of languages reordered by importance per config file.
									 Language found by analyzing the config file is used as target.
	Parameters:
		languages: list of languages to be reordered
		languageByConfig: target language
	Returns:
		list of languages reordered
*/
func getLanguagesWeightedByConfigFile(languages []language.Language, languageByConfig string) []language.Language {
	if languageByConfig == "" {
		return languages
	}

	for index, lang := range languages {
		if strings.EqualFold(lang.Name, languageByConfig) {
			sliceWithoutLang := append(languages[:index], languages[index+1:]...)
			return append([]language.Language{lang}, sliceWithoutLang...)
		}
	}
	return languages
}

func isConfigurationValid(language string, file string) bool {
	langEnricher := enricher.GetEnricherByLanguage(language)
	if langEnricher != nil {
		return langEnricher.IsConfigValidForComponentDetection(language, file)
	}
	return false
}
