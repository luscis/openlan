FROM debian:bullseye

ARG linux_bin

WORKDIR /root

# RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list

ADD ${linux_bin} /tmp
RUN /tmp/${linux_bin} --nosysm --verbose
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/*

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/var/openlan/script/switch.sh", "start"]
