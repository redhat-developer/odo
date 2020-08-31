#!/bin/bash

# This will cross-compile odo for all platforms:
# Windows, Linux and macOS


COMMON_FLAGS=${@}

if [[ -z "${COMMON_FLAGS}" ]]; then
    echo "Common build flags is missing"
    exit 1
fi

for platform in linux-amd64 linux-arm64 linux-ppc64le linux-s390x darwin-amd64 windows-amd64 ; do
  echo "Cross compiling $platform and placing binary at dist/bin/$platform/"
  if [ $platform == "windows-amd64" ]; then
    GOARCH=amd64 GOOS=windows go build -o dist/bin/$platform/odo.exe -ldflags="-s -w $COMMON_FLAGS" ./cmd/odo/
  else
    GOARCH=${platform#*-} GOOS=${platform%-*} go build -o dist/bin/$platform/odo -ldflags="-s -w $COMMON_FLAGS" ./cmd/odo/
  fi
done
