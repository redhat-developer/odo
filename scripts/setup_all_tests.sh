#!/usr/bin/env bash

# Base Setup

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

mkdir bin
BINDIR="`pwd`/bin"

if [[ $BASE_OS == "linux"  ]]; then
    set +x
	curl -k ${OC4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.tar -o ./oc.tar
    set -x
	tar -C $BINDIR -xvf ./oc.tar && rm -rf ./oc.tar
else
    set +x
    curl -k ${OC4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.zip -o ./oc.zip
    set -x
    gunzip -c ./oc.zip > $BINDIR/oc && rm -rf ./oc.zip && chmod +x $BINDIR/oc
    if [[ $BASE_OS == "windows" ]]; then
        mv -f $BINDIR/oc $BINDIR/oc.exe
    fi
fi

## 4.x tests
shout "Logging into 4x cluster as developer (logs hidden)"
set +x
oc login -u developer -p password@123 --insecure-skip-tls-verify  ${OCP4X_API_URL}
set -x

