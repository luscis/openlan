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

OpenLAN 是一套多租户网络解决方案，可在广域网链路上传输局域网报文，帮助您在跨地域、云环境与分支站点之间构建并运营多个相互隔离的虚拟以太网络。

## 🤔 为什么选择 OpenLAN？

如果您需要灵活的 VPN 方案来实现企业内网安全访问、流量代理转发或经公网云主机建立隧道，OpenLAN 可以显著简化部署并提升运维效率。

## ✨ 核心功能

- 🔒 **多网络空间隔离**：支持划分多个独立的网络空间，为不同业务提供逻辑网络隔离；
- 🔗 **Central Switch 互联与路由转发**：支持跨站点互联、三层转发与 SNAT/DNAT，覆盖分支到中心的访问与跨网络发布场景；
- 🖥️ **OpenVPN 接入能力**：支持 OpenVPN 接入、路由重定向、客户端互通，以及通过 SNAT 访问远端 VIP；
- 🛡️ **隧道与叠加网络**：支持 TCP/UDP 传输与 IPSec+VxLAN/GRE 叠加隧道，满足跨地域组网与租户隔离需求；
- 🔑 **认证与分级加密**：支持用户名/密码认证、同账号互斥控制、多管理员并发登录，以及网络级预共享密钥加密；
- 🧭 **策略控制面**：内置 ACL、零信任控制（Guest/Knock）、FindHop 路由绑定、External BGP 邻居与前缀过滤策略；
- ⚙️ **可运维流量治理**：支持限速规则动态调整、规则重载一致性、NAT/路由状态可观测；
- 🔄 **灵活代理转发**：支持 HTTP/TCP/DNS 代理与按域名匹配后端，便于按业务策略分流流量。

## 🗺️ 典型应用场景

### 🏢 分支中心接入

```text
        Central Switch(企业中心) - 10.16.1.10/24
                           ^
                           |
                        Wifi(DNAT)
                           |
                           |
      -----------------Internet-----------------
      ^                    ^                   ^
      |                    |                   |
    分支 1                分支 2               分支 3
      |                    |                   |
  OpenLAN              OpenLAN             OpenLAN
10.16.1.11/24        10.16.1.12/24       10.16.1.13/24
```

### 🌍 多区域互联

```text
192.168.1.20/24                                    192.168.1.21/24
      |                                                  |
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
      ^                     ^                              ^
      |                     |                              |
   办公 Wifi             家庭 Wifi                      酒店 Wifi
      |                     |                              |
   OpenLAN               OpenLAN                        OpenLAN
192.168.1.11/24       192.168.1.12/24                192.168.1.13/24
```

### 🔐 零信任接入控制

```text
      访客终端                  员工终端                   运维终端
         |                        |                         |
      OpenVPN                  OpenVPN                   OpenVPN
         \                        |                         /
          \                       |                        /
           ----------------------互联网----------------------
                                   |
                                   |
                         Central Switch(策略中心)
                     ZTrust + ACL + Knock + Auth
                     /                         \
                    /                           \
         Guest Network(仅受限访问)     Trusted Network(按策略访问业务)
             172.16.100.0/24               10.16.1.0/24
```

## 📚 文档指南

- 📦 [软件安装](docs/install.md)
- 🏢 [分支接入](docs/central.md)
- 🌍 [多区域互联](docs/multiarea.md)
- 🔐 [零信任网络](docs/ztrust.md)
- 🐳 [Docker Compose](docs/docker.md)

## 🧪 场景测试

OpenLAN 提供了 33 个可直接执行的场景测试脚本，位于 `tests/cases`，
共组织为 59 个验证函数，累计包含 796 条断言。
统一入口为 `tests/start.sh`。

常用命令：

```bash
# 列出所有场景
bash tests/start.sh --list

# 运行全部场景
bash tests/start.sh

# 运行指定场景
bash tests/start.sh switch_tcp access_success

# 生成测试报告（txt/html/tar）
bash tests/start.sh --report
```

报告查看：[run.html](docs/report/latest/run.html)

功能覆盖（按能力分组）：

- **Access 认证与会话（核心）**：`access_success`、`access_fail`、`access_admin_multi_login`、`access_same_user_mutex`；
- **Access 加密与策略作用域**：`access_pre_network_crypt`、`access_snat_scope_matrix`；
- **OpenVPN 功能**：`access_openvpn`、`access_openvpn_redirect`、`access_openvpn_client_ping`、`access_openvpn_tcp_reset`、`access_openvpn_snat_vip`；
- **OpenVPN 性能**：`access_openvpn_perf`（时延/吞吐/协议维度对比）；
- **Proxy 能力**：`proxy_http`、`proxy_tcp`、`proxy_name`、`proxy_name_backends`；
- **Switch 基础隧道**：`switch_tcp`、`switch_udp`；
- **Switch IPSec 叠加互联**：`switch_ipsec_vxlan`、`switch_ipsec_gre`；
- **Switch IPSec 叠加性能**：`switch_ipsec_vxlan_perf`；
- **Switch ACL 与访问控制**：`switch_acl`、`switch_acl_default_action`、`switch_ztrust`；
- **Switch 路由与转发**：`switch_bgp`、`switch_route3`、`switch_findhop`；
- **Switch NAT 与流控**：`switch_dnat`、`switch_ratelimit`；
- **Switch Namespace/VRF 与隔离**：`switch_namespace`、`switch_namespace_snat`、`switch_namespace_openvpn`；
- **Switch Output 综合性能**：`switch_output_perf`（混合 TCP/UDP 的连通、时延、丢包、带宽）。
