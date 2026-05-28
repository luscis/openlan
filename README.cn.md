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

- `Access 认证与会话`：[`access_success`](tests/cases/access_success.sh)、[`access_fail`](tests/cases/access_fail.sh)、[`access_admin_multi_login`](tests/cases/access_admin_multi_login.sh)、[`access_same_user_mutex`](tests/cases/access_same_user_mutex.sh)；
- `Access 加密与策略作用域`：[`access_pre_network_crypt`](tests/cases/access_pre_network_crypt.sh)、[`access_snat_scope_matrix`](tests/cases/access_snat_scope_matrix.sh)；
- `OpenVPN 功能`：[`access_openvpn`](tests/cases/access_openvpn.sh)、[`access_openvpn_redirect`](tests/cases/access_openvpn_redirect.sh)、[`access_openvpn_client_ping`](tests/cases/access_openvpn_client_ping.sh)、[`access_openvpn_tcp_reset`](tests/cases/access_openvpn_tcp_reset.sh)、[`access_openvpn_snat_vip`](tests/cases/access_openvpn_snat_vip.sh)；
- `OpenVPN 性能`：[`access_openvpn_perf`](tests/cases/access_openvpn_perf.sh)（时延/吞吐/协议维度对比）；
- `Proxy 能力`：[`proxy_http`](tests/cases/proxy_http.sh)、[`proxy_tcp`](tests/cases/proxy_tcp.sh)、[`proxy_name`](tests/cases/proxy_name.sh)、[`proxy_name_backends`](tests/cases/proxy_name_backends.sh)；
- `Switch 基础隧道`：[`switch_tcp`](tests/cases/switch_tcp.sh)、[`switch_udp`](tests/cases/switch_udp.sh)；
- `Switch IPSec 叠加互联`：[`switch_ipsec_vxlan`](tests/cases/switch_ipsec_vxlan.sh)、[`switch_ipsec_gre`](tests/cases/switch_ipsec_gre.sh)；
- `Switch IPSec 叠加性能`：[`switch_ipsec_vxlan_perf`](tests/cases/switch_ipsec_vxlan_perf.sh)；
- `Switch ACL 与访问控制`：[`switch_acl`](tests/cases/switch_acl.sh)、[`switch_acl_default_action`](tests/cases/switch_acl_default_action.sh)、[`switch_ztrust`](tests/cases/switch_ztrust.sh)；
- `Switch 路由与转发`：[`switch_bgp`](tests/cases/switch_bgp.sh)、[`switch_route3`](tests/cases/switch_route3.sh)、[`switch_findhop`](tests/cases/switch_findhop.sh)；
- `Switch NAT 与流控`：[`switch_dnat`](tests/cases/switch_dnat.sh)、[`switch_ratelimit`](tests/cases/switch_ratelimit.sh)；
- `Switch Namespace/VRF 与隔离`：[`switch_namespace`](tests/cases/switch_namespace.sh)、[`switch_namespace_snat`](tests/cases/switch_namespace_snat.sh)、[`switch_namespace_openvpn`](tests/cases/switch_namespace_openvpn.sh)；
- `Switch Output 综合性能`：[`switch_output_perf`](tests/cases/switch_output_perf.sh)（混合 TCP/UDP 的连通、时延、丢包、带宽）。
