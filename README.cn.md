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
      -------------------------------------------------------
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

OpenLAN 提供了 30 个可直接执行的场景测试脚本，位于 `tests/cases`，
共组织为 45 个验证函数，累计包含 732 条断言。
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

功能覆盖（按能力分组）：

- `access_*`：验证认证链路正确性，覆盖成功/失败、多端登录与同账号互斥；
- `access_pre_network_crypt`：验证网络级预共享密钥生效，确保隔离网络独立加密；
- `access_openvpn*`：验证 OpenVPN 生命周期、路由重定向、客户端互通、TCP reset 处理，以及通过 SNAT 访问 VIP；
- `access_snat_scope_matrix`：验证不同入口（OpenVPN/Access）下 SNAT 作用域符合预期；
- `proxy_*`：验证 HTTP/TCP/DNS 代理转发能力及按域名路由后端的策略正确性；
- `switch_tcp|switch_udp`：验证 Central Switch 间基础隧道连通能力；
- `switch_ipsec_*`：验证 IPSec 叠加 VxLAN/GRE 的跨站点互联能力；
- `switch_acl*`：验证 ACL 增删改查、默认动作切换与重载后的规则一致性；
- `switch_bgp`：验证 External BGP 邻居建立、路由发布过滤及配置持久化；
- `switch_dnat`：验证 DNAT 配置变更与 NAT 表规则同步；
- `switch_findhop`：验证 FindHop 绑定、删除保护与重载后状态恢复；
- `switch_namespace`：验证基于 namespace/VRF 的网络绑定与 overlay 连通性；
- `switch_namespace_snat`：验证 namespace 网络的 SNAT 源地址改写，并确认未启用 SNAT 的网络即使存在 VIP 路由也保持隔离；
- `switch_namespace_openvpn`：验证 VRF 中的 OpenVPN 流量、OpenVPN 作用域 SNAT 访问远端 VIP，以及未启用 SNAT 的独立网络在存在 VIP 路由时仍不可达；
- `switch_ztrust`：验证零信任开关及 Guest/Knock 访问控制行为；
- `switch_ratelimit`：验证限速规则增删改与内核 tc 状态一致；
- `switch_route3`：验证经中间节点（sw2）转发时的三层路由可达性。
