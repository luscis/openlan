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
  local os=$(cat /etc/os-release | grep ^ID= | sed 's/"//g')
  if echo $os | grep -q -e centos -e redhat; then
    yum install -y net-tools iptables iputils openvpn openssl openvswitch dnsmasq
  elif echo $os | grep -q -e debian -e ubuntu; then
    apt-get install -y net-tools iptables iproute2 openvpn openssl openvswitch-switch dnsmasq
  fi
}

function install() {
  local source=$(find $tmp -name 'openlan-linux-*')
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
