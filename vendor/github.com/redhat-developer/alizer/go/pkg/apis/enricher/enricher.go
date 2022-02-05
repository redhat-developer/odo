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
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/language"
)

type Enricher interface {
	GetSupportedLanguages() []string
	DoEnrichLanguage(language *language.Language, files *[]string)
}

type FrameworkDetectorWithConfigFile interface {
	DoFrameworkDetection(language *language.Language, config string)
}

type FrameworkDetectorWithoutConfigFile interface {
	DoFrameworkDetection(language *language.Language, files *[]string)
}

func getEnrichers() []Enricher {
	return []Enricher{
		&JavaEnricher{},
		&JavaScriptEnricher{},
		&PythonEnricher{},
	}
}

func GetEnricherByLanguage(language *language.Language) Enricher {
	for _, enricher := range getEnrichers() {
		if isLanguageSupportedByEnricher(language.Name, enricher) {
			return enricher
		}
	}
	return nil
}

func isLanguageSupportedByEnricher(nameLanguage string, enricher Enricher) bool {
	for _, language := range enricher.GetSupportedLanguages() {
		if strings.ToLower(language) == strings.ToLower(nameLanguage) {
			return true
		}
	}
	return false
}
