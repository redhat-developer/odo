#!/bin/bash

# source: https://github.com/codecov/example-go
# go test can't generate code coverage for multiple packages in one command

set -e
echo "" > coverage.txt
go test -i -race ./cmd/odo
for d in $(go list ./... | grep -v vendor | grep -v tests | grep -v testingutil); do
    # For watch related tests, race check causes issue so disabling them here as race is already tested in other tests when used with `-coverprofile=profile.out`
    if [ "$d" = "github.com/openshift/odo/pkg/component" ]; then
        go test -coverprofile=profile.out -covermode=atomic $d
    else
        go test -race -coverprofile=profile.out -covermode=atomic $d
    fi
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done