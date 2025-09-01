#!/bin/bash

sed -i 's/is\ not\ "/\!=\ "/' /usr/lib/frr/frr-reload.py

if [ ! -e /var/run/frr ]; then
	 mkdir -p /var/run/frr
	 chown frr:frr /var/run/frr
fi

for file in daemons frr.conf vtysh.conf; do
	if [ -e "/etc/frr/$file" ]; then
		continue
	fi
	if [ -e "/var/openlan/frr/$file" ]; then
		cp -rvf /var/openlan/frr/$file /etc/frr/$file
	fi
done

# Start reloader server for FRR
exec /var/openlan/script/frr-server &

# Start daemons
source /usr/lib/frr/frrcommon.sh
/usr/lib/frr/watchfrr $(daemon_list)