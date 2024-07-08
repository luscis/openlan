FROM centos:7

ARG linux_bin

WORKDIR /root

# mirrorlist for centos7 is EOF.
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/*.repo && \
    sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/*.repo
RUN yum update -y && yum install -y epel-release

ADD ${linux_bin} /tmp
RUN DOCKER=yes /tmp/${linux_bin}
RUN rm -rvf /tmp/* && yum clean all

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
