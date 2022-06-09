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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	enricher "github.com/redhat-developer/alizer/go/pkg/apis/enricher"
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	langfile "github.com/redhat-developer/alizer/go/pkg/utils/langfiles"
)

type languageItem struct {
	item       langfile.LanguageItem
	percentage int
}

func Analyze(path string) ([]language.Language, error) {
	languagesFile := langfile.Get()
	languagesDetected := make(map[string]languageItem)

	paths, err := getFilePathsFromRoot(path)
	if err != nil {
		return []language.Language{}, err
	}
	extensionsGrouped := extractExtensions(paths)
	extensionHasProgrammingLanguage := false
	totalProgrammingOccurrences := 0
	for extension := range extensionsGrouped {
		languages := languagesFile.GetLanguagesByExtension(extension)
		if len(languages) == 0 {
			continue
		}
		for _, language := range languages {
			if language.Kind == "programming" {
				var languageFileItem langfile.LanguageItem
				var err error
				if len(language.Group) == 0 {
					languageFileItem = language
				} else {
					languageFileItem, err = languagesFile.GetLanguageByName(language.Group)
					if err != nil {
						continue
					}
				}
				tmpLanguageItem := languageItem{languageFileItem, 0}
				percentage := languagesDetected[tmpLanguageItem.item.Name].percentage + extensionsGrouped[extension]
				tmpLanguageItem.percentage = percentage
				languagesDetected[tmpLanguageItem.item.Name] = tmpLanguageItem
				extensionHasProgrammingLanguage = true
			}
		}
		if extensionHasProgrammingLanguage {
			totalProgrammingOccurrences += extensionsGrouped[extension]
			extensionHasProgrammingLanguage = false
		}
	}

	var languagesFound []language.Language
	for name, item := range languagesDetected {
		tmpPercentage := float64(item.percentage) / float64(totalProgrammingOccurrences)
		tmpPercentage = float64(int(tmpPercentage*10000)) / 10000
		if tmpPercentage > 0.02 {
			tmpLanguage := language.Language{
				Name:              name,
				Aliases:           item.item.Aliases,
				UsageInPercentage: tmpPercentage * 100,
				Frameworks:        []string{},
				Tools:             []string{},
				CanBeComponent:    item.item.Component}
			langEnricher := enricher.GetEnricherByLanguage(name)
			if langEnricher != nil {
				langEnricher.DoEnrichLanguage(&tmpLanguage, &paths)
			}
			languagesFound = append(languagesFound, tmpLanguage)
		}
	}

	sort.SliceStable(languagesFound, func(i, j int) bool {
		return languagesFound[i].UsageInPercentage > languagesFound[j].UsageInPercentage
	})

	return languagesFound, nil
}

func extractExtensions(paths []string) map[string]int {
	extensions := make(map[string]int)
	for _, path := range paths {
		extension := filepath.Ext(path)
		if len(extension) == 0 {
			continue
		}
		count := extensions[extension] + 1
		extensions[extension] = count
	}
	return extensions
}

func getFilePathsFromRoot(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && isFileInRoot(root, path) {
				files = append([]string{path}, files...)
			} else {
				files = append(files, path)
			}
			return nil
		})
	return files, err
}

func isFileInRoot(root string, file string) bool {
	dir, _ := filepath.Split(file)
	return strings.EqualFold(filepath.Clean(dir), filepath.Clean(root))
}

func getFilePathsInRoot(root string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, fileInfo := range fileInfos {
		files = append(files, filepath.Join(root, fileInfo.Name()))
	}
	return files, nil
}
