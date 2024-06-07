English | [简体中文](./README.cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan)
[![Codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![CodeQL](https://github.com/luscis/openlan/actions/workflows/codeql.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/codeql.yml)
[![Build](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://github.com/luscis/openlan/tree/master/docs)
[![Releases](https://img.shields.io/github/release/luscis/openlan/all.svg?style=flat-square)](https://github.com/luscis/openlan/releases)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

## What's OpenLAN?

OpenLAN provides a realization of the transmission of LAN data packets in the WAN, and can establish a virtual Ethernet network in multiple user spaces. 

## Why is OpenLAN?

If you have more flexible VPN business needs and need to use VPN to access the enterprise, or use public network cloud hosts for network proxy and network penetration, you can try OpenLAN, which can make deployment easier.

## What is the function of OpenLAN?

* Users can use OpenLAN to divide multiple network spaces to provide logical network isolation for different services;
* Multiple Central Switchs can use the OpenLAN protocol to communicate on the ethernet layer, and SNAT routes can be added to the second layer network to easily access the internal network of the enterprise;
* Users can use OpenVPN to access the User Network, OpenVPN supports multiple platforms such as Android/MacOS/Windows, etc;
* IPSec tunnel network can also be used between multiple Central Switchs, and it supports further division of VxLAN/STT tenant networks on this network;
* Use a simple username and password as the access authentication method, and you can set a pre-shared key to encrypt data packets;
* The OpenLAN protocol can work on various transmission protocols such as TCP/TLS/UDP/KCP/WS/WSS, TCP has high performance, and TLS/WSS can provide better encryption security;
* OpenLAN also provides simple HTTP/HTTPS/SOCKS5 and other HTTP forward proxy technology, users can flexibly configure proxy for network penetration according to needs;

## Working scenario of OpenLAN?
### Branch central access

                              Central Switch - 10.16.1.10/24
                                      ^
                                      |
                                   Wifi(DNAT)
                                      |
                                      |
             ----------------------Internet-------------------------
             ^                        ^                           ^
             |                        |                           |
           Branch1                  Branch2                     Branch3     
             |                        |                           |
         OpenLAN                  OpenLAN                      OpenLAN
      10.16.1.11/24             10.16.1.12/24                10.16.1.13/24

### Multi-region interconnection

     192.168.1.20/24                                                  192.168.1.21/24
            |                                                                |
        OpenLAN -- Hotel Wifi --> Central Switch(NanJing) <--- Other Wifi --- OpenLAN
                                         |
                                         |
                                       Internet
                                         |
                                         |
                                 Central Switch(Shanghai) - 192.168.1.10/24
                                         |
                                         |
                ------------------------------------------------------
                ^                        ^                           ^
                |                        |                           |
             Office Wifi              Home Wifi                 Hotel Wifi     
                |                        |                           |
            OpenLAN                  OpenLAN                     OpenLAN
        192.168.1.11/24           192.168.1.12/24             192.168.1.13/24


## Help documents
- [Software Installation](docs/install.md)
- [Branch Access](docs/central.md)
- [Multi-region Interconnection](docs/multiarea.md)
- [Zero Trust Network](docs/ztrust.md)
- [Docker Compose](docs/docker.md)
