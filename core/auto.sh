#!/bin/bash

set -ex

action=$1
version=$(cat VERSION)
cd $(dirname $0)

check_and_update() {
  file0=$1
  file1=$2
  md5f0=$(md5sum $file0 | awk '{print $1}')
  md5f1=$(md5sum $file1 | awk '{print $1}')
  if [ "$md5f0"x != "$md5f1"x ]; then
    mv $file0 $file1
  fi
}

python_bin=python
type $python_bin || python_bin="python3"
ovs_dir="../3rd/ovs"

build_idlc() {
  idlc_bin="$ovs_dir/ovsdb/ovsdb-idlc.in"
  [ -e "idlc/confd.ovsschema" ] || ln -s -f ../../dist/resource/confd.schema.json idlc/confd.ovsschema
  PYTHONPATH="$ovs_dir/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes $python_bin $idlc_bin annotate idlc/confd.ovsschema idlc/confd-idl.ann > /tmp/confd-idl.ovsidl
  check_and_update /tmp/confd-idl.ovsidl idlc/confd-idl.ovsidl
  PYTHONPATH="$ovs_dir/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes $python_bin $idlc_bin c-idl-source idlc/confd-idl.ovsidl > /tmp/confd-idl.c
  check_and_update /tmp/confd-idl.c idlc/confd-idl.c
  PYTHONPATH="$ovs_dir/python:"$PYTHONPATH PYTHONDONTWRITEBYTECODE=yes $python_bin $idlc_bin c-idl-header idlc/confd-idl.ovsidl > /tmp/confd-idl.h
  check_and_update /tmp/confd-idl.h idlc/confd-idl.h
}

update_version() {
  cp version.h /tmp/version.h
  sed -i  "s/#define CORE_PACKAGE_STRING .*/#define CORE_PACKAGE_STRING  \"opencore $version\"/g" /tmp/version.h
  sed -i  "s/#define CORE_PACKAGE_VERSION .*/#define CORE_PACKAGE_VERSION \"$version\"/g" /tmp/version.h
  check_and_update /tmp/version.h version.h
}

if [ "$action"x == "build"x ] || [ "$action"x == ""x ]; then
  update_version
  build_idlc
elif [ "$action"x == "clean"x ]; then
  echo "TODO"
fi
