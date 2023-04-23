#!/bin/bash

set -ex

## Upgrade ovsdb
# ovsdb-client convert unix:///var/openlan/confd/confd.sock /var/openlan/confd.schema.json

## Enable check for DDNS
# [root@centos ~]# crontab -l
# 0,5,10,15,20,25,30,35,40,45,50,55 * * * * /var/openlan/script/update.sh
# [root@centos ~]#

## Update your DDNS
export VERSION=v6
names=$(openlan na ls | grep -w 'name:' | sed 's/name://g')
for name in $names; do
  openlan name add --name $name
done
