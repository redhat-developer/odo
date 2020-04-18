# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13

RUN yum -y install make wget gcc git httpd-tools docker
RUN sudo systemctl start docker

RUN mkdir -p /tmp/secret