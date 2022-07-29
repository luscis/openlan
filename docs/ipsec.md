Topology
========
We use 192.168.7.0/24 as underlay network for IPSec. And S1 has public address with 192.168.7.11, C1 and C2 under firewall without public address.

                                             +----+
                                             | s1 |     -- .10.1/24
                                             +----+
                                             /    \
                                           /        \
                                         /            \
                                      +----+          +----+
               192.168.2.0/24    --   | c2 |          | c3 |  -- 192.168.3.0/24
                                      +----+          +----+
                                        |               |
                                     .10.2/32        .10.3/32

Server
======
```
$ openlan network add --name ipsec --provider esp --address 10.10.10.1/24
$ openlan link add --network ipsec --device spi:12 --remote-address 10.10.10.2
$ openlan link add --network ipsec --device spi:13 --remote-address 10.10.10.3
```
```
$ openlan route add --network ipsec --prefix 192.168.2.0/24 --gateway spi:12
$ openlan route add --network ipsec --prefix 192.168.3.0/24 --gateway spi:13
```

Client
======

C2
--
```
$ openlan network add --name ipsec --provider esp --address 10.10.10.2
$ openlan link add --network ipsec --connection udp:192.168.7.11 --device spi:12 --remote-address 10.10.10.1/24
$ openlan link ls
```
```
$ ping 10.10.10.1
```
```
$ openlan route add --network ipsec --prefix 192.168.3.0/24 --gateway spi:12
```

C3
--

```
$ openlan network add --name ipsec --provider esp --address 10.10.10.3
$ openlan link add --network ipsec --connection udp:192.168.7.11 --device spi:13 --remote-address 10.10.10.1/24
```
```
$ ping 10.10.10.2
```
```
$ openlan route add --network ipsec --prefix 192.168.2.0/24 --gateway spi:13
```