# Preface

OpenLAN软件包含下面部分：

* Central Switch : 具有公网地址的CentOS服务器、云主机或者DMZ主机
* Access Point : 运行在企业内部的CentOS主机或者移动办公的PC上，无需公网地址

# CentOS

## Central Switch

您可以在CentOS7上通过下面步骤部署Central Switch软件：
1. 安装依赖的软件；
   ```
   $ yum install -y epel-release
   ```
2. 使用bin安装Central Switch软件；
   ```
   $ wget https://github.com/luscis/openlan/releases/download/v25.4.1/openlan-v25.4.1.x86_64.bin
   $ chmod +x ./openlan-v25.4.1.x86_64.bin && ./openlan-v25.4.1.x86_64.bin
   ```
3. 配置Central Switch服务自启动；
   ```
   $ systemctl enable --now openlan-switch
   ```
4. 配置预共享密钥以及加密算法；
   ```
   $ cd /etc/openlan/switch
   $ cp ./switch.yaml.example ./switch.yaml
   $ vim ./switch.yaml
   protocol: tcp
   crypt:
     algorithm: aes-128
     secret: ea64d5b0c96c
   ```

1. 添加一个新的OpenLAN网络；
   ```
   $ cd ./network
   $ cp ./network.yaml.example ./example.yaml
   $ vim ./example.yaml 
   name: example
   bridge: 
     address: 172.32.10.10/24
   subnet:
     startAt: 172.32.10.100
     endAt: 172.32.10.150
   routes: 
   - prefix: 192.168.10.0/24
   openvpn:
     protocol: tcp
     listen: 0.0.0.0:1194
   ```
2. 重启Central Switch服务；
   ```
   $ systemctl restart openlan-switch
   $ journalctl -u openlan-switch
   ```
3. 添加一个新的接入认证的用户；
   ```
   $ openlan user add --name hi@example
   hi@example  l6llot97yx  guest
   ```
4. 导出OpenVPN的客户端配置文件；

   在浏览器直接访问接口获取VPN Profile，弹出框中输入账户密码，替换<access-ip>为公网IP地址。
   ```
   $ curl -k https://<access-ip>:10000/get/network/example/ovpn
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
   ```
2. 使用bin安装Access Point软件；
   ```
   $ wget https://github.com/luscis/openlan/releases/download/v25.4.1/openlan-v25.4.1.x86_64.bin
   $ chmod +x ./openlan-v25.4.1.x86_64.bin && ./openlan-v25.4.1.x86_64.bin nodeps
   ```
2. 添加一个新的网络配置；
   ```
   $ cd /etc/openlan/access
   $ cp access.yaml.example example.yaml
   $ vim example.yaml
   protocol: tcp
   crypt:
     algorithm: aes-128
     secret: ea64d5b0c96c
   forward:
   - 192.168.10.0/24
   connection: example.net
   username: <your-name>@example
   password: <your-password>
   ```
3. 配置Access Point服务自启动；
   ```
   $ systemctl enable --now openlan-access@example
   $ journalctl -u openlan-access@example
   ```
4. 检测网络是否可达；
   ```
   $ ping 172.32.10.10 -c 3
   $ ping 192.168.10.1 -c 3
   ```
