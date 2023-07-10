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
	"github.com/devfile/alizer/pkg/utils"
)

// hasFramework uses all files to check for framework
func hasFramework(files *[]string, tag string) bool {
	for _, file := range *files {
		if hasTag, _ := utils.IsTagInFile(file, tag); hasTag {
			return true
		}
	}
	return false
}
