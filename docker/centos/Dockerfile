FROM centos:7

ARG linux_bin

WORKDIR /root

# mirrorlist for centos7 is EOF.
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/*.repo && \
    sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/*.repo

RUN yum update -y
RUN yum install -y epel-release
RUN yum install -y openssl net-tools iptables iputils iperf3 tcpdump \
    openvpn dnsmasq bridge-utils ipset procps wget socat
RUN wget -O /tmp/libreswan-4.10-1.el7.x86_64.rpm https://github.com/luscis/packages/raw/main/redhat/centos7/libreswan-4.10-1.el7.x86_64.rpm
RUN yum install -y /tmp/libreswan-4.10-1.el7.x86_64.rpm || yum install -y libreswan
RUN wget -O /tmp/frr-stable-repo.el7.noarch.rpm https://rpm.frrouting.org/repo/frr-stable-repo.el7.noarch.rpm
RUN yum install -y /tmp/frr-stable-repo.el7.noarch.rpm
RUN yum install -y frr frr-pythontools

ADD ${linux_bin} /tmp
RUN /tmp/${linux_bin} --nosysm --verbose
RUN yum clean all && rm -rf /var/cache/yum/* /tmp/*

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
