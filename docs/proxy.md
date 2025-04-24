# Setup Proxy

```
             Google <------------ Internet -------------> Githup
                                     ^
                                     |
                                     |      
                           Central Switch(Singapo)        - 192.168.1.88
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
## Http Proxy
```
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.yaml << EOF
http:
- listen: 192.168.1.88:11082

EOF
root@openlan::/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```
## Socks Proxy
## TCP Reverse Proxy    
```
root@openlan:/opt/openlan/etc/openlan# cd /opt/openlan/etc/openlan
root@openlan:/opt/openlan/etc/openlan# cat > proxy.yaml << EOF
tcp:
- listen: 192.168.1.66:11082,  
  target: [192.168.1.88:11082

EOF
root@openlan:/opt/openlan/etc/openlan# docker restart openlan_proxy_1
```
