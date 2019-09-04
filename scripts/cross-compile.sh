#!/bin/bash

# This will cross-compile odo for all platforms:
# Windows, Linux and macOS


COMMON_FLAGS=${@}

if [[ -z "${COMMON_FLAGS}" ]]; then
    echo "Common build flags is missing"
    exit 1
fi

for platform in linux darwin windows ; do
  echo "Cross compiling $platform-amd64 and placing binary at dist/bin/$platform-amd64/"
  if [ $platform == "windows" ]; then
    GOARCH=amd64 GOOS=$platform go build -o dist/bin/$platform-amd64/odo.exe -ldflags="$COMMON_FLAGS" ./cmd/odo/
  else
    GOARCH=amd64 GOOS=$platform go build -o dist/bin/$platform-amd64/odo -ldflags="$COMMON_FLAGS" ./cmd/odo/
  fi
done
