# Central Branch Example

## Topology

```
                                     Switch(Central) - 10.16.1.10/24
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
               Access Point            Access Point                 Access Point
             10.16.1.11/24             10.16.1.12/24                10.16.1.13/24

```

## Configure Central Switch

Generage a pre-shared key:

```
[root@switch ~]# uuidgen 
e108fe36-a2cd-43bc-82e2-f367aa429ed2
[root@switch ~]# 
```

Global configure with pre-share key:

```
[root@switch ~]# cd /etc/openlan/switch
[root@switch ~]# cat > switch.yaml <<EOF
crypt:
  secret: f367aa429ed2
EOF
```

Add a user network configuration:

```
[root@switch ~]# cd network
[root@switch ~]# cat > central.yaml <<EOF
name: central
bridge: 
  name: br-em1
  address: 10.16.1.10/24
subnet: 
  endAt: 10.16.1.100
  startAt: 10.16.1.44
hosts: 
- hostname: access1.hostname
  address: 10.16.1.11
openvpn: 
  listen: 0.0.0.0:1194
  subnet: 172.32.194.0/24
EOF
```

Add three access users on central network:

```

[root@switch ~]# openlan user add --name admin@central --role admin
[root@switch ~]# openlan user add --name access1@central
[root@switch ~]# openlan user add --name access2@central
[root@switch ~]# openlan user add --name access3@central
```



## Configure Access Point

Add a user network configuration:

```
[root@access1 ~]# cd /etc/openlan
[root@access1 ~]# cat > central.yaml <<EOF                          
crypt: 
  secret: f367aa429ed2
connection: public-ip-of-switch
username: access1@central
password: get-password-of-switch-administrator
EOF
[root@access1 ~]# cat central.yaml | python -m json.tool
```

Enable Access Point for central network:

```
systemctl enable --now openlan-access@central
```

Check journal log:

```
journalctl -u openlan-access@central
```



