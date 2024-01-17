FROM debian:buster

ARG BIN

WORKDIR /root

RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list
RUN apt-get update -y

ADD ${BIN} /tmp
RUN DOCKER=yes /tmp/${BIN}

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
