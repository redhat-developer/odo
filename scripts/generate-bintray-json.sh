#!/bin/bash

DATE=`date --iso-8601=date`
TIME=`date --iso-8601=seconds`


# generate ./dist/bin/info.txt
# this file contains information about what files were build and what was source of build

commit_id=$(git rev-parse HEAD)
origin=$(git config --get remote.origin.url)

cat > "./dist/bin/info.txt" <<EOF
date: ${TIME}
build_from: ${origin}
commit_id: ${commit_id}
files:
EOF

for f in $(ls -1 ./dist/bin/*/* | grep -v info.txt); do
  sha256sum=$(sha256sum $f | cut -d ' ' -f 1);
  name=$(echo $f | sed "s/^\.\/dist\/bin\///")
  updated_on=$(stat -c %y $f)
  cat >> "./dist/bin/info.txt" <<EOF
  - name: ${name}"
    sha256sum: ${sha256sum}"
    updated_on: ${updated_on}
EOF
done


PACKAGE="odo"
VERSION="latest"
UPLOAD_DIR="latest"
DESC="odo build for ${TRAVIS_BRANCH} branch"
if [[ "${TRAVIS_PULL_REQUEST}" == "false" ]]; then
    DESC="odo build for PR #${TRAVIS_PULL_REQUEST}"
    UPLOAD_DIR="pr${TRAVIS_PULL_REQUEST}"
    VERSION="pr"
fi



# generate .bintray.json
# this file contains all information on what will be upload to bintray
# for mor info: https://docs.travis-ci.com/user/deployment/bintray/
cat > "./.bintray.json" <<EOF
{
    "package": {
        "name": "${PACKAGE}",
        "repo": "odo",
        "subject": "odo",
        "desc": "OpenShift Command line for Developers",
        "website_url": "https://github.com/redhat-developer/odo",
        "issue_tracker_url": "https://github.com/redhat-developer/odo/issues",
        "vcs_url": "https://github.com/redhat-developer/odo.git",
        "licenses": ["Apache-2.0"],
        "public_download_numbers": false,
        "public_stats": false
    },

    "version": {
        "name": "${VERSION}",
        "desc": "${DESC}",
        "released": "${DATE}",
        "vcs_tag": "${TRAVIS_COMMIT}",
        "attributes": [{"name": "TRAVIS_JOB_NUMBER", "values" : ["${TRAVIS_JOB_NUMBER}"], "type": "string"},
                       {"name": "TRAVIS_JOB_ID", "values" : ["${TRAVIS_JOB_ID}"], "type": "string"},
                       {"name": "TRAVIS_COMMIT", "values" : ["${TRAVIS_COMMIT}"], "type": "string"},
                       {"name": "TRAVIS_BRANCH", "values" : ["${TRAVIS_BRANCH}"], "type": "string"},
                       {"name": "TRAVIS_PULL_REQUEST", "values" : ["${TRAVIS_PULL_REQUEST}"], "type": "string"},
                       {"name": "date", "values" : ["${TIME}"], "type": "date"}],
        "gpgSign": false
    },

    "files":
        [
            {"includePattern": "dist\/bin\/([^\/]+)\/(.*)",
             "uploadPattern": "./${UPLOAD_DIR}/\$1/\$2", 
             "matrixParams": {"override": 1 }
            },
            {"includePattern": "dist/bin/info.txt",
             "uploadPattern": "./${UPLOAD_DIR}/info.txt",
             "matrixParams": {"override": 1 }
            }
        ],
    "publish": true
}
EOF
