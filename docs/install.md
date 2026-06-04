# 📦 Installation Guide

OpenLAN软件包含下面部分：

* 🖥️ **Central Switch** : 具有公网地址的CentOS服务器、云主机或者DMZ主机
* 💻 **Access Point** : 运行在企业内部的CentOS主机或者移动办公的PC上，无需公网地址

## 🖥️ CentOS

## 🖥️ Central Switch

您可以在CentOS7上通过下面步骤部署Central Switch软件：

1. 安装依赖的软件；

   ```bash
   yum install -y epel-release ebtables
   ```

2. 使用bin安装Central Switch软件；

   ```bash
   wget https://github.com/luscis/openlan/releases/download/v26.5.1/openlan-v26.5.1.amd64.bin
   chmod +x ./openlan-v26.5.1.amd64.bin && ./openlan-v26.5.1.amd64.bin
   ```

3. 配置 Central Switch 服务自启动；

   ```bash
   systemctl enable --now openlan-switch
   journalctl -u openlan-switch
   ```

4. 使用 `openlan` CLI 配置加密参数并保存；

   ```bash
   openlan crypt update --algorithm aes-128 --secret ea64d5b0c96c
   openlan crypt ls
   openlan reload --save
   ```

5. 使用 `openlan` CLI 添加 OpenLAN 网络；

   ```bash
   openlan network --name example add --address 192.11.0.1/24
   openlan network ls
   ```

6. 使用 `openlan` CLI 添加路由与 OpenVPN 接入；

   ```bash
   openlan network --name example route add --prefix 192.168.10.0/24
   openlan network --name example openvpn add \
     --listen :1194 \
     --protocol tcp \
     --subnet 10.99.0.0/24
   ```

7. 保存当前配置；

   ```bash
   openlan reload --save
   ```

8. 添加一个新的接入认证的用户；

   ```bash
   $ openlan user add --name hi@example
   hi@example  l6llot97yx  guest
   ```

9. 导出OpenVPN的客户端配置文件；

   在浏览器直接访问接口获取VPN Profile，弹出框中输入账户密码，替换`<access-ip>`为公网IP地址。

   ```bash
   curl -k https://<access-ip>:10000/get/network/example/ovpn
   ```

   在OpenVPN的客户端，`via URL`的方式自动导入，输入框中录入用户名密码。

   ```text
   https://<access-ip>:10000
   ```

## 💻 Access Point

同样的您也可以在CentOS7上通过下面步骤部署Access Point软件：

1. 安装依赖的软件；

   ```bash
   yum install -y epel-release
   ```

2. 使用bin安装Access Point软件；

   ```bash
   wget https://github.com/luscis/openlan/releases/download/v26.5.1/openlan-v26.5.1.amd64.bin
   chmod +x ./openlan-v26.5.1.amd64.bin && ./openlan-v26.5.1.amd64.bin nodeps
   ```

3. 添加一个新的网络配置；

   ```bash
   cd /etc/openlan/access
   cp access.yaml.example example.yaml
   vim example.yaml
   ```

   ```yaml
   protocol: tcp
   crypt:
     algorithm: aes-128
     secret: ea64d5b0c96c
   forward:
   - 192.168.10.0/24
   connection: example.net
   username: <your-name>@example
   password: <your-password>
   interface:
     address: 192.11.0.11/24
   ```

4. 配置Access Point服务自启动；

   ```bash
   systemctl enable --now openlan-access@example
   journalctl -u openlan-access@example
   ```

5. 检测网络是否可达；

   ```bash
   ping 192.11.0.1 -c 3
   ping 192.168.10.1 -c 3
   ```
