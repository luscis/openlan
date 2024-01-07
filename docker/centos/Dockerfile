FROM centos:7

ARG BIN

WORKDIR /root

ADD ${BIN} /tmp

RUN yum install -y epel-release centos-release-openstack-train \
    && yum install -y rdma-core libibverbs
RUN DOCKER=yes /tmp/${BIN}

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]