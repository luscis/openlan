English | [简体中文](./README.cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan)
[![Codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![CodeQL](https://github.com/luscis/openlan/actions/workflows/codeql.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/codeql.yml)
[![Build](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://github.com/luscis/openlan/tree/master/docs)
[![Releases](https://img.shields.io/github/release/luscis/openlan/all.svg?style=flat-square)](https://github.com/luscis/openlan/releases)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

## What is OpenLAN?

OpenLAN enables LAN packet transmission over WAN, allowing you to establish virtual Ethernet networks across multiple user spaces.

## Why Choose OpenLAN?

OpenLAN is designed for flexible VPN scenarios. Whether you need to access enterprise networks remotely, or leverage public cloud instances for network proxying and penetration, OpenLAN simplifies deployment and management.

## Key Features

- **Network Segmentation**: Create multiple isolated network spaces for different services with logical network isolation.
- **Central Switch Interconnection**: Multiple Central Switches communicate at the Ethernet layer using the OpenLAN protocol. Add SNAT routes at Layer 2 for seamless access to enterprise internal networks.
- **OpenVPN Integration**: Connect user networks via OpenVPN, with support for multiple platforms including Android, macOS, and Windows.
- **IPSec Tunnel Support**: Establish IPSec tunnels between Central Switches, with support for VxLAN tenant networks on top.
- **Simple Authentication**: Username/password-based access authentication with optional pre-shared key encryption for data packets.
- **Multi-Protocol Support**: OpenLAN operates over TCP, TLS, UDP, KCP, WS, and WSS. TCP delivers high performance, while TLS/WSS provides enhanced encryption security.
- **Proxy Capabilities**: Built-in HTTP/HTTPS/SOCKS5 proxy support with flexible domain-based routing rules for traffic forwarding.

## Use Cases

### Branch-to-Center Access

```
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
```

### Multi-Region Interconnection

```
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
```

## Documentation

- [Software Installation](docs/install.md)
- [Branch Access](docs/central.md)
- [Multi-Region Interconnection](docs/multiarea.md)
- [Zero Trust Network](docs/ztrust.md)
- [Docker Compose](docs/docker.md)
