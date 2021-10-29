#!/usr/bin/env bash

set -e

echo "Preping rpm"
scripts/rpm-prepare.sh

echo "Building rpm"
scripts/rpm-local-build.sh

rm -rf dist/rpmtest
mkdir -p dist/rpmtest/{odo,redistributable}

echo "Validating odo rpm"
rpm2cpio dist/rpmbuild/RPMS/x86_64/`ls dist/rpmbuild/RPMS/x86_64/ | grep -v redistributable` > dist/rpmtest/odo/odo.cpio
pushd dist/rpmtest/odo
cpio -idv < odo.cpio
ls ./usr/bin | grep odo
./usr/bin/odo version
popd

RL="odo-darwin-amd64 odo-darwin-arm64 odo-linux-ppc64le odo-linux-arm64 odo-windows-amd64.exe odo-linux-amd64 odo-linux-s390x"
echo "Validating odo-redistributable rpm"
rpm2cpio dist/rpmbuild/RPMS/x86_64/`ls dist/rpmbuild/RPMS/x86_64/ | grep redistributable` > dist/rpmtest/redistributable/odo-redistribuable.cpio
pushd dist/rpmtest/redistributable
cpio -idv < odo-redistribuable.cpio
for i in $RL; do
	ls ./usr/share/odo-redistributable | grep $i
done
./usr/share/odo-redistributable/odo-linux-amd64 version
popd
