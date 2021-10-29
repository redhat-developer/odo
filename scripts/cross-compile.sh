#!/bin/bash

# This will cross-compile odo for all platforms:
# Windows, Linux and macOS

if [[ -z "${*}" ]]; then
    echo "Build flags are missing"
    exit 1
fi

for platform in linux-amd64 linux-arm64 linux-ppc64le linux-s390x darwin-amd64 darwin-arm64 windows-amd64 ; do
  echo "Cross compiling $platform and placing binary at dist/bin/$platform/"
  if [ $platform == "windows-amd64" ]; then
    GOARCH=amd64 GOOS=windows go build -o dist/bin/$platform/odo.exe "${@}" ./cmd/odo/
  else
    GOARCH=${platform#*-} GOOS=${platform%-*} go build -o dist/bin/$platform/odo "${@}" ./cmd/odo/
  fi
done
