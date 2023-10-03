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
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/schema"
	"github.com/devfile/alizer/pkg/utils"
)

type DotNetDetector struct{}

func (d DotNetDetector) GetSupportedFrameworks() []string {
	return []string{""}
}

// DoFrameworkDetection uses configFilePath to check for the name of the framework
func (d DotNetDetector) DoFrameworkDetection(language *model.Language, configFilePath string) {
	framework := getFrameworks(configFilePath)
	if framework == "" {
		return
	}
	var frameworks []string
	if strings.Contains(framework, ";") {
		frameworks = strings.Split(framework, ";")
	} else {
		frameworks = []string{framework}
	}

	for _, frm := range frameworks {
		if !utils.Contains(language.Frameworks, frm) {
			language.Frameworks = append(language.Frameworks, frm)
		}
	}
}

func (d DotNetDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
}

func getFrameworks(configFilePath string) string {
	cleanConfigPath := filepath.Clean(configFilePath)
	xmlFile, err := os.Open(cleanConfigPath)
	if err != nil {
		return ""
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var proj schema.DotNetProject
	err = xml.Unmarshal(byteValue, &proj)
	if err != nil {
		return ""
	}
	defer func() error {
		if err := xmlFile.Close(); err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
		return nil
	}()
	if proj.PropertyGroup.TargetFramework != "" {
		return proj.PropertyGroup.TargetFramework
	} else if proj.PropertyGroup.TargetFrameworkVersion != "" {
		return proj.PropertyGroup.TargetFrameworkVersion
	} else if proj.PropertyGroup.TargetFrameworks != "" {
		return proj.PropertyGroup.TargetFrameworks
	}
	return ""
}
