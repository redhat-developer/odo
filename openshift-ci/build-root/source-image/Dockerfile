FROM src

COPY oc /usr/bin/oc

RUN yum -y install epel-release && yum -y install python-devel python-pip && sudo pip install pika --upgrade
