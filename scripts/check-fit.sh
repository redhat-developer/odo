#!/bin/bash

set -e

if grep -nr "FIt(" tests/; then
  echo "Not OK. FIt exists somewhere in the testing code. Please remove it."
  exit 1
else
  echo "OK"
fi
