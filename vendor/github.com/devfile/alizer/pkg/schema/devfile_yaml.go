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

package schema

type DevfileYaml struct {
	StarterProjects []struct {
		Git struct {
			CheckoutFrom struct {
				Remote   string `yaml:"remote"`
				Revision string `yaml:"revision"`
			} `yaml:"checkoutFrom"`
			Remotes struct {
				Origin string `yaml:"origin"`
			} `yaml:"remotes"`
		} `yaml:"git"`
		SubDir string `yaml:"subDir"`
	} `yaml:"starterProjects"`
}
