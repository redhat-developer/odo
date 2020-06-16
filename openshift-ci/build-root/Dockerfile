# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.12

RUN yum -y install make wget gcc git httpd-tools

# This is a temporary change to make sure that it should run on both template based as well as multi-stage test
# will remove this change once we completely migrate to multi-stage test infra
COPY ./check-oc.sh .
RUN chmod +x ./check-oc.sh
