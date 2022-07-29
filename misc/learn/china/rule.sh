#!/bin/bash

set -ex

ips="china.list"

origin="3398"
outside="3396"

originGw="192.168.7.1"
outsideGw="192.168.10.11"

clean_rule() {
	local table=$1
	local tmp="${table}.rules"

	ip rule show | grep "lookup ${table}" | awk -F ':' '{print $2}' > $tmp
	while read -r line; do ip rule del ${line}; done < $tmp
}

ip route flush table ${outside}
ip route add default via ${outsideGw} table ${outside}

ip route flush table ${origin}
ip route add default via ${originGw} table ${origin}

clean_rule ${outside}
ip rule add from 172.33.196.0/24 lookup ${outside}

clean_rule ${origin}
for i in $(cat ${ips}); do 
	ip rule add from 172.33.196.0/24 to $i lookup ${origin}; 
done
