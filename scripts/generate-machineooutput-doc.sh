#!/bin/bash

set -e

asciioutput() {
  OUTPUT=`${@}`
  ALLOUTPUT="== ${@}

[source,json]
----
${OUTPUT}
----
"
  echo ${ALLOUTPUT} > test.adoc
}

tmpdir=`mktemp -d`
cd $tmpdir
git clone https://github.com/openshift/nodejs-ex
cd nodejs-ex

# Commands that don't have json support
# app delete
# catalog describe
# catalog search
# component create
# component delete
# component link
# component log
# component push
# component unlink
# component update
# component watch
# config set
# config unset
# config view
# debug port-forward 
# preference set
# preference unset
# preference view
# service create
# service delete
# storage delete
# url create
# url delete
# login
# logout
# utils *
# version


# Alphabetical order for json output...

# Preliminary?
odo project delete foobar -f || true
sleep 5
odo project create foobar
sleep 5
odo create nodejs
odo push

# app
asciioutput odo app describe app -o json
odo app list -o json

# catalog
odo catalog list components -o json
odo catalog list services -o json

# component
odo component delete -o json
odo component push 

# project
odo project create foobar -o json
odo project delete foobar -o json
odo project list -o json

# service

## preliminary
odo service create mongodb-persistent mongodb --plan default --wait -p DATABASE_SERVICE_NAME=mongodb -p MEMORY_LIMIT=512Mi -p MONGODB_DATABASE=sampledb -p VOLUME_CAPACITY=1Gi
odo service list -o json

# storage
odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi -o json
odo storage list -o json
odo storage delete

# url
odo url create myurl
odo url list -o json
odo url delete myurl
