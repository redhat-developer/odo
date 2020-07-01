#!/usr/bin/env bash
set -e

if [[ ! -d dist/rpmbuild  ]]; then
	echo "Cannot build as artifacts are not generated. Run scrips/rpm-prepare.sh first"
	exit 1
fi

echo "Cleaning up old rpmcontent"
rm -f ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

echo "Copying content to local rpmbuild"
cp -avrf dist/rpmbuild/SOURCES/* $HOME/rpmbuild/SOURCES/
cp -avrf dist/rpmbuild/SPECS/* $HOME/rpmbuild/SPECS/

echo "Building locally"
rpmbuild -ba $HOME/rpmbuild/SPECS/openshift-odo.spec
