# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13

RUN yum -y install sudo make wget gcc git httpd-tools ruby ruby-devel gcc-c++ patch \
    readline readline-devel zlib zlib-devel libffi-devel openssl-devel make bzip2 \
    autoconf automake libtool bison sqlite-devel

RUN sudo curl -sSL https://rvm.io/mpapis.asc | gpg2 --import - \
    sudo curl -sSL https://rvm.io/pkuczynski.asc | gpg2 --import -

RUN curl -L get.rvm.io | bash -s stable

RUN source /etc/profile.d/rvm.sh \
    rvm reload

RUN rvm requirements run \
    rvm install 2.7

RUN ruby --version

RUN gem install travis --pre --no-document

