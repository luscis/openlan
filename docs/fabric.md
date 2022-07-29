# Setup Fabric Network
We using 192.168.100.0/24 to emulate a internet network. And underlay: 100.65.0.0/24 over Internet by IPSec: SPI-117118/117119/118119.
```       
                                                      192.168.100.117
                                                              |
                                                              |
                                                              |
                                                         +---------+    
                                                         | dev-117 |
                                                         +---------+
                                                            /   \
                                                          /       \
                                   SPI-117118           /           \         SPI-117119
                                                      /               \
                                                    /                   \
                                                  /                       \
                                             +---------+              +---------+     
                                             | dev-118 | ------------ | kvm-119 |
                                             +---------+              +---------+
                                                  |                       | 
                                                  |        SPI-118119     |
                                                  |                       |
                                            192.168.100.118          192.168.100.119
                            
                                          
                            
```

Data Center Interconnect with Subnet 192.168.30-40.0/24 Over IPSec network: 100.65.0.0/24 by VxLAN/STT.
```

                                                          100.65.0.117
                                                                |
                                               eth1.200 ---     |    --- eth1.100
                                                             \  |  /
                                                           +---------+  
                                                           | dev-117 | 
                                                           +---------+ 
                                                              /   \
                                                            /       \                   
                                                          /           \                 
                                                        /               \
                                enp2s4.100 ---        /                   \        --- eth4.30
                                               \    /                       \     /
                                              +---------+               +---------+
                                              | dev-118 | ------------- | kvm-119 |
                                              +---------+               +---------+
                                                /    |                      |   \
                                enp2s4.101 ---       |                      |     --- eth4.200
                                                     |                      |
                                              100.65.0.118            100.65.0.119
                                              
                               
                       VNI-1023 192.168.30.0/24 [dev-117_eth1.100, dev-118_enp2s4.100, kvm-119_eth4.30]
                       VNI-1024 192.168.40.0/24 [dev-117_eth1.200, dev-118_enp2s4.101, kvm-119_eth4.200]
                            
```

## Install Software
```
[root@dev-117 network]# yum install -y epel-release
[root@dev-117 network]# yum install -y centos-release-openstack-train
[root@dev-117 network]# yum install -y libibverbs bridge-utils iproute openvswitch
[root@dev-117 network]# 
[root@dev-117 network]# systemctl enable openvswitch
[root@dev-117 network]# systemctl start openvswitch
[root@dev-117 network]# ovs-vsctl show
6bea41ef-b177-4e5c-81b4-fe1f8b90cbac
    Bridge br-tun
        fail_mode: secure
        Port "vx-100650118"
            Interface "vx-100650118"
                type: vxlan
                options: {df_default="false", dst_port="4789", key=flow, remote_ip="100.65.0.118"}
        Port "vnt-3ff"
            Interface "vnt-3ff"
        Port br-tun
            Interface br-tun
                type: internal
        Port "vx-100650117"
            Interface "vx-100650119"
                type: vxlan
                options: {df_default="false", dst_port="4789", key=flow, remote_ip="100.65.0.119"}
        Port "vnt-400"
            Interface "vnt-400"
    ovs_version: "2.12.0"

```

## Configuration on Node: dev-117
```
[root@dev-117 network]# cat ./esp.json
{
    "name": "esp",
    "provider": "esp",
    "specifies": {
        "address": "100.65.0.117",
        "members": [
            {
                "peer": "100.65.0.118",
                "spi": 117118,
                "state": {
                    "remote": "192.168.100.118"
                }
            },
            {
                "peer": "100.65.0.119",
                "spi": 117119,
                "state": {
                    "remote": "192.168.100.119"
                }
            }
        ]
    }
}
[root@dev-117 network]# 
[root@dev-117 network]# cat ./fabric.json
{
    "name": "fabric",
    "provider": "fabric",
    "bridge": {
        "name": "br-tun"
    },
    "specifies": {
        "mss": 1332,
        "tunnels": [
            {
                "dport": 4789,
                "remote": "100.65.0.118"
            },
            {
                "dport": 4789,
                "remote": "100.65.0.119"
            }
        ],
        "networks": [
            {
                "vni": 1023,
                "bridge": "br-100",
                "outputs": [
                    {
                        "vlan": 100,
                        "interface": "eth1"
                    }
                ]
            },
            {
                "vni": 1024,
                "outputs": [
                    {
                        "vlan": 200,
                        "interface": "eth1"
                    }
                ]
            }
        ]
    }
}
[root@dev-117 network]# 
[root@dev-117 network]# ip route
100.65.0.118 via 100.65.0.117 dev spi117118 
100.65.0.119 via 100.65.0.117 dev spi117119  
192.168.30.0/24 dev br-100 proto kernel scope link src 192.168.30.117 
192.168.40.0/24 dev br-400 proto kernel scope link src 192.168.40.117 
192.168.100.0/24 dev eth2 proto kernel scope link src 192.168.100.117 
[root@dev-117 network]# 
[root@dev-117 network]# 

```

## Configuration on Node: dev-118
```
[root@dev-118 network]# cat ./fabric.json
{
    "name": "fabric",
    "provider": "fabric",
    "bridge": {
        "name": "br-tun"
    },
    "specifies": {
        "tunnels": [
            {
                "remote": "100.65.0.117"
            },
            {
                "remote": "100.65.0.119"
            }
        ],
        "networks": [
            {
                "vni": 1023,
                "bridge": "br-100",
                "outputs": [
                    {
                        "vlan": 100,
                        "interface": "enp2s4"
                    }
                ]
            },
            {
                "vni": 1024,
                "outputs": [
                    {
                        "vlan": 101,
                        "interface": "enp2s4"
                    }
                ]
            }
        ]
    }
}
[root@dev-118 network]# 
[root@dev-118 network]# ip route
100.65.0.117 via 100.65.0.118 dev spi117118 
100.65.0.119 via 100.65.0.118 dev spi118119 
192.168.30.0/24 dev br-100 proto kernel scope link src 192.168.30.118 
192.168.40.0/24 dev br-400 proto kernel scope link src 192.168.40.118 
192.168.100.0/24 dev enp2s3 proto kernel scope link src 192.168.100.118 metric 101 
[root@dev-118 network]# 

```

## Configuration on Node: kvm-119
```
[root@kvm-119 switch]# cat ./network/esp.json
{
    "name": "esp",
    "provider": "esp",
    "specifies": {
        "address": "100.65.0.119",
        "members": [
            {
                "peer": "100.65.0.117",
                "spi": 117119,
                "state": {
                    "remote": "192.168.100.117"
                }
            },
            {
                "peer": "100.65.0.118",
                "spi": 118119,
                "state": {
                    "remote": "192.168.100.118"
                }
            }
        ]
    }
}
[root@kvm-119 switch]# 
[root@kvm-119 switch]# cat ./network/fabric.json
{
    "name": "fabric",
    "provider": "fabric",
    "bridge": {
        "name": "br-tun"
    },
    "specifies": {
        "tunnels": [
            {
                "dport": 4789,
                "remote": "100.65.0.117"
            },
            {
                "dport": 4789,
                "remote": "100.65.0.118"
            }
        ],
        "networks": [
            {
                "vni": 1023,
                "bridge": "br-100",
                "outputs": [
                    {
                        "vlan": 30,
                        "interface": "eth4"
                    }
                ]
            },
            {
                "vni": 1024,
                "outputs": [
                    {
                        "vlan": 200,
                        "interface": "eth4"
                    }
                ]
            }
        ]
    }
}

[root@kvm-119 switch]# 
[root@kvm-119 switch]# ip route
100.65.0.117 via 100.65.0.119 dev spi117119 
100.65.0.118 via 100.65.0.119 dev spi118119  
192.168.30.0/24 dev br-100 proto kernel scope link src 192.168.30.119 
192.168.40.0/24 dev br-400 proto kernel scope link src 192.168.40.119 
192.168.100.0/24 dev eth1 proto kernel scope link src 192.168.100.119 
[root@kvm-119 switch]# 
```
