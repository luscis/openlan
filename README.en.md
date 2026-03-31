English | [简体中文](./README.cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan)
[![Codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![CodeQL](https://github.com/luscis/openlan/actions/workflows/codeql.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/codeql.yml)
[![Build](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://github.com/luscis/openlan/tree/master/docs)
[![Releases](https://img.shields.io/github/release/luscis/openlan/all.svg?style=flat-square)](https://github.com/luscis/openlan/releases)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

## 🌐 What is OpenLAN?

OpenLAN is a solution for transmitting LAN packets over WAN, enabling you to create multiple virtual Ethernet networks in user space.

## 🤔 Why Choose OpenLAN?

If you need a flexible VPN solution — such as accessing enterprise internal networks, or proxying and tunneling traffic through public cloud instances — OpenLAN makes deployment simpler and more efficient.

## ✨ Key Features

- 🔒 **Network Segmentation**: Divide the network into multiple isolated spaces, providing logical network isolation for different services.
- 🔗 **Central Switch Interconnection**: Multiple Central Switches communicate at the link layer via the OpenLAN protocol, with SNAT route support for seamless access to enterprise internal networks.
- 🖥️ **OpenVPN Integration**: Connect user networks via OpenVPN, with support for Android, macOS, Windows, and other platforms.
- 🛡️ **IPSec & VxLAN Support**: Establish IPSec tunnels between Central Switches, with VxLAN tenant network segmentation on top.
- 🔑 **Simple Authentication**: Username/password-based access authentication with optional pre-shared key encryption for data packets.
- 📡 **Multi-Protocol Support**: OpenLAN runs over TCP, TLS, UDP, KCP, WS, and WSS — TCP for high performance, TLS/WSS for stronger encryption security.
- 🔄 **Flexible Proxy Forwarding**: Built-in HTTP, HTTPS, and SOCKS5 forward proxy support with domain-based routing rules for flexible traffic forwarding.

## 🗺️ Use Cases

### 🏢 Branch-to-Center Access

```text
                       Central Switch (Enterprise Center) - 10.16.1.10/24
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

### 🌍 Multi-Region Interconnection

```text
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

## 📚 Documentation

- 📦 [Software Installation](docs/install.md)
- 🏢 [Branch Access](docs/central.md)
- 🌍 [Multi-Region Interconnection](docs/multiarea.md)
- 🔐 [Zero Trust Network](docs/ztrust.md)
- 🐳 [Docker Compose](docs/docker.md)
