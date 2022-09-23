
简体中文 | [English](./README.en.md)

# 概述 
[![codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan-go)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

OpenLAN提供一种局域网数据报文在广域网的传输实现，并能够建立多个用户空间的虚拟以太网络。

## 缩略语

* OLSW: OpenLAN Switch，开放局域网交换机
* OLAP: OpenLAN Access Point，开放局域网接入点
* NAT: Network Address Translation, 网络地址转换
* VxLAN: Virtual eXtensible Local Area Network，虚拟扩展局域网
* STT: Stateless Transport Tunneling，无状态传输隧道

## 功能清单

* 支持多个网络空间划分，为不同的业务提供逻辑网络隔离；
* 支持OLAP或者OpenVPN接入，提供网桥把局域网共享出去；
* 支持IPSec隧道网络，以及基于VxLAN/STT的租户网络划分；
* 支持基于用户名密码的接入认证，使用预共享密钥对数据报文进行加密；
* 支持TCP/TLS，UDP/KCP，WS/WSS等多种传输协议实现，TCP模式具有较高的性能；
* 支持HTTP/HTTPS，以及SOCKS5等HTTP的正向代理技术，灵活配置代理进行网络穿透；
* 支持基于TCP的端口转发，为防火墙下的主机提供TCP端口代理。


## 分支中心接入

                                       OLSW(企业中心) - 10.16.1.10/24
                                                ^
                                                |
                                             Wifi(DNAT)
                                                |
                                                |
                       ----------------------Internet-------------------------
                       ^                        ^                           ^
                       |                        |                           |
                     分支1                    分支2                        分支3     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                 10.16.1.11/24             10.16.1.12/24                10.16.1.13/24
                 

## 多区域互联

                   192.168.1.20/24                                 192.168.1.21/24
                         |                                                 |
                       OLAP -- 酒店 Wifi --> OLSW(南京) <--- 其他 Wifi --- OLAP
                                                |
                                                |
                                             互联网
                                                |
                                                |
                                             OLSW(上海) - 192.168.1.10/24
                                                |
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   办公 Wifi               家庭 Wifi                 酒店 Wifi     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                192.168.1.11/24           192.168.1.12/24             192.168.1.13/24

## 数据中心全互联网络

* Underlay for VxLAN over Internet by IPSec.

                                         47.example.com
                                                |
                                                |
                                                |
                                            +-------+
                                            | vps-47|  -- 100.65.0.117
                                            +-------+
                                              /   \
                                            /       \
                     SPI-117118           /           \         SPI-117119
                                        /               \
                                      /                   \
                                +-------+                +-------+
                                | vps-92| -------------- | vps-12|
                                +-------+                +-------+
                                /   |                       |  \ 
                               /    |    SPI-118119         |   \
                  100.65.0.118      |                       |    100.65.0.119
                                    |                       |
                              92.example.com          12.example.com
                                            
                                            

* DCI Subnet: 192.168.x.x over IPSec Network: 100.65.0.x.

                                          100.65.0.117
                                                |
                               eth1.200 ---     |    --- eth1.100
                                             \  |  /
                                            +--------+
                                            | vps-47 |
                                            +--------+
                                              /   \
                                            /       \                   
                                          /           \                 
                                        /               \
                enp2s4.100 ---        /                   \        --- eth4.30
                               \    /                       \     /
                               +--------+                 +--------+
                               | vps-92 | --------------- | vps-12 |
                               +--------+                 +--------+
                                /    |                      |   \
                enp2s4.101 ---       |                      |     --- eth4.200
                                     |                      |
                              100.65.0.118            100.65.0.119


           VNI-1023 192.168.30.0/24 [vps-47_eth1.100, vps-92_enp2s4.100, vps-12_eth4.30]
           VNI-1024 192.168.40.0/24 [vps-47_eth1.200, vps-92_enp2s4.101, vps-12_eth4.200]


## 文档
- [软件安装](docs/install.md)
- [分支接入](docs/central.md)
- [多区域互联](docs/multiarea.md)
- [全互连网络](docs/fabric.md)
- [IPSec网络](docs/ipsec.md)
