#!/usr/bin/sh

############################################################################
#  PREREQUISITES FOR THIS SCRIPT
# 1. Redistributable-binary(.rpm) should be passed as the first argument
# 2. Login to the cluster should be done prior to running this script
# 3. The cluster should be in a state where it can be used for testing
#
# USAGE:
# ./release-bit-verification.sh redistributable-binary
#
# Example: ./release-bit-verification.sh ~/Downloads/odo-redistributable-2.4.3-1.el8.x86_64.rpm
#

shout() {
  echo "--------------------------------------------------------------------------"
}
# Check SHASUM for all the binary files and there should be no difference

# Create a Temp directory 
WORKING_DIR=`mktemp -d`

# Extract from rpm file 
rpm2cpio ${1} | cpio -idmvD $WORKING_DIR
pushd $WORKING_DIR/usr/share/odo-redistributable/

# Check sha256sum for all the files
while IFS= read -r line; do
    read -r SHA FILE <<<"$line"
    read -r SHATOCHECK FILE <<< `sha256sum $FILE`
    if [[ $SHA == $SHATOCHECK ]]; then
        # Print if the file is correct
        printf '%-50s\U0002705\n' $FILE 
    fi
done < SHA256_SUM

shout

# Copy binary for testing purpose
OS=`uname -s`
ARCH=`uname -m`

if [[ $OS == "Linux" ]]; then
    if [[ $ARCH == "x86_64" ]]; then
        cp ./odo-linux-amd64 $GOBIN/odo
    fi
fi

# Check odo verion and if it is correct 
VERSION=`cat VERSION`
ODOVERSIONCHECK=`odo version`
if [[ "$ODOVERSIONCHECK" == *"$VERSION"*  ]]; then
    echo "odo binary is installed correctly"
fi

#clone repo for testing and checkout release tag
pushd $WORKING_DIR
if [  -d "odo" ]; then
    rm -rf odo
fi
git clone https://github.com/redhat-developer/odo.git  && cd $WORKING_DIR/odo && git checkout "v$VERSION"

#Run tests
make test-integration-devfile
make test-integration
make test-operator-hub
make test-e2e-all
make test-cmd-project

# Cleanup
rm -rf /tmp/odo /tmp/usr