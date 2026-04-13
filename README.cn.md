简体中文 | [English](./README.en.md)

<p align="center">
  <img src="./pkg/public/openlan.png" alt="OpenLAN Logo" width="180" />
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan)
[![Codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![CodeQL](https://github.com/luscis/openlan/actions/workflows/codeql.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/codeql.yml)
[![Build](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml/badge.svg)](https://github.com/luscis/openlan/actions/workflows/ubuntu.yml)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://github.com/luscis/openlan/tree/master/docs)
[![Releases](https://img.shields.io/github/release/luscis/openlan/all.svg?style=flat-square)](https://github.com/luscis/openlan/releases)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

## 🌐 什么是 OpenLAN？

OpenLAN 是一种实现局域网数据报文在广域网传输的解决方案，支持在用户空间创建多个虚拟以太网。

## 🤔 为什么选择 OpenLAN？

如果您需要更灵活的 VPN 解决方案——例如访问企业内部网络、通过公网云主机进行网络代理或穿透——OpenLAN 能够让部署变得更加简单高效。

## ✨ 核心功能

- 🔒 **多网络空间隔离**：支持划分多个独立的网络空间，为不同业务提供逻辑网络隔离；
- 🔗 **Central Switch 互联**：多个 Central Switch 之间可通过 OpenLAN 协议在链路层互联互通，并支持配置 SNAT 路由，轻松访问企业内部网络；
- 🖥️ **OpenVPN 接入**：支持通过 OpenVPN 接入用户网络，兼容 Android、macOS、Windows 等多平台；
- 🛡️ **IPSec 隧道与 VxLAN**：支持在多个 Central Switch 之间建立 IPSec 隧道网络，并可在该网络上进一步划分 VxLAN 租户网络；
- 🔑 **简洁的认证机制**：采用用户名/密码方式进行接入认证，支持配置预共享密钥对数据报文加密；
- 📡 **多传输协议支持**：OpenLAN 协议可运行于 TCP、TLS、UDP、KCP、WS、WSS 等多种传输协议之上——TCP 性能优异，TLS/WSS 提供更强的加密安全；
- 🔄 **灵活代理转发**：提供 HTTP、HTTPS、SOCKS5 等正向代理功能，支持按域名匹配策略灵活配置流量转发。

## 🗺️ 典型应用场景

### 🏢 分支中心接入

```text
                           Central Switch(企业中心) - 10.16.1.10/24
                                      ^
                                      |
                                   Wifi(DNAT)
                                      |
                                      |
             ----------------------Internet-------------------------
             ^                        ^                           ^
             |                        |                           |
           分支 1                    分支 2                        分支 3
             |                        |                           |
         OpenLAN                  OpenLAN                      OpenLAN
      10.16.1.11/24             10.16.1.12/24                10.16.1.13/24
```

### 🌍 多区域互联

```text
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
```

## 📚 文档指南

- 📦 [软件安装](docs/install.md)
- 🏢 [分支接入](docs/central.md)
- 🌍 [多区域互联](docs/multiarea.md)
- 🔐 [零信任网络](docs/ztrust.md)
- 🐳 [Docker Compose](docs/docker.md)
