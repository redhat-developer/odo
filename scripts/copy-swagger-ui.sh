#!/bin/sh

set -e

TEMPDIR=$(mktemp -d)
(
    cd ${TEMPDIR} &&
    git clone -b master --depth 1 --single-branch https://github.com/swagger-api/swagger-ui/ .
)
rm -rf pkg/apiserver-impl/swagger-ui/*
cp -R ${TEMPDIR}/dist/* pkg/apiserver-impl/swagger-ui/
rm -rf ${TEMPDIR}
sed -i "s|https://petstore.swagger.io/v2/swagger.json|./swagger.yaml|" pkg/apiserver-impl/swagger-ui/swagger-initializer.js 
