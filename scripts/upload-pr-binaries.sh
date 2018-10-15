#!/bin/bash
set -e

# upload binaries build from PR
# required $BINTRAY_USER and $BINTRAY_KEY

BIN_DIR="./dist/bin/"


if [[ -z "${BINTRAY_USER}" ]] || [[ -z "${BINTRAY_KEY}" ]]; then
    echo "Required variables \$BINTRAY_USER, \$BINTRAY_KEY"
    exit 1
fi
if [[ -z "${TRAVIS_PULL_REQUEST}" ]]; then
    echo "This script should run on travis-ci. (TRAVIS_PULL_REQUEST env variable is not set)"
    exit 1
fi

if [[ "${TRAVIS_PULL_REQUEST}" == "false" ]]; then
    echo "Not a pull request. (TRAVIS_PULL_REQUEST=${TRAVIS_PULL_REQUEST})"
    exit 0
fi

bintray_version="pr${TRAVIS_PULL_REQUEST}"


for file in `gfind ${BIN_DIR} -type f -printf '%P\n'`; do  
    upload_path="/${bintray_version}/${file}.gz"
    download_url="https://dl.bintray.com/odo/odo/${upload_path}"

    # todo remove first pr
    echo "Uploading ${BIN_DIR}/${file} as ${file}.gz to Bintray (${download_url})"
    gzip ${BIN_DIR}/${file} -c | curl -T - -u $BINTRAY_USER:$BINTRAY_KEY "https://api.bintray.com/content/odo/odo/odo/${bintray_version}/${upload_path}?publish=1&override=1"
    echo ""
    echo ""
done
