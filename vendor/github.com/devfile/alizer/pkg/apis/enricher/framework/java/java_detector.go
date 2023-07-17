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

	"github.com/devfile/alizer/pkg/utils"
)

type ApplicationPropertiesFile struct {
	Dir  string
	File string
}

// hasFramework uses the build.gradle, groupId, and artifactId to check for framework
func hasFramework(configFile, groupId, artifactId string) (bool, error) {
	if utils.IsPathOfWantedFile(configFile, "build.gradle") {
		return utils.IsTagInFile(configFile, groupId)
	} else if artifactId != "" {
		return utils.IsTagInPomXMLFileArtifactId(configFile, groupId, artifactId)
	} else {
		return utils.IsTagInPomXMLFile(configFile, groupId)
	}
}

// GetPortsForJBossFrameworks tries to detect any port information inside javaOpts of configuration
// of a given profiles plugin
func GetPortsForJBossFrameworks(pomFilePath, pluginArtifactId, pluginGroupId string) string {
	portPlaceholder := ""
	pom, err := utils.GetPomFileContent(pomFilePath)
	if err != nil {
		return portPlaceholder
	}

	re := regexp.MustCompile(`jboss.https?.port=\d*`)
	// Check for port configuration inside profiles
	for _, profile := range pom.Profiles.Profile {
		for _, plugin := range profile.Build.Plugins.Plugin {
			if !(strings.Contains(plugin.ArtifactId, pluginArtifactId) && strings.Contains(plugin.GroupId, pluginGroupId)) {
				continue
			}
			matchIndexesSlice := re.FindAllStringSubmatchIndex(plugin.Configuration.JavaOpts, -1)
			for _, matchIndexes := range matchIndexesSlice {
				if len(matchIndexes) > 1 {
					portPlaceholder = plugin.Configuration.JavaOpts[matchIndexes[0]:matchIndexes[1]]
					for _, httpArg := range []string{"jboss.http.port=", "jboss.https.port="} {
						portPlaceholder = strings.Replace(portPlaceholder, httpArg, "", -1)
					}
				}
			}
		}
	}
	return portPlaceholder
}
