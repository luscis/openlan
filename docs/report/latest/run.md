# OpenLAN Test Report

- Title: OpenLAN test run started
- Generated: 26-05-28 07:42:49
- Passed: 33
- Failed: 0
- Total: 33

## Scenarios

| Time | Status | Scenario | Topology | Cost |
|---|---|---|---|---|
| 26-05-28 07:13:30 | PASS | [`access_admin_multi_login`](cases/01-access_admin_multi_login_.log) | Docker mgmt network: 172.251.0.0/24 | 13.757s |
| 26-05-28 07:13:49 | PASS | [`access_fail`](cases/02-access_fail_.log) | Docker mgmt network: 172.253.0.0/24 | 18.259s |
| 26-05-28 07:13:55 | PASS | [`access_openvpn_client_ping`](cases/03-access_openvpn_client_ping_.log) | Docker mgmt network: 172.253.0.0/24 | 5.798s |
| 26-05-28 07:14:27 | PASS | [`access_openvpn_perf`](cases/04-access_openvpn_perf_.log) | Docker mgmt network: 172.251.0.0/24 | 32.330s |
| 26-05-28 07:14:47 | PASS | [`access_openvpn_redirect`](cases/05-access_openvpn_redirect_.log) | Docker mgmt network: 172.249.0.0/24 | 19.954s |
| 26-05-28 07:14:58 | PASS | [`access_openvpn`](cases/06-access_openvpn_.log) | Docker mgmt network: 172.252.0.0/24 | 11.000s |
| 26-05-28 07:15:38 | PASS | [`access_openvpn_snat_vip`](cases/07-access_openvpn_snat_vip_.log) | Docker mgmt network: 172.250.0.0/24 | 39.798s |
| 26-05-28 07:15:43 | PASS | [`access_openvpn_tcp_reset`](cases/08-access_openvpn_tcp_reset_.log) | Docker mgmt network: 172.248.0.0/24 | 4.718s |
| 26-05-28 07:17:11 | PASS | [`access_pre_network_crypt`](cases/09-access_pre_network_crypt_.log) | Docker mgmt network: 172.251.0.0/24 | 88.580s |
| 26-05-28 07:17:27 | PASS | [`access_same_user_mutex`](cases/10-access_same_user_mutex_.log) | Docker mgmt network: 172.252.0.0/24 | 15.368s |
| 26-05-28 07:21:11 | PASS | [`access_snat_scope_matrix`](cases/11-access_snat_scope_matrix_.log) | Docker mgmt network: 172.249.0.0/24 | 224.442s |
| 26-05-28 07:21:32 | PASS | [`access_success`](cases/12-access_success_.log) | Docker mgmt network: 172.255.0.0/24 | 20.766s |
| 26-05-28 07:22:01 | PASS | [`proxy_http`](cases/13-proxy_http_.log) | Docker mgmt network: 172.252.0.0/24 | 28.562s |
| 26-05-28 07:22:36 | PASS | [`proxy_name_backends`](cases/14-proxy_name_backends_.log) | Docker mgmt network: 172.248.0.0/24 | 35.718s |
| 26-05-28 07:23:03 | PASS | [`proxy_name`](cases/15-proxy_name_.log) | Docker mgmt network: 172.249.0.0/24 | 26.378s |
| 26-05-28 07:23:32 | PASS | [`proxy_tcp`](cases/16-proxy_tcp_.log) | Docker mgmt network: 172.250.0.0/24 | 28.595s |
| 26-05-28 07:25:26 | PASS | [`switch_acl_default_action`](cases/17-switch_acl_default_action_.log) | Docker mgmt network: 172.254.1.0/24 | 114.901s |
| 26-05-28 07:28:15 | PASS | [`switch_acl`](cases/18-switch_acl_.log) | Docker mgmt network: 172.254.0.0/24 | 167.974s |
| 26-05-28 07:28:32 | PASS | [`switch_bgp`](cases/19-switch_bgp_.log) | Docker mgmt network: 172.244.0.0/24 | 17.185s |
| 26-05-28 07:29:24 | PASS | [`switch_dnat`](cases/20-switch_dnat_.log) | Docker mgmt network: 172.246.0.0/24 | 52.094s |
| 26-05-28 07:31:29 | PASS | [`switch_findhop`](cases/21-switch_findhop_.log) | Docker mgmt network: 172.243.0.0/24 | 124.539s |
| 26-05-28 07:31:46 | PASS | [`switch_ipsec_gre`](cases/22-switch_ipsec_gre_.log) | Docker mgmt network: 172.247.0.0/24 | 17.151s |
| 26-05-28 07:32:39 | PASS | [`switch_ipsec_vxlan_perf`](cases/23-switch_ipsec_vxlan_perf_.log) | Docker mgmt network: 172.247.0.0/24 | 53.443s |
| 26-05-28 07:33:29 | PASS | [`switch_ipsec_vxlan`](cases/24-switch_ipsec_vxlan_.log) | Docker mgmt network: 172.248.0.0/24 | 49.188s |
| 26-05-28 07:36:33 | PASS | [`switch_namespace_openvpn`](cases/25-switch_namespace_openvpn_.log) | Docker mgmt network: 172.240.0.0/24 | 184.114s |
| 26-05-28 07:36:53 | PASS | [`switch_namespace`](cases/26-switch_namespace_.log) | Docker mgmt network: 172.242.0.0/24 | 19.774s |
| 26-05-28 07:39:06 | PASS | [`switch_namespace_snat`](cases/27-switch_namespace_snat_.log) | Docker mgmt network: 172.241.0.0/24 | 133.837s |
| 26-05-28 07:40:02 | PASS | [`switch_output_perf`](cases/28-switch_output_perf_.log) | One center switch sw1 accepts mixed output dial-ins. | 55.572s |
| 26-05-28 07:40:32 | PASS | [`switch_ratelimit`](cases/29-switch_ratelimit_.log) | Docker mgmt network: 172.253.0.0/24 | 29.448s |
| 26-05-28 07:41:09 | PASS | [`switch_route3`](cases/30-switch_route3_.log) | Docker mgmt network: 172.251.0.0/24 | 36.933s |
| 26-05-28 07:41:53 | PASS | [`switch_tcp`](cases/31-switch_tcp_.log) | Docker mgmt network: 172.255.0.0/24 | 44.408s |
| 26-05-28 07:42:16 | PASS | [`switch_udp`](cases/32-switch_udp_.log) | Docker mgmt network: 172.254.0.0/24 | 23.215s |
| 26-05-28 07:42:49 | PASS | [`switch_ztrust`](cases/33-switch_ztrust_.log) | Docker mgmt network: 172.245.0.0/24 | 32.371s |
