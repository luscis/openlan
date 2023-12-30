# Multiple Area Example

## Topology

```
              192.168.1.20/24                                 192.168.1.21/24
                     |                                                 |
               Access1 -- Hotal Wifi --> Switch(NJ) <--- Other Wifi --- Access2
                                            |
                                            |
                                         Internet
                                            |
                                            |
                                         Switch(SH) - 192.168.1.10/24
                                            |
                                            |
                   +------------------------+---------------------------+
                   ^                        ^                           ^
                   |                        |                           |
              Office Wifi               Home Wifi                  Hotal Wifi     
                   |                        |                           |
                Access3                 Access4                       Access5
            192.168.1.11/24           192.168.1.12/24             192.168.1.13/24
```

## Configure Central Switch for Nanjing

Global configure:

```
[root@switch-nj ~]# cd /etc/openlan/switch
[root@switch-nj ~]# cat > switch.json <<EOF
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

Network configure:

```
[root@switch-nj ~]# cd network
[root@switch-nj ~]# cat > private.json <<EOF
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
[root@switch-nj ~]# openlan cfg co
[root@switch-sh ~]# systemctl restart openlan-switch
```

Add two access users on private network:

```
[root@switch-nj ~]# openlan us add --name admin@private --role admin
[root@switch-nj ~]# openlan us add --name access1@private
[root@switch-nj ~]# openlan us add --name access2@private
```

## Configure Central Switch for ShangHai

Global configure:

```
[root@switch-sh ~]# cd /etc/openlan/switch
[root@switch-sh ~]# cat > switch.json <<EOF
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

Network configure:

```
[root@switch-sh ~]# cd network
[root@switch-sh ~]# cat > private.json <<EOF
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
      "connection": "address-of-switch-nj",
      "password": "get-it-from-switch-nj",
      "username": "admin",
      "crypt": { 
         "secret": "f367aa429ed2" 
      }
    }
  ]
}
EOF
[root@switch-sh ~]# openlan cfg co
[root@switch-sh ~]# systemctl restart openlan-switch
```

Add three access users on private network:

```
[root@switch-sh ~]# openlan us add --name admin@private --role admin
[root@switch-sh ~]# openlan us add --name access3@private
[root@switch-sh ~]# openlan us add --name access4@private
[root@switch-sh ~]# openlan us add --name access5@private
```

