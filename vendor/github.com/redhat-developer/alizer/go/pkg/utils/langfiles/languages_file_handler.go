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
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/schema"
	"gopkg.in/yaml.v3"
)

type LanguageItem struct {
	Name               string
	Aliases            []string
	Kind               string
	Group              string
	ConfigurationFiles []string
	ExcludeFolders     []string
	Component          bool
	disabled           bool
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
		customizeLanguage(&languageItem)
		if !languageItem.disabled {
			languages[name] = languageItem
			extensions := properties.Extensions
			for _, ext := range extensions {
				languagesByExtension := extensionsXLanguage[ext]
				languagesByExtension = append(languagesByExtension, languageItem)
				extensionsXLanguage[ext] = languagesByExtension
			}
		}
	}

	return &LanguageFile{
		languages:           languages,
		extensionsXLanguage: extensionsXLanguage,
	}
}

func customizeLanguage(languageItem *LanguageItem) {
	languagesCustomizations := getLanguageCustomizations()
	if customization, hasCustomization := languagesCustomizations[(*languageItem).Name]; hasCustomization {
		(*languageItem).ConfigurationFiles = customization.ConfigurationFiles
		(*languageItem).ExcludeFolders = customization.ExcludeFolders
		(*languageItem).Component = customization.Component
		(*languageItem).Aliases = appendSlice((*languageItem).Aliases, customization.Aliases)
		(*languageItem).disabled = customization.Disabled
	}
}

func appendSlice(values []string, toBeAdded []string) []string {
	for _, item := range toBeAdded {
		values = appendIfMissing(values, item)
	}
	return values
}

func appendIfMissing(values []string, item string) []string {
	for _, value := range values {
		if strings.EqualFold(value, item) {
			return values
		}
	}
	return append(values, item)
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

func getLanguageCustomizations() schema.LanguagesCustomizations {
	yamlFile, err := res.ReadFile("resources/languages-customization.yml")
	if err != nil {
		return schema.LanguagesCustomizations{}
	}

	var data schema.LanguagesCustomizations
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

func (l *LanguageFile) GetLanguageByAlias(alias string) (LanguageItem, error) {
	for _, langItem := range l.languages {
		for _, aliasItem := range langItem.Aliases {
			if strings.EqualFold(alias, aliasItem) {
				return langItem, nil
			}
		}
	}
	return LanguageItem{}, errors.New("no language found with this alias")
}

func (l *LanguageFile) GetLanguageByNameOrAlias(name string) (LanguageItem, error) {
	langItem, err := l.GetLanguageByName(name)
	if err == nil {
		return langItem, nil
	}

	return l.GetLanguageByAlias(name)
}

func (l *LanguageFile) GetConfigurationPerLanguageMapping() map[string][]string {
	configurationPerLanguage := make(map[string][]string)
	for langName, langItem := range l.languages {
		configurationFiles := langItem.ConfigurationFiles
		for _, configFile := range configurationFiles {
			configurationPerLanguage[configFile] = append(configurationPerLanguage[configFile], langName)
		}
	}
	return configurationPerLanguage
}
