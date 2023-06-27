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
	"context"
	"encoding/json"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/utils"
)

type VertxDetector struct{}

type VertxConf struct {
	Port         int          `json:"http.port,omitempty"`
	ServerConfig ServerConfig `json:"http.server,omitempty"`
}
type ServerConfig struct {
	Port int `json:"http.server.port,omitempty"`
}

func (v VertxDetector) GetSupportedFrameworks() []string {
	return []string{"Vertx"}
}

// DoFrameworkDetection uses the groupId to check for the framework name
func (v VertxDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.vertx", ""); hasFwk {
		language.Frameworks = append(language.Frameworks, "Vertx")
	}
}

// DoPortsDetection searches for the port in json files under src/main/conf/
func (v VertxDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "src/main/conf",
			File: ".*.json",
		},
	}, ctx)
	if err != nil {
		return
	}
	var data VertxConf
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return
	}
	if utils.IsValidPort(data.Port) {
		component.Ports = []int{data.Port}
	} else if utils.IsValidPort(data.ServerConfig.Port) {
		component.Ports = []int{data.ServerConfig.Port}
	}

}
