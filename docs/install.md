# Preface

openlan软件包含下面部分：

* openlan switch具有公网地址的centos服务器、云主机或者dmz主机
* openlan point运行在企业内部的centos主机或者移动办公的pc上，没有公网地址
* openlan network管理员定义的逻辑网络

# CentOS

## OpenLAN Switch

您可以在centos7上通过下面步骤部署openlan switch软件：
1. 安装依赖的软件；
   ```
   yum install -y epel-release
   yum search centos-release-openstack
   yum install -y centos-release-openstack-train
   yum install -y rdma-core libibverbs
   ```
2. 使用yum安装openlan switch软件；
   ```
   yum install -y https://github.com/luscis/openlan/releases/download/v5.8.22/openlan-switch-5.8.22-1.el7.x86_64.rpm
   ```
3. 配置openlan switch服务自启动；
   ```
   systemctl enable openlan-switch
   systemctl start  openlan-switch
   ```
4. 配置预共享密钥以及加密算法；
   ```
   cd /etc/openlan/switch
   cp ./switch.json.example ./switch.json
   vim ./switch.json               ## 编辑switch.json配置文件
   {
     "protocol": "tcp",
     "crypt": {
       "algo": "aes-128",          ## 支持xor,aes-128,aes-192等对称加密算法
       "secret": "ea64d5b0c96c"
     }
   }
   openlan cfg co                  ## 配置预检查
   ```
   
5. 添加一个新的openlan网络；
   ```
   cd ./network
   cp ./network.json.example ./example.json
   vim ./example.json 
   {
       "name": "example",
       "provider": "openlan",
       "bridge": {
           "address": "172.32.10.10/24"       ## 本地地址
       },
       "subnet": {                            ## example网络的子网配置
           "start": "172.32.10.100",          ## 用于动态分配给point的起始地址
           "end": "172.32.10.150",            ## 截止地址
           "netmask": "255.255.255.0"         ## 子网掩码
       },
       "hosts": [                             ## 为point添加静态地址分配
           {
               "hostname": "pc-99",           ## point的主机名称
               "address": "172.32.10.99"      ## 分配的地址
           }
       ],
       "routes": [                            ## 注入给point的路由信息
           {
               "prefix": "192.168.10.0/24",
               "mode": "snat"                 ## 默认转发模型为snat，route模式将不会下发nat规则
           }
       ],
       "openvpn": {                           ## 配置网络支持OpenVPN接入
           "protocol": "tcp",                 ## 访问的协议类型如tcp或者udp
           "listen": "0.0.0.0:1194",          ## 对外提供的访问端口
           "subnet": "172.16.194.0/24"        ## OpenVPN的子网地址，可以是任意的内网地址
       }
   }
   openlan cfg co                             ## 配置预检查
   ```
6. 重启openlan switch服务；
   ```
   systemctl restart openlan-switch
   journalctl -u openlan-switch               ## 查看日志信息
   ```
7. 导出openvpn的客户端配置文件；
   ```
   cd /var/openlan/openvpn/example            ## openvpn的配置信息存放目录
   cat ./client.ovpn                          ## 导出后编辑remote配置项，替换0.0.0.0为公网IP地址
   ```
   或者通过http接口获取
   ```
   cat /etc/openlan/switch/token | md5sum | cut -b 1-12
   a01234abc00                                ## 获取口令
   curl -k https://a01234abc00@<access-ip>:10000/get/network/example/tcp1194.ovpn
                                              ## 替换access-ip为公网IP地址
   ```
8. 添加一个新的接入认证的用户；
   ```
   openlan us add --name hi@example               ## <用户名>@<网络>
   openlan us ls | grep example                   ## 查看随机密码
   hi@example  l6llot97yxulsw1qqbm07vn1 guest     ## <用户名>@<网络> 密码 角色 租期
   
   openlan us rm --name hi@example                ## 删除一个用户
   ```
## OpenLAN Point

同样的您也可以在centos7上通过下面步骤部署openlan point软件：
1. 使用yum安装openlan point软件；
   ```
   yum install -y https://github.com/luscis/openlan/releases/download/v5.6.4/openlan-point-5.6.4-1.el7.x86_64.rpm
   ```
2. 添加一个新的网络配置；
   ```
   cd /etc/openlan
   cp point.json.example example.json
   vim example.json                           ## <网络名>.json
   {
     "protocol": "tcp",                       ## 同上
     "crypt": {                               ## 同上
       "algo": "aes-128",
       "secret": "ea64d5b0c96c"
     },
     "connection": "example.net",             ## 默认端口10002，格式:<adderss>:<port>
     "username": "hi@example",                ## <用户名>@<网络>
     "password": "l6llot97yxulsw1qqbm07vn1"   ## 认证的密码
   }
   cat example.json | python -m json.tool     ## 配置预检查
   ```
3. 配置openlan point服务自启动；
   ```
   systemctl enable openlan-point@example
   systemctl start  openlan-point@example
   journalctl -u openlan-point@example        ## 查看日志信息
   ```
4. 检测网络是否可达；
   ```
   ping 172.32.10.10 -c 3
   ping 192.168.10.1 -c 3
   ```
