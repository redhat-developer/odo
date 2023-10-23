/*******************************************************************************
 * Copyright (c) 2021 Red Hat, Inc.
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
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/devfile/alizer/pkg/apis/enricher"
	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	langfile "github.com/devfile/alizer/pkg/utils/langfiles"
)

type languageItem struct {
	item   langfile.LanguageItem
	weight int
}

func Analyze(path string) ([]model.Language, error) {
	ctx := context.Background()
	return analyze(path, &ctx)
}

func analyze(path string, ctx *context.Context) ([]model.Language, error) {
	languagesFile := langfile.Get()
	languagesDetected := make(map[string]languageItem)
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(1).Info("Searching for files in root")
	paths, err := utils.GetCachedFilePathsFromRoot(path, ctx)
	if err != nil {
		return []model.Language{}, err
	}
	alizerLogger.V(0).Info(fmt.Sprintf("Found %d cached file paths from root", len(paths)))
	alizerLogger.V(1).Info("Searching for language file extensions in given paths")
	extensionsGrouped := extractExtensions(paths)
	alizerLogger.V(0).Info(fmt.Sprintf("Found %d file extensions in given paths", len(extensionsGrouped)))
	extensionHasProgrammingLanguage := false
	totalProgrammingPoints := 0
	for extension := range extensionsGrouped {
		alizerLogger.V(1).Info(fmt.Sprintf("Checking extension %s", extension))
		languages := languagesFile.GetLanguagesByExtension(extension)
		if len(languages) == 0 {
			alizerLogger.V(1).Info(fmt.Sprintf("Not able to match %s extension with any known language", extension))
			continue
		}
		alizerLogger.V(1).Info(fmt.Sprintf("Found %d languages for extension %s", len(languages), extension))
		alizerLogger.V(1).Info(fmt.Sprintf("Accessing languages for extension %s", extension))
		for _, language := range languages {
			alizerLogger.V(1).Info(fmt.Sprintf("Accessing %s language", language.Name))
			if language.Kind == "programming" {
				var languageFileItem langfile.LanguageItem
				var err error
				if len(language.Group) == 0 {
					languageFileItem = language
				} else {
					languageFileItem, err = languagesFile.GetLanguageByName(language.Group)
					if err != nil {
						alizerLogger.V(1).Info(fmt.Sprintf("Cannot get language item for %s", language.Name))
						continue
					}
				}
				tmpLanguageItem := languageItem{languageFileItem, 0}
				alizerLogger.V(1).Info(fmt.Sprintf("Extension %s was found %d times. Adding %s to detected languages", extension, extensionsGrouped[extension], language.Name))
				weight := languagesDetected[tmpLanguageItem.item.Name].weight + extensionsGrouped[extension]
				tmpLanguageItem.weight = weight
				languagesDetected[tmpLanguageItem.item.Name] = tmpLanguageItem
				extensionHasProgrammingLanguage = true
			} else {
				alizerLogger.V(1).Info(fmt.Sprintf("%s is not a programming language", language.Name))
			}
		}
		if extensionHasProgrammingLanguage {
			totalProgrammingPoints += extensionsGrouped[extension]
			extensionHasProgrammingLanguage = false
		}
	}

	var languagesFound []model.Language
	if len(languagesDetected) > 0 {
		alizerLogger.V(0).Info(fmt.Sprintf("Accessing %d detected programming languages", len(languagesDetected)))
	} else {
		alizerLogger.V(0).Info("No programming language was detected")
	}
	for name, item := range languagesDetected {
		tmpWeight := float64(item.weight) / float64(totalProgrammingPoints)
		tmpWeight = float64(int(tmpWeight*100)) / 100
		if tmpWeight > 0.02 {
			tmpLanguage := model.Language{
				Name:           name,
				Aliases:        item.item.Aliases,
				Weight:         tmpWeight * 100,
				Frameworks:     []string{},
				Tools:          []string{},
				CanBeComponent: item.item.Component}
			langEnricher := enricher.GetEnricherByLanguage(name)
			if langEnricher != nil {
				langEnricher.DoEnrichLanguage(&tmpLanguage, &paths)
			}
			alizerLogger.V(0).Info(fmt.Sprintf("%s weight is %f. Detecting frameworks", tmpLanguage.Name, tmpLanguage.Weight))
			languagesFound = append(languagesFound, tmpLanguage)
		}
	}

	sort.SliceStable(languagesFound, func(i, j int) bool {
		return languagesFound[i].Weight > languagesFound[j].Weight
	})

	return languagesFound, nil
}

func AnalyzeFile(configFile string, targetLanguage string) (model.Language, error) {
	lang, err := langfile.Get().GetLanguageByName(targetLanguage)
	if err != nil {
		return model.Language{}, err
	}
	tmpLanguage := model.Language{
		Name:                    lang.Name,
		Aliases:                 lang.Aliases,
		Frameworks:              []string{},
		Tools:                   []string{},
		Weight:                  100,
		CanBeComponent:          lang.Component,
		CanBeContainerComponent: lang.ContainerComponent,
	}
	langEnricher := enricher.GetEnricherByLanguage(targetLanguage)
	if langEnricher != nil {
		langEnricher.DoEnrichLanguage(&tmpLanguage, &[]string{configFile})
	}
	return tmpLanguage, nil
}

func isStaticFileExtension(path string) bool {
	staticDirs := [2]string{"static/", "templates/"}
	for _, dir := range staticDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	return false
}

func extractExtensions(paths []string) map[string]int {
	extensions := make(map[string]int)
	for _, path := range paths {
		extension := filepath.Ext(path)
		if len(extension) == 0 {
			continue
		}
		extensionPoints := extensions[extension]
		if !isStaticFileExtension(path) {
			extensionPoints = extensionPoints + 100
		} else {
			extensionPoints = extensionPoints + 10
		}
		extensions[extension] = extensionPoints
	}
	return extensions
}
