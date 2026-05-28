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

- **Access authentication and sessions (core)**: `access_success`, `access_fail`, `access_admin_multi_login`, `access_same_user_mutex`;
- **Access encryption and scope**: `access_pre_network_crypt`, `access_snat_scope_matrix`;
- **OpenVPN functional flows**: `access_openvpn`, `access_openvpn_redirect`, `access_openvpn_client_ping`, `access_openvpn_tcp_reset`, `access_openvpn_snat_vip`;
- **OpenVPN performance**: `access_openvpn_perf` (latency, throughput, and protocol-level comparison);
- **Proxy capabilities**: `proxy_http`, `proxy_tcp`, `proxy_name`, `proxy_name_backends`;
- **Switch baseline tunnels**: `switch_tcp`, `switch_udp`;
- **Switch IPSec overlays**: `switch_ipsec_vxlan`, `switch_ipsec_gre`;
- **Switch IPSec overlay performance**: `switch_ipsec_vxlan_perf`;
- **Switch ACL and access control**: `switch_acl`, `switch_acl_default_action`, `switch_ztrust`;
- **Switch routing and forwarding**: `switch_bgp`, `switch_route3`, `switch_findhop`;
- **Switch NAT and traffic control**: `switch_dnat`, `switch_ratelimit`;
- **Switch namespace/VRF and isolation**: `switch_namespace`, `switch_namespace_snat`, `switch_namespace_openvpn`;
- **Switch output performance**: `switch_output_perf` (mixed TCP/UDP connectivity, latency, loss, and bandwidth).
