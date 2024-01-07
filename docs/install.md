# Preface

OpenLAN软件包含下面部分：

* Central Switch : 具有公网地址的CentOS服务器、云主机或者DMZ主机
* Access Point : 运行在企业内部的CentOS主机或者移动办公的PC上，没有公网地址

# CentOS

## Central Switch

您可以在CentOS7上通过下面步骤部署Central Switch软件：
1. 安装依赖的软件；
   ```
   $ yum install -y epel-release
   $ yum search centos-release-openstack
   $ yum install -y centos-release-openstack-train
   $ yum install -y rdma-core libibverbs
   ```
2. 使用bin安装Central Switch软件；
   ```
   $ wget https://github.com/luscis/openlan/releases/download/v24.01.01/openlan-CentOS7.8.2003-v24.01.01.x86_64.bin
   $ chmod +x ./openlan-CentOS7.8.2003-v24.01.01.x86_64.bin
   $ ./openlan-CentOS7.8.2003-v24.01.01.x86_64.bin
   ```
3. 配置Central Switch服务自启动；
   ```
   $ systemctl enable --now openlan-confd
   $ systemctl enable --now openlan-switch
   ```
4. 配置预共享密钥以及加密算法；
   ```
   $ cd /etc/openlan/switch
   $ cp ./switch.json.example ./switch.json
   $ vim ./switch.json               ## 编辑switch.json配置文件
   {
     "protocol": "tcp",
     "crypt": {
       "algo": "aes-128",          ## 支持xor,aes-128,aes-192等对称加密算法
       "secret": "ea64d5b0c96c"
     }
   }
   $ openlan cfg co                  ## 配置预检查
   ```
   
5. 添加一个新的OpenLAN网络；
   ```
   $ cd ./network
   $ cp ./network.json.example ./example.json
   $ vim ./example.json 
   {
       "name": "example",
       "bridge": {
           "address": "172.32.10.10/24"       ## 一个唯一的子网地址，如共享二层网络填充本地地址
       },
       "subnet": {                            ## 网络的子网配置，如果没有动态地址分配可以忽略
           "start": "172.32.10.100",          ## 用于动态分配给接入point的起始地址
           "end": "172.32.10.150",            ## 用于动态分配的截止地址
           "netmask": "255.255.255.0"         ## 网络子网的掩码
       },
       "hosts": [                             ## 为point添加静态地址分配
           {
               "hostname": "pc-99",           ## 接入point的主机名称
               "address": "172.32.10.99"      ## 固定的地址
           }
       ],
       "routes": [                            ## 注入给point的路由信息
           {
               "prefix": "192.168.10.0/24"
           }
       ],
       "openvpn": {                           ## 配置网络支持OpenVPN接入
           "protocol": "tcp",                 ## 访问的协议类型如tcp或者udp
           "listen": "0.0.0.0:1194",          ## 对外提供的访问端口
       }
   }
   $ openlan cfg co                             ## 配置预检查
   ```
6. 重启Central Switch服务；
   ```
   $ systemctl restart openlan-switch
   $ journalctl -u openlan-switch               ## 查看日志信息
   ```
7. 添加一个新的接入认证的用户；
   ```
   $ openlan us add --name hi@example               ## <用户名>@<网络>
   $ openlan us ls --network example                ## 查看随机密码
   hi@example  l6llot97yx  guest                    ## <用户名>@<网络> 密码 角色 租期

   $ openlan us rm --name hi@example                ## 删除一个用户
   ```
8. 导出OpenVPN的客户端配置文件；

   在浏览器直接访问接口获取VPN Profile，弹出框中输入账户密码。
   ```
   $ curl -k https://<access-ip>:10000/get/network/example/ovpn
                                                    ## 替换access-ip为公网IP地址
   ```
   在OpenVPN的客户端，`via URL`的方式自动导入，输入框中录入用户名密码。
   ```
   https://<access-ip>:10000
   ```
## Access Point

同样的您也可以在CentOS7上通过下面步骤部署Access Point软件：

1. 安装依赖的软件；
   ```
   $ yum install -y epel-release
   $ yum search centos-release-openstack
   $ yum install -y centos-release-openstack-train
   $ yum install -y rdma-core libibverbs
   ```
2. 使用bin安装Access Point软件；
   ```
   $ wget https://github.com/luscis/openlan/releases/download/v24.01.01/openlan-CentOS7.8.2003-v24.01.01.x86_64.bin
   $ chmod +x ./openlan-CentOS7.8.2003-v24.01.01.x86_64.bin
   $ ./openlan-CentOS7.8.2003-v24.01.01.x86_64.bin
   ```
2. 添加一个新的网络配置；
   ```
   $ cd /etc/openlan
   $ cp point.json.example example.json
   $ vim example.json                           ## <网络名>.json
   {
     "protocol": "tcp",                         ## 同上
     "crypt": {                                 ## 同上
       "algo": "aes-128",
       "secret": "ea64d5b0c96c"
     },
     "connection": "example.net",               ## 默认端口10002，格式:<adderss>:<port>
     "username": "hi@example",                  ## <用户名>@<网络>
     "password": "l6llot97yxulsw1qqbm07vn1"     ## 认证的密码
   }
   $ cat example.json | python -m json.tool     ## 配置预检查
   ```
3. 配置Access Point服务自启动；
   ```
   $ systemctl enable --now openlan-point@example
   $ journalctl -u openlan-point@example        ## 查看日志信息
   ```
4. 检测网络是否可达；
   ```
   $ ping 172.32.10.10 -c 3
   $ ping 192.168.10.1 -c 3
   ```
