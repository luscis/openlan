# Setup Proxy
## Http Proxy
```
root@cloud:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@cloud:/opt/openlan/etc/openlan# cat > proxy.json << EOF
{
    "http": [{"listen": "172.168.1.88:11082"}] 
}
EOF
root@cloud: docker restart openlan_proxy_1
```
## Socks Proxy
## TCP Reverse Proxy
```
root@i:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@i:/opt/openlan/etc/openlan# cat > proxy.json << EOF
{
    "tcp": [
        {
            "listen": "192.168.1.66:11082",  # Penlan switch example network IP
            "target": ["192.168.1.88:11082"] #  Proxy Device IP
        }
    ]
}
EOF
root@i:/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```
