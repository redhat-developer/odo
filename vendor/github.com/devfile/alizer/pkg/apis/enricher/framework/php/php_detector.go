/*******************************************************************************
 * Copyright (c) 2023 Red Hat, Inc.
 * Distributed under license by Red Hat, Inc. All rights reserved.
 * This program is made available under the terms of the
 * Eclipse Public License v2.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 * Red Hat, Inc.
 ******************************************************************************/

package enricher

import "github.com/devfile/alizer/pkg/utils"

// hasFramework uses the composer.json to check for framework
func hasFramework(configFile string, tag string) bool {
	return utils.IsTagInComposerJsonFile(configFile, tag)
}
