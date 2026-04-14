# 🔀 Proxy Setup Example

```text
             Google <------------ Internet -------------> GitHub
                                     ^
                                     |
                                     |
                           Central Switch(Singapore)      - 192.168.1.88
                                     ^
                                     |
                                     |
                                  Internet
                                     |
                                     |
                           Central Switch(Shanghai)       - 192.168.1.66
                                     ^
                                     |
                                     |
              Curl ------------> HTTP Proxy <------------- Chrome
```

## 🌐 Http Proxy

```bash
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.yaml << EOF
http:
- listen: 192.168.1.88:11082
  network: default

EOF
root@openlan:/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```

## 🧦 Socks Proxy

```bash
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.yaml << EOF
socks5:
- listen: 192.168.1.88:11083

EOF
root@openlan:/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```

## 🔁 TCP Reverse Proxy

```bash
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.yaml << EOF
tcp:
- listen: 192.168.1.66:11082
  target: [192.168.1.88:11082]

EOF
root@openlan:/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```
