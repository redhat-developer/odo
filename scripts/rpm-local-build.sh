#!/usr/bin/env bash

if [[ ! -d dist/rpmbuild  ]]; then
	echo "Cannot build as artifacts are not generated. Run scrips/rpm-prepare.sh first"
	exit 1
fi

echo "Copying content to local rpmbuild"
cp -avrf dist/rpmbuild/SOURCES/* $HOME/rpmbuild/SOURCES/
cp -avrf dist/rpmbuild/SPECS/* $HOME/rpmbuild/SPECS/

echo "Building locally"
rpmbuild -ba $HOME/rpmbuild/SPECS/atomic-openshift-odo.spec
