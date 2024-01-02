
简体中文 | [English](./README.en.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan)
[![Codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![CodeQL](https://github.com/luscis/openlan/actions/workflows/codeql.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/codeql.yml)
[![Build](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://github.com/luscis/openlan/tree/master/docs)
[![Releases](https://img.shields.io/github/release/luscis/openlan/all.svg?style=flat-square)](https://github.com/luscis/openlan/releases)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

## 什么是OpenLAN？

OpenLAN提供一种局域网数据报文在广域网的传输实现，并能够建立多个用户空间的虚拟以太网络。

## 为什么是OpenLAN？

如果你有更加灵活的VPN业务需求，需要使用VPN访问企业内部，或者借用公网云主机等进行网络代理、网络穿透等，可以试试OpenLAN，它能让部署变得更简单。

## OpenLAN有什么功能？

* 用户可以使用OpenLAN划分多个网络空间，为不同的业务提供逻辑网络隔离；
* 多个Central Switch之间可以使用OpenLAN协议在链路层上互联互通，在链路网络上可以添加SNAT路由轻松的访问企业内部网络；
* 用户可以使用OpenVPN接入用户网络，OpenVPN支持多平台如Android/MacOS/Windows等；
* 多个Central Switch之间也可以使用IPSec隧道网络，并且支持在该网络上进一步划分VxLAN/STT的租户网络；
* 使用简单的用户名密码的作为接入认证方式，并且可以设置预共享密钥对数据报文进行加密；
* OpenLAN协议可以工作在TCP/TLS/UDP/KCP/WS/WSS等多种传输协议上，TCP具有较高的性能，TLS/WSS能够提供更好的加密安全；
* OpenLAN也提供了简单的HTTP/HTTPS/SOCKS5等HTTP的正向代理技术，用户可以根据需要灵活配置代理进行网络穿透；


## OpenLAN的工作场景？
### 分支中心接入

                           Central Switch(企业中心) - 10.16.1.10/24
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
         OpenLAN                  OpenLAN                      OpenLAN
      10.16.1.11/24             10.16.1.12/24                10.16.1.13/24
       

### 多区域互联

     192.168.1.20/24                                                  192.168.1.21/24
            |                                                                |
        OpenLAN -- 酒店 Wifi --> Central Switch(南京) <--- 其他 Wifi --- OpenLAN
                                         |
                                         |
                                       互联网
                                         |
                                         |
                                 Central Switch(上海) - 192.168.1.10/24
                                         |
                                         |
                ------------------------------------------------------
                ^                        ^                           ^
                |                        |                           |
             办公 Wifi               家庭 Wifi                 酒店 Wifi     
                |                        |                           |
            OpenLAN                  OpenLAN                     OpenLAN
        192.168.1.11/24           192.168.1.12/24             192.168.1.13/24

### 数据中心全互联网络

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


## 帮助文档
- [软件安装](docs/install.md)
- [分支接入](docs/central.md)
- [多区域互联](docs/multiarea.md)
- [全互连网络](docs/fabric.md)
- [IPSec网络](docs/ipsec.md)
- [零信任网络](docs/ztrust.md)
