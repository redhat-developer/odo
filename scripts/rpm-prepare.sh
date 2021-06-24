#!/usr/bin/env bash

set +ex

echo "Reading ODO_VERSION, ODO_RELEASE and GIT_COMMIT env, if they are set"
# Change version as needed. In most cases ODO_RELEASE would not be touched unless
# we want to do a re-lease of same version as we are not backporting
export ODO_VERSION=${ODO_VERSION:=2.2.2}
export ODO_RELEASE=${ODO_RELEASE:=1}

export GIT_COMMIT=${GIT_COMMIT:=`git rev-parse --short HEAD 2>/dev/null`}
export ODO_RPM_VERSION=${ODO_VERSION//-}

# Golang version variables, if you are bumping this, please contact redhat maintainers to ensure that internal
# build systems can handle these versions
export GOLANG_VERSION=${GOLANG_VERSION:-1.15}
export GOLANG_VERSION_NODOT=${GOLANG_VERSION_NODOT:-115}

# Print env for verifcation
echo "Printing envs for verification"
echo "ODO_VERSION=$ODO_VERSION"
echo "ODO_RELEASE=$ODO_RELEASE"
echo "GIT_COMMIT=$GIT_COMMIT"
echo "ODO_RPM_VERSION=$ODO_RPM_VERSION"
echo "GOLANG_VERSION=$GOLANG_VERSION"
echo "GOLANG_VERSION_NODO=$GOLANG_VERSION_NODOT"

OUT_DIR=".rpmbuild"
DIST_DIR="$(pwd)/dist"

SPEC_DIR="$OUT_DIR/SPECS"
SOURCES_DIR="$OUT_DIR/SOURCES"
FINAL_OUT_DIR="$DIST_DIR/rpmbuild"

NAME="openshift-odo-$ODO_RPM_VERSION-$ODO_RELEASE"

echo "Making release for $NAME, git commit $GIT_COMMIT"

echo "Cleaning up old content"
rm -rf $DIST_DIR
rm -rf $FINAL_OUT_DIR

echo "Configuring output directory $OUT_DIR"
rm -rf $OUT_DIR
mkdir -p $SPEC_DIR
mkdir -p $SOURCES_DIR/$NAME
mkdir -p $FINAL_OUT_DIR

echo "Generating spec file $SPEC_DIR/openshift-odo.spec"
envsubst <rpms/openshift-odo.spec > $SPEC_DIR/openshift-odo.spec

echo "Generating tarball $SOURCES_DIR/$NAME.tar.gz"
# Copy code for manipulation
cp -arf ./* $SOURCES_DIR/$NAME
pushd $SOURCES_DIR
pushd $NAME
# Remove bin if it exists, we dont need it in tarball
rm -rf ./odo
popd

# Create tarball
tar -czf $NAME.tar.gz $NAME
# Removed copied content
rm -rf $NAME
popd

echo "Finalizing..."
# Store version information in file for reference purposes
echo "ODO_VERSION=$ODO_VERSION" > $OUT_DIR/version
echo "ODO_RELEASE=$ODO_RELEASE" >> $OUT_DIR/version
echo "GIT_COMMIT=$GIT_COMMIT" >> $OUT_DIR/version
echo "ODO_RPM_VERSION=$ODO_RPM_VERSION" >> $OUT_DIR/version
echo "GOLANG_VERSION=$GOLANG_VERSION" >> $OUT_DIR/version
echo "GOLANG_VERSION_NODOT=$GOLANG_VERSION_NODOT" >> $OUT_DIR/version


# After success copy stuff to actual location
mv $OUT_DIR/* $FINAL_OUT_DIR
# Remove out dir
rm -rf $OUT_DIR	
echo "Generated content in $FINAL_OUT_DIR"
