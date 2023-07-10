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
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
	"github.com/hashicorp/go-version"
)

const MinimumAllowedVersion = "2.0.0"

func SelectDevFilesFromTypes(path string, devFileTypes []model.DevFileType) ([]int, error) {
	alizerLogger := utils.GetOrCreateLogger()
	ctx := context.Background()
	alizerLogger.V(0).Info("Applying component detection to match a devfile")
	devFilesIndexes := selectDevFilesFromComponentsDetectedInPath(path, devFileTypes)
	if len(devFilesIndexes) > 0 {
		alizerLogger.V(0).Info(fmt.Sprintf("Found %d potential matches", len(devFilesIndexes)))
		return devFilesIndexes, nil
	}
	alizerLogger.V(0).Info("No components found, applying language analysis for devfile matching")
	languages, err := analyze(path, &ctx)
	if err != nil {
		return []int{}, err
	}
	devfile, err := SelectDevFileUsingLanguagesFromTypes(languages, devFileTypes)
	if err != nil {
		return []int{}, errors.New("No valid devfile found for project in " + path)
	}
	return []int{devfile}, nil
}

func selectDevFilesFromComponentsDetectedInPath(path string, devFileTypes []model.DevFileType) []int {
	components, _ := DetectComponentsInRoot(path)
	devFilesIndexes := selectDevFilesFromComponents(components, devFileTypes)
	if len(devFilesIndexes) > 0 {
		return devFilesIndexes
	}

	components, _ = DetectComponents(path)
	return selectDevFilesFromComponents(components, devFileTypes)
}

func selectDevFilesFromComponents(components []model.Component, devFileTypes []model.DevFileType) []int {
	var devFilesIndexes []int
	for _, component := range components {
		devFiles, err := selectDevFilesByLanguage(component.Languages[0], devFileTypes)
		if err == nil {
			devFilesIndexes = append(devFilesIndexes, devFiles...)
		}
	}
	return devFilesIndexes
}

func SelectDevFileFromTypes(path string, devFileTypes []model.DevFileType) (int, error) {
	devfiles, err := SelectDevFilesFromTypes(path, devFileTypes)
	if err != nil {
		return -1, err
	}
	return devfiles[0], nil
}

func SelectDevFilesUsingLanguagesFromTypes(languages []model.Language, devFileTypes []model.DevFileType) ([]int, error) {
	var devFilesIndexes []int
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(1).Info("Searching potential matches from detected languages")
	for _, language := range languages {
		alizerLogger.V(1).Info(fmt.Sprintf("Accessing %s language", language.Name))
		devFiles, err := selectDevFilesByLanguage(language, devFileTypes)
		if err == nil {
			alizerLogger.V(1).Info(fmt.Sprintf("Found %d potential matches for language %s", len(devFiles), language.Name))
			devFilesIndexes = append(devFilesIndexes, devFiles...)
		}
	}
	if len(devFilesIndexes) > 0 {
		return devFilesIndexes, nil
	}
	return []int{}, errors.New("no valid devfile found by using those languages")
}

func SelectDevFileUsingLanguagesFromTypes(languages []model.Language, devFileTypes []model.DevFileType) (int, error) {
	devFilesIndexes, err := SelectDevFilesUsingLanguagesFromTypes(languages, devFileTypes)
	if err != nil {
		return -1, err
	}
	return devFilesIndexes[0], nil
}

func MatchDevfiles(path string, url string, filter model.DevfileFilter) ([]model.DevFileType, error) {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info("Starting devfile matching")
	alizerLogger.V(1).Info(fmt.Sprintf("Downloading devfiles from registry %s", url))
	devFileTypesFromRegistry, err := downloadDevFileTypesFromRegistry(url, filter)
	if err != nil {
		return []model.DevFileType{}, err
	}

	return selectDevfiles(path, devFileTypesFromRegistry)
}

func SelectDevFilesFromRegistry(path string, url string) ([]model.DevFileType, error) {
	alizerLogger := utils.GetOrCreateLogger()
	alizerLogger.V(0).Info("Starting devfile matching")
	alizerLogger.V(1).Info(fmt.Sprintf("Downloading devfiles from registry %s", url))
	devFileTypesFromRegistry, err := downloadDevFileTypesFromRegistry(url, model.DevfileFilter{MinVersion: "", MaxVersion: ""})
	if err != nil {
		return []model.DevFileType{}, err
	}

	return selectDevfiles(path, devFileTypesFromRegistry)
}

func selectDevfiles(path string, devFileTypesFromRegistry []model.DevFileType) ([]model.DevFileType, error) {
	indexes, err := SelectDevFilesFromTypes(path, devFileTypesFromRegistry)
	if err != nil {
		return []model.DevFileType{}, err
	}

	var devFileTypes []model.DevFileType
	for _, index := range indexes {
		devFileTypes = append(devFileTypes, devFileTypesFromRegistry[index])
	}

	return devFileTypes, nil

}

func SelectDevFileFromRegistry(path string, url string) (model.DevFileType, error) {
	devFileTypes, err := downloadDevFileTypesFromRegistry(url, model.DevfileFilter{MinVersion: "", MaxVersion: ""})
	if err != nil {
		return model.DevFileType{}, err
	}

	index, err := SelectDevFileFromTypes(path, devFileTypes)
	if err != nil {
		return model.DevFileType{}, err
	}
	return devFileTypes[index], nil
}

func GetUrlWithVersions(url, minVersion, maxVersion string) (string, error) {
	minAllowedVersion, err := version.NewVersion(MinimumAllowedVersion)
	if err != nil {
		return "", nil
	}

	if minVersion != "" && maxVersion != "" {
		minV, err := version.NewVersion(minVersion)
		if err != nil {
			return url, nil
		}
		maxV, err := version.NewVersion(maxVersion)
		if err != nil {
			return url, nil
		}
		if maxV.LessThan(minV) {
			return "", fmt.Errorf("max-version cannot be lower than min-version")
		}
		if maxV.LessThan(minAllowedVersion) || minV.LessThan(minAllowedVersion) {
			return "", fmt.Errorf("min and/or max version are lower than the minimum allowed version (2.0.0)")
		}

		return fmt.Sprintf("%s?minSchemaVersion=%s&maxSchemaVersion=%s", url, minVersion, maxVersion), nil
	} else if minVersion != "" {
		minV, err := version.NewVersion(minVersion)
		if err != nil {
			return "", nil
		}
		if minV.LessThan(minAllowedVersion) {
			return "", fmt.Errorf("min version is lower than the minimum allowed version (2.0.0)")
		}
		return fmt.Sprintf("%s?minSchemaVersion=%s", url, minVersion), nil
	} else if maxVersion != "" {
		maxV, err := version.NewVersion(maxVersion)
		if err != nil {
			return "", nil
		}
		if maxV.LessThan(minAllowedVersion) {
			return "", fmt.Errorf("max version is lower than the minimum allowed version (2.0.0)")
		}
		return fmt.Sprintf("%s?maxSchemaVersion=%s", url, maxVersion), nil
	} else {
		return url, nil
	}
}

func downloadDevFileTypesFromRegistry(url string, filter model.DevfileFilter) ([]model.DevFileType, error) {
	url = adaptUrl(url)
	tmpUrl := appendIndexPath(url)
	url, err := GetUrlWithVersions(tmpUrl, filter.MinVersion, filter.MaxVersion)
	if err != nil {
		return nil, err
	}
	// This value is set by the user in order to configure the registry
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return []model.DevFileType{}, err
	}
	defer func() error {
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
		return nil
	}()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return []model.DevFileType{}, errors.New("unable to fetch devfiles from the registry")
	}

	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return []model.DevFileType{}, errors.New("unable to fetch devfiles from the registry")
	}

	var devFileTypes []model.DevFileType
	err = json.Unmarshal(body, &devFileTypes)
	if err != nil {
		return []model.DevFileType{}, errors.New("unable to fetch devfiles from the registry")
	}

	return devFileTypes, nil
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

// selectDevFilesByLanguage detects devfiles that fit best with a project.
//
// Performs a search in two steps looping through all devfiles available.
// When a framework is detected, this is stored in a map but still not saved. A check is made eventually as there could be that a future or
// previous devfile is more appropriate based on other infos (e.g. the tool -> quarkus gradle vs quarkus maven).
// If no framework is detected, the devfiles are picked based on their score (language +1, tool +5). The largest score wins.
//
// At the end, if some framework is supported by some devfile, they are returned. Otherwise, Alizer was not able to find any
// specific devfile for the frameworks detected and returned the devfiles which got the largest score.
func selectDevFilesByLanguage(language model.Language, devFileTypes []model.DevFileType) ([]int, error) {
	var devFileIndexes []int
	frameworkPerDevFile := make(map[string]model.DevFileScore)
	scoreTarget := 0

	for index, devFile := range devFileTypes {
		score := 0
		frameworkPerDevfileTmp := make(map[string]interface{})
		if strings.EqualFold(devFile.Language, language.Name) || matches(language.Aliases, devFile.Language) != "" {
			score++
			if frw := matchesFormatted(language.Frameworks, devFile.ProjectType); frw != "" {
				frameworkPerDevfileTmp[frw] = nil
				score += utils.FRAMEWORK_WEIGHT
			}
			for _, tag := range devFile.Tags {
				if frw := matchesFormatted(language.Frameworks, tag); frw != "" {
					frameworkPerDevfileTmp[frw] = nil
					score += utils.FRAMEWORK_WEIGHT
				}
				if matches(language.Tools, tag) != "" {
					score += utils.TOOL_WEIGHT
				}
			}

			for framework := range frameworkPerDevfileTmp {
				devFileObj := frameworkPerDevFile[framework]
				if score > devFileObj.Score {
					frameworkPerDevFile[framework] = model.DevFileScore{
						DevFileIndex: index,
						Score:        score,
					}
				}
			}

			if len(frameworkPerDevFile) == 0 {
				if score == scoreTarget {
					devFileIndexes = append(devFileIndexes, index)
				} else if score > scoreTarget {
					scoreTarget = score
					devFileIndexes = []int{index}
				}
			}
		}
	}

	if len(frameworkPerDevFile) > 0 {
		devFileIndexes = []int{}
		for _, val := range frameworkPerDevFile {
			devFileIndexes = append(devFileIndexes, val.DevFileIndex)
		}
	}

	if len(devFileIndexes) == 0 {
		return devFileIndexes, errors.New("No valid devfile found for current language " + language.Name)
	}
	return devFileIndexes, nil
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
