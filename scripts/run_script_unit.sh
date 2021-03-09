#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}
shout "Running unit tests"

# Run unit tests
if [[ $BASE_OS == "windows" ]]; then
  GOFLAGS='-mod=vendor' powershell -Command "make test"
else
  GOFLAGS='-mod=vendor' make test
fi