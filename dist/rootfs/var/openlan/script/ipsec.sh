#!/bin/bash

set -ex

## Clear xfrm
/sbin/ip xfrm policy flush
/sbin/ip xfrm state flush

## Checking ipsec
/usr/libexec/ipsec/addconn --config /etc/ipsec.conf --checkconfig

/usr/libexec/ipsec/_stackmanager start
/usr/sbin/ipsec --checknss

## Start pluto
/usr/libexec/ipsec/pluto --leak-detective --config /etc/ipsec.conf --nofork