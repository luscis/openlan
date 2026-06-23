# OpenLAN Test Report

- Title: OpenLAN test run started
- Generated: 26-06-22 06:48:07
- Passed: 42
- Failed: 0
- Total: 42

## Scenarios

| Time | Status | Scenario | Topology | Cost |
|---|---|---|---|---|
| 26-06-22 06:02:45 | PASS | [`access_admin_multi_login`](cases/01-access_admin_multi_login_.log) | sw1(center) 100.100.0.241 / example \| tcp access \| tcp access \| ac1(admin) ac2(admin) | 13.916s |
| 26-06-22 06:03:30 | PASS | [`access_client_qos`](cases/02-access_client_qos_.log) | sw1(center) 192.91.0.1 \| OpenVPN tcp/1194, 10.91.0.0/24 \| vpn1 10.91.0.10 \| QoS rule add/update/remove on vpn1@example | 44.886s |
| 26-06-22 06:03:48 | PASS | [`access_fail`](cases/03-access_fail_.log) | sw1(center) 100.100.0.241 / 192.31.0.1 \| tcp access with bad credentials \| acbad asks for 192.31.0.11 | 18.381s |
| 26-06-22 06:04:41 | PASS | [`access_openvpn_acl`](cases/04-access_openvpn_acl_.log) | sw1(center) 192.64.0.1 \| OpenVPN tcp/1194, 10.88.0.0/24 \| vpn1 10.88.0.10 \| ACL drop rule is enforced by original iptables tun hook | 52.660s |
| 26-06-22 06:04:47 | PASS | [`access_openvpn_client_ping`](cases/05-access_openvpn_client_ping_.log) | sw1(center) 192.42.0.1 \| OpenVPN tcp/1194 \| OpenVPN tcp/1194 \| vpn1 10.97.0.10 <----------> vpn2 10.97.0.11 | 6.014s |
| 26-06-22 06:06:21 | PASS | [`access_openvpn_multi_route`](cases/06-access_openvpn_multi_route_.log) | vpn1@sw1/a 10.82.0.10 \| sw1 a 192.82.0.1 -- TCP output --> sw2 a 192.82.0.2 + sw2 b 192.83.0.2 | 93.772s |
| 26-06-22 06:07:22 | PASS | [`access_openvpn_multi_snat`](cases/07-access_openvpn_multi_snat_.log) | vpn1@sw1/a 10.84.0.10 \| sw1 a 192.84.0.1 -- TCP output + SNAT --> sw2 a 192.84.0.2 + sw2 b 192.85.0.2 | 60.997s |
| 26-06-22 06:07:55 | PASS | [`access_openvpn_perf`](cases/08-access_openvpn_perf_.log) | sw1(center) 192.55.0.1 \| OpenVPN tcp/1194 + udp/1195 \| vpn1 10.95.0.0/24 \| ping RTT and iperf3 TCP/UDP samples | 32.583s |
| 26-06-22 06:08:15 | PASS | [`access_openvpn_redirect`](cases/09-access_openvpn_redirect_.log) | vpn1 10.97.0.10 \| v OpenVPN tcp/1194 \| sw1 192.53.0.1 -- output/route --> sw2 192.53.0.2 \| VIP 10.253.0.11 VIP 10.253.0.12 | 19.914s |
| 26-06-22 06:08:26 | PASS | [`access_openvpn`](cases/10-access_openvpn_.log) | sw1(center) 192.41.0.1 \| OpenVPN AES tcp/1194 \| OpenVPN SM4 tcp/1194 \| vpn1 vpn2 | 10.982s |
| 26-06-22 06:09:06 | PASS | [`access_openvpn_snat_vip`](cases/11-access_openvpn_snat_vip_.log) | vpn1 10.96.0.10 \| v OpenVPN tcp/1194 \| sw1 192.52.0.1 -- output + SNAT --> sw2 192.52.0.2 \| VIP 10.252.0.12 | 39.691s |
| 26-06-22 06:09:10 | PASS | [`access_openvpn_tcp_reset`](cases/12-access_openvpn_tcp_reset_.log) | sw1(center) 192.54.0.1:8082 \| OpenVPN tcp/1194, 10.92.0.0/24 \| vpn1 10.92.0.10 \| INPUT tcp-reset rule toggles HTTP reachability | 4.711s |
| 26-06-22 06:10:39 | PASS | [`access_pre_network_crypt`](cases/13-access_pre_network_crypt_.log) | sw1(center) 100.100.0.241 \| network a crypt \| network b global crypt \| ac clients 192.61.0.11-17 ac clients 192.62.0.11-13 | 88.523s |
| 26-06-22 06:10:55 | PASS | [`access_same_user_mutex`](cases/14-access_same_user_mutex_.log) | sw1(center) 100.100.0.241 / example \| tcp access \| tcp access \| ac1(t1) ac2(t1) | 16.500s |
| 26-06-22 06:14:40 | PASS | [`access_snat_scope_matrix`](cases/15-access_snat_scope_matrix_.log) | vpn/access on sw1 network a,b \| sw1: a=192.53.0.1 b=192.54.0.1 int=192.55.0.1 \| int output uplink | 224.468s |
| 26-06-22 06:16:35 | PASS | [`access_success`](cases/16-access_success_.log) | sw1(center) 100.100.0.241 / 192.11.0.1 \| tcp access \| tcp access \| ac1 192.11.0.11 ac2 192.11.0.12 | 114.478s |
| 26-06-22 06:17:03 | PASS | [`proxy_http`](cases/17-proxy_http_.log) | sw1 proxy client 192.52.0.1 \| wget via local Ceci HTTP proxy \| sw1 openceci(http) -- output --> sw2 192.52.0.2:18081 | 28.656s |
| 26-06-22 06:17:39 | PASS | [`proxy_name_backends`](cases/18-proxy_name_backends_.log) | sw1 openceci(name) \| domain A \| domain B \| sw2 dnsmasq sw3 dnsmasq \| 192.55.0.2 192.55.0.3 | 35.793s |
| 26-06-22 06:18:06 | PASS | [`proxy_name`](cases/19-proxy_name_.log) | sw1 name client 192.54.0.1 \| nslookup via local Ceci name proxy \| sw1 openceci(name) -- output --> sw2 dnsmasq 192.54.0.2:5300 | 26.529s |
| 26-06-22 06:18:35 | PASS | [`proxy_tcp`](cases/20-proxy_tcp_.log) | sw1 proxy client 192.53.0.1 \| wget via local Ceci TCP proxy \| sw1 openceci(tcp) -- output --> sw2 192.53.0.2:18082 | 28.678s |
| 26-06-22 06:19:20 | PASS | [`service_http`](cases/21-service_http_.log) | client wget with Host header on sw1 \| sw1 Ceci HTTP service \| ^ route groups ^ global backend | 45.634s |
| 26-06-22 06:20:03 | PASS | [`service_tcp`](cases/22-service_tcp_.log) | client wget on sw1 \| sw1 Ceci TCP service -- output --> sw2 backend 192.56.0.2:18083 | 42.406s |
| 26-06-22 06:22:01 | PASS | [`switch_acl_default`](cases/23-switch_acl_default_.log) | sw1 192.62.0.1 -- UDP output --> sw2 192.62.0.2 \| +-- default drop/accept ----> VIP 10.254.1.12:80/ICMP | 118.567s |
| 26-06-22 06:22:11 | PASS | [`switch_acl_network`](cases/24-switch_acl_network_.log) | sw1 192.63.0.1 \| +-- ACL hook checks on br-example | 9.536s |
| 26-06-22 06:25:09 | PASS | [`switch_acl`](cases/25-switch_acl_.log) | sw1 192.61.0.1 -- UDP output --> sw2 192.61.0.2 \| +------ ACL checks ----------> VIP 10.254.0.12:80/ICMP | 177.674s |
| 26-06-22 06:25:26 | PASS | [`switch_bgp`](cases/26-switch_bgp_.log) | sw1 100.100.0.241 / AS65101 <---- BGP ----> sw2 100.100.0.242 / AS65102 \| svc 192.54.0.1 svc 192.54.0.2 | 17.323s |
| 26-06-22 06:26:26 | PASS | [`switch_dhcp`](cases/27-switch_dhcp_.log) | sw1 DHCP server 192.67.0.1 \| physical output \| tcp access tunnel \| veth-dhcp <-> ns client ac1 tap bridge | 59.970s |
| 26-06-22 06:27:19 | PASS | [`switch_dnat`](cases/28-switch_dnat_.log) | sw1 192.58.0.1 -- UDP output --> sw2 192.58.0.2 \| +----------- DNAT example:80 -> 127.0.0.1:8080 | 52.208s |
| 26-06-22 06:29:21 | PASS | [`switch_findhop`](cases/29-switch_findhop_.log) | sw0 VIP 10.243.0.10 \| network a \| network b \| sw1.0 ------------------- sw1.1 \| +--------- sw2 ---------+ | 122.614s |
| 26-06-22 06:31:11 | PASS | [`switch_ipsec_gre`](cases/30-switch_ipsec_gre_.log) | sw1 100.100.0.241 <==== IPSec ====> sw2 100.100.0.242 \| svc 192.57.0.1 <---- GRE output -- svc 192.57.0.2 | 109.503s |
| 26-06-22 06:32:05 | PASS | [`switch_ipsec_vxlan_perf`](cases/31-switch_ipsec_vxlan_perf_.log) | sw1 100.100.0.241 <==== optional IPSec ====> sw2 100.100.0.242 \| svc 192.57.0.1 <---- VxLAN output ------ svc 192.57.0.2 | 53.657s |
| 26-06-22 06:34:44 | PASS | [`switch_ipsec_vxlan`](cases/32-switch_ipsec_vxlan_.log) | sw1 100.100.0.241 <==== IPSec ====> sw2 100.100.0.242 \| svc 192.56.0.1 <---- VxLAN output - svc 192.56.0.2 | 159.326s |
| 26-06-22 06:37:48 | PASS | [`switch_namespace_openvpn`](cases/33-switch_namespace_openvpn_.log) | vpn1 10.241.0.10 \| sw1 example [vrf-vpn] -- TCP output --> sw2 example VIP 10.240.2.12 \| sw1 network b 192.66.0.1 -- acb 192.66.0.11 | 184.143s |
| 26-06-22 06:38:08 | PASS | [`switch_namespace`](cases/34-switch_namespace_.log) | sw1 192.63.0.1 [vrf-example] <-- UDP output -- sw2 192.63.0.2 [vrf-example] \| both service L3 devices are enslaved to the same VRF name | 19.689s |
| 26-06-22 06:40:22 | PASS | [`switch_namespace_snat`](cases/35-switch_namespace_snat_.log) | ac1 -> sw2 example [vrf-snat] -- UDP output --> sw1 VIP 10.242.2.11 \| acb -> sw2 network b (no VRF) ----------------x same VIP | 133.957s |
| 26-06-22 06:41:18 | PASS | [`switch_output_perf`](cases/36-switch_output_perf_.log) | sw1 center 192.53.0.1 \| UDP output \| TCP output \| sw2 192.53.0.2 sw3 192.53.0.3 \| mixed output auth, ping RTT, and bandwidth samples | 55.540s |
| 26-06-22 06:41:47 | PASS | [`switch_ratelimit`](cases/37-switch_ratelimit_.log) | sw1 192.60.0.1 \| bridge device hi-example \| OpenVPN tcp/1194, tun1194, 10.60.0.0/24 \| rate limits are applied to bridge and OpenVPN devices | 29.473s |
| 26-06-22 06:42:24 | PASS | [`switch_route3`](cases/38-switch_route3_.log) | sw1 VIP 10.251.0.11 \| output \| sw2 VIP 10.251.0.12 \| output + static routes \| sw3 reaches sw1/sw2 loopback VIPs through nexthops | 36.882s |
| 26-06-22 06:42:38 | PASS | [`switch_setaddress`](cases/39-switch_setaddress_.log) | sw1 changes example bridge address from 192.72.0.1/24 to 192.73.0.1/24 and openvpn restart pushes the new route range | 13.033s |
| 26-06-22 06:45:21 | PASS | [`switch_tcp`](cases/40-switch_tcp_.log) | sw1 192.41.0.1 \| TCP output \| TCP output \| sw2 192.41.0.2 sw3 192.41.0.3 \| then sw3 output is moved from sw1 to sw2 | 163.565s |
| 26-06-22 06:47:32 | PASS | [`switch_udp`](cases/41-switch_udp_.log) | sw1 192.51.0.1 <----- UDP output ----- sw2 192.51.0.2 \| center switch accepts branch output over Docker mgmt network | 130.602s |
| 26-06-22 06:48:07 | PASS | [`switch_ztrust`](cases/42-switch_ztrust_.log) | vpn1 10.93.0.10 \| v OpenVPN tcp/1194 \| sw1 192.59.0.1:8081 \| ZTrust guest + knock gates service access | 34.905s |
