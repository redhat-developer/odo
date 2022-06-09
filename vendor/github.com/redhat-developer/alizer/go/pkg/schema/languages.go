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
package schema

type LanguageProperties struct {
	Type               string   `yaml:"type,omitempty"`
	Color              string   `yaml:"color,omitempty"`
	Extensions         []string `yaml:"extensions,omitempty"`
	TmScope            string   `yaml:"tm_scope,omitempty"`
	AceMode            string   `yaml:"ace_mode,omitempty"`
	LanguageID         int      `yaml:"language_id,omitempty"`
	Aliases            []string `yaml:"aliases,omitempty"`
	CodemirrorMode     string   `yaml:"codemirror_mode,omitempty"`
	CodemirrorMimeType string   `yaml:"codemirror_mime_type,omitempty"`
	Group              string   `yaml:"group"`
	Filenames          []string `yaml:"filenames"`
}

type LanguagesProperties map[string]LanguageProperties

type LanguageCustomization struct {
	ConfigurationFiles []string `yaml:"configuration_files"`
	Component          bool     `yaml:"component"`
	ExcludeFolders     []string `yaml:"exclude_folders,omitempty"`
	Aliases            []string `yaml:"aliases"`
}

type LanguagesCustomizations map[string]LanguageCustomization
