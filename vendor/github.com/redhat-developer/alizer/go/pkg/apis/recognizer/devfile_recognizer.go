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
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
)

func SelectDevFileFromTypes(path string, devFileTypes []model.DevFileType) (int, error) {
	components, _ := DetectComponentsInRoot(path)
	if len(components) > 0 {
		devfile, err := selectDevFileByLanguage(components[0].Languages[0], devFileTypes)
		if err == nil {
			return devfile, nil
		}
	}

	components, _ = DetectComponents(path)
	if len(components) > 0 {
		devfile, err := selectDevFileByLanguage(components[0].Languages[0], devFileTypes)
		if err == nil {
			return devfile, nil
		}
	}

	languages, err := Analyze(path)
	if err != nil {
		return -1, err
	}
	devfile, err := SelectDevFileUsingLanguagesFromTypes(languages, devFileTypes)
	if err != nil {
		return -1, errors.New("No valid devfile found for project in " + path)
	}
	return devfile, nil
}

func SelectDevFileUsingLanguagesFromTypes(languages []model.Language, devFileTypes []model.DevFileType) (int, error) {
	for _, language := range languages {
		devfile, err := selectDevFileByLanguage(language, devFileTypes)
		if err == nil {
			return devfile, nil
		}
	}
	return -1, errors.New("no valid devfile found by using those languages")
}

func SelectDevFileFromRegistry(path string, url string) (model.DevFileType, error) {
	devFileTypes, err := downloadDevFileTypesFromRegistry(url)
	if err != nil {
		return model.DevFileType{}, err
	}

	index, err := SelectDevFileFromTypes(path, devFileTypes)
	if err != nil {
		return model.DevFileType{}, err
	}
	return devFileTypes[index], nil
}

func downloadDevFileTypesFromRegistry(url string) ([]model.DevFileType, error) {
	url = adaptUrl(url)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		// retry by appending index to url
		url = appendIndexPath(url)
		resp, err = http.Get(url)
		if err != nil {
			return []model.DevFileType{}, err
		}
	}
	defer resp.Body.Close()

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
		return url + "index"
	}
	return url + "/index"
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

func selectDevFileByLanguage(language model.Language, devFileTypes []model.DevFileType) (int, error) {
	scoreTarget := 0
	devfileTarget := -1
	FRAMEWORK_WEIGHT := 10
	TOOL_WEIGHT := 5
	for index, devFile := range devFileTypes {
		score := 0
		if strings.EqualFold(devFile.Language, language.Name) || matches(language.Aliases, devFile.Language) {
			score++
			if matches(language.Frameworks, devFile.ProjectType) {
				score += FRAMEWORK_WEIGHT
			}
			for _, tag := range devFile.Tags {
				if matches(language.Frameworks, tag) {
					score += FRAMEWORK_WEIGHT
				}
				if matches(language.Tools, tag) {
					score += TOOL_WEIGHT
				}
			}
		}
		if score > scoreTarget {
			scoreTarget = score
			devfileTarget = index
		}
	}

	if scoreTarget == 0 {
		return devfileTarget, errors.New("No valid devfile found for current language " + language.Name)
	}
	return devfileTarget, nil
}

func matches(values []string, valueToFind string) bool {
	for _, value := range values {
		if strings.EqualFold(value, valueToFind) {
			return true
		}
	}
	return false
}
