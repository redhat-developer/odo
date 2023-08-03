---
title: Using the Web UI to edit the Devfile
sidebar_position: 3
---

When the command `odo dev` is running, a Graphical User Interface (GUI) is accessible (generally at http://localhost:20000). 

The interface can be used to edit the Devfile used by the `odo dev` session.

The interface proposes three representations of the Devfile: a textual representation in YAML format, a Chart, and a graphical representation.

You can edit any representation independently, the other representations will be modified accordingly. The chart is read-only, the two other representations can be freely edited.

The YAML representation reflects exactly the content of the `devfile.yaml` file found into the directory where the `odo dev` session is running.

The page *Chart* contains a chart describing the different steps of the `odo dev` session.

The following pages of the UI contain a graphical representation of the Devfile. From these pages, you can edit the Devfile by adding and deleting objects (commands, events, containers, images, resources and volumes). From the *Commands* page, it is possible to change the *Kind* (Build, Run, Test, Debug or Deploy) of each command. 

When you *Save* the Devfile, the content of the YAML representation is saved to the disk, replacing the previous version of the Devfile. The `odo dev` session will react accordingly, depending on the changes done into the Devfile.

When the file `devfile.yaml` is modified, its content is sent to the GUI, which will alert you and give you the opportunity to accept the changes done into the file. If you accept, the changes you may have done in the interface will be lost.

## Limitations

### Limited Devfile Schema Versions

The only supported Devfile Schema version is 2.2.0.

### Limited support of parameters during object creation

When you create an object (either a command, a container or an image) from the graphical representation, you can fill in a limited number of parameters for the object, as only the more common parameters are presented in the creation form.

You can add parameters to the object after it has been created by editing it from the YAML representation.

### Limited support for parent devfile

When the current Devfile is referencing a parent Devfile (using the `.parent` field into the YAML), this parent is not represented into the GUI.

It is still possible to add a parent information to the YAML representation. It will be taken into account by the `odo dev` session once saved.

### Experimental Chart representation

The chart representation is experimental. It is possible that, for some complex Devfile, the chart is not accurate or is not displayed.
