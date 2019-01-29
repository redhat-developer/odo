#!/bin/bash

# fpm is required installed (https://github.com/jordansissel/fpm)

BIN_DIR="./dist/bin/"
PKG_DIR="./dist/pkgs/"

mkdir -p $PKG_DIR

# package version, use current date by default (if build from master)
PKG_VERSION=$(date "+%Y%m%d%H%M%S")

# if this is run on travis make sure that binary was build with corrent version
if [[ -n $TRAVIS_TAG ]]; then
    echo "Checking if odo version was set to the same version as current tag"
    # use sed to get only semver part
    bin_version=$(${BIN_DIR}/linux-amd64/odo version --client | head -1 | sed "s/^odo \(.*\) (.*)$/\1/")
    if [ "$TRAVIS_TAG" == "${bin_version}" ]; then
        echo "OK: odo version output is matching current tag"
    else
        echo "ERR: TRAVIS_TAG ($TRAVIS_TAG) is not matching 'odo version' (v${bin_version})"
        exit 1
    fi
    # this is build from tag, that means it is proper relase, use version for PKG_VERSION
    PKG_VERSION=$(echo ${bin_version} | sed "s/^v\(.*\)$/\1/")
fi

# create packages using fpm
fpm -h  >/dev/null 2>&1 || { 
    echo "ERROR: fpm (https://github.com/jordansissel/fpm) is not installed. Can't create linux packages"
    exit 1
}

TMP_DIR=$(mktemp -d)
mkdir -p $TMP_DIR/usr/local/bin/
cp $BIN_DIR/linux-amd64/odo $TMP_DIR/usr/local/bin/

echo "creating DEB package"
fpm \
  --input-type dir --output-type deb \
  --chdir $TMP_DIR \
  --name odo --version $PKG_VERSION \
  --architecture amd64 \
  --maintainer "Tomas Kral <tkral@redhat.com>" \
  --package $PKG_DIR

echo "creating RPM package"
fpm \
  --input-type dir --output-type rpm \
  --chdir $TMP_DIR \
  --name odo --version $PKG_VERSION \
  --architecture x86_64 --rpm-os linux \
  --maintainer "Tomas Kral <tkral@redhat.com>" \
  --package $PKG_DIR