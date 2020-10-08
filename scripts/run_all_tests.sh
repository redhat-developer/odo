#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "Setting up"

mkdir bin
GOBIN="`pwd`/bin"
if [[ $BASE_OS == "windows" ]]; then
    GOBIN="$(cygpath -pw $GOBIN)"
fi
PATH=$PATH:$GOBIN

#-----------------------------------------------------------------------------

shout "Testing against 4x cluster"

if [[ $BASE_OS == "linux"  ]]; then
    set +x
	curl -k ${OC4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.tar -o ./oc.tar
    set -x
	tar -C $GOBIN -xvf ./oc.tar && rm -rf ./oc.tar
else
    set +x
    curl -k ${OC4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.zip -o ./oc.zip
    set -x
    gunzip -c ./oc.zip > $GOBIN/oc && rm -rf ./oc.zip && chmod +x $GOBIN/oc
    if [[ $BASE_OS == "windows" ]]; then
        mv -f $GOBIN/oc $GOBIN/oc.exe
    fi
fi

shout "Logging into 4x cluster as developer (logs hidden)"
set +x
oc login -u developer -p password@123 --insecure-skip-tls-verify  ${OCP4X_API_URL}
set -x

# Run unit tests
GOFLAGS='-mod=vendor' make test

# Prep for int
echo "getting ginkgo"
GOBIN="$GOBIN" make goget-ginkgo

set +e
ls -a $GOBIN
ginkgo version
run_all=$?
set -e
# Integration tests
if [[ $run_all -eq 0 ]]; then
    make test-e2e-all
else
    echo "Ginkgo does not exist, skipping integration/e2e"
fi
