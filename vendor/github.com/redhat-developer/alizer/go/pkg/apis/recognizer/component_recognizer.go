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
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	enricher "github.com/redhat-developer/alizer/go/pkg/apis/enricher"
	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	"github.com/redhat-developer/alizer/go/pkg/utils/langfiles"
)

func DetectComponentsInRoot(path string) ([]model.Component, error) {
	ctx := context.Background()
	return detectComponentsInRootWithPathAndPortStartegy(path, []model.PortDetectionAlgorithm{model.DockerFile, model.Compose, model.Source}, &ctx)
}

func DetectComponents(path string) ([]model.Component, error) {
	ctx := context.Background()
	return detectComponentsWithPathAndPortStartegy(path, []model.PortDetectionAlgorithm{model.DockerFile, model.Compose, model.Source}, &ctx)
}

func DetectComponentsInRootWithPathAndPortStartegy(path string, portDetectionStrategy []model.PortDetectionAlgorithm) ([]model.Component, error) {
	ctx := context.Background()
	return detectComponentsInRootWithPathAndPortStartegy(path, portDetectionStrategy, &ctx)
}

func detectComponentsInRootWithPathAndPortStartegy(path string, portDetectionStrategy []model.PortDetectionAlgorithm, ctx *context.Context) ([]model.Component, error) {
	return detectComponentsInRootWithSettings(model.DetectionSettings{
		BasePath:              path,
		PortDetectionStrategy: portDetectionStrategy,
	}, ctx)
}

func DetectComponentsWithPathAndPortStartegy(path string, portDetectionStrategy []model.PortDetectionAlgorithm) ([]model.Component, error) {
	ctx := context.Background()
	return detectComponentsWithPathAndPortStartegy(path, portDetectionStrategy, &ctx)
}

func detectComponentsWithPathAndPortStartegy(path string, portDetectionStrategy []model.PortDetectionAlgorithm, ctx *context.Context) ([]model.Component, error) {
	return detectComponentsWithSettings(model.DetectionSettings{
		BasePath:              path,
		PortDetectionStrategy: portDetectionStrategy,
	}, ctx)
}

func DetectComponentsInRootWithSettings(settings model.DetectionSettings) ([]model.Component, error) {
	ctx := context.Background()
	return detectComponentsInRootWithSettings(settings, &ctx)
}

func detectComponentsInRootWithSettings(settings model.DetectionSettings, ctx *context.Context) ([]model.Component, error) {
	files, err := utils.GetFilePathsInRoot(settings.BasePath)
	if err != nil {
		return []model.Component{}, err
	}
	components := DetectComponentsFromFilesList(files, settings, ctx)

	return components, nil
}

func DetectComponentsWithSettings(settings model.DetectionSettings) ([]model.Component, error) {
	ctx := context.Background()
	return detectComponentsWithSettings(settings, &ctx)
}

func detectComponentsWithSettings(settings model.DetectionSettings, ctx *context.Context) ([]model.Component, error) {
	files, err := utils.GetCachedFilePathsFromRoot(settings.BasePath, ctx)
	if err != nil {
		return []model.Component{}, err
	}
	components := DetectComponentsFromFilesList(files, settings, ctx)

	// it may happen that a language has no a specific configuration file (e.g opposite to JAVA -> pom.xml and Nodejs -> package.json)
	// we then rely on the language recognizer
	directoriesNotBelongingToExistingComponent := getDirectoriesWithoutConfigFile(settings.BasePath, components)
	components = append(components, getComponentsWithoutConfigFile(directoriesNotBelongingToExistingComponent, settings, ctx)...)

	return components, nil
}

/*
	getComponentsWithoutConfigFile retrieves the components which are written with a language that does not require a config file
	Parameters:
		directories: list of directories to analyze
		settings: settings to perform detection
	Returns:
		components found
*/
func getComponentsWithoutConfigFile(directories []string, settings model.DetectionSettings, ctx *context.Context) []model.Component {
	var components []model.Component
	for _, dir := range directories {
		component, _ := detectComponentByFolderAnalysis(dir, []string{}, settings, ctx)
		if component.Path != "" && isLangForNoConfigComponent(component.Languages[0]) {
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
func isLangForNoConfigComponent(language model.Language) bool {
	lang, err := langfiles.Get().GetLanguageByNameOrAlias(language.Name)
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
func getDirectoriesWithoutConfigFile(root string, components []model.Component) []string {
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

/*
	getParentFolders return all paths which are not sub-folders of some other path within the list
	Target will be added to the list if it is not a sub-folder of any other path within the list
	If a path in the list is sub-folder of Target, that path will be removed.
	Parameters:
		target: new path to be added
		directories: list of all previously added paths
	Returns:
		the list containing all paths which are not sub-folders of any other
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
func isAnyComponentInPath(path string, components []model.Component) bool {
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
	DetectComponentsFromFilesList detect components by analyzing all files
	Parameters:
		files: list of files to analyze
		settings: settings to perform detection
	Returns:
		list of components detected or err if any error occurs
*/
func DetectComponentsFromFilesList(files []string, settings model.DetectionSettings, ctx *context.Context) []model.Component {
	configurationPerLanguage := langfiles.Get().GetConfigurationPerLanguageMapping()
	var components []model.Component
	for _, file := range files {
		languages, err := getLanguagesByConfigurationFile(configurationPerLanguage, file)
		if err != nil {
			continue
		}

		component, err := detectComponentUsingConfigFile(file, languages, settings, ctx)
		if err != nil {
			continue
		}
		components = appendIfMissing(components, component)
	}
	return components
}

func appendIfMissing(components []model.Component, component model.Component) []model.Component {
	for _, existing := range components {
		if strings.EqualFold(existing.Path, component.Path) && strings.EqualFold(existing.Languages[0].Name, component.Languages[0].Name) {
			return components
		}
	}
	return append(components, component)
}

func getLanguagesByConfigurationFile(configurationPerLanguage map[string][]string, file string) ([]string, error) {
	for regex, languages := range configurationPerLanguage {
		if match, _ := regexp.MatchString(regex, file); match {
			return languages, nil
		}
	}
	return nil, errors.New("no languages found for configuration file " + file)
}

/*
	detectComponent returns a Component if found:
							- language must be enabled for component detection
					, error otherwise
	Parameters:
		root: path to be used as root where to start the detection
		configLanguages: languages associated to the config file found and to be used as target for detection
		settings: settings to perform detection
	Returns:
		component detected or error if any error occurs
*/
func detectComponentByFolderAnalysis(root string, configLanguages []string, settings model.DetectionSettings, ctx *context.Context) (model.Component, error) {
	languages, err := analyze(root, ctx)
	if err != nil {
		return model.Component{}, err
	}
	languages = getLanguagesWeightedByConfigFile(languages, configLanguages)
	if len(languages) > 0 {
		if mainLang := languages[0]; mainLang.CanBeComponent {
			component := model.Component{
				Path:      root,
				Languages: languages,
			}
			enrichComponent(&component, settings, ctx)
			return component, nil
		}
	}

	return model.Component{}, errors.New("no component detected")

}

func detectComponentByAnalyzingConfigFile(file string, language string, settings model.DetectionSettings, ctx *context.Context) (model.Component, error) {
	if !isConfigurationValid(language, file) {
		return model.Component{}, errors.New("language not valid for component detection")
	}
	dir, _ := utils.NormalizeSplit(file)
	lang, err := AnalyzeFile(file, language)
	if err != nil {
		return model.Component{}, err
	}
	component := model.Component{
		Path: dir,
		Languages: []model.Language{
			lang,
		},
	}
	enrichComponent(&component, settings, ctx)
	return component, nil
}

func doBelongToSameFamily(languages []string) bool {
	return (len(languages) == 2 &&
		languages[0] != languages[1] &&
		(strings.ToLower(languages[0]) == "typescript" || strings.ToLower(languages[0]) == "javascript") &&
		(strings.ToLower(languages[1]) == "typescript" || strings.ToLower(languages[1]) == "javascript"))
}

func detectComponentUsingConfigFile(file string, languages []string, settings model.DetectionSettings, ctx *context.Context) (model.Component, error) {
	if len(languages) == 1 || doBelongToSameFamily(languages) {
		return detectComponentByAnalyzingConfigFile(file, languages[0], settings, ctx)
	} else {
		dir, _ := utils.NormalizeSplit(file)
		for _, language := range languages {
			if isConfigurationValid(language, file) {
				return detectComponentByFolderAnalysis(dir, languages, settings, ctx)
			}
		}
	}
	return model.Component{}, errors.New("no component detected")
}

func enrichComponent(component *model.Component, settings model.DetectionSettings, ctx *context.Context) {
	componentEnricher := enricher.GetEnricherByLanguage(component.Languages[0].Name)
	if componentEnricher != nil {
		componentEnricher.DoEnrichComponent(component, settings, ctx)
	}
}

/*
	getLanguagesWeightedByConfigFile returns the list of languages reordered by importance per config file.
									Language found by analyzing the config file is used as target.
	Parameters:
		languages: list of languages to be reordered
		configLanguages: languages associated to the config file found and to be used as target languages
	Returns:
		list of languages reordered
*/
func getLanguagesWeightedByConfigFile(languages []model.Language, configLanguages []string) []model.Language {
	if len(configLanguages) == 0 {
		return languages
	}

	for index, lang := range languages {
		for _, configLanguage := range configLanguages {
			if strings.EqualFold(lang.Name, configLanguage) {
				sliceWithoutLang := append(languages[:index], languages[index+1:]...)
				return append([]model.Language{lang}, sliceWithoutLang...)
			}
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
