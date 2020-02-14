#!/usr/bin/env bash

# Copyright 2019 The Tekton Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

function cleanup {
  tmp_files=$(find docs -name "${TMP_PREFIX}*.md")
  for file in ${tmp_files}; do
    rm "${file}"
  done
}
trap cleanup EXIT

TMP_PREFIX="tmpOverride-"
# Whether this script is being used for validation or not
DIFF=
# Syntax format for code fences
SYNTAX="YAML"
CODE_FENCE='^```'
# Comment that will specify the file to embed within $SYNTAX code fence
COMMENT='<!-- *FILE: *(.*) *-->'
EMPTY='^[[:space:]]*$'

source $(dirname $0)/../vendor/github.com/tektoncd/plumbing/scripts/library.sh
cd ${REPO_ROOT_DIR}

# Parse flags
while getopts ":d" opt; do
  case ${opt} in
    d )
        DIFF="true"
        ;;
    * )
        echo "Invalid Option: -$OPTARG" 1>&2
        exit 1
        ;;
  esac
done
shift $((OPTIND -1))

doc_files="$(find docs -name "*.md")"
for file in ${doc_files};do
    new_file="${file%/*}/${TMP_PREFIX}${file##*/}"
    > ${new_file}
    fenced="false"
    # Read each markdown file for replacements
    while IFS= read -r line; do
        # File has been embedded
        if [[ ${fenced} == "maybe" ]];then
            # Look for proceeding codefence
            if [[ ${line} =~ ${CODE_FENCE} ]];then
                fenced="true"
                continue
            fi
            # If a non-empty line is deteced
            if [[ ! ${line} =~ ${EMPTY} ]];then
                fenced="false"
            fi
        fi
        # Turn off code fencing
        if [[ ${fenced} == "true" ]];then
            if [[ ${line} =~ ${CODE_FENCE} ]];then
                fenced="false"
                continue
            fi
        fi
        # Write to replacement file
        if [[ ${fenced} != "true" ]];then
            # Copy line
            echo "${line}" >> ${new_file}
            # Inline file
            if [[ "$line" =~ ${COMMENT} ]];then
                echo '```'$SYNTAX >> ${new_file}
                cat ${BASH_REMATCH[1]} >> ${new_file}
                echo '```' >> ${new_file}
                fenced="maybe"
            fi
        fi
    done < "${file}"
    
    if [[ ${DIFF} == "true" ]];then
        # Check if up to date
        set +o errexit
        diff -B ${file} ${new_file}
        if [[ $? != 0 ]];then
            echo 'Run `./hack/update-docs.sh` to update the docs'
            exit 1
        fi
    else
        # Overwrite file 
        rm ${file}
        mv ${new_file} ${file}
    fi
done
