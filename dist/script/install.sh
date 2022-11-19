#!/bin/bash

set -e

tmp=""
installer="$0"
archive=$(grep -a -n "__ARCHIVE_BELOW__:$" $installer | cut -f1 -d:)

function download() {
  tmp=$(mktemp -d)
  tail -n +$((archive + 1)) $installer | gzip -dc - | tar -xf - -C $tmp
}

function requires() {
  if type yum > /dev/null; then
    yum install -y net-tools iptables iputils openvpn openvswitch dnsmasq
  elif type apt > /dev/null; then
    apt-get install -y net-tools iptables iproute2 openvpn openvswitch-switch dnsmasq
  else
    echo "We didn't find yum and apt."
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
    /usr/bin/ovsdb-tool create /etc/openlan/switch/confd.db /etc/openlan/switch/confd.schema.json
  }
  [ -e "/var/openlan/confd.sock" ] && {
    /usr/bin/ovsdb-client convert unix:///var/openlan/confd.sock /etc/openlan/switch/confd.schema.json
  }
  sysctl -p /etc/sysctl.d/90-openlan.conf
}

function finish() {
  rm -rf $tmp
  systemctl daemon-reload
  echo "success"
}

download
requires
install
post
finish
exit 0
