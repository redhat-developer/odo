#!/bin/bash

BIN_DIR="./dist/bin/"
RELEASE_DIR="./dist/release/"

mkdir -p $RELEASE_DIR

for arch in `ls -1 $BIN_DIR/`;do
    suffix=""
    if [[ $arch == windows-* ]]; then
        suffix=".exe"
    fi
    gzip --keep --to-stdout $BIN_DIR/$arch/ocdev$suffix > $RELEASE_DIR/ocdev-$arch$suffix.gz
done
