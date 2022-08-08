# PPPOE Client
```
$ yum -y install rp-pppoe
```

## Configure
```
$ cp /usr/share/doc/rp-pppoe-3.11/configs/pppoe.conf  /etc/ppp
$ cat /etc/ppp/chap-secrets
# Secrets for authentication using CHAP
# client	server	secret			IP addresses
username	*	password            *

$ 
$ cat /etc/ppp/pppoe.conf  | grep -e ETH -e USER
ETH=eth1
USER=username
$
```

## Start
```
$ pppoe-start
$ pppoe-status
pppoe-status: Link is up and running on interface ppp0
6: ppp0: <POINTOPOINT,MULTICAST,NOARP,UP,LOWER_UP> mtu 1480 qdisc pfifo_fast state UNKNOWN group default qlen 3
link/ppp
inet 192.168.33.83 peer 192.168.33.1/32 scope global ppp0
valid_lft forever preferred_lft forever

$ iptables -t nat -A POSTROUTING -o ppp0 -j MASQUERADE
```

# PPPOE Server
```
apt-get install pppoe
```

## Configure
```
$ cat /etc/ppp/options | grep -e +chap -e -pap -e dns
ms-dns 192.168.33.1
ms-dns 192.168.33.2
-pap
+chap
$ cat > /etc/ppp/pppoe-server-options << 'EOF'
# PPP options for the PPPoE server
require-chap
lcp-echo-interval 60
lcp-echo-failure 5
logfile /var/log/pppd.log
EOF

$ cat /etc/ppp/chap-secrets
# Secrets for authentication using CHAP
# client	server	secret			IP addresses
username	*	password            *

```

## Start 
```
modprobe pppoe
pppoe-server -I br-private -L 192.168.33.1 -R 192.168.33.20 -N 20
```
