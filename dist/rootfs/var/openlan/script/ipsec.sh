#!/bin/bash

set -ex

/sbin/ip xfrm policy flush
/sbin/ip xfrm state flush

/usr/libexec/ipsec/addconn --config /etc/ipsec.conf --checkconfig
/usr/libexec/ipsec/pluto --leak-detective --config /etc/ipsec.conf --nofork