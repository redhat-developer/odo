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
package enricher

import (
	"regexp"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
	"golang.org/x/mod/modfile"
)

func hasFramework(modules []*modfile.Require, tag string) bool {
	for _, module := range modules {
		if strings.EqualFold(module.Mod.Path, tag) || strings.HasPrefix(module.Mod.Path, tag) {
			return true
		}
	}
	return false
}

func GetPortFromFileGo(rules model.PortMatchRules, text string) []int {
	ports := []int{}
	for _, matchIndexRegex := range rules.MatchIndexRegexes {
		matchIndexesSlice := matchIndexRegex.Regex.FindAllStringSubmatchIndex(text, -1)
		for _, matchIndexes := range matchIndexesSlice {
			if len(matchIndexes) > 1 {
				port := GetPortWithMatchIndexesGo(text, matchIndexes, matchIndexRegex.ToReplace)
				if port != -1 {
					ports = append(ports, port)
				}
			}
		}
	}

	for _, matchRegex := range rules.MatchRegexes {
		matchesSlice := matchRegex.Regex.FindAllStringSubmatch(text, -1)
		for _, matches := range matchesSlice {
			if len(matches) > 0 {
				// hostPortValue should be host:port
				hostPortValue := matches[len(matches)-1]
				if port := utils.FindPortSubmatch(matchRegex.SubRegex, hostPortValue, 1); port != -1 {
					ports = append(ports, port)
				}
			}
		}
	}

	return ports
}

func GetPortWithMatchIndexesGo(content string, matchIndexes []int, toBeReplaced string) int {
	portPlaceholder := content[matchIndexes[0]:matchIndexes[1]]
	//we should end up with something like ".ListenAndServe(PORT"
	portPlaceholder = strings.Replace(portPlaceholder, toBeReplaced, "", -1)
	// if we are lucky enough portPlaceholder contains a real HOST:PORT otherwise it is a variable/expression
	re := regexp.MustCompile(`:*(\d+)`)
	if port := utils.FindPortSubmatch(re, portPlaceholder, 1); port != -1 {
		return port
	}

	// we are not dealing with a host:port, let's try to find a variable set before the listen function
	contentBeforeMatch := content[0:matchIndexes[0]]
	re = regexp.MustCompile(portPlaceholder + `\s+[:=]+\s"([^"]*)`)
	matches := re.FindStringSubmatch(contentBeforeMatch)
	if len(matches) > 0 {
		// hostPortValue should be host:port
		hostPortValue := matches[len(matches)-1]
		re = regexp.MustCompile(`:*(\d+)$`)
		if port := utils.FindPortSubmatch(re, hostPortValue, 1); port != -1 {
			return port
		}
	}

	return -1
}
