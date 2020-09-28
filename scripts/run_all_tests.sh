#!/usr/bin/env bash
set -e
GOFLAGS="-mod=vendor" make test
