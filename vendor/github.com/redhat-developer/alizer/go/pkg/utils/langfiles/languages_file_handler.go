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
package langfiles

import (
	"embed"
	"errors"

	"github.com/redhat-developer/alizer/go/pkg/schema"
	"gopkg.in/yaml.v3"
)

type LanguageItem struct {
	Name    string
	Aliases []string
	Kind    string
	Group   string
}

type LanguageFile struct {
	languages           map[string]LanguageItem
	extensionsXLanguage map[string][]LanguageItem
}

var (
	instance *LanguageFile

	//go:embed resources
	res embed.FS
)

func Get() *LanguageFile {
	if instance == nil {
		instance = create()
	}
	return instance
}

func create() *LanguageFile {
	languages := make(map[string]LanguageItem)
	extensionsXLanguage := make(map[string][]LanguageItem)

	languagesProperties := getLanguagesProperties()

	for name, properties := range languagesProperties {
		languageItem := LanguageItem{
			Name:    name,
			Aliases: properties.Aliases,
			Kind:    properties.Type,
			Group:   properties.Group,
		}
		languages[name] = languageItem
		extensions := properties.Extensions
		for _, ext := range extensions {
			languagesByExtension := extensionsXLanguage[ext]
			languagesByExtension = append(languagesByExtension, languageItem)
			extensionsXLanguage[ext] = languagesByExtension
		}
	}

	return &LanguageFile{
		languages:           languages,
		extensionsXLanguage: extensionsXLanguage,
	}
}

func getLanguagesProperties() schema.LanguagesProperties {
	yamlFile, err := res.ReadFile("resources/languages.yml")
	if err != nil {
		return schema.LanguagesProperties{}
	}
	var data schema.LanguagesProperties
	yaml.Unmarshal(yamlFile, &data)
	return data
}

func (l *LanguageFile) GetLanguagesByExtension(extension string) []LanguageItem {
	return l.extensionsXLanguage[extension]
}

func (l *LanguageFile) GetLanguageByName(name string) (LanguageItem, error) {
	for langName, langItem := range l.languages {
		if langName == name {
			return langItem, nil
		}
	}
	return LanguageItem{}, errors.New("no language found with this name")
}
