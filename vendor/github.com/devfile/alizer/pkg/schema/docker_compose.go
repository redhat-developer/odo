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
package schema

type DockerComposeFile struct {
	Services SVCS `yaml:"services"`
}

type SVCS struct {
	Web WebService `yaml:"web"`
}

type WebService struct {
	Ports  []interface{} `yaml:"ports"`
	Expose []string      `yaml:"expose"`
}
