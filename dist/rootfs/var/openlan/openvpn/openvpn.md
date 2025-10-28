# OpenVPN
```
yum install -y epel-release
yum install -y openvpn
```
## Generate Diffie-Hellman
```
openssl dhparam -out /var/openlan/openvpn/dh.pem 1024  
```
## Generate TLS Auth Key
```
openvpn --genkey --secret /var/openlan/openvpn/ta.key
```

# Configure OpenVPN in Network
```
{
    "name": "example",
    "openvpn": {
        "listen": "0.0.0.0:1194",
        "subnet": "10.9.9.0/24"
    }
}
```

## Restart OpenLAN Switch Service
```
systemctl reload openlan-switch
```
