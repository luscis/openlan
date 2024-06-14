#!/bin/bash

set -e

tmp=""
installer="$0"
nodeps="no"
if [ "$1"x == "nodeps"x ]; then
  nodeps="yes"
fi

archive=$(grep -a -n "__ARCHIVE_BELOW__:$" $installer | cut -f1 -d:)

OS="linux"
if type yum 2> /dev/null; then
  OS="centos"
elif type apt 2> /dev/null; then
  OS="debian"
fi

function download() {
  echo "Uncompress files ..."
  tmp=$(mktemp -d)
  tail -n +$((archive + 1)) $installer | gzip -dc - | tar -xf - -C $tmp
}

function requires() {
  echo "Install dependents ..."
  if [ "$OS"x == "centos"x ]; then
    yum install -y openssl net-tools iptables iputils iperf3 tcpdump
    yum install -y openvpn dnsmasq bridge-utils ipset libreswan procps
  elif [ "$OS"x == "debian"x ]; then
    apt install -y net-tools iptables iproute2 tcpdump ca-certificates iperf3
    apt install -y openvpn dnsmasq bridge-utils ipset procps wget
    wget -O /tmp/libreswan_4.10-1_amd64.deb https://github.com/luscis/packages/raw/main/debian/bullseye/libreswan_4.10-1_amd64.deb
    apt install -y /tmp/libreswan_4.10-1_amd64.deb
  else
    echo "We didn't find any packet tool: $OS"
  fi
}

function install() {
  echo "Installing files ..."
  local source=$(find $tmp -maxdepth 1 -name 'openlan-*')
  pushd $source
  /usr/bin/env \cp -rf ./{etc,usr,var} /
  chmod +x /var/openlan/script/*.sh
  /usr/bin/env find ./ -type f > /usr/share/openlan.db
  popd
}

function post() {
  echo "Initializing ..."
  if [ x"$DOCKER" == x"no" ] || [ x"$DOCKER" == x"" ]; then
    sysctl -p /etc/sysctl.d/90-openlan.conf
  fi

  if [ "$OS"x == "centos"x ]; then
    ## Prepare openvpn.
    [ -e "/var/openlan/openvpn/dh.pem" ] || {
      openssl dhparam -out /var/openlan/openvpn/dh.pem 1024
    }
    [ -e "/var/openlan/openvpn/ta.key" ] || {
      openvpn --genkey --secret /var/openlan/openvpn/ta.key
    }
    ## Install CA.
    cp -rf /var/openlan/cert/ca.crt /etc/pki/ca-trust/source/anchors/OpenLAN_CA.crt
    update-ca-trust
  elif [ "$OS"x == "debian"x ]; then
    ## Prepare openvpn.
    [ -e "/var/openlan/openvpn/dh.pem" ] || {
      openssl dhparam -out /var/openlan/openvpn/dh.pem 2048
    }
    [ -e "/var/openlan/openvpn/ta.key" ] || {
      openvpn --genkey > /var/openlan/openvpn/ta.key
    }
    ## Install CA.
    cp -rf /var/openlan/cert/ca.crt /usr/local/share/ca-certificates/OpenLAN_CA.crt
    update-ca-certificates
  fi
}

function finish() {
  rm -rf $tmp
  if [ x"$DOCKER" == x"no" ] || [ x"$DOCKER" == x"" ]; then
    systemctl daemon-reload
  fi
  echo "Finished ..."
}


download
if [ "$nodeps"x == "no"x ]; then
  requires
fi
install
if [ "$nodeps"x == "no"x ]; then
  post
fi
finish
exit 0
