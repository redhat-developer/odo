#/bin/bash

set -eo pipefail

if [ "$#" -lt 2 ]; then
  echo "Wrong number of arguments. Usage: ./build-registry.sh /path/to/registry/dir /path/to/empty/build/dir"
  exit 1
fi

registryDir=$1
outputDir=$2

TEMPDIR=$(mktemp -d)
(
    cd ${TEMPDIR} &&
    git clone -b deterministic_stack_tar_archives_in_build_script --depth 1 --single-branch https://github.com/rm3l/devfile-registry-support .
)

bash "${TEMPDIR}"/build-tools/build.sh "$registryDir" "$outputDir"
rm -rf "${TEMPDIR}"
