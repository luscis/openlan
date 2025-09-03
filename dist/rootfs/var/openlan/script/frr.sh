#!/bin/bash

if [ -r "/lib/lsb/init-functions" ]; then
	. /lib/lsb/init-functions
else
	log_success_msg() {
		echo "$@"
	}
	log_warning_msg() {
		echo "$@" >&2
	}
	log_failure_msg() {
		echo "$@" >&2
	}
fi

set -ex

sed -i 's/is\ not\ "/\!=\ "/' /usr/lib/frr/frr-reload.py

if [ ! -e "/var/run/frr" ]; then
	 mkdir -p /var/run/frr
	 chown frr:frr /var/run/frr
fi

for file in daemons frr.conf vtysh.conf; do
	if [ -e "/etc/frr/$file" ]; then
		continue
	fi
	if [ -e "/usr/share/frr/$file" ]; then
		cp -rf "/usr/share/frr/$file" "/etc/frr/$file"
	fi
done

# Start reloader server for FRR
exec /var/openlan/script/frr-server &

set +x

# Start daemons
source /usr/lib/frr/frrcommon.sh
/usr/lib/frr/watchfrr $(daemon_list)