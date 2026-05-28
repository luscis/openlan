English | [简体中文](./README.cn.md)

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

## 🌐 What is OpenLAN?

OpenLAN is a multiple-tenant networking solution that carries LAN packets over WAN links, allowing you to build and operate multiple isolated virtual Ethernet networks across regions, clouds, and branch sites.

## 🤔 Why Choose OpenLAN?

If you need a flexible VPN solution for secure enterprise access, traffic proxying, or tunneling through public cloud instances, OpenLAN simplifies deployment and improves operational efficiency.

## ✨ Key Features

- 🔒 **Network Segmentation**: Divide the network into multiple isolated spaces, providing logical network isolation for different services.
- 🔗 **Inter-Switch Routing & NAT**: Build cross-site interconnection with L3 forwarding, SNAT, and DNAT for branch-to-center and service publishing scenarios.
- 🖥️ **OpenVPN Access**: Support OpenVPN onboarding, route redirect, client-to-client reachability, and remote VIP access through SNAT paths.
- 🛡️ **Tunnels & Overlays**: Support TCP/UDP transports plus IPSec with VxLAN/GRE overlays for multi-region networking and tenant separation.
- 🔑 **Authentication & Layered Crypt**: Support username/password auth, same-user mutex, concurrent admin logins, and per-network pre-shared crypt.
- 🧭 **Policy Control Plane**: Built-in ACLs, zero-trust controls (Guest/Knock), FindHop route binding, and External BGP peering with prefix filtering.
- ⚙️ **Operational Traffic Governance**: Dynamic rate-limit updates, reload consistency, and observable NAT/route state for day-2 operations.
- 🔄 **Flexible Proxy Forwarding**: HTTP/TCP/DNS proxy with domain-based backend routing for policy-driven traffic splitting.

## 🗺️ Use Cases

### 🏢 Branch-to-Center Access

```text
         Central Switch (Enterprise Center) - 10.16.1.10/24
                               ^
                               |
                            Wifi(DNAT)
                               |
                               |
         --------------------Internet-------------------
         ^                     ^                       ^
         |                     |                       |
      Branch1                Branch2                 Branch3
         |                     |                       |
      OpenLAN               OpenLAN                 OpenLAN
   10.16.1.11/24          10.16.1.12/24           10.16.1.13/24
```

### 🌍 Multi-Region Interconnection

```text
192.168.1.20/24                                      192.168.1.21/24
     |                                                    |
OpenLAN -- Hotel Wifi --> Central Switch(NanJing) <--- Other Wifi --- OpenLAN
                                |
                                |
                              Internet
                                |
                                |
                  Central Switch(Shanghai) - 192.168.1.10/24
                                |
                                |
      --------------------------------------------------------
      ^                         ^                            ^
      |                         |                            |
   Office Wifi               Home Wifi                    Hotel Wifi
      |                         |                            |
   OpenLAN                    OpenLAN                     OpenLAN
192.168.1.11/24            192.168.1.12/24             192.168.1.13/24
```

### 🔐 Zero-Trust Access Control

```text
       Guest Endpoint            Staff Endpoint            Ops Endpoint
             |                        |                        |
          OpenVPN                  OpenVPN                  OpenVPN
             \                        |                        /
              \                       |                       /
               ---------------------Internet-------------------
                                      |
                                      |
                         Central Switch (Policy Hub)
                        ZTrust + ACL + Knock + Auth
                        /                         \
                       /                           \
      Guest Network (restricted)      Trusted Network (policy access)
            172.16.100.0/24                 10.16.1.0/24
```

## 📚 Documentation

- 📦 [Software Installation](docs/install.md)
- 🏢 [Branch Access](docs/central.md)
- 🌍 [Multi-Region Interconnection](docs/multiarea.md)
- 🔐 [Zero Trust Network](docs/ztrust.md)
- 🐳 [Docker Compose](docs/docker.md)

## 🧪 Scenario Tests

OpenLAN provides 33 executable scenario scripts under `tests/cases`,
organized into 59 validation functions with 796 assertions in total.
The unified entrypoint is `tests/start.sh`.

Common commands:

```bash
# List all scenarios
bash tests/start.sh --list

# Run all scenarios
bash tests/start.sh

# Run selected scenarios
bash tests/start.sh switch_tcp access_success

# Generate test reports (txt/html/tar)
bash tests/start.sh --report
```

Report: [run.html](docs/report/latest/run.html)

Capability coverage:

- `Access authentication and sessions`: [`access_success`](tests/cases/access_success.sh), [`access_fail`](tests/cases/access_fail.sh), [`access_admin_multi_login`](tests/cases/access_admin_multi_login.sh), [`access_same_user_mutex`](tests/cases/access_same_user_mutex.sh);
- `Access encryption and scope`: [`access_pre_network_crypt`](tests/cases/access_pre_network_crypt.sh), [`access_snat_scope_matrix`](tests/cases/access_snat_scope_matrix.sh);
- `OpenVPN functional flows`: [`access_openvpn`](tests/cases/access_openvpn.sh), [`access_openvpn_redirect`](tests/cases/access_openvpn_redirect.sh), [`access_openvpn_client_ping`](tests/cases/access_openvpn_client_ping.sh), [`access_openvpn_tcp_reset`](tests/cases/access_openvpn_tcp_reset.sh), [`access_openvpn_snat_vip`](tests/cases/access_openvpn_snat_vip.sh);
- `OpenVPN performance`: [`access_openvpn_perf`](tests/cases/access_openvpn_perf.sh) (latency, throughput, and protocol-level comparison);
- `Proxy capabilities`: [`proxy_http`](tests/cases/proxy_http.sh), [`proxy_tcp`](tests/cases/proxy_tcp.sh), [`proxy_name`](tests/cases/proxy_name.sh), [`proxy_name_backends`](tests/cases/proxy_name_backends.sh);
- `Switch baseline tunnels`: [`switch_tcp`](tests/cases/switch_tcp.sh), [`switch_udp`](tests/cases/switch_udp.sh);
- `Switch IPSec overlays`: [`switch_ipsec_vxlan`](tests/cases/switch_ipsec_vxlan.sh), [`switch_ipsec_gre`](tests/cases/switch_ipsec_gre.sh);
- `Switch IPSec overlay performance`: [`switch_ipsec_vxlan_perf`](tests/cases/switch_ipsec_vxlan_perf.sh);
- `Switch ACL and access control`: [`switch_acl`](tests/cases/switch_acl.sh), [`switch_acl_default_action`](tests/cases/switch_acl_default_action.sh), [`switch_ztrust`](tests/cases/switch_ztrust.sh);
- `Switch routing and forwarding`: [`switch_bgp`](tests/cases/switch_bgp.sh), [`switch_route3`](tests/cases/switch_route3.sh), [`switch_findhop`](tests/cases/switch_findhop.sh);
- `Switch NAT and traffic control`: [`switch_dnat`](tests/cases/switch_dnat.sh), [`switch_ratelimit`](tests/cases/switch_ratelimit.sh);
- `Switch namespace/VRF and isolation`: [`switch_namespace`](tests/cases/switch_namespace.sh), [`switch_namespace_snat`](tests/cases/switch_namespace_snat.sh), [`switch_namespace_openvpn`](tests/cases/switch_namespace_openvpn.sh);
- `Switch output performance`: [`switch_output_perf`](tests/cases/switch_output_perf.sh) (mixed TCP/UDP connectivity, latency, loss, and bandwidth).
