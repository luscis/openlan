# Deployment OpenLAN by docker compose

Please ensure you already installed the following softwares:
* docker
* docker-compose 

## Download config's source package

```
wget https://github.com/luscis/openlan/releases/download/v25.4.1/config.tar.gz
```

## Unarchive it to your roootfs

```
tar -xvf config.tar.gz -C /opt
```

## Now you can edit your network

```
[root@example openlan]# cd /opt/openlan/etc/openlan/switch/network
[root@example network]#
[root@example network]# cat ./example.json
{
    "name": "example",
    "bridge": {
        "address": "172.32.100.40/24"
    }
}
[root@example network]#
```

## Update image version

you can find latest version on [docker hub](<https://hub.docker.com/r/luscis/openlan/tags>)
```
[root@example network]# cd /opt/openlan
[root@example openlan]# sed -i -e 's/:latest.x86_64/:v25.4.1.x86_64/' docker-compose.yml
[root@example openlan]# 
```

## Bootstrap OpenLAN by compose

```
[root@example openlan]# docker-compose up -d
[root@example openlan]#
[root@example openlan]# docker-compose ps
      Name                    Command               State   Ports
-----------------------------------------------------------------
openlan_ipsec_1    /var/openlan/script/ipsec.sh     Up
openlan_proxy_1    /usr/bin/openlan-proxy -co ...   Up
openlan_switch_1   /var/openlan/script/switch ...   Up
```

## Upgrating OpenLAN and backup OpenVPN

```
[root@example openlan]# cd /opt/openlan
[root@example openlan]# mkdir -p var/openlan
[root@example openlan]# docker cp openlan_switch_1:/var/openlan/openvpn ./
[root@example openlan]# docker-compose down
[root@example openlan]# vi docker-compose.yml
version: "2.3"
services:
  ipsec:
    restart: always
    image: "luscis/openlan:v25.4.1.x86_64.deb"
    privileged: true
    network_mode: host
    entrypoint: ["/var/openlan/script/ipsec.sh"]
    volumes:
      - /opt/openlan/etc/ipsecd.d:/etc/ipsec.d
      - /opt/openlan/run/pluto:/run/pluto
  switch:
    restart: always
    image: "luscis/openlan:v25.4.1.x86_64.deb"
    privileged: true
    network_mode: "host"
    entrypoint: ["/var/openlan/script/switch.sh", "start"]
    volumes:
      - /opt/openlan/etc/openlan:/etc/openlan
      - /opt/openlan/etc/ipsecd.d:/etc/ipsec.d
      - /opt/openlan/run/pluto:/run/pluto
  proxy:
    restart: always
    image: "luscis/openlan:v25.4.1.x86_64.deb"
    privileged: true
    network_mode: "host"
    entrypoint: ["/usr/bin/openlan-proxy", "-conf", "/etc/openlan/proxy.json", "-log:file", "/dev/null"]
    volumes:
      - /opt/openlan/etc/openlan:/etc/openlan

[root@example openlan]# docker-compose up -d

```
