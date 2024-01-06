# Deployment OpenLAN by docker compose

Please ensure you already installed the following softwares:
* docker
* docker-compose 

## Download config's source package

```
wget https://github.com/luscis/openlan/releases/download/v24.01.01/config.tar.gz
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
[root@example openlan]# sed -i -e 's/:latest.x86_64/:v24.01.01.x86_64/' docker-compose.yml
[root@example openlan]# 
```

## Bootstrap OpenLAN by compose

```
[root@example openlan]# docker-compose up -d
Recreating openlan_confd_1 ... done
Recreating openlan_ovsdb-server_1 ... done
Recreating openlan_ovs-vswitchd_1 ... done
Recreating openlan_switch_1 ... done
Recreating openlan_proxy_1 ...
[root@example openlan]#
[root@example openlan]# docker ps
CONTAINER ID        IMAGE                             COMMAND                  CREATED             STATUS              PORTS               NAMES
aafb3cc2b8f9        luscis/openlan:v24.01.01.x86_64   "/usr/bin/openlan-..."   12 seconds ago      Up 12 seconds                           openlan_proxy_1
0bb9b586ed53        luscis/openlan:v24.01.01.x86_64   "/var/openlan/scri..."   13 seconds ago      Up 13 seconds                           openlan_switch_1
d5543d22db6e        luscis/openlan:v24.01.01.x86_64   "/var/openlan/scri..."   18 seconds ago      Up 17 seconds                           openlan_ovs-vswitchd_1
a1d86acdb6b4        luscis/openlan:v24.01.01.x86_64   "/var/openlan/scri..."   18 seconds ago      Up 18 seconds                           openlan_ovsdb-server_1
e42f200f6694        luscis/openlan:v24.01.01.x86_64   "/var/openlan/scri..."   19 seconds ago      Up 19 seconds                           openlan_confd_1
[root@example openlan]#
```