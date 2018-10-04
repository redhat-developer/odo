#
# This is an "initContainer" image using the base "source-to-image" OpenShift template
# in order to apprpriately inject the supervisord binary into the application container.
#

FROM centos:7

ENV SUPERVISORD_DIR /opt/supervisord

RUN mkdir -p ${SUPERVISORD_DIR}/conf ${SUPERVISORD_DIR}/bin

ADD supervisor.conf ${SUPERVISORD_DIR}/conf/
ADD https://raw.githubusercontent.com/sclorg/s2i-base-container/master/core/root/usr/bin/fix-permissions  /usr/bin/fix-permissions
RUN chmod +x /usr/bin/fix-permissions

ADD https://github.com/ochinchina/supervisord/releases/download/v0.5/supervisord_0.5_linux_amd64 ${SUPERVISORD_DIR}/bin/supervisord

ADD assemble-and-restart ${SUPERVISORD_DIR}/bin
# ADD assemble ${SUPERVISORD_DIR}/bin
# RUN ${SUPERVISORD_DIR}/bin/assemble
ADD run ${SUPERVISORD_DIR}/bin

RUN chgrp -R 0 ${SUPERVISORD_DIR}  && \
    chmod -R g+rwX ${SUPERVISORD_DIR} && \
    chmod -R 666 ${SUPERVISORD_DIR}/conf/* && \
    chmod 775 ${SUPERVISORD_DIR}/bin/supervisord
