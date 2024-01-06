#!/bin/bash

set -ex

version=$(./dist/version.sh)
package=openlan-$version

# build dist.tar
if tmp=$(mktemp -d); then
  rsync -r --exclude '.git' . $tmp/$package
  cd $tmp
  tar cf - $package | gzip -c > ~/rpmbuild/SOURCES/$package-source.tar.gz
  rm -rf $tmp
fi
