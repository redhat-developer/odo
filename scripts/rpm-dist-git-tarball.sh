#!/usr/bin/env bash

SPEC_FILE="atomic-openshift-odo.spec"
SPEC_FILE_SRC="dist/rpmbuild/SPECS/$SPEC_FILE"
VERSION_INFO_FILE="dist/rpmbuild/version"
ODO_VERSION=`cat $VERSION_INFO_FILE | grep "ODO_VERSION" | cut -d "=" -f2`
ODO_RELEASE=`cat $VERSION_INFO_FILE | grep "ODO_RELEASE" | cut -d "=" -f2`
NAME="atomic-openshift-odo-$ODO_VERSION-$ODO_RELEASE"
echo "Preping dist-git tarball for $NAME"
cp -arf $SPEC_FILE_SRC $SPEC_FILE
pushd ..
echo "Creating tarball at $(pwd)/$NAME.tar.gz"
tar -czf $NAME.tar.gz odo/*
popd
rm -rf $SPEC_FILE
