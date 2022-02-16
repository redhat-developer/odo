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
package recognizer

import (
	"strings"

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
