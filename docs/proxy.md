# Setup Proxy

```
              Google--------------Internet---------------Githup
                                     |
                                     |      
                           Central Switch(Singapo)  - 192.168.1.88/24  
                                     |
                                     |
                                   互联网
                                     |
                                     |
                           Central Switch(Shanghai) - 192.168.1.66/24
                                     |
                                     |
              Curl--------------HTTP Proxy---------------Chrome
```
## Http Proxy
```
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.json << EOF
{
    "http": [{"listen": "192.168.1.88:11082"}] 
}
EOF
root@cloud: docker restart openlan_proxy_1
```
## Socks Proxy
## TCP Reverse Proxy    
```
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.json << EOF
{
    "tcp": [
        {
            "listen": "192.168.1.66:11082",  
            "target": ["192.168.1.88:11082"] 
        }
    ]
}
EOF
root@i:/opt/openlan/etc/openlan# cat proxy.json | python -m json.tool
root@i:/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```
