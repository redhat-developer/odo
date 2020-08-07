#!/bin/bash
set -x

# Image stream source https://github.com/openshift/library/tree/master/community 

# odo supported nodejs image stream
oc apply -n openshift -f tests/image-streams/supported-nodejs.json
# odo supported java image stream
oc apply -n openshift -f tests/image-streams/supported-java.json
