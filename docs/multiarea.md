# Multiple Area Example

## Topology

```
              192.168.1.20/24                                 192.168.1.21/24
                     |                                                 |
                  OLAP1 -- Hotal Wifi --> OLSW(NJ) <--- Other Wifi --- OLAP2
                                            |
                                            |
                                         Internet
                                            |
                                            |
                                         OLSW(SH) - 192.168.1.10/24
                                            |
                                            |
                   +------------------------+---------------------------+
                   ^                        ^                           ^
                   |                        |                           |
              Office Wifi               Home Wifi                  Hotal Wifi     
                   |                        |                           |
                 OLAP3                    OLAP4                       OLAP5
            192.168.1.11/24           192.168.1.12/24             192.168.1.13/24
```

## Configure OLSW for Nanjing

配置交换机：

```
[root@olsw-nj ~]# cd /etc/openlan/switch
[root@olsw-nj ~]# cat > switch.json <<EOF
{
  "cert": {
    "dir": "/var/openlan/cert"
  },
  "http": {
    "public": "/var/openlan/public"
  },
  "crypt": {
    "secret": "f367aa429ed2"
  }
}
EOF
```

配置网络：

```
[root@olsw-nj ~]# cd network
[root@olsw-nj ~]# cat > private.json <<EOF
{
  "name": "private",
  "bridge": {
    "name": "br-em2",
    "address": "192.168.1.66/24"
  },
  "subnet": {
    "end": "192.168.1.99",
    "netmask": "255.255.255.0",
    "start": "192.168.1.80"
  },
  "openvpn": {
    "listen": "0.0.0.0:1166",
    "subnet": "172.32.66.0/24"
  }
}
EOF
[root@olsw-nj ~]# openlan cfg co
[root@olsw-sh ~]# systemctl restart openlan-switch
```

添加认证用户：

```
[root@olsw-nj ~]# openlan us add --name admin@private --role admin
[root@olsw-nj ~]# openlan us add --name olap1@private
[root@olsw-nj ~]# openlan us add --name olap2@private
```

## Configure OLSW for ShangHai

配置交换机：

```
[root@olsw-sh ~]# cd /etc/openlan/switch
[root@olsw-sh ~]# cat > switch.json <<EOF
{
  "cert": {
    "dir": "/var/openlan/cert"
  },
  "http": {
    "public": "/var/openlan/public"
  },
  "crypt": {
    "secret": "7519e54d12c5"
  }
}
EOF
```

配置网络：

```
[root@olsw-sh ~]# cd network
[root@olsw-sh ~]# cat > private.json <<EOF
{
  "name": "private",
  "bridge": {
    "name": "br-em2",
    "address": "192.168.1.88/24"
  },
  "subnet": {
    "end": "192.168.1.150",
    "netmask": "255.255.255.0",
    "start": "192.168.1.100"
  },
  "openvpn": {
    "listen": "0.0.0.0:1188",
    "subnet": "172.32.88.0/24"
  },
  "links": [
    {
      "connection": "address-of-olsw-nj",
      "password": "get-it-from-olsw-nj",
      "username": "admin",
      "crypt": { 
         "secret": "f367aa429ed2" 
      }
    }
  ]
}
EOF
[root@olsw-sh ~]# openlan cfg co
[root@olsw-sh ~]# systemctl restart openlan-switch
```

添加认证用户：

```
[root@olsw-sh ~]# openlan us add --name admin@private --role admin
[root@olsw-sh ~]# openlan us add --name olap3@private
[root@olsw-sh ~]# openlan us add --name olap4@private
[root@olsw-sh ~]# openlan us add --name olap5@private
```

