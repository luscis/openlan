#!/bin/bash

set -ex

tmp=""
installer="$0"
archive=$(grep -a -n "__ARCHIVE_BELOW__:$" $installer | cut -f1 -d:)

function download() {
  tmp=$(mktemp -d)
  tail -n +$((archive + 1)) $installer | gzip -dc - | tar -xf - -C $tmp
}

function requires() {
  if type yum > /dev/null; then
    yum install -y xl2tpd openssl net-tools iptables iputils openvpn openvswitch dnsmasq bridge-utils iperf3 tcpdump ipset
  elif type apt > /dev/null; then
    apt-get install -y xl2tpd net-tools iptables iproute2 openvpn openvswitch-switch dnsmasq bridge-utils iperf3 tcpdump ipset
  else
    echo "We didn't find any packet tool: yum or apt."
  fi
}

function install() {
  local source=$(find $tmp -maxdepth 1 -name 'openlan-*')
  cd $source && {
    /usr/bin/env \cp -rf ./{etc,usr,var} /
    /usr/bin/env find ./ -type f > /usr/share/openlan.db
  }
}

function post() {
  if [ x"$DOCKER" == x"no" ] || [ x"$DOCKER" == x"" ]; then
    sysctl -p /etc/sysctl.d/90-openlan.conf
  fi
  [ -e "/etc/openlan/switch/switch.json" ] || {
    cp -rf /etc/openlan/switch/switch.json.example /etc/openlan/switch/switch.json
  }
  [ -e "/var/openlan/openvpn/dh.pem" ] || {
    openssl dhparam -out /var/openlan/openvpn/dh.pem 2048
  }
  [ -e "/var/openlan/openvpn/ta.key" ] || {
    openvpn --genkey --secret /var/openlan/openvpn/ta.key
  }
  [ -e "/etc/openlan/switch/confd.db" ] || {
    /usr/bin/ovsdb-tool create /etc/openlan/switch/confd.db /var/openlan/confd.schema.json
  }
  [ ! -e "/var/openlan/confd/confd.sock" ] || {
    /usr/bin/ovsdb-client convert unix:///var/openlan/confd/confd.sock /var/openlan/confd.schema.json
  }
}

function finish() {
  rm -rf $tmp
  if [ x"$DOCKER" == x"no" ] || [ x"$DOCKER" == x"" ]; then
    systemctl daemon-reload
  fi
  echo "success"
}

download
requires
install
post
finish
exit 0
