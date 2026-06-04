FROM debian:bullseye

ARG linux_bin

WORKDIR /root

RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list

RUN apt-get update -y
RUN apt-get install -y net-tools iptables ebtables iproute2 tcpdump ca-certificates iperf3 socat
RUN apt-get install -y openvpn dnsmasq bridge-utils ipset procps wget iputils-ping frr
RUN apt-get install -y libreswan dnsutils haproxy isc-dhcp-client

ADD ${linux_bin} /tmp
RUN /tmp/${linux_bin} --nosysm --verbose
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/*

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
