#!/bin/bash

set -e

PATH=${PATH}:${GOPATH}/bin
make goget-tools
make validate
