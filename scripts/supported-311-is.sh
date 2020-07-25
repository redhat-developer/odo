#!/bin/bash
set -x

# odo supported nodejs image stream
oc apply -n openshift -f tests/image-streams/supported-nodejs.json

# odo supported java image stream
oc apply -n openshift -f tests/image-streams/supported-java.json
