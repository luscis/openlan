#!/bin/bash

if [ -e "VERSION" ]; then
	cat VERSION
	exit 0
fi

ver=$(git describe --tags --abbrev=0 --match 'v*')
if [ $? -eq 0 ]; then
	echo $ver
	exit 0
fi

date +%y%m%d
