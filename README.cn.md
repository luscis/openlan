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

- 🔒 **多网络与命名空间隔离**：支持多网络空间、VRF/Namespace 绑定、跨网络隔离、按作用域启用 SNAT，以及网络内 DHCP 地址分配；
- 🔗 **Central Switch 互联与路由转发**：支持 TCP/UDP output、三节点转发、静态路由、FindHop 主备/负载均衡和 External BGP 前缀过滤；
- 🖥️ **OpenVPN 接入能力**：支持 OpenVPN 接入、静态客户端地址、客户端互通、路由重定向、TCP reset 处理，以及经 SNAT 访问远端 VIP；
- 🛡️ **隧道与叠加网络**：支持 TCP/UDP 传输、VxLAN/GRE output 与 IPSec 隧道，并覆盖无 IPSec/启用 IPSec 的连通和性能采样；
- 🔑 **认证与分级加密**：支持用户名/密码认证、同账号互斥、多管理员并发登录、全局与网络级预共享密钥，以及 AES/SM4 OpenVPN cipher 协商；
- 🧭 **策略控制面**：内置 ACL 默认动作、精细规则保存/重载、零信任 Guest/Knock 控制、DNAT 端口发布和客户端 QoS 规则；
- ⚙️ **可运维流量治理**：支持限速规则动态调整、Linux tc/iptables 状态观测、reload 持久性，以及 ping/RTT/iperf3 性能采样；
- 🔄 **Ceci 代理与服务转发**：支持 HTTP/TCP/DNS Proxy、按域名匹配多后端，以及 TCP/HTTP Service 的路由后端、全局后端和重启恢复。

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

OpenLAN 提供了 42 个可直接执行的场景测试脚本，位于 `tests/cases`，
共组织为 75+ 个验证函数，累计包含 1000+ 条断言。
统一入口为 `tests/start.sh`。

常用命令：

```bash
# 列出所有场景
bash tests/start.sh --list

# 运行全部场景
bash tests/start.sh

# 运行指定场景
bash tests/start.sh switch_tcp access_success

# 生成测试报告（md/html）
bash tests/start.sh --report
```

报告查看：[run.md](./docs/report/latest/run.md)

功能覆盖（按测试场景分组）：

- **Access 认证与会话**
  - `access_success`：验证双客户端登录与互通，并覆盖全局加密更新后的重连。
  - `access_fail`：验证错误密码拒绝。
  - `access_admin_multi_login`：验证管理员多端并发。
  - `access_same_user_mutex`：验证普通用户同账号互斥登录。
- **Access 加密、SNAT 与 QoS**
  - `access_pre_network_crypt`：验证网络级预共享密钥与更新后新旧客户端行为。
  - `access_snat_scope_matrix`：覆盖 OpenVPN、Network A 与 Network B 的 SNAT 作用域矩阵。
  - `access_client_qos`：验证客户端 QoS 规则的新增、更新、列表、保存和删除。
- **OpenVPN 接入链路**
  - `access_openvpn`：覆盖 OpenVPN 添加/删除、客户端 CCD 文件、非法 cipher 拒绝以及 AES/SM4 数据通道协商。
  - `access_openvpn_acl`：验证 OpenVPN ACL 走 iptables，bridge ACL 走 ebtables。
  - `access_openvpn_client_ping`：验证静态地址客户端互 ping。
  - `access_openvpn_redirect`：验证源路由重定向到二级 Switch 后的 VIP 访问。
  - `access_openvpn_tcp_reset`：验证服务端 TCP reset 场景。
  - `access_openvpn_snat_vip`：验证 OpenVPN 客户端经 SNAT 访问远端 VIP。
  - `access_openvpn_multi_route`：验证 sw1 network a 的 OpenVPN 客户端在 sw2 补充回程路由前不可达、补充后可访问 sw2 network a 与 b。
  - `access_openvpn_multi_snat`：验证 sw1 network a 的 OpenVPN 客户端在无 sw2 回程路由时，启用 OpenVPN SNAT 后可访问 sw2 network a 与 b。
- **OpenVPN 性能采样**
  - `access_openvpn_perf`：覆盖 TCP/UDP OpenVPN 的连通性、0% 丢包 RTT 摘要、iperf3 带宽采样和 reload 持久性。
- **Ceci Proxy 与 Service**
  - `proxy_http`、`proxy_tcp`、`proxy_name`、`proxy_name_backends`：覆盖 HTTP/TCP/DNS 代理、域名匹配多后端路由和 reload 后恢复。
  - `service_tcp`、`service_http`：覆盖 Ceci Service 的 TCP/HTTP 转发、后端路由/全局后端和重启恢复。
- **Switch 基础输出链路**
  - `switch_tcp`、`switch_udp`：覆盖 TCP/UDP output 的认证、连通、reload 和 output 删除后的隔离。
- **Switch IPSec 与叠加隧道**
  - `switch_ipsec_vxlan`、`switch_ipsec_gre`：覆盖 VxLAN/GRE output 与 IPSec 隧道建立、reload 和删除。
  - `switch_ipsec_vxlan_perf`：对比无 IPSec 与启用 IPSec 后的 ping、RTT 和 TCP/UDP iperf3 采样。
- **Switch ACL 与零信任**
  - `switch_acl`：验证 VIP TCP/80 与 ICMP 的 ACL 新增、列表、保存、reload 和删除。
  - `switch_acl_default`：验证 ACL 默认动作在 drop 和 accept 之间切换。
  - `switch_acl_network`：验证 ACL ebtables hook 只作用于 bridge ingress。
  - `switch_ztrust`：验证 ZTrust 启停、Guest 添加、无地址 client 错误输出、Knock add/list token 推导、其他用户 knock 失败和 reload 持久性。
- **Switch 路由与转发控制**
  - `switch_bgp`：验证 BGP 邻居建立、前缀接收/发布和 reload。
  - `switch_route3`：验证三节点转发与静态路由可达。
  - `switch_findhop`：验证 FindHop 路由绑定、删除保护、主备和负载均衡。
- **Switch NAT、DHCP、限速与命名空间隔离**
  - `switch_dnat`：验证 DNAT 新增、访问、reload 和删除。
  - `switch_dhcp`：验证 DHCP enable/disable API、独立 dhcpConfig、dnsmasq 启停、命名空间和 Access 客户端获取租约、互 ping 以及 reload 持久性。
  - `switch_ratelimit`：验证桥接与 OpenVPN 设备限速规则更新和 tc 状态。
  - `switch_setaddress`：验证修改网桥地址后，地址下发、SNAT 源网段和 OpenVPN push route 随新地址范围刷新。
  - `switch_namespace`、`switch_namespace_snat`、`switch_namespace_openvpn`：覆盖 VRF 绑定、SNAT 源地址改写、OpenVPN 设备入 VRF、跨网络隔离和 reload 持久性。
- **Switch Output 综合性能**
  - `switch_output_perf`：覆盖中心 Switch 同时接入 UDP/TCP output 的认证、连通性、0% 丢包 RTT 摘要、带宽采样和 reload 恢复。
