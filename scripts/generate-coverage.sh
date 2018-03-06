#!/bin/bash

# source: https://github.com/codecov/example-go
# go test can't generate code coverage for multiple packages in one command

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -v -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
