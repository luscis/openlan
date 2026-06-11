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

- 🔒 **Multi-Network and Namespace Isolation**: Support multiple network spaces, VRF/Namespace binding, cross-network isolation, scope-based SNAT, and in-network DHCP address allocation.
- 🔗 **Central Switch Interconnect and Routing**: Support TCP/UDP outputs, three-node forwarding, static routes, FindHop active-backup/load balancing, and External BGP prefix filters.
- 🖥️ **OpenVPN Access**: Support OpenVPN onboarding, static client addresses, client-to-client reachability, route redirect, TCP reset handling, and remote VIP access through SNAT.
- 🛡️ **Tunnels and Overlays**: Support TCP/UDP transports, VxLAN/GRE outputs, and IPSec tunnels, with connectivity and performance sampling before and after IPSec is enabled.
- 🔑 **Authentication and Layered Crypt**: Support username/password auth, same-user mutex, concurrent admin logins, global and per-network pre-shared crypt, plus AES/SM4 OpenVPN cipher negotiation.
- 🧭 **Policy Control Plane**: Built-in ACL default actions, persisted fine-grained rules, zero-trust Guest/Knock controls, DNAT service publishing, and client QoS rules.
- ⚙️ **Operational Traffic Governance**: Dynamic rate-limit updates, observable Linux tc/iptables state, reload persistence, and ping/RTT/iperf3 performance sampling.
- 🔄 **Ceci Proxy and Service Forwarding**: HTTP/TCP/DNS Proxy, domain-matched multi-backend routing, and TCP/HTTP Service forwarding with route/global backends and restart recovery.

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

OpenLAN can use a central switch as the enterprise access hub. Branch devices
or edge gateways authenticate to the center over public networks and join the
same virtual LAN with assigned addresses. This keeps branch-to-center and
branch-to-branch traffic on a managed overlay, while the central switch can
apply user authentication, shared crypt settings, ACLs, SNAT, DNAT, DHCP, and
rate-limit policies from one control point.

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

OpenLAN can connect switches across cities, clouds, and temporary networks such
as hotel or home Wi-Fi. Each region can keep its local OpenLAN switch while
outputs build authenticated TCP/UDP tunnels between regions. Static routes or
FindHop routes then make remote subnets and VIPs reachable through the proper
nexthop, so applications can move across sites without exposing every backend
directly to the Internet.

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

OpenLAN Zero Trust turns a virtual network into a default-deny access plane for
new traffic while keeping established flows intact. Remote users can connect
through OpenVPN, but protected services stay unreachable until the user is
registered as a ZTrust guest and creates a temporary knock rule for a specific
protocol and socket, such as `tcp/192.59.0.1:8081`. This makes it useful for
contractor access, emergency operations, and service-by-service approvals where
network reachability, identity, and time-limited permission must be controlled
together.

### 🔀 HTTP Route Forwarding

```text
        Client request              Client request              Client request
        Host: group.test            Host: single.test           Host: unknown.test
              |                           |                           |
              +---------------------------+---------------------------+
                                          |
                                          v
                                  Ceci HTTP Service
                              listen: 192.168.1.10:13083
                                          |
                       +------------------+-----------------+
                       /                  |                  \
                      / group route       | single route      \ global backend
                     /                    |                    \
        sw2 group backends        sw2 single backend      sw3 fallback backend
         192.56.0.2:18084          192.56.0.2:18086         192.56.0.3:18088
         192.56.0.2:18085
         192.56.0.2:18087
```

OpenLAN can publish one local HTTP service endpoint and route requests to
different backends by the HTTP `Host` header. A hostname such as `single.test`
can map to one backend, while `group.test` can map to multiple backends with
round-robin balancing. Requests without a matched hostname can fall back to a
global backend. The backends may live behind other OpenLAN switches, so the
service endpoint stays stable while applications are distributed across sites.

## 📚 Documentation

- 📦 [Software Installation](docs/install.md)
- 🏢 [Branch Access](docs/central.md)
- 🌍 [Multi-Region Interconnection](docs/multiarea.md)
- 🔐 [Zero Trust Network](docs/ztrust.md)
- 🔀 [Ceci Proxy Examples](docs/proxy.md)
- 🐳 [Docker Compose](docs/docker.md)

## 🧪 Scenario Tests

OpenLAN provides 42 executable scenario scripts under `tests/cases`,
organized into 75+ validation functions with 1000+ assertions in total.
The unified entrypoint is `tests/start.sh`.

Common commands:

```bash
# List all scenarios
bash tests/start.sh --list

# Run all scenarios
bash tests/start.sh

# Run selected scenarios
bash tests/start.sh switch_tcp access_success

# Generate test reports (md/html)
bash tests/start.sh --report
```

Report: [run.md](./docs/report/latest/run.md)

Capability coverage by test scenario:

- **Access authentication and sessions**
  - `access_success`: verifies two-client login, reachability, and reconnect after global crypt update.
  - `access_fail`: verifies wrong-password rejection.
  - `access_admin_multi_login`: verifies concurrent admin logins.
  - `access_same_user_mutex`: verifies same-user mutex for regular users.
- **Access crypt, SNAT, and QoS**
  - `access_pre_network_crypt`: verifies per-network pre-shared crypt and client behavior after key updates.
  - `access_snat_scope_matrix`: covers the SNAT scope matrix for OpenVPN, Network A, and Network B.
  - `access_client_qos`: verifies client QoS rule add, update, list, save, and remove flows.
- **OpenVPN access paths**
  - `access_openvpn`: covers OpenVPN add/remove, CCD files, invalid cipher rejection, and AES/SM4 data-channel negotiation.
  - `access_openvpn_acl`: verifies OpenVPN ACL uses iptables while bridge ACL uses ebtables.
  - `access_openvpn_client_ping`: verifies static-address client-to-client ping.
  - `access_openvpn_redirect`: verifies source-route redirect to a second switch for VIP access.
  - `access_openvpn_tcp_reset`: verifies server-side TCP reset handling.
  - `access_openvpn_snat_vip`: verifies OpenVPN client access to a remote VIP through SNAT.
  - `access_openvpn_multi_route`: verifies a sw1 network-a OpenVPN client cannot reach sw2 network a/b before sw2 return routing is added, then can reach both after the route is added.
  - `access_openvpn_multi_snat`: verifies a sw1 network-a OpenVPN client can reach sw2 network a/b through OpenVPN-scoped SNAT without a sw2 return route.
- **OpenVPN performance sampling**
  - `access_openvpn_perf`: covers TCP/UDP OpenVPN connectivity, 0% packet-loss RTT summaries, iperf3 bandwidth sampling, and reload persistence.
- **Ceci Proxy and Service**
  - `proxy_http`, `proxy_tcp`, `proxy_name`, and `proxy_name_backends`: cover HTTP/TCP/DNS proxying, domain-matched multi-backend routing, and reload recovery.
  - `service_tcp` and `service_http`: cover Ceci Service TCP/HTTP forwarding, route/global backends, and restart recovery.
- **Switch baseline output links**
  - `switch_tcp` and `switch_udp`: cover TCP/UDP output authentication, reachability, reload behavior, and isolation after output removal.
- **Switch IPSec and overlays**
  - `switch_ipsec_vxlan` and `switch_ipsec_gre`: cover VxLAN/GRE outputs with IPSec tunnel establishment, reload, and removal.
  - `switch_ipsec_vxlan_perf`: compares ping, RTT, and TCP/UDP iperf3 samples before and after IPSec is enabled.
- **Switch ACL and zero trust**
  - `switch_acl`: verifies ACL add/list/save/reload/remove with VIP TCP/80 and ICMP.
  - `switch_acl_default`: verifies ACL default action switching between drop and accept.
  - `switch_acl_network`: verifies ACL ebtables hook is bridge ingress only.
  - `switch_ztrust`: verifies ZTrust enable/disable, Guest add, no-address client error output, token-derived Guest/Knock add/list, other-user knock rejection, and reload persistence.
- **Switch routing and forwarding control**
  - `switch_bgp`: verifies BGP peering, prefix advertise/receive filters, and reload.
  - `switch_route3`: verifies three-node forwarding and static-route reachability.
  - `switch_findhop`: verifies FindHop route binding, remove guards, active-backup, and load balancing.
- **Switch NAT, DHCP, rate limit, and namespace isolation**
  - `switch_dnat`: verifies DNAT add, reachability, reload, and remove.
  - `switch_dhcp`: verifies DHCP enable/disable APIs, independent dhcpConfig, dnsmasq start/stop, namespace and access client lease allocation, ping reachability, and reload persistence.
  - `switch_ratelimit`: verifies bridge/OpenVPN device rate-limit updates and Linux tc state.
  - `switch_setaddress`: verifies bridge address updates refresh address assignment, SNAT source ranges, and OpenVPN pushed routes.
  - `switch_namespace`, `switch_namespace_snat`, and `switch_namespace_openvpn`: cover VRF binding, SNAT source rewriting, OpenVPN device VRF membership, cross-network isolation, and reload persistence.
- **Switch output performance**
  - `switch_output_perf`: covers one center switch with mixed UDP/TCP outputs, authentication, connectivity, 0% packet-loss RTT summaries, bandwidth sampling, and reload recovery.
