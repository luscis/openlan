FROM debian:buster

ARG linux_bin

WORKDIR /root

#RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list
RUN apt-get update -y

ADD ${linux_bin} /tmp
RUN DOCKER=yes /tmp/${linux_bin}

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
