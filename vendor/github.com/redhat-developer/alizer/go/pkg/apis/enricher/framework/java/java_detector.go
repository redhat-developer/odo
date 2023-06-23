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
	"github.com/redhat-developer/alizer/go/pkg/utils"
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
