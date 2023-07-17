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

// Package recognizer implements functions that are used by every cobra cli command.
// Uses the enricher and model packages to return a result.
package recognizer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devfile/alizer/pkg/apis/enricher"
	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"github.com/devfile/alizer/pkg/utils/langfiles"
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
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info("Starting component with settings detection")
	alizerLogger.V(0).Info("Getting cached filepaths from root")
	files, err := utils.GetCachedFilePathsFromRoot(settings.BasePath, ctx)
	if err != nil {
		alizerLogger.V(0).Info("Not able to get cached file paths from root: exiting")
		return []model.Component{}, err
	}
	components := DetectComponentsFromFilesList(files, settings, ctx)

	// it may happen that a language has no a specific configuration file (e.g opposite to JAVA -> pom.xml and Nodejs -> package.json)
	// we then rely on the language recognizer
	alizerLogger.V(0).Info("Checking for components without configuration file")
	directoriesNotBelongingToExistingComponent := getDirectoriesWithoutConfigFile(settings.BasePath, components)
	components = append(components, getComponentsWithoutConfigFile(directoriesNotBelongingToExistingComponent, settings, ctx)...)

	return components, nil
}

// getComponentsWithoutConfigFile retrieves the components which are written with a language that does not require a config file.
// Uses the settings to perform detection on the list of directories to analyze.
func getComponentsWithoutConfigFile(directories []string, settings model.DetectionSettings, ctx *context.Context) []model.Component {
	alizerLogger := utils.GetOrCreateLogger()
	var components []model.Component
	for _, dir := range directories {
		alizerLogger.V(1).Info(fmt.Sprintf("Accessing %s dir", dir))
		component, _ := detectComponentByFolderAnalysis(dir, []string{}, settings, ctx)
		if component.Path != "" && isLangForNoConfigComponent(component.Languages[0]) {
			alizerLogger.V(1).Info(fmt.Sprintf("Component %s found for %s dir", component.Name, dir))
			components = append(components, component)
		} else {
			alizerLogger.V(1).Info(fmt.Sprintf("No component found for %s dir", dir))
		}
	}
	alizerLogger.V(0).Info(fmt.Sprintf("Found %d components without configuration file", len(components)))
	return components
}

// isLangForNoConfigComponent verifies if main language requires any config file.
// Returns true if language does not require any config file.
func isLangForNoConfigComponent(language model.Language) bool {
	lang, err := langfiles.Get().GetLanguageByNameOrAlias(language.Name)
	if err != nil {
		return false
	}

	return len(lang.ConfigurationFiles) == 0
}

// getDirectoriesPathsWithoutConfigFile retrieves all directories that do not contain any Component.
// Search starts from the root and returns a list of directory paths that do not contain any component.
func getDirectoriesWithoutConfigFile(root string, components []model.Component) []string {
	if len(components) == 0 {
		return []string{root}
	}
	var directories []string
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

// getParentFolders return all paths which are not sub-folders of some other path within the list.
// target will be added to the list if it is not a sub-folder of any other path within the list.
// If a path in the list is sub-folder of target, that path will be removed.
func getParentFolders(target string, directories []string) []string {
	var updatedDirectories []string
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

// isAnyComponentInPath checks if a component is present in path.
// Search starts from path and will return true if a component is found.
func isAnyComponentInPath(path string, components []model.Component) bool {
	for _, component := range components {
		if strings.EqualFold(path, component.Path) || isFirstPathParentOfSecond(component.Path, path) || isFirstPathParentOfSecond(path, component.Path) {
			return true
		}
	}
	return false
}

// isFirstPathParentOfSecond check if first path is parent (direct or not) of second path.
func isFirstPathParentOfSecond(firstPath string, secondPath string) bool {
	return strings.Contains(secondPath, firstPath)
}

// DetectComponentsFromFilesList detect components by analyzing all files.
// Uses the settings to perform component detection on files.
func DetectComponentsFromFilesList(files []string, settings model.DetectionSettings, ctx *context.Context) []model.Component {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info(fmt.Sprintf("Detecting components for %d fetched file paths", len(files)))
	configurationPerLanguage := langfiles.Get().GetConfigurationPerLanguageMapping()
	var components []model.Component
	for _, file := range files {
		alizerLogger.V(1).Info(fmt.Sprintf("Accessing %s", file))
		languages, err := getLanguagesByConfigurationFile(configurationPerLanguage, file)

		if err != nil {
			alizerLogger.V(1).Info(err.Error())
			continue
		}

		alizerLogger.V(0).Info(fmt.Sprintf("File %s detected as configuration file for %d languages", file, len(languages)))
		alizerLogger.V(1).Info("Searching for components based on this configuration file")
		component, err := detectComponentUsingConfigFile(file, languages, settings, ctx)
		if err != nil {
			alizerLogger.V(1).Info(err.Error())
			continue
		}
		alizerLogger.V(0).Info(fmt.Sprintf("Component %s found", component.Name))
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
	return nil, errors.New("no languages found for file " + file)
}

// detectComponentByFolderAnalysis returns a Component if found.
// Using settings, detection starts from root and uses configLanguages as a target.
func detectComponentByFolderAnalysis(root string, configLanguages []string, settings model.DetectionSettings, ctx *context.Context) (model.Component, error) {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info("Detecting component by folder language analysis")
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
	alizerLogger.V(0).Info("No component detected")
	return model.Component{}, errors.New("no component detected")

}

// detectComponentByAnalyzingConfigFile returns a Component if found.
func detectComponentByAnalyzingConfigFile(file string, language string, settings model.DetectionSettings, ctx *context.Context) (model.Component, error) {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(1).Info("Analyzing config file for singe language or family of languages")
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
	return len(languages) == 2 &&
		languages[0] != languages[1] &&
		(strings.ToLower(languages[0]) == "typescript" || strings.ToLower(languages[0]) == "javascript") &&
		(strings.ToLower(languages[1]) == "typescript" || strings.ToLower(languages[1]) == "javascript")
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

// getLanguagesWeightedByConfigFile returns the list of languages reordered by importance per config file.
// Language found by analyzing the config file is used as target.
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
