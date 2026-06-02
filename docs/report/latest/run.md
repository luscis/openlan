# OpenLAN Test Report

- Title: OpenLAN test run started
- Generated: 26-06-02 09:53:04
- Passed: 37
- Failed: 0
- Total: 37

## Scenarios

| Time | Status | Scenario | Topology | Cost |
|---|---|---|---|---|
| 26-06-02 09:12:19 | PASS | [`access_admin_multi_login`](cases/01-access_admin_multi_login_.log) | sw1(center) 172.251.0.241 / example; \| tcp access          \| tcp access; ac1(admin)            ac2(admin) | 14.309s |
| 26-06-02 09:13:04 | PASS | [`access_client_qos`](cases/02-access_client_qos_.log) | sw1(center) 192.91.0.1; \| OpenVPN tcp/1194, 10.91.0.0/24; vpn1 10.91.0.10; QoS rule add/update/remove on vpn1@example | 44.856s |
| 26-06-02 09:13:22 | PASS | [`access_fail`](cases/03-access_fail_.log) | sw1(center) 172.253.0.241 / 192.31.0.1; \| tcp access with bad credentials; acbad asks for 192.31.0.11 | 18.383s |
| 26-06-02 09:13:28 | PASS | [`access_openvpn_client_ping`](cases/04-access_openvpn_client_ping_.log) | sw1(center) 192.42.0.1; \| OpenVPN tcp/1194         \| OpenVPN tcp/1194; vpn1 10.97.0.10 <----------> vpn2 10.97.0.11 | 5.924s |
| 26-06-02 09:14:01 | PASS | [`access_openvpn_perf`](cases/05-access_openvpn_perf_.log) | sw1(center) 192.55.0.1; \| OpenVPN tcp/1194 + udp/1195; vpn1 10.95.0.0/24; \| ping RTT and iperf3 TCP/UDP samples | 32.637s |
| 26-06-02 09:14:21 | PASS | [`access_openvpn_redirect`](cases/06-access_openvpn_redirect_.log) | vpn1 10.97.0.10; v OpenVPN tcp/1194; sw1 192.53.0.1  -- output/route -->  sw2 192.53.0.2; VIP 10.253.0.11                   VIP 10.253.0.12 | 19.913s |
| 26-06-02 09:14:32 | PASS | [`access_openvpn`](cases/07-access_openvpn_.log) | sw1(center) 192.41.0.1; \| OpenVPN AES tcp/1194     \| OpenVPN SM4 tcp/1194; vpn1                      vpn2 | 11.060s |
| 26-06-02 09:15:12 | PASS | [`access_openvpn_snat_vip`](cases/08-access_openvpn_snat_vip_.log) | vpn1 10.96.0.10; v OpenVPN tcp/1194; sw1 192.52.0.1  -- output + SNAT -->  sw2 192.52.0.2; VIP 10.252.0.12 | 39.741s |
| 26-06-02 09:15:17 | PASS | [`access_openvpn_tcp_reset`](cases/09-access_openvpn_tcp_reset_.log) | sw1(center) 192.54.0.1:8082; \| OpenVPN tcp/1194, 10.92.0.0/24; vpn1 10.92.0.10; \| INPUT tcp-reset rule toggles HTTP reachability | 4.785s |
| 26-06-02 09:16:45 | PASS | [`access_pre_network_crypt`](cases/10-access_pre_network_crypt_.log) | sw1(center) 172.251.0.241; \| network a crypt          \| network b global crypt; ac clients 192.61.0.11-17   ac clients 192.62.0.11-13 | 88.616s |
| 26-06-02 09:17:01 | PASS | [`access_same_user_mutex`](cases/11-access_same_user_mutex_.log) | sw1(center) 172.252.0.241 / example; \| tcp access          \| tcp access; ac1(t1)              ac2(t1) | 15.422s |
| 26-06-02 09:20:45 | PASS | [`access_snat_scope_matrix`](cases/12-access_snat_scope_matrix_.log) | vpn/access on sw1 network a,b; sw1: a=192.53.0.1 b=192.54.0.1 int=192.55.0.1; \| int output uplink | 224.705s |
| 26-06-02 09:22:40 | PASS | [`access_success`](cases/13-access_success_.log) | sw1(center) 172.255.0.241 / 192.11.0.1; \| tcp access          \| tcp access; ac1 192.11.0.11       ac2 192.11.0.12 | 114.409s |
| 26-06-02 09:23:09 | PASS | [`proxy_http`](cases/14-proxy_http_.log) | sw1 proxy client 192.52.0.1; \| wget via local Ceci HTTP proxy; sw1 openceci(http) -- output --> sw2 192.52.0.2:18081 | 28.625s |
| 26-06-02 09:23:45 | PASS | [`proxy_name_backends`](cases/15-proxy_name_backends_.log) | sw1 openceci(name); \| domain A         \| domain B; sw2 dnsmasq        sw3 dnsmasq; 192.55.0.2         192.55.0.3 | 36.275s |
| 26-06-02 09:24:12 | PASS | [`proxy_name`](cases/16-proxy_name_.log) | sw1 name client 192.54.0.1; \| nslookup via local Ceci name proxy; sw1 openceci(name) -- output --> sw2 dnsmasq 192.54.0.2:5300 | 26.522s |
| 26-06-02 09:24:40 | PASS | [`proxy_tcp`](cases/17-proxy_tcp_.log) | sw1 proxy client 192.53.0.1; \| wget via local Ceci TCP proxy; sw1 openceci(tcp) -- output --> sw2 192.53.0.2:18082 | 28.549s |
| 26-06-02 09:25:26 | PASS | [`service_http`](cases/18-service_http_.log) | client wget with Host header on sw1; sw1 Ceci HTTP service; ^ route groups          ^ global backend | 45.509s |
| 26-06-02 09:26:08 | PASS | [`service_tcp`](cases/19-service_tcp_.log) | client wget on sw1; sw1 Ceci TCP service -- output --> sw2 backend 192.56.0.2:18083 | 42.373s |
| 26-06-02 09:28:03 | PASS | [`switch_acl_default_action`](cases/20-switch_acl_default_action_.log) | sw1 192.62.0.1  -- UDP output -->  sw2 192.62.0.2; +-- default drop/accept ----> VIP 10.254.1.12:80/ICMP | 115.082s |
| 26-06-02 09:30:51 | PASS | [`switch_acl`](cases/21-switch_acl_.log) | sw1 192.61.0.1  -- UDP output -->  sw2 192.61.0.2; +------ ACL checks ----------> VIP 10.254.0.12:80/ICMP | 167.956s |
| 26-06-02 09:31:09 | PASS | [`switch_bgp`](cases/22-switch_bgp_.log) | sw1 172.244.0.241 / AS65101  <---- BGP ---->  sw2 172.244.0.242 / AS65102; svc 192.54.0.1                              svc 192.54.0.2 | 17.276s |
| 26-06-02 09:32:17 | PASS | [`switch_dhcp`](cases/23-switch_dhcp_.log) | sw1 DHCP server 192.67.0.1; \| physical output          \| tcp access tunnel; veth-dhcp <-> ns client      ac1 tap bridge | 68.690s |
| 26-06-02 09:33:10 | PASS | [`switch_dnat`](cases/24-switch_dnat_.log) | sw1 192.58.0.1  -- UDP output -->  sw2 192.58.0.2; +----------- DNAT example:80 -> 127.0.0.1:8080 | 52.139s |
| 26-06-02 09:35:12 | PASS | [`switch_findhop`](cases/25-switch_findhop_.log) | sw0 VIP 10.243.0.10; \| network a             \| network b; sw1.0 ------------------- sw1.1; +--------- sw2 ---------+ | 122.434s |
| 26-06-02 09:37:02 | PASS | [`switch_ipsec_gre`](cases/26-switch_ipsec_gre_.log) | sw1 172.247.0.241  <==== IPSec ====>  sw2 172.247.0.242; svc 192.57.0.1     <---- GRE output -- svc 192.57.0.2 | 109.573s |
| 26-06-02 09:37:55 | PASS | [`switch_ipsec_vxlan_perf`](cases/27-switch_ipsec_vxlan_perf_.log) | sw1 172.247.0.241  <==== optional IPSec ====>  sw2 172.247.0.242; svc 192.57.0.1       <---- VxLAN output ------ svc 192.57.0.2 | 53.520s |
| 26-06-02 09:40:35 | PASS | [`switch_ipsec_vxlan`](cases/28-switch_ipsec_vxlan_.log) | sw1 172.248.0.241  <==== IPSec ====>  sw2 172.248.0.242; svc 192.56.0.1    <---- VxLAN output - svc 192.56.0.2 | 159.400s |
| 26-06-02 09:43:39 | PASS | [`switch_namespace_openvpn`](cases/29-switch_namespace_openvpn_.log) | vpn1 10.241.0.10; sw1 example [vrf-vpn]  -- TCP output -->  sw2 example VIP 10.240.2.12; sw1 network b 192.66.0.1 -- acb 192.66.0.11 | 184.204s |
| 26-06-02 09:43:59 | PASS | [`switch_namespace`](cases/30-switch_namespace_.log) | sw1 192.63.0.1 [vrf-example]  <-- UDP output --  sw2 192.63.0.2 [vrf-example]; both service L3 devices are enslaved to the same VRF name | 19.825s |
| 26-06-02 09:46:13 | PASS | [`switch_namespace_snat`](cases/31-switch_namespace_snat_.log) | ac1 -> sw2 example [vrf-snat] -- UDP output --> sw1 VIP 10.242.2.11; acb -> sw2 network b (no VRF) ----------------x same VIP | 133.892s |
| 26-06-02 09:47:09 | PASS | [`switch_output_perf`](cases/32-switch_output_perf_.log) | sw1 center 192.53.0.1; \| UDP output       \| TCP output; sw2 192.53.0.2     sw3 192.53.0.3; mixed output auth, ping RTT, and bandwidth samples | 55.493s |
| 26-06-02 09:47:38 | PASS | [`switch_ratelimit`](cases/33-switch_ratelimit_.log) | sw1 192.60.0.1; \| bridge device hi-example; \| OpenVPN tcp/1194, tun1194, 10.60.0.0/24; rate limits are applied to bridge and OpenVPN devices | 29.412s |
| 26-06-02 09:48:15 | PASS | [`switch_route3`](cases/34-switch_route3_.log) | sw1 VIP 10.251.0.11; \| output; sw2 VIP 10.251.0.12; \| output + static routes; sw3 reaches sw1/sw2 loopback VIPs through nexthops | 36.776s |
| 26-06-02 09:50:45 | PASS | [`switch_tcp`](cases/35-switch_tcp_.log) | sw1 192.41.0.1; \| TCP output    \| TCP output; sw2 192.41.0.2   sw3 192.41.0.3; then sw3 output is moved from sw1 to sw2 | 150.312s |
| 26-06-02 09:52:29 | PASS | [`switch_udp`](cases/36-switch_udp_.log) | sw1 192.51.0.1  <----- UDP output -----  sw2 192.51.0.2; center switch accepts branch output over Docker mgmt network | 103.665s |
| 26-06-02 09:53:04 | PASS | [`switch_ztrust`](cases/37-switch_ztrust_.log) | vpn1 10.93.0.10; v OpenVPN tcp/1194; sw1 192.59.0.1:8081; ZTrust guest + knock gates service access | 34.976s |
