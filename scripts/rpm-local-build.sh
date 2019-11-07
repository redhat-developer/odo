#!/usr/bin/env bash
set -e

if [[ ! -d dist/rpmbuild  ]]; then
	echo "Cannot build as artifacts are not generated. Run scripts/rpm-prepare.sh first"
	exit 1
fi

echo "Copying content to local rpmbuild"
mkdir -p $HOME/rpmbuild/SOURCES $HOME/rpmbuild/SPECS
cp -avrf dist/rpmbuild/SOURCES/* $HOME/rpmbuild/SOURCES/
cp -avrf dist/rpmbuild/SPECS/* $HOME/rpmbuild/SPECS/

echo "Building locally"
rpmbuild -ba $HOME/rpmbuild/SPECS/openshift-odo.spec
