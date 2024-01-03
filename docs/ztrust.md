# Zero Trust 

## Enable ztrust on a network
```
$ cat /etc/openlan/switch/network/example.json
{
	...
	"ztrust": "enable"
}
$
$ systemctl restart openlan-switch
$
```

## Access network via OpenVPN

* Open your OpenVPN Connect application;
* Click `Import Profile` button and Select `via URL`;
* Input value: `https://<your-central-switch-address>:10000`, Click `Next`;
* Input your name: `daniel@exmaple` and password: `18a102852f28`;
* Click `Connect` button to access network: `example`.

## Add yourself to ztrust
```
$ export TOKEN="daniel@example:<password>"
$ export URL="https://<your-central-switch-address>:10000"
$ openlan guest add
$ openlan guest ls
# total 1
username                 address
daniel@example          169.254.15.6
$
```

## Knock a host service
```
$ openlan knock add --protocol icmp --socket 192.168.20.10
$ openlan knock add --protocol tcp --socket 192.168.20.10:22
$ openlan knock ls
# total 2
username                 protocol socket                   age  createAt
daniel@example          tcp      192.168.20.10:22         57   2024-01-02 12:42:06 +0000 UTC
daniel@example          icmp     192.168.20.10:           46   2024-01-02 12:41:55 +0000 UTC
$
```

## Connect to a host service
```
$ ping 192.168.20.10 -c 3
PING 192.168.20.10 (192.168.20.10): 56 data bytes
64 bytes from 192.168.20.10: icmp_seq=0 ttl=63 time=5.969 ms
64 bytes from 192.168.20.10: icmp_seq=1 ttl=63 time=6.317 ms
64 bytes from 192.168.20.10: icmp_seq=2 ttl=63 time=5.694 ms

--- 192.168.20.10 ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 5.694/5.993/6.317/0.255 ms
$
$ ssh root@192.168.20.10 hostname
hostservice.luscis
$
```