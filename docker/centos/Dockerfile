FROM centos:7

ARG linux_bin

WORKDIR /root

RUN yum update -y && yum install -y epel-release

ADD ${linux_bin} /tmp
RUN DOCKER=yes /tmp/${linux_bin}
RUN rm -rvf /tmp/* && yum clean dbcache

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
