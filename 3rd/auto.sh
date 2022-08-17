#!/bin/bash

set -ex

action=$1
version=$(cat VERSION)
cd $(dirname $0)

build_openvswitch() {
  obj_dir=$(pwd)/../build/obj
  cd ovs && {
    [ -e './configure' ] || ./boot.sh
    [ -e './Makefile' ] || ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var --disable-libcapng --disable-ssl
    make -j4 && make install DESTDIR=$obj_dir
    cd -
  }
}

build_openvpn() {
  obj_dir=$(pwd)/../build/obj
  cd openvpn && {
    [ -e './configure' ] || autoreconf -i -v -f
    [ -e './Makefile' ] || ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var
    make -j4 && make install DESTDIR=$obj_dir
    cd -
  }
}


clean_openvswitch() {
  cd ovs && {
    if [ -e Makefile ]; then
      make clean
      rm ./Makefile
    fi
    cd -
  }
}

clean_openvpn() {
  cd openvpn && {
    if [ -e Makefile ]; then
      make clean
      rm ./Makefile
    fi
    cd -
  }
}


if [ "$action"x == "build"x ] || [ "$action"x == ""x ]; then
  build_openvswitch
  build_openvpn
elif [ "$action"x == "clean"x ]; then
  clean_openvswitch
  clean_openvpn
fi
