#!/bin/bash

cd ../..

go install golang.org/x/tools/cmd/godoc@latest
go install gitlab.com/tslocum/godoc-static@latest

export GOPATH=$(go env GOPATH)
PATH=$PATH:${GOPATH}/bin

mkdir -p docs/website/build/godoc
godoc-static -destination docs/website/build/godoc .
