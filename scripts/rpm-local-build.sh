#!/usr/bin/env bash
set -e

if [[ ! -d dist/rpmbuild  ]]; then
	echo "Cannot build as artifacts are not generated. Run scrips/rpm-prepare.sh first"
	exit 1
fi

top_dir="`pwd`/dist/rpmbuild"
echo "Building locally"
rpmbuild --define "_topdir `echo $top_dir`" -ba dist/rpmbuild/SPECS/odo.spec
