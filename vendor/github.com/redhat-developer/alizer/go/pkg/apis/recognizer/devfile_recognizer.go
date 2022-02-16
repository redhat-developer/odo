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
	"errors"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/language"
)

type DevFileType struct {
	Name        string
	Language    string
	ProjectType string
	Tags        []string
}

func SelectDevFileFromTypes(path string, devFileTypes []DevFileType) (DevFileType, error) {
	languages, err := Analyze(path)
	if err != nil {
		return DevFileType{}, err
	}
	devfile, err := SelectDevFileUsingLanguagesFromTypes(languages, devFileTypes)
	if err != nil {
		return DevFileType{}, errors.New("No valid devfile found for project in " + path)
	}
	return devfile, nil
}

func SelectDevFileUsingLanguagesFromTypes(languages []language.Language, devFileTypes []DevFileType) (DevFileType, error) {
	for _, language := range languages {
		devfile, err := selectDevFileByLanguage(language, devFileTypes)
		if err == nil {
			return devfile, nil
		}
	}
	return DevFileType{}, errors.New("no valid devfile found by using those languages")
}

func selectDevFileByLanguage(language language.Language, devFileTypes []DevFileType) (DevFileType, error) {
	scoreTarget := 0
	devfileTarget := DevFileType{}
	FRAMEWORK_WEIGHT := 10
	TOOL_WEIGHT := 5
	for _, devFile := range devFileTypes {
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
			devfileTarget = devFile
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
