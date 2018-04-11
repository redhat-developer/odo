#!/bin/bash

# upload linux packages to bintray repositories
# required $BINTRAY_USER and $BINTRAY_KEY

PKG_DIR="./dist/pkgs/"


if [[ -z "${BINTRAY_USER}" ]] || [[ -z "${BINTRAY_KEY}" ]]  ; then
    echo "Required variables \$BINTRAY_USER and \$BINTRAY_KEY"
    exit 1
fi


#  for deb
for pkg in `ls -1 $PKG_DIR/*.deb`; do 
    filename=$(basename $pkg)
    # get version from filename
    version=$(expr "$filename" : '.*_\([^_]*\)_.*')

    repo="odo-deb-dev"
    # if version is semver format upload to releases
    if [[ $version =~ [0-9]+\.[0-9]+\.[0-9]+ ]] ; then 
        repo="odo-deb-releases"
    fi
    
    echo "Uploading DEB package $pkg version $version to Bintray $repo"

    curl -T $pkg -u $BINTRAY_USER:$BINTRAY_KEY "https://api.bintray.com/content/odo/${repo}/odo/${version}/${filename};deb_distribution=stretch;deb_component=main;deb_architecture=amd64;publish=1"
    echo ""
    echo ""
done

#  for rpm
for pkg in `ls -1 $PKG_DIR/*.rpm`; do 
    filename=$(basename $pkg)
    # get version from filename
    version=$(expr "$filename" : '.*-\(.*-[0-9]*\)\.x86_64.*')

    repo="odo-rpm-dev"
    # if version is semver format upload to releases
    if [[ $version =~ [0-9]+\.[0-9]+\.[0-9]+ ]] ; then 
        repo="odo-rpm-releases"
    fi
    
    echo "Uploading RPM package $pkg version $version to Bintray $repo"
    curl -T $pkg -u $BINTRAY_USER:$BINTRAY_KEY "https://api.bintray.com/content/odo/${repo}/odo/${version}/${filename};publish=1"
    echo ""
    echo ""
done
