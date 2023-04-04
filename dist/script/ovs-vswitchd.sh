#!/bin/bash

set -ex

/usr/sbin/ovs-vswitchd unix:/run/openvswitch/db.sock \
  -vconsole:info -vsyslog:off -vfile:off --mlockall \
  --pidfile
