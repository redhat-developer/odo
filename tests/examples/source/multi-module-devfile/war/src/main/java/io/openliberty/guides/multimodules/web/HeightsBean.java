// tag::copyright[]
/*******************************************************************************
 * Copyright (c) 2017, 2019 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *     IBM Corporation - Initial implementation
 *******************************************************************************/
// end::copyright[]
package io.openliberty.guides.multimodules.web;

public class HeightsBean implements java.io.Serializable {
    private String heightCm = null;
    private String heightFeet = null;
    private String heightInches = null;
    private int cm = 0;
    private int feet = 0;
    private int inches = 0;

    public HeightsBean() {
    }

    // Capitalize the first letter of the name i.e. first letter after get
    // If first letter is not capitalized, it must match the property name in
    // index.jsp
    public String getHeightCm() {
        return heightCm;
    }

    public String getHeightFeet() {
        return heightFeet;
    }

    public String getHeightInches() {
        return heightInches;
    }

    public void setHeightCm(String heightcm) {
        this.heightCm = heightcm;
    }

    // Need an input as placeholder, you can choose not to use the input
    // tag::setHeightFeet[]
    public void setHeightFeet(String heightfeet) {
        this.cm = Integer.valueOf(heightCm);
        // tag::getFeet[]
        this.feet = io.openliberty.guides.multimodules.lib.Converter.getFeet(cm);
        // end::getFeet[]
        String result = String.valueOf(feet);
        this.heightFeet = result;
    }
    // end::setHeightFeet[]

    // tag::setHeightInches[]
    public void setHeightInches(String heightinches) {
        this.cm = Integer.valueOf(heightCm);
        // tag::getInches[]
        this.inches = io.openliberty.guides.multimodules.lib.Converter.getInches(cm);
        // end::getInches[]
        String result = String.valueOf(inches);
        this.heightInches = result;
    }
    // end::setHeightInches[]

}
