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

OpenLAN is a solution for transmitting LAN packets over WAN, enabling you to create multiple virtual Ethernet networks in user space.

## 🤔 Why Choose OpenLAN?

If you need a flexible VPN solution — such as accessing enterprise internal networks, or proxying and tunneling traffic through public cloud instances — OpenLAN makes deployment simpler and more efficient.

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
      --------------------Internet--------------------
      ^                      ^                       ^
      |                      |                       |
   Branch1                Branch2                 Branch3
      |                      |                       |
   OpenLAN                OpenLAN                 OpenLAN
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
      ---------------------------------------------------------
      ^                      ^                               ^
      |                      |                               |
   Office Wifi            Home Wifi                      Hotel Wifi
      |                      |                               |
   OpenLAN                OpenLAN                         OpenLAN
192.168.1.11/24        192.168.1.12/24                 192.168.1.13/24
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
                       /                        \
                      /                          \
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

OpenLAN provides executable scenario scripts under `tests/cases/*.sh`,
with a unified entrypoint at `tests/start.sh`.

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

Capability coverage:

- `access_*`: validates auth path correctness for success/failure, multi-login, and same-user mutex;
- `access_pre_network_crypt`: validates per-network pre-shared crypt for isolated network encryption;
- `access_openvpn*`: validates OpenVPN lifecycle, route redirect, client reachability, and VIP access;
- `access_snat_scope_matrix`: validates SNAT scope behavior across OpenVPN and Access entry paths;
- `proxy_*`: validates HTTP/TCP/DNS proxy forwarding and domain-based backend routing behavior;
- `switch_tcp|switch_udp`: validates baseline inter-switch tunnel connectivity;
- `switch_ipsec_*`: validates cross-site interconnection with IPSec + VxLAN/GRE overlays;
- `switch_acl*`: validates ACL CRUD, default action switching, and reload consistency;
- `switch_bgp`: validates External BGP peering, route filtering, and config persistence;
- `switch_dnat`: validates DNAT updates and NAT table rule synchronization;
- `switch_findhop`: validates FindHop binding, remove guard, and state recovery after reload;
- `switch_ztrust`: validates zero-trust toggling and Guest/Knock access controls;
- `switch_ratelimit`: validates rate-limit CRUD and kernel tc state consistency;
- `switch_route3`: validates L3 route reachability via an intermediate switch (sw2).
