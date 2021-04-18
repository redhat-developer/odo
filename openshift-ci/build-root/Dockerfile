# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.ci.openshift.org/openshift/release:golang-1.13

RUN yum -y install make wget gcc git httpd-tools curl

# Download kubectl, a cli tool for accessing Kubernetes cluster
RUN curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.16.1/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/


