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
	"path/filepath"
	"regexp"

	"github.com/redhat-developer/alizer/go/pkg/schema"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
)

type packageScriptFunc func(schema.PackageJson) string

func hasFramework(configFile string, tag string) bool {
	return utils.IsTagInPackageJsonFile(configFile, tag)
}

func getPortFromStartScript(root string, regexes []string) int {
	getStartScript := func(packageJson schema.PackageJson) string {
		return packageJson.Scripts.Start
	}
	return getPortFromScript(root, getStartScript, regexes)
}

func getPortFromDevScript(root string, regexes []string) int {
	getDevScript := func(packageJson schema.PackageJson) string {
		return packageJson.Scripts.Dev
	}
	return getPortFromScript(root, getDevScript, regexes)
}

func getPortFromScript(root string, getScript packageScriptFunc, regexes []string) int {
	packageJson, err := getPackageJson(root)
	if err != nil {
		return -1
	}

	for _, regex := range regexes {
		re := regexp.MustCompile(regex)
		port := utils.FindPortSubmatch(re, getScript(packageJson), 1)
		if port != -1 {
			return port
		}
	}

	return -1
}

func getPackageJson(root string) (schema.PackageJson, error) {
	packageJsonPath := filepath.Join(root, "package.json")
	return utils.GetPackageJsonSchemaFromFile(packageJsonPath)
}
