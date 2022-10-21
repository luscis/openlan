# Central Branch Example

## Topology

```
                                    OLSW(Central) - 10.16.1.10/24
                                            ^
                                            |
                                         Wifi(DNAT)
                                            |
                                            |
                   +---------------------Internet-----------------------+
                   ^                        ^                           ^
                   |                        |                           |
                 Branch1                  Branch2                     Branch3     
                   |                        |                           |
                 OLAP1                    OLAP2                       OLAP3
             10.16.1.11/24             10.16.1.12/24                10.16.1.13/24

```

## Configure OLSW

生成预共享密钥：

```
[root@olsw ~]# uuidgen 
e108fe36-a2cd-43bc-82e2-f367aa429ed2
[root@olsw ~]# 
```

交换机配置：

```
[root@olsw ~]# cd /etc/openlan/switch
[root@olsw ~]# cat > switch.json <<EOF
{
  "cert": {
    "dir": "/var/openlan/cert"
  },
  "http": {
    "public": "/var/openlan/public"
  },
  "inspect": [
    "neighbor", 
    "online"
  ],
  "crypt": {
    "secret": "f367aa429ed2"
  }
}
EOF
```

添加网络配置：

```
[root@olsw ~]# cd network
[root@olsw ~]# cat > central.json <<EOF
{
  "name": "central",
  "bridge": {
    "name": "br-em1",
    "address": "10.16.1.10/24"
  },
  "subnet": {
    "end": "10.16.1.100",
    "netmask": "255.255.255.0",
    "start": "10.16.1.44"
  },
  "hosts": [
     {
       "hostname": "olap1.hostname",
       "address": "10.16.1.11"
     }
  ],
  "openvpn": {
    "listen": "0.0.0.0:1194",
    "subnet": "172.32.194.0/24"
  }
}
EOF
```

添加接入认证的用户：

```

[root@olsw ~]# openlan us add --name admin@central --role admin
[root@olsw ~]# openlan us add --name olap1@central
[root@olsw ~]# openlan us add --name olap2@central
[root@olsw ~]# openlan us add --name olap3@central
```



## Configure OLAP

添加一个网络：

```
[root@olap1 ~]# cd /etc/openlan
[root@olap1 ~]# cat > central.json <<EOF                          
{
  "crypt": {
    "secret": "f367aa429ed2"
  },
  "connection": "public-ip-of-olsw",
  "username": "olap1@central",
  "password": "get-password-of-olsw-administrator"
}
EOF
[root@olap1 ~]# cat central.json | python -m json.tool
```

配置网络服务：

```
systemctl enable --now openlan-point@central
```

检查启动日志：

```
journalctl -u openlan-point@central
```



