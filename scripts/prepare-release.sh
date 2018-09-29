#!/bin/bash

# this script assumes that runs on linux

BIN_DIR="./dist/bin/"
RELEASE_DIR="./dist/release/"

mkdir -p $RELEASE_DIR

# if this is run on travis make sure that binary was build with corrent version
if [[ -n $TRAVIS_TAG ]]; then
    echo "Checking if odo version was set to the same version as current tag"
    # use sed to get only semver part
    bin_version=$(${BIN_DIR}/linux-amd64/odo version | head -1 | sed "s/^odo \(.*\) (.*)$/\1/")
    if [ "$TRAVIS_TAG" == "${bin_version}" ]; then
        echo "OK: odo version output is matching current tag"
    else
        echo "ERR: TRAVIS_TAG ($TRAVIS_TAG) is not matching 'odo version' (v${bin_version})"
        exit 1
    fi
fi

# gziped binaries
for arch in `ls -1 $BIN_DIR/`;do
    suffix=""
    if [[ $arch == windows-* ]]; then
        suffix=".exe"
    fi
    source_file=$BIN_DIR/$arch/odo$suffix
    target_file=$RELEASE_DIR/odo-$arch$suffix

    # Create a gzip of the binary
    echo "gzipping binary $source_file as $target_file"
    gzip --keep --to-stdout $source_file > $target_file.gz

    # Move binaries to the release directory as well
    echo "copying binary $source_file to release directory"
    cp $source_file $target_file
done

function release_sha() {
    release_dir_files=`find $RELEASE_DIR -maxdepth 1 ! -name SHA256_SUM -type f -printf "%f\n"`
    for filename in $release_dir_files; do
        sha_sum=`sha256sum $RELEASE_DIR${filename}|awk '{ print $1 }'`; echo $sha_sum  $filename;
    done > ${RELEASE_DIR}SHA256_SUM
}

release_sha