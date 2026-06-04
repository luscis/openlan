# 🐳 Deployment OpenLAN by Docker Compose

Please ensure you already installed the following softwares:

* docker
* docker-compose

## 📥 Download config's source package

```bash
wget https://github.com/luscis/openlan/releases/download/v26.5.1/config.tar.gz
```

## 📂 Unarchive it to your rootfs

```bash
tar -xvf config.tar.gz -C /opt
```

## ✏️ Now you can edit your network

```bash
[root@example openlan]# cd /opt/openlan/etc/openlan/switch/network
[root@example network]#
[root@example network]# cat ./example.yaml
name: example
bridge:
  address: 192.11.0.1/24
[root@example network]#
```

## 🔄 Update image version

You can find the latest version on [docker hub](https://hub.docker.com/r/luscis/openlan/tags).

```bash
[root@example network]# cd /opt/openlan
[root@example openlan]# sed -i -e 's/:latest.amd64/:v26.5.1.amd64/' docker-compose.yml
[root@example openlan]#
```

## 🚀 Bootstrap OpenLAN by compose

```bash
[root@example openlan]# docker-compose up -d
[root@example openlan]#
[root@example openlan]# docker-compose ps
      Name                    Command               State   Ports
-----------------------------------------------------------------
openlan_ipsec_1    /var/openlan/script/ipsec.sh     Up
openlan_switch_1   /var/openlan/script/switch ...   Up
```

## ⬆️ Upgrading OpenLAN and backup OpenVPN

```bash
[root@example openlan]# cd /opt/openlan
[root@example openlan]# mkdir -p var/openlan
[root@example openlan]# docker cp openlan_switch_1:/var/openlan/openvpn ./
[root@example openlan]# docker-compose down
[root@example openlan]# vi docker-compose.yml
```

```yaml
version: 2.3
services:
  ipsec:
    restart: always
    image: luscis/openlan:v26.5.1.amd64.deb
    privileged: true
    network_mode: host
    entrypoint: [/var/openlan/script/ipsec.sh]
    volumes:
      - /opt/openlan/etc/ipsecd.d:/etc/ipsec.d
      - /opt/openlan/run/pluto:/run/pluto
  switch:
    restart: always
    image: luscis/openlan:v26.5.1.amd64.deb
    privileged: true
    network_mode: host
    entrypoint: [/var/openlan/script/switch.sh, start]
    volumes:
      - /opt/openlan/etc/openlan:/etc/openlan
      - /opt/openlan/etc/ipsecd.d:/etc/ipsec.d
      - /opt/openlan/run/pluto:/run/pluto
```

```bash
[root@example openlan]# docker-compose up -d
```
