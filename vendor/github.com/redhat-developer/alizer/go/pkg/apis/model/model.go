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
package model

type Language struct {
	Name           string
	Aliases        []string
	Weight         float64
	Frameworks     []string
	Tools          []string
	CanBeComponent bool
}

type Component struct {
	Name      string
	Path      string
	Languages []Language
}
