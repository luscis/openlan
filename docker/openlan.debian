FROM debian:buster

ARG BIN

WORKDIR /root

ADD ${BIN} /tmp

RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list
RUN apt-get update -y

RUN DOCKER=yes /tmp/${BIN}

RUN rm -rf /tmp/*

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
