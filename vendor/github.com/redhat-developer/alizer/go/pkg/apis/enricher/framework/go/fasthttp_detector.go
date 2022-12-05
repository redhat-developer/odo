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
package enricher

import (
	"os"
	"regexp"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	"golang.org/x/mod/modfile"
)

type FastHttpDetector struct{}

func (f FastHttpDetector) GetSupportedFrameworks() []string {
	return []string{"FastHttp"}
}

func (f FastHttpDetector) DoFrameworkDetection(language *model.Language, goMod *modfile.File) {
	if hasFramework(goMod.Require, "github.com/valyala/fasthttp") {
		language.Frameworks = append(language.Frameworks, "FastHttp")
	}
}

func (f FastHttpDetector) DoPortsDetection(component *model.Component) {
	files, err := utils.GetFilePathsFromRoot(component.Path)
	if err != nil {
		return
	}

	matchRegexRule := model.PortMatchRules{
		MatchIndexRegexes: []model.PortMatchRule{
			{
				Regex:     regexp.MustCompile(`.ListenAndServe\([^,)]*`),
				ToReplace: ".ListenAndServe(",
			},
		},
	}
	for _, file := range files {
		bytes, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		ports := GetPortFromFileGo(matchRegexRule, string(bytes))
		if len(ports) > 0 {
			component.Ports = ports
			return
		}
	}

}
