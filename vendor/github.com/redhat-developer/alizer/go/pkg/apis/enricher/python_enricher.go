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
	framework "github.com/redhat-developer/alizer/go/pkg/apis/enricher/framework/python"
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
)

type PythonEnricher struct{}

func getPythonFrameworkDetectors() []FrameworkDetectorWithoutConfigFile {
	return []FrameworkDetectorWithoutConfigFile{
		&framework.DjangoDetector{},
	}
}

func (p PythonEnricher) GetSupportedLanguages() []string {
	return []string{"python"}
}

func (p PythonEnricher) DoEnrichLanguage(language *language.Language, files *[]string) {
	language.Tools = []string{}
	detectPythonFrameworks(language, files)
}

func detectPythonFrameworks(language *language.Language, files *[]string) {
	for _, detector := range getPythonFrameworkDetectors() {
		detector.DoFrameworkDetection(language, files)
	}
}
