#!/bin/bash

set -x

# probe kernel mod.
/usr/bin/env modprobe bridge
/usr/bin/env modprobe br_netfilter
/usr/bin/env modprobe xfrm4_mode_tunnel
/usr/bin/env modprobe vxlan

# clean older files.
/usr/bin/env find /var/openlan/point -type f -delete
/usr/bin/env find /var/openlan/openvpn -name '*.status' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.ovpn' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.tmpl' -delete

# upgrade database.
