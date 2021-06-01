#!/bin/bash

# This script generates the baseline release notes without the changelog
# User must copy contents of `Changelog.md` into appropriate location in
# release notes

if [ -z "$1" ]  || [ -z "$2" ]
then
  echo -e "Must provide first and next release numbers..\nex: ./changelog-script.sh v1.0.0 v1.0.1"
  exit 1
fi

MIRROR="https://mirror.openshift.com/pub/openshift-v4/clients/odo/$2/"
INSTALLATION_GUIDE="https://docs.openshift.com/container-platform/latest/cli_reference/developer_cli_odo/installing-odo.html"
GIT_TREE="https://github.com/openshift/odo/tree/$2"
FULL_CHANGELOG="https://github.com/openshift/odo/compare/$1...$2"

echo -e "
# Release of $2

## [$2]($GIT_TREE) ($(date '+%Y-%m-%d'))

[Full Changelog]($FULL_CHANGELOG)

-- COPY CONTENT FROM Changelog.md here --

# Installation of $2

To install odo, follow our installation guide at [docs.openshift.com]($INSTALLATION_GUIDE)

After each release, binaries are synced to [mirror.openshift.com]($MIRROR)" > /tmp/changelog

echo "The changelog is located at: /tmp/changelog"
echo ""
echo "Contents of changelog : "
cat /tmp/changelog
echo ""