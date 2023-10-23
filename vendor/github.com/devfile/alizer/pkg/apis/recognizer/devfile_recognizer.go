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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"github.com/hashicorp/go-version"
)

const MinimumAllowedVersion = "2.0.0"

// DEPRECATION WARNING: This function is deprecated, please use devfile_recognizer.MatchDevfiles
// instead.
// func SelectDevFilesFromTypes: Returns a list of devfiles matched for the given application
func SelectDevFilesFromTypes(path string, devfileTypes []model.DevfileType) ([]int, error) {
	alizerLogger := utils.GetOrCreateLogger()
	ctx := context.Background()
	alizerLogger.V(0).Info("Applying component detection to match a devfile")
	devfilesIndexes := selectDevfilesFromComponentsDetectedInPath(path, devfileTypes)
	if len(devfilesIndexes) > 0 {
		alizerLogger.V(0).Info(fmt.Sprintf("Found %d potential matches", len(devfilesIndexes)))
		return devfilesIndexes, nil
	}
	alizerLogger.V(0).Info("No components found, applying language analysis for devfile matching")
	languages, err := analyze(path, &ctx)
	if err != nil {
		return []int{}, err
	}
	mainLanguage, err := getMainLanguage(languages)
	if err != nil {
		return []int{}, err
	}
	devfiles, err := selectDevfilesByLanguage(mainLanguage, devfileTypes)
	if err != nil {
		return []int{}, errors.New("No valid devfile found for project in " + path)
	}
	return devfiles, nil
}

func getMainLanguage(languages []model.Language) (model.Language, error) {
	if len(languages) == 0 {
		return model.Language{}, fmt.Errorf("cannot detect main language due to empty languages list")
	}

	mainLanguage := languages[0]
	for _, language := range languages {
		if language.Weight > mainLanguage.Weight {
			mainLanguage = language
		}
	}
	return mainLanguage, nil
}

func selectDevfilesFromComponentsDetectedInPath(path string, devfileTypes []model.DevfileType) []int {
	components, _ := DetectComponentsInRoot(path)
	devfilesIndexes := selectDevfilesFromComponents(components, devfileTypes)
	if len(devfilesIndexes) > 0 {
		return devfilesIndexes
	}

	components, _ = DetectComponents(path)
	return selectDevfilesFromComponents(components, devfileTypes)
}

func selectDevfilesFromComponents(components []model.Component, devfileTypes []model.DevfileType) []int {
	var devfilesIndexes []int
	for _, component := range components {
		devfiles, err := selectDevfilesByLanguage(component.Languages[0], devfileTypes)
		if err == nil {
			devfilesIndexes = append(devfilesIndexes, devfiles...)
		}
	}
	return devfilesIndexes
}

// DEPRECATION WARNING: This function is deprecated, please use devfile_recognizer.MatchDevfiles
// instead.
// func SelectDevFileFromTypes: Returns the first devfile from the list of devfiles returned
// from SelectDevFilesFromTypes func. It also returns an error if exists.
func SelectDevFileFromTypes(path string, devfileTypes []model.DevfileType) (int, error) {
	devfiles, err := SelectDevFilesFromTypes(path, devfileTypes)
	if err != nil {
		return -1, err
	}
	return devfiles[0], nil
}

func SelectDevfilesUsingLanguagesFromTypes(languages []model.Language, devfileTypes []model.DevfileType) ([]int, error) {
	var devfilesIndexes []int
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(1).Info("Searching potential matches from detected languages")
	for _, language := range languages {
		alizerLogger.V(1).Info(fmt.Sprintf("Accessing %s language", language.Name))
		devfiles, err := selectDevfilesByLanguage(language, devfileTypes)
		if err == nil {
			alizerLogger.V(1).Info(fmt.Sprintf("Found %d potential matches for language %s", len(devfiles), language.Name))
			devfilesIndexes = append(devfilesIndexes, devfiles...)
		}
	}
	if len(devfilesIndexes) > 0 {
		return devfilesIndexes, nil
	}
	return []int{}, errors.New("no valid devfile found by using those languages")
}

func SelectDevfileUsingLanguagesFromTypes(languages []model.Language, devfileTypes []model.DevfileType) (int, error) {
	devfilesIndexes, err := SelectDevfilesUsingLanguagesFromTypes(languages, devfileTypes)
	if err != nil {
		return -1, err
	}
	return devfilesIndexes[0], nil
}

func MatchDevfiles(path string, url string, filter model.DevfileFilter) ([]model.DevfileType, error) {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info("Starting devfile matching")
	alizerLogger.V(1).Info(fmt.Sprintf("Downloading devfiles from registry %s", url))
	devfileTypesFromRegistry, err := DownloadDevfileTypesFromRegistry(url, filter)
	if err != nil {
		return []model.DevfileType{}, err
	}

	return selectDevfiles(path, devfileTypesFromRegistry)
}

func SelectDevfilesFromRegistry(path string, url string) ([]model.DevfileType, error) {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info("Starting devfile matching")
	alizerLogger.V(1).Info(fmt.Sprintf("Downloading devfiles from registry %s", url))
	devfileTypesFromRegistry, err := DownloadDevfileTypesFromRegistry(url, model.DevfileFilter{MinSchemaVersion: "", MaxSchemaVersion: ""})
	if err != nil {
		return []model.DevfileType{}, err
	}

	return selectDevfiles(path, devfileTypesFromRegistry)
}

// selectDevfiles is exposed as global var in the purpose of mocking tests
var selectDevfiles = func(path string, devfileTypesFromRegistry []model.DevfileType) ([]model.DevfileType, error) {
	indexes, err := SelectDevFilesFromTypes(path, devfileTypesFromRegistry)
	if err != nil {
		return []model.DevfileType{}, err
	}

	var devfileTypes []model.DevfileType
	for _, index := range indexes {
		devfileTypes = append(devfileTypes, devfileTypesFromRegistry[index])
	}

	return devfileTypes, nil

}

func SelectDevfileFromRegistry(path string, url string) (model.DevfileType, error) {
	devfileTypes, err := DownloadDevfileTypesFromRegistry(url, model.DevfileFilter{MinSchemaVersion: "", MaxSchemaVersion: ""})
	if err != nil {
		return model.DevfileType{}, err
	}

	index, err := SelectDevFileFromTypes(path, devfileTypes)
	if err != nil {
		return model.DevfileType{}, err
	}
	return devfileTypes[index], nil
}

func GetUrlWithVersions(url, minSchemaVersion, maxSchemaVersion string) (string, error) {
	minAllowedVersion, err := version.NewVersion(MinimumAllowedVersion)
	if err != nil {
		return "", nil
	}

	if minSchemaVersion != "" && maxSchemaVersion != "" {
		minV, err := version.NewVersion(minSchemaVersion)
		if err != nil {
			return url, nil
		}
		maxV, err := version.NewVersion(maxSchemaVersion)
		if err != nil {
			return url, nil
		}
		if maxV.LessThan(minV) {
			return "", fmt.Errorf("max-schema-version cannot be lower than min-schema-version")
		}
		if maxV.LessThan(minAllowedVersion) || minV.LessThan(minAllowedVersion) {
			return "", fmt.Errorf("min and/or max version are lower than the minimum allowed version (2.0.0)")
		}

		return fmt.Sprintf("%s?minSchemaVersion=%s&maxSchemaVersion=%s", url, minSchemaVersion, maxSchemaVersion), nil
	} else if minSchemaVersion != "" {
		minV, err := version.NewVersion(minSchemaVersion)
		if err != nil {
			return "", nil
		}
		if minV.LessThan(minAllowedVersion) {
			return "", fmt.Errorf("min version is lower than the minimum allowed version (2.0.0)")
		}
		return fmt.Sprintf("%s?minSchemaVersion=%s", url, minSchemaVersion), nil
	} else if maxSchemaVersion != "" {
		maxV, err := version.NewVersion(maxSchemaVersion)
		if err != nil {
			return "", nil
		}
		if maxV.LessThan(minAllowedVersion) {
			return "", fmt.Errorf("max version is lower than the minimum allowed version (2.0.0)")
		}
		return fmt.Sprintf("%s?maxSchemaVersion=%s", url, maxSchemaVersion), nil
	} else {
		return url, nil
	}
}

// DownloadDevfileTypesFromRegistry is exposed as a global variable for the purpose of running mock tests
var DownloadDevfileTypesFromRegistry = func(url string, filter model.DevfileFilter) ([]model.DevfileType, error) {
	url = adaptUrl(url)
	tmpUrl := appendIndexPath(url)
	url, err := GetUrlWithVersions(tmpUrl, filter.MinSchemaVersion, filter.MaxSchemaVersion)
	if err != nil {
		return nil, err
	}
	// This value is set by the user in order to configure the registry
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return []model.DevfileType{}, err
	}
	defer func() error {
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
		return nil
	}()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return []model.DevfileType{}, errors.New("unable to fetch devfiles from the registry")
	}

	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		return []model.DevfileType{}, errors.New("unable to fetch devfiles from the registry")
	}

	var devfileTypes []model.DevfileType
	err = json.Unmarshal(body, &devfileTypes)
	if err != nil {
		return []model.DevfileType{}, errors.New("unable to fetch devfiles from the registry")
	}

	return devfileTypes, nil
}

func appendIndexPath(url string) string {
	if strings.HasSuffix(url, "/") {
		return url + "v2index"
	}
	return url + "/v2index"
}

func adaptUrl(url string) string {
	re := regexp.MustCompile(`^https://github.com/(.*)/blob/(.*)$`)
	url = re.ReplaceAllString(url, `https://raw.githubusercontent.com/$1/$2`)

	re = regexp.MustCompile(`^https://gitlab.com/(.*)/-/blob/(.*)$`)
	url = re.ReplaceAllString(url, `https://gitlab.com/$1/-/raw/$2`)

	re = regexp.MustCompile(`^https://bitbucket.org/(.*)/src/(.*)$`)
	url = re.ReplaceAllString(url, `https://bitbucket.org/$1/raw/$2`)

	return url
}

// selectDevfilesByLanguage detects devfiles that fit best with a project.
//
// Performs a search in two steps looping through all devfiles available.
// When a framework is detected, this is stored in a map but still not saved. A check is made eventually as there could be that a future or
// previous devfile is more appropriate based on other infos (e.g. the tool -> quarkus gradle vs quarkus maven).
// If no framework is detected, the devfiles are picked based on their score (language +1, tool +5). The largest score wins.
//
// At the end, if some framework is supported by some devfile, they are returned. Otherwise, Alizer was not able to find any
// specific devfile for the frameworks detected and returned the devfiles which got the largest score.
func selectDevfilesByLanguage(language model.Language, devfileTypes []model.DevfileType) ([]int, error) {
	var devfileIndexes []int
	frameworkPerDevfile := make(map[string]model.DevfileScore)
	scoreTarget := 0

	for index, devfile := range devfileTypes {
		score := 0
		frameworkPerDevfileTmp := make(map[string]interface{})
		if strings.EqualFold(devfile.Language, language.Name) || matches(language.Aliases, devfile.Language) != "" {
			score++
			if frw := matchesFormatted(language.Frameworks, devfile.ProjectType); frw != "" {
				frameworkPerDevfileTmp[frw] = nil
				score += utils.FRAMEWORK_WEIGHT
			}
			for _, tag := range devfile.Tags {
				if frw := matchesFormatted(language.Frameworks, tag); frw != "" {
					frameworkPerDevfileTmp[frw] = nil
					score += utils.FRAMEWORK_WEIGHT
				}
				if matches(language.Tools, tag) != "" {
					score += utils.TOOL_WEIGHT
				}
			}

			for framework := range frameworkPerDevfileTmp {
				devfileObj := frameworkPerDevfile[framework]
				if score > devfileObj.Score {
					frameworkPerDevfile[framework] = model.DevfileScore{
						DevfileIndex: index,
						Score:        score,
					}
				}
			}

			if len(frameworkPerDevfile) == 0 {
				if score == scoreTarget {
					devfileIndexes = append(devfileIndexes, index)
				} else if score > scoreTarget {
					scoreTarget = score
					devfileIndexes = []int{index}
				}
			}
		}
	}

	if len(frameworkPerDevfile) > 0 {
		devfileIndexes = []int{}
		for _, val := range frameworkPerDevfile {
			devfileIndexes = append(devfileIndexes, val.DevfileIndex)
		}
	}

	if len(devfileIndexes) == 0 {
		return devfileIndexes, errors.New("No valid devfile found for current language " + language.Name)
	}
	return devfileIndexes, nil
}

func matchesFormatted(values []string, valueToFind string) string {
	for _, value := range values {
		if strings.EqualFold(trim(value), trim(valueToFind)) {
			return value
		}
	}
	return ""
}

func trim(value string) string {
	formattedValueNoSpaces := strings.ReplaceAll(value, " ", "")
	return strings.ReplaceAll(formattedValueNoSpaces, ".", "")
}

func matches(values []string, valueToFind string) string {
	for _, value := range values {
		if strings.EqualFold(value, valueToFind) {
			return value
		}
	}
	return ""
}
