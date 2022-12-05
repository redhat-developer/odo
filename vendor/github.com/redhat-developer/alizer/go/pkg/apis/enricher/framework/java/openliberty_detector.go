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
	"encoding/xml"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	utils "github.com/redhat-developer/alizer/go/pkg/utils"
)

type OpenLibertyDetector struct{}

type Server_Xml struct {
	Server struct {
		HttpEndpoint struct {
			HttpPort  string `xml:"httpPort,attr"`
			HttpsPort string `xml:"httpsPort,attr"`
		} `xml:"httpEndpoint"`
	} `xml:"server"`
}

func (o OpenLibertyDetector) GetSupportedFrameworks() []string {
	return []string{"OpenLiberty"}
}

func (o OpenLibertyDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFwk, _ := hasFramework(config, "io.openliberty"); hasFwk {
		language.Frameworks = append(language.Frameworks, "OpenLiberty")
	}
}

func (o OpenLibertyDetector) DoPortsDetection(component *model.Component) {
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "server.xml",
		},
		{
			Dir:  "src/main/liberty/config",
			File: "server.xml",
		},
	})
	if err != nil {
		return
	}
	var data Server_Xml
	xml.Unmarshal(bytes, &data)
	ports := utils.GetValidPorts([]string{data.Server.HttpEndpoint.HttpPort, data.Server.HttpEndpoint.HttpsPort})
	if len(ports) > 0 {
		component.Ports = ports
	}
}
