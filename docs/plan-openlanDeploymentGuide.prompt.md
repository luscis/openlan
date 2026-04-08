# OpenLAN 完整部署指南

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [部署前准备](#部署前准备)
- [详细部署指南](#详细部署指南)
  - [Linux 源码编译部署](#linux-源码编译部署)
  - [Linux 二进制安装](#linux-二进制安装)
  - [Docker 单机部署](#docker-单机部署)
  - [Docker Compose 多服务部署](#docker-compose-多服务部署)
  - [Kubernetes 部署](#kubernetes-部署)
  - [Windows 部署](#windows-部署)
  - [macOS 部署](#macos-部署)
- [配置参考](#配置参考)
  - [Switch 配置详解](#switch-配置详解)
  - [Access 配置详解](#access-配置详解)
  - [Network 配置详解](#network-配置详解)
  - [Proxy 配置详解](#proxy-配置详解)
  - [高级配置](#高级配置)
- [常见场景配置](#常见场景配置)
  - [分支互联场景](#分支互联场景)
  - [多区域互联场景](#多区域互联场景)
  - [零信任网络场景](#零信任网络场景)
  - [HTTP/SOCKS5 代理场景](#httpsocks5-代理场景)
- [故障排查指南](#故障排查指南)
- [运维管理](#运维管理)
- [性能调优](#性能调优)
- [安全最佳实践](#安全最佳实践)
- [常见问题 FAQ](#常见问题-faq)

---

## 概述

OpenLAN 是一个开源的虚拟以太网技术项目，能够在广域网（WAN）上实现局域网（LAN）数据包的传输，并能够在多个用户空间中建立虚拟以太网络。OpenLAN 具有以下核心特性：

### 核心组件

| 组件 | 说明 | 部署位置 |
|------|------|---------|
| **Central Switch** | 中央交换机，具有公网地址，作为网络中心 | 公网云主机或DMZ区域 |
| **Access Point** | 接入点，运行在企业内部或移动办公设备 | 企业内网或移动设备 |
| **OpenVPN** | VPN 协议支持，多平台客户端 | 内置于 Central Switch |
| **IPSec** | IPSec 隧道加密 | 可选，用于中心互联 |
| **Proxy** | HTTP/SOCKS5 代理服务 | 可选模块 |

### 主要功能

- **虚拟网络划分**：创建多个网络空间，提供逻辑网络隔离
- **多种传输协议**：支持 TCP、TLS、UDP、KCP、WS、WSS 等
- **灵活加密**：支持 AES-128、AES-192、AES-256 等加密算法
- **用户认证**：简单的用户名密码认证机制
- **QoS 管理**：支持带宽限速和流量管理
- **零信任网络**：支持零信任安全模型
- **代理能力**：内置 HTTP/SOCKS5 代理功能
- **路由管理**：支持 SNAT、DNAT、静态路由等

---

## 快速开始

### 3分钟快速启动（基于 Docker Compose）

#### 前置要求

- 安装 Docker 和 Docker Compose
- Linux 系统（CentOS 7+ 或 Ubuntu 18.04+）
- 至少 2GB RAM 和 1GB 磁盘空间

#### 快速启动步骤

```bash
# 1. 下载配置包
wget https://github.com/luscis/openlan/releases/download/v25.4.1/config.tar.gz

# 2. 解压到系统目录
tar -xvf config.tar.gz -C /opt

# 3. 进入配置目录
cd /opt/openlan

# 4. 启动服务
docker-compose up -d

# 5. 验证服务状态
docker-compose ps

# 6. 查看日志
docker-compose logs -f switch
```

#### 验证部署成功

```bash
# 检查服务是否运行
docker-compose ps

# 预期输出：
# NAME              STATUS      PORTS
# openlan_switch_1  Up          
# openlan_proxy_1   Up
# openlan_ipsec_1   Up

# 检查交换机日志
docker-compose logs switch | grep "listening"
```

#### 创建第一个用户

```bash
# 进入 switch 容器
docker-compose exec switch bash

# 添加用户
openlan user add --name demo@example

# 输出示例：
# demo@example  l6llot97yx  guest
```

#### 获取 OpenVPN 配置

```bash
# 方式一：通过 curl 下载
curl -k https://<your-central-switch-ip>:10000/get/network/example/ovpn > example.ovpn

# 方式二：通过浏览器访问
# https://<your-central-switch-ip>:10000
# 选择 "Download Profile" 并输入用户名密码
```

> **提示**：默认网络名称为 `example`，网络地址为 `172.32.100.40/24`，可在配置文件中修改

---

## 部署前准备

### 系统要求

#### 最低配置

| 项目 | 最低要求 | 推荐配置 |
|------|---------|---------|
| CPU | 1 核 | 2 核+ |
| 内存 | 512 MB | 2 GB+ |
| 磁盘 | 100 MB | 500 MB+ |
| 网络 | 1 Mbps | 10 Mbps+ |
| 内核 | Linux 3.10+ | Linux 4.15+ |

#### 操作系统支持

- **Linux**：CentOS 7+、Ubuntu 18.04+、Debian 10+
- **Windows**：Windows 10 (21H2)+、Windows Server 2019+
- **macOS**：macOS 10.15+
- **Docker**：支持容器化部署

### 网络要求

#### 防火墙规则

```bash
# Central Switch 所需的端口
22   # SSH 管理
1194 # OpenVPN 默认端口
10000 # 管理接口 HTTPS
10002 # 接入点通信端口
500  # IPSec IKE（可选）
4500 # IPSec NAT-T（可选）
```

#### 网络连接要求

```bash
# 从 Access Point 到 Central Switch
- 需要出站连接到 Central Switch
- 支持 TCP、UDP、KCP 等多种协议
- 可通过代理或 NAT 穿透

# Central Switch 之间
- 建议 VPC 内网直连或 IPSec 隧道
- 支持多链路负载均衡
```

### 依赖软件安装

#### CentOS/RHEL

```bash
# 基础工具
sudo yum install -y epel-release
sudo yum install -y gcc gcc-c++ make
sudo yum install -y libmnl-devel libnftnl-devel
sudo yum install -y openssl-devel

# 可选：用于编译
sudo yum install -y git go-toolset
```

#### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install -y build-essential
sudo apt-get install -y libmnl-dev libnftnl-dev
sudo apt-get install -y libssl-dev pkg-config

# 可选：用于编译
sudo apt-get install -y git golang-go
```

#### Docker 安装

```bash
# CentOS
sudo yum install -y docker-ce docker-compose-plugin

# Ubuntu
sudo apt-get install -y docker.io docker-compose

# 启动 Docker
sudo systemctl enable --now docker

# 添加当前用户到 docker 组
sudo usermod -aG docker $USER
```

### 下载与验证

```bash
# 下载最新版本
VERSION="v25.4.1"
ARCH="amd64"  # 或 arm64
wget https://github.com/luscis/openlan/releases/download/${VERSION}/openlan-${VERSION}.${ARCH}.bin

# 验证完整性（可选）
# wget https://github.com/luscis/openlan/releases/download/${VERSION}/SHA256SUMS
# sha256sum -c SHA256SUMS

# 赋予执行权限
chmod +x openlan-${VERSION}.${ARCH}.bin
```

---

## 详细部署指南

### Linux 源码编译部署

#### 获取源代码

```bash
# 克隆项目
git clone https://github.com/luscis/openlan.git
cd openlan

# 初始化依赖
make init

# 更新依赖
make update
```

#### 编译

```bash
# 编译所有平台
make pkg

# 或仅编译 Linux
make linux-bin

# 输出目录
# build/
# ├── openlan-v25.4.1.amd64/  # Linux 二进制目录
# ├── openlan-v25.4.1.amd64.tar.gz
# └── ...
```

#### 编译参数

```bash
# 指定架构
make ARCH=arm64 linux-bin

# 指定输出目录
BD="/custom/build/path" make linux-bin

# 查看 Makefile 中的版本信息
cat Makefile | grep -E "VER|LDFLAGS"
```

#### 安装编译产物

```bash
# 进入编译输出目录
cd build/openlan-v25.4.1.amd64

# 安装
sudo ./install.sh

# 或手动复制
sudo cp -r * /

# 验证安装
which openlan-switch
which openlan-access
```

#### 配置系统服务

```bash
# 使用 systemd 管理
sudo systemctl daemon-reload
sudo systemctl enable openlan-switch
sudo systemctl enable openlan-access@example

# 启动服务
sudo systemctl start openlan-switch

# 查看状态
systemctl status openlan-switch
journalctl -u openlan-switch -f
```

### Linux 二进制安装

#### 下载和安装

```bash
# 下载二进制包
VERSION="v25.4.1"
ARCH="amd64"
wget https://github.com/luscis/openlan/releases/download/${VERSION}/openlan-${VERSION}.${ARCH}.bin

# 赋予执行权限
chmod +x openlan-${VERSION}.${ARCH}.bin

# 执行安装程序（假设依赖已安装）
sudo ./openlan-${VERSION}.${ARCH}.bin

# 如果系统缺少依赖，可以跳过依赖检查
sudo ./openlan-${VERSION}.${ARCH}.bin nodeps
```

#### 验证安装

```bash
# 检查安装路径
ls -la /usr/bin/openlan-*
ls -la /etc/openlan/

# 检查版本
openlan-switch --version
openlan-access --version

# 查看配置目录
tree /etc/openlan/
```

#### 初始化配置

```bash
# Central Switch 配置
cd /etc/openlan/switch
cp ./switch.yaml.example ./switch.yaml
vim ./switch.yaml

# Access Point 配置
cd /etc/openlan/access
cp access.yaml.example example.yaml
vim example.yaml
```

#### 启动服务

```bash
# 启动 Central Switch
sudo systemctl enable --now openlan-switch

# 验证
sudo systemctl status openlan-switch
sudo journalctl -u openlan-switch -n 50

# 启动 Access Point
sudo systemctl enable --now openlan-access@example
sudo systemctl status openlan-access@example
```

### Docker 单机部署

#### 使用 Docker 运行 Switch

```bash
# 创建配置目录
mkdir -p /opt/openlan/etc/openlan/switch/network
mkdir -p /opt/openlan/var/openlan

# 复制配置示例
cd /opt/openlan/etc/openlan/switch
wget https://raw.githubusercontent.com/luscis/openlan/master/config/switch.yaml.example
wget https://raw.githubusercontent.com/luscis/openlan/master/config/network.yaml.example

cp switch.yaml.example switch.yaml
cp network.yaml.example network/example.yaml

# 运行 Switch 容器
docker run -d \
  --name openlan-switch \
  --privileged \
  --network host \
  -v /opt/openlan/etc/openlan:/etc/openlan \
  -v /opt/openlan/var/openlan:/var/openlan \
  luscis/openlan:v25.4.1.amd64.deb \
  openlan-switch -conf /etc/openlan/switch/switch.yaml

# 查看日志
docker logs -f openlan-switch
```

#### 使用 Docker 运行 Access Point

```bash
# 准备配置文件
mkdir -p /opt/openlan/etc/openlan/access
cd /opt/openlan/etc/openlan/access

cat > example.yaml << EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: central-switch.example.com:10002
username: user@example
password: your-password
network: example
forward:
  - 192.168.1.0/24
EOF

# 运行 Access Point 容器
docker run -d \
  --name openlan-access \
  --privileged \
  --network host \
  -v /opt/openlan/etc/openlan/access:/etc/openlan/access \
  luscis/openlan:v25.4.1.amd64.deb \
  openlan-access -conf /etc/openlan/access/example.yaml

# 查看日志
docker logs -f openlan-access
```

#### 容器管理命令

```bash
# 查看运行中的容器
docker ps | grep openlan

# 停止容器
docker stop openlan-switch

# 重启容器
docker restart openlan-switch

# 查看容器内部日志
docker logs --tail 100 openlan-switch

# 进入容器
docker exec -it openlan-switch bash

# 删除容器
docker rm openlan-switch
```

### Docker Compose 多服务部署

#### 准备配置文件

```bash
# 创建目录结构
mkdir -p /opt/openlan/{etc/openlan/{switch,access,proxy},run/{pluto,frr},var/openlan/frr}

# 下载完整配置包
wget https://github.com/luscis/openlan/releases/download/v25.4.1/config.tar.gz
tar -xvf config.tar.gz -C /opt
```

#### Docker Compose 配置

```yaml
# /opt/openlan/docker-compose.yml
version: "2.3"
services:
  ipsec:
    restart: always
    image: "luscis/openlan:v25.4.1.amd64.deb"
    privileged: true
    network_mode: host
    entrypoint: ["/var/openlan/script/ipsec.sh"]
    volumes:
      - /opt/openlan/etc/ipsec.d:/etc/ipsec.d
      - /opt/openlan/run/pluto:/run/pluto

  frr:
    restart: always
    image: "luscis/openlan:v25.4.1.amd64.deb"
    network_mode: host
    privileged: true
    entrypoint: ["/var/openlan/script/frr.sh"]
    volumes:
      - /opt/openlan/etc/frr:/etc/frr
      - /opt/openlan/var/openlan/frr:/var/openlan/frr

  switch:
    restart: always
    image: "luscis/openlan:v25.4.1.amd64.deb"
    privileged: true
    network_mode: host
    entrypoint: ["/var/openlan/script/switch.sh", "start"]
    volumes:
      - /opt/openlan/etc/openlan:/etc/openlan
      - /opt/openlan/etc/ipsec.d:/etc/ipsec.d
      - /opt/openlan/run/pluto:/run/pluto
      - /opt/openlan/etc/frr:/etc/frr
      - /opt/openlan/var/openlan/frr:/var/openlan/frr
    depends_on:
      - ipsec
      - frr

  proxy:
    restart: always
    image: "luscis/openlan:v25.4.1.amd64.deb"
    network_mode: host
    entrypoint: ["/usr/bin/openlan-proxy", "-conf", "/etc/openlan/proxy.json"]
    volumes:
      - /opt/openlan/etc/openlan:/etc/openlan
    depends_on:
      - switch
```

#### 启动和管理

```bash
# 启动所有服务
cd /opt/openlan
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f switch

# 查看特定服务日志
docker-compose logs -f proxy

# 重启服务
docker-compose restart switch

# 停止服务
docker-compose stop

# 删除容器（保留数据）
docker-compose down

# 更新镜像版本
sed -i 's/:v25.4.1.amd64.deb/:v25.4.2.amd64.deb/g' docker-compose.yml
docker-compose up -d
```

#### 备份 OpenVPN 配置

```bash
# 备份 switch 容器中的 OpenVPN 配置
docker cp openlan_switch_1:/var/openlan/openvpn ./openvpn_backup/

# 升级后恢复
docker cp ./openvpn_backup/openvpn openlan_switch_1:/var/openlan/
```

### Kubernetes 部署

#### 前置条件

- Kubernetes 集群（v1.19+）
- kubectl 命令行工具
- 存储类（StorageClass）

#### 准备 ConfigMap

```bash
# 创建 namespace
kubectl create namespace openlan

# 创建配置 ConfigMap
kubectl create configmap openlan-switch-config \
  --from-file=/opt/openlan/etc/openlan/switch \
  -n openlan

kubectl create configmap openlan-proxy-config \
  --from-file=/opt/openlan/etc/openlan/proxy.json \
  -n openlan
```

#### 部署 Switch

```yaml
# switch-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openlan-switch
  namespace: openlan
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openlan-switch
  template:
    metadata:
      labels:
        app: openlan-switch
    spec:
      hostNetwork: true
      hostPID: true
      containers:
      - name: switch
        image: luscis/openlan:v25.4.1.amd64.deb
        imagePullPolicy: IfNotPresent
        securityContext:
          privileged: true
        command: ["/var/openlan/script/switch.sh", "start"]
        volumeMounts:
        - name: config
          mountPath: /etc/openlan
        - name: data
          mountPath: /var/openlan
        - name: runtime
          mountPath: /run
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
      volumes:
      - name: config
        configMap:
          name: openlan-switch-config
      - name: data
        emptyDir: {}
      - name: runtime
        emptyDir:
          medium: Memory
```

#### 部署 Proxy

```yaml
# proxy-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openlan-proxy
  namespace: openlan
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openlan-proxy
  template:
    metadata:
      labels:
        app: openlan-proxy
    spec:
      hostNetwork: true
      containers:
      - name: proxy
        image: luscis/openlan:v25.4.1.amd64.deb
        securityContext:
          privileged: true
        command: 
        - /usr/bin/openlan-proxy
        - -conf
        - /etc/openlan/proxy.json
        volumeMounts:
        - name: config
          mountPath: /etc/openlan
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: openlan-proxy-config
```

#### 应用部署

```bash
# 应用部署配置
kubectl apply -f switch-deployment.yaml
kubectl apply -f proxy-deployment.yaml

# 查看部署状态
kubectl get pods -n openlan

# 查看日志
kubectl logs -n openlan -l app=openlan-switch -f

# 进入 Pod
kubectl exec -it -n openlan deployment/openlan-switch -- bash
```

### Windows 部署

#### 系统要求

- Windows 10 (21H2+) 或 Windows Server 2019+
- TAP-Windows 虚拟网卡驱动
- 管理员权限

#### 安装步骤

```powershell
# 1. 下载安装程序
$url = "https://github.com/luscis/openlan/releases/download/v25.4.1/openceci-windows-v25.4.1.amd64.msi"
Invoke-WebRequest -Uri $url -OutFile "openceci-v25.4.1.amd64.msi"

# 2. 安装 TAP 驱动（如果尚未安装）
# 从 https://build.openvpn.net/downloads/releases/ 下载 TAP-Windows

# 3. 执行安装
msiexec /i openceci-v25.4.1.amd64.msi /quiet

# 4. 验证安装
Get-ChildItem "C:\Program Files\OpenLAN"
```

#### 配置接入点

```powershell
# 配置文件位置
$confDir = "C:\ProgramData\OpenLAN\access"

# 创建配置文件
@"
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: central.example.com:10002
username: user@example
password: your-password
network: example
"@ | Out-File -FilePath "$confDir\example.yaml" -Encoding UTF8

# 启动服务
Start-Service OpenLANAccess
Get-Service OpenLANAccess
```

#### 验证连接

```powershell
# 测试网络连接
ping 172.32.10.10

# 查看虚拟网卡
Get-NetAdapter | Where-Object {$_.Name -like "*OpenLAN*"}

# 查看路由表
route print | grep -A5 172.32
```

### macOS 部署

#### 系统要求

- macOS 10.15+
- Homebrew
- 安装 utun 虚拟网卡

#### 安装步骤

```bash
# 1. 使用 Homebrew 安装
brew install luscis/openlan/openlan

# 2. 或手动下载
curl -O https://github.com/luscis/openlan/releases/download/v25.4.1/openceci-darwin-v25.4.1.amd64.zip
unzip openceci-darwin-v25.4.1.amd64.zip
chmod +x openceci-darwin-v25.4.1.amd64/openlan-access

# 3. 复制到应用目录
sudo cp -r openceci-darwin-v25.4.1.amd64 /Applications/OpenLAN
```

#### 配置和运行

```bash
# 创建配置目录
mkdir -p ~/.openlan/access

# 创建配置文件
cat > ~/.openlan/access/example.yaml << EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: central.example.com:10002
username: user@example
password: your-password
network: example
interface:
  name: utun0
  address: 172.32.10.11/24
EOF

# 运行接入点
sudo /Applications/OpenLAN/openlan-access -conf ~/.openlan/access/example.yaml

# 后台运行（使用 LaunchAgent）
cat > ~/Library/LaunchAgents/com.openlan.access.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.openlan.access</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Applications/OpenLAN/openlan-access</string>
        <string>-conf</string>
        <string>~/.openlan/access/example.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOF

# 加载服务
launchctl load ~/Library/LaunchAgents/com.openlan.access.plist
```

---

## 配置参考

### Switch 配置详解

#### 基础配置

```yaml
# /etc/openlan/switch/switch.yaml

# 协议配置
protocol: tcp              # 支持: tcp, tls, udp, kcp, ws, wss
listen: 0.0.0.0:10002     # 监听地址和端口

# 加密配置
crypt:
  algorithm: aes-128      # 支持: aes-128, aes-192, aes-256
  secret: ea64d5b0c96c    # 预共享密钥（十六进制）

# 日志配置
log:
  file: /var/log/openlan/switch.log
  verbose: false          # 详细日志

# HTTP 管理接口
http:
  listen: 0.0.0.0:10000   # HTTPS 管理端口
  cacert: /etc/openlan/certs/ca.crt
  cert: /etc/openlan/certs/server.crt
  key: /etc/openlan/certs/server.key

# 证书配置
cert:
  # 自动生成或手动指定

# 队列配置（性能调优）
queue:
  sockWr: 128             # Socket 写队列深度
  sockRd: 128             # Socket 读队列深度
  tapWr: 64              # TAP 写队列深度
  tapRd: 2               # TAP 读队列深度
  virSnd: 256            # 虚拟设备发送队列
  virWrt: 128            # 虚拟设备写队列

# 资源限制
limit:
  access: 128            # 最大接入点数量
  user: 1024             # 最大用户数
  link: 128              # 最大链接数
```

#### 完整示例

```yaml
protocol: tcp
listen: 0.0.0.0:10002
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
http:
  listen: 0.0.0.0:10000
log:
  file: /var/log/openlan/switch.log
  verbose: false
```

### Network 配置详解

#### 虚拟网络配置

```yaml
# /etc/openlan/switch/network/example.yaml

# 网络基础信息
name: example                      # 网络唯一标识
provider: bridge                   # 网络类型: bridge(默认), router, ipsec, bgp, ceci

# 桥接配置
bridge:
  name: br-example               # 网桥名称（Linux 独有）
  address: 172.32.10.10/24       # 网桥 IP 和掩码
  mtu: 1500                      # MTU 大小

# 子网配置（DHCP）
subnet:
  startAt: 172.32.10.100         # DHCP 起始地址
  endAt: 172.32.10.150           # DHCP 结束地址
  lease: 3600                    # DHCP 租约时间（秒）

# OpenVPN 配置
openvpn:
  protocol: tcp                  # OpenVPN 协议
  listen: 0.0.0.0:1194           # OpenVPN 监听端口
  dev: tap0                       # TAP 设备名

# 接入点链接配置
links:
  - connection: branch1.net:10002
    username: branch1@example
    password: branch1-password
    protocol: tcp

# 主机静态分配
hosts:
  - address: 172.32.10.20
    lease: 0                      # 永久分配
  - address: 172.32.10.21

# 静态路由
routes:
  - prefix: 192.168.10.0/24       # 目标网络
    metric: 100                   # 路由优先级
    via: 172.32.10.11            # 下一跳地址
    interface: example

# 访问控制列表
acl: acl-example                  # 引用 ACL 配置

# QoS 配置
qos: qos-example                  # 引用 QoS 配置

# SNAT 配置
snat: enable                      # 启用 SNAT
snat_addr: 172.32.10.10          # SNAT 源地址

# 零信任配置
ztrust: ztrust-example           # 引用零信任配置

# 输出配置（监控）
outputs:
  - type: prometheus              # Prometheus 监控
    listen: 0.0.0.0:8080

# 网络空间隔离
namespace: openlan-example

# DNAT 配置
dnat:
  - protocol: tcp
    destination: 0.0.0.0          # 目标 IP（0.0.0.0 表示任意）
    dport: 8080                   # 目标端口
    todestination: 192.168.10.10  # 实际目标 IP
    todport: 80                    # 实际目标端口
```

#### 完整示例

```yaml
name: example
provider: bridge
bridge:
  address: 172.32.10.10/24
subnet:
  startAt: 172.32.10.100
  endAt: 172.32.10.150
  lease: 3600
openvpn:
  protocol: tcp
  listen: 0.0.0.0:1194
routes:
  - prefix: 192.168.10.0/24
    metric: 100
```

### Access 配置详解

#### 基础配置

```yaml
# /etc/openlan/access/example.yaml

# 连接配置
connection: central.example.com:10002  # Central Switch 地址和端口
username: user@example                 # 用户名（user@network 格式）
password: your-password                # 密码

# 网络配置
network: example                       # 虚拟网络名称
protocol: tcp                          # 传输协议: tcp, tls, udp, kcp, ws, wss
timeout: 30                            # 连接超时（秒）

# 加密配置
crypt:
  algorithm: aes-128                   # 加密算法
  secret: ea64d5b0c96c                # 预共享密钥

# 虚拟网卡配置
interface:
  name: tap0                          # 网卡名称
  address: 172.32.10.11/24           # 分配的 IP 地址
  mtu: 1500                          # MTU 大小
  provider: kernel                   # 网卡驱动: kernel, water(Windows/macOS)

# 转发路由
forward:
  - 192.168.10.0/24                  # 转发的本地子网
  - 10.0.0.0/8

# 日志配置
log:
  file: /var/log/openlan/access.log
  verbose: false

# HTTP 代理配置
http:
  listen: 127.0.0.1:8080            # HTTP 代理地址

# 证书配置
cert:
  cacert: /etc/openlan/certs/ca.crt
  cert: /etc/openlan/certs/client.crt
  key: /etc/openlan/certs/client.key

# 故障转移
fallback: fallback-server.net:10002   # 备用服务器

# 启动/停止脚本
run1: /etc/openlan/scripts/up.sh      # 网卡启动时执行
run0: /etc/openlan/scripts/down.sh    # 网卡关闭时执行

# 状态文件
status: /var/run/openlan/access.status
pid: /var/run/openlan/access.pid
```

#### 完整示例

```yaml
connection: central.example.com:10002
username: branch1@example
password: branch1-password
network: example
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
interface:
  address: 172.32.10.11/24
forward:
  - 192.168.10.0/24
```

### Proxy 配置详解

#### HTTP 代理

```yaml
# /etc/openlan/proxy.json

{
  "http": [
    {
      "listen": "192.168.1.88:11082",
      "timeout": 30,
      "auth": {
        "method": "basic",
        "username": "user",
        "password": "pass"
      },
      "rules": [
        {
          "pattern": "^http://.*\\.example\\.com.*",
          "target": "http://internal-proxy.example.com:3128"
        }
      ]
    }
  ],
  "socks5": [
    {
      "listen": "0.0.0.0:1080",
      "timeout": 30,
      "auth": {
        "method": "basic",
        "username": "user",
        "password": "pass"
      }
    }
  ],
  "tcp": [
    {
      "listen": "192.168.1.66:11082",
      "target": "192.168.1.88:11082",
      "timeout": 30
    }
  ]
}
```

#### 代理规则说明

| 代理类型 | 说明 | 用途 |
|---------|------|------|
| HTTP | HTTP/HTTPS 正向代理 | 浏览器、企业应用代理 |
| SOCKS5 | SOCKS5 通用代理 | 各类应用代理 |
| TCP | TCP 反向代理 | 服务发布、内网穿透 |

### 高级配置

#### LDAP 认证

```yaml
# /etc/openlan/switch/ldap.yaml

ldap:
  server: ldap.example.com:389
  bindDn: cn=admin,dc=example,dc=com
  bindPassword: admin-password
  baseDn: dc=example,dc=com
  filter: (uid=%s)
  attr:
    name: uid
    mail: mail
```

#### 防火墙/ACL 配置

```yaml
# /etc/openlan/switch/acl/example.yaml

name: acl-example
rules:
  - name: allow-office
    protocol: all
    srcPrefix: 192.168.1.0/24
    dstPrefix: 172.32.10.0/24
    action: accept
  
  - name: deny-external
    protocol: all
    srcPrefix: 0.0.0.0/0
    dstPrefix: 172.32.10.0/24
    action: drop
```

#### QoS 配置

```yaml
# /etc/openlan/switch/qos/example.yaml

name: qos-example
queues:
  - name: users
    # 限速规则
    limitIn: 10485760   # 下行 10 Mbps
    limitOut: 10485760  # 上行 10 Mbps
    priority: 100
    
  - name: servers
    limitIn: 104857600  # 100 Mbps
    limitOut: 104857600
    priority: 50
```

#### IPSec 隧道

```yaml
# /etc/openlan/switch/network/ipsec.yaml

name: ipsec
provider: ipsec
specifies:
  # IPSec 隧道参数
  leftAddress: 203.0.113.1       # 本地公网 IP
  rightAddress: 198.51.100.1     # 对端公网 IP
  leftSubnet: 192.168.1.0/24     # 本地私网
  rightSubnet: 192.168.2.0/24    # 对端私网
  psk: your-preshared-key        # 预共享密钥
  authAlgorithm: sha256
  encryptAlgorithm: aes256
```

#### BGP 路由

```yaml
# /etc/openlan/switch/network/bgp.yaml

name: bgp
provider: bgp
specifies:
  routerId: 203.0.113.1
  localAsn: 65000
  neighbors:
    - address: 10.0.0.1
      asn: 65001
      networks:
        - 192.168.1.0/24
        - 192.168.2.0/24
```

---

## 常见场景配置

### 分支互联场景

#### 拓扑结构

```
                    Central Switch (203.0.113.1)
                    Network: 172.32.10.0/24
                            |
                  ----------|----------
                  |          |        |
              Branch1    Branch2   Branch3
            Access Pt   Access Pt  Access Pt
         192.168.1.0/24 192.168.2.0/24 192.168.3.0/24
```

#### Central Switch 配置

```yaml
# /etc/openlan/switch/switch.yaml
protocol: tcp
listen: 0.0.0.0:10002
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c

# /etc/openlan/switch/network/main.yaml
name: main
bridge:
  address: 172.32.10.10/24
subnet:
  startAt: 172.32.10.100
  endAt: 172.32.10.150
routes:
  - prefix: 192.168.1.0/24
    via: 172.32.10.11
  - prefix: 192.168.2.0/24
    via: 172.32.10.12
  - prefix: 192.168.3.0/24
    via: 172.32.10.13
snat: enable
```

#### Branch1 Access Point 配置

```yaml
# /etc/openlan/access/branch1.yaml
connection: central.example.com:10002
username: branch1@main
password: branch1-password
network: main
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
interface:
  address: 172.32.10.11/24
forward:
  - 192.168.1.0/24
```

#### 测试连通性

```bash
# Branch1 访问 Branch2
ssh user@192.168.2.1

# Branch2 访问 Branch3
ssh user@192.168.3.1

# 查看路由表
route -n | grep 172.32
```

### 多区域互联场景

#### 拓扑结构

```
    Branch Network           Central Network          Cloud Network
   (Shanghai)               (Beijing)                 (Singapore)
    192.168.1.0/24          10.0.0.0/24             10.1.0.0/24
         |                       |                       |
    OpenLAN Pt              OpenLAN Sw              OpenLAN Pt
   (172.32.10.11)          (172.32.10.10)        (172.32.10.12)
         |                       |                       |
         ----------- Virtual Network (172.32.10.0/24) -----------
                       IPSec Tunnel for Encryption
```

#### Central Switch 配置

```yaml
# /etc/openlan/switch/network/main.yaml
name: main
bridge:
  address: 172.32.10.10/24
subnet:
  startAt: 172.32.10.100
  endAt: 172.32.10.150
routes:
  - prefix: 192.168.1.0/24
    via: 172.32.10.11
  - prefix: 10.1.0.0/24
    via: 172.32.10.12
```

#### Shanghai Branch 配置

```yaml
connection: central.example.com:10002
username: shanghai@main
password: shanghai-password
network: main
interface:
  address: 172.32.10.11/24
forward:
  - 192.168.1.0/24
```

#### Singapore Cloud 配置

```yaml
connection: central.example.com:10002
username: singapore@main
password: singapore-password
network: main
interface:
  address: 172.32.10.12/24
forward:
  - 10.1.0.0/24
```

### 零信任网络场景

#### 启用零信任

```bash
# Central Switch
openlan ztrust --network main enable
systemctl restart openlan-switch
```

#### 用户注册和授权

```bash
# 管理员添加用户
openlan user add --name alice@main

# 用户登录后初始化
export TOKEN="alice@main:password"
export URL="https://central.example.com:10000"

# 用户注册为零信任端点
openlan ztrust guest add

# 查看注册的端点
openlan ztrust guest ls
```

#### 主机服务发现和访问

```bash
# 用户发现目标主机
openlan ztrust knock add --protocol icmp --socket 192.168.1.10

# 添加 SSH 访问请求
openlan ztrust knock add --protocol tcp --socket 192.168.1.10:22

# 查看已授予的访问权限
openlan ztrust knock ls

# 访问服务
ssh user@192.168.1.10
```

#### 零信任配置示例

```yaml
# /etc/openlan/switch/ztrust/example.yaml
name: example
enabled: true
requireAuth: true
timeoutDuration: 3600
authMethods:
  - type: password
    enabled: true
```

### HTTP/SOCKS5 代理场景

#### Proxy 服务配置

```json
{
  "http": [
    {
      "listen": "0.0.0.0:11080",
      "timeout": 60,
      "auth": {
        "method": "basic",
        "username": "proxyuser",
        "password": "proxypass"
      },
      "rules": [
        {
          "pattern": "^https?://.*\\.internal\\..*",
          "action": "forward",
          "target": "http://internal-service.internal:8080"
        },
        {
          "pattern": "^https?://.*",
          "action": "direct"
        }
      ]
    }
  ],
  "socks5": [
    {
      "listen": "0.0.0.0:11081",
      "timeout": 60,
      "auth": {
        "method": "basic",
        "username": "socksuser",
        "password": "sockspass"
      }
    }
  ]
}
```

#### 客户端使用

```bash
# HTTP 代理
export http_proxy=http://proxyuser:proxypass@proxy.example.com:11080
export https_proxy=http://proxyuser:proxypass@proxy.example.com:11080
curl https://api.example.com

# SOCKS5 代理
ssh -o ProxyCommand='nc -X 5 -x socksuser:sockspass@proxy.example.com:11081 %h %p' user@internal-host

# 浏览器配置
# 设置 SOCKS5 代理: proxy.example.com:11081
# 用户名: socksuser
# 密码: sockspass
```

---

## 故障排查指南

### 常见问题诊断

#### 连接问题

```bash
# 1. 验证 Central Switch 是否运行
systemctl status openlan-switch

# 2. 检查监听端口
netstat -tlnp | grep openlan
lsof -i :10002

# 3. 测试网络连通性
ping central.example.com
telnet central.example.com 10002

# 4. 检查防火墙规则
iptables -L -n | grep 10002
firewall-cmd --list-ports

# 5. 查看 Switch 日志
journalctl -u openlan-switch -n 100 -f
```

#### Access Point 连接失败

```bash
# 1. 检查配置文件
cat /etc/openlan/access/example.yaml

# 2. 验证用户名和密码
openlan user ls  # 在 Central Switch 上

# 3. 检查接入点日志
journalctl -u openlan-access@example -f

# 4. 测试指定的连接
openlan-access -conf /etc/openlan/access/example.yaml

# 5. 检查虚拟网卡
ip link show tap0
ifconfig tap0
```

#### 网络不可达

```bash
# 1. 检查虚拟网络配置
cat /etc/openlan/switch/network/example.yaml

# 2. 验证网桥状态
brctl show br-example
ip link show br-example

# 3. 检查路由表
ip route show
route -n

# 4. 测试网络连通性
ping 172.32.10.10
ping 192.168.10.1 (from access point)

# 5. 使用 traceroute 诊断
traceroute -n 192.168.10.1
```

### 日志分析

#### 查看详细日志

```bash
# Switch 日志
journalctl -u openlan-switch --since "10 minutes ago" -f

# Access 日志
journalctl -u openlan-access@example -f

# 查看文件日志（如果配置了）
tail -f /var/log/openlan/switch.log
tail -f /var/log/openlan/access.log

# 启用详细日志
# 修改配置文件中的 verbose: true
# 重启服务
systemctl restart openlan-switch
```

#### 日志级别说明

| 级别 | 说明 | 用途 |
|------|------|------|
| ERROR | 错误信息 | 问题诊断 |
| WARN | 警告信息 | 潜在问题 |
| INFO | 信息日志 | 正常运行 |
| DEBUG | 调试信息 | 详细诊断 |
| TRACE | 跟踪信息 | 性能分析 |

### 性能检查

```bash
# 1. 检查 CPU 和内存使用
top -p $(pgrep openlan-switch)
ps aux | grep openlan

# 2. 检查网络流量
iftop -i tap0
nethogs -i tap0

# 3. 检查磁盘 I/O
iostat -x 1 5

# 4. 检查连接数
ss -tan | grep -c ESTABLISHED
ss -tan | grep -c TIME_WAIT
```

---

## 运维管理

### 用户管理

#### 添加用户

```bash
# 添加用户
openlan user add --name user1@example

# 输出包含：用户名、临时密码、角色
# user1@example  l6llot97yx  guest

# 用户首次登录时需要修改密码
```

#### 用户列表

```bash
# 查看所有用户
openlan user ls

# 查看特定网络的用户
openlan user ls --network example
```

#### 删除用户

```bash
# 删除用户
openlan user del --name user1@example

# 从所有网络中删除
openlan user del --name user1@example --all
```

#### 修改用户角色

```bash
# 提升为管理员
openlan user edit --name user1@example --role admin

# 降级为普通用户
openlan user edit --name user1@example --role guest
```

### 升级和回滚

#### 升级步骤

```bash
# 1. 备份当前配置
tar -czf /opt/backup/openlan-$(date +%Y%m%d).tar.gz \
  /etc/openlan /var/openlan

# 2. 下载新版本
wget https://github.com/luscis/openlan/releases/download/v25.4.2/openlan-v25.4.2.x86_64.bin

# 3. 停止服务
systemctl stop openlan-switch
systemctl stop openlan-access@example

# 4. 安装新版本
chmod +x openlan-v25.4.2.x86_64.bin
sudo ./openlan-v25.4.2.x86_64.bin

# 5. 检查配置兼容性
openlan-switch --config-check /etc/openlan/switch/switch.yaml

# 6. 启动服务
systemctl start openlan-switch
systemctl start openlan-access@example

# 7. 验证升级
openlan-switch --version
systemctl status openlan-switch
```

#### 回滚到旧版本

```bash
# 1. 停止服务
systemctl stop openlan-switch
systemctl stop openlan-access@example

# 2. 恢复备份
tar -xzf /opt/backup/openlan-20250227.tar.gz -C /

# 3. 获取旧版本二进制
wget https://github.com/luscis/openlan/releases/download/v25.4.1/openlan-v25.4.1.x86_64.bin
chmod +x openlan-v25.4.1.x86_64.bin
sudo ./openlan-v25.4.1.x86_64.bin

# 4. 启动服务
systemctl start openlan-switch
```

### 备份和恢复

#### 备份配置

```bash
# 备份所有配置
tar -czf openlan-config-backup-$(date +%Y%m%d).tar.gz \
  /etc/openlan

# 备份数据
tar -czf openlan-data-backup-$(date +%Y%m%d).tar.gz \
  /var/openlan

# 合并备份
tar -czf openlan-full-backup-$(date +%Y%m%d).tar.gz \
  /etc/openlan /var/openlan
```

#### 恢复配置

```bash
# 恢复配置
tar -xzf openlan-config-backup-20250227.tar.gz -C /

# 恢复数据（注意：会覆盖现有数据）
tar -xzf openlan-data-backup-20250227.tar.gz -C /

# 重启服务
systemctl restart openlan-switch
```

#### 定期备份脚本

```bash
#!/bin/bash
# /usr/local/bin/openlan-backup.sh

BACKUP_DIR="/opt/backups/openlan"
RETENTION_DAYS=30

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份
BACKUP_FILE="$BACKUP_DIR/openlan-$(date +%Y%m%d-%H%M%S).tar.gz"
tar -czf "$BACKUP_FILE" /etc/openlan /var/openlan

# 清理旧备份
find $BACKUP_DIR -name "openlan-*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $BACKUP_FILE"
```

在 crontab 中每天执行：

```bash
0 2 * * * /usr/local/bin/openlan-backup.sh
```

---

## 性能调优

### 队列优化

```yaml
# /etc/openlan/switch/switch.yaml
queue:
  sockWr: 256       # 提高 Socket 写入性能
  sockRd: 256       # 提高 Socket 读取性能
  tapWr: 128        # TAP 写入性能
  tapRd: 4          # 减少 TAP 读取延迟
  virSnd: 512       # 虚拟设备发送能力
  virWrt: 256       # 虚拟设备写入能力
```

### 网络优化

```bash
# 增加系统网络缓冲区
sysctl -w net.core.rmem_max=134217728
sysctl -w net.core.wmem_max=134217728
sysctl -w net.ipv4.tcp_rmem="4096 87380 67108864"
sysctl -w net.ipv4.tcp_wmem="4096 65536 67108864"

# 增加网络连接处理能力
sysctl -w net.core.netdev_max_backlog=5000
sysctl -w net.ipv4.tcp_max_syn_backlog=5000

# 永久配置
cat >> /etc/sysctl.conf << EOF
net.core.rmem_max=134217728
net.core.wmem_max=134217728
net.ipv4.tcp_rmem=4096 87380 67108864
net.ipv4.tcp_wmem=4096 65536 67108864
net.core.netdev_max_backlog=5000
net.ipv4.tcp_max_syn_backlog=5000
EOF

sysctl -p
```

### 连接优化

```yaml
# /etc/openlan/switch/switch.yaml
limit:
  access: 512       # 增加接入点限制
  user: 4096        # 增加用户限制
  link: 512         # 增加链接限制
```

### 资源监控

```bash
# 监控 Switch 资源使用
watch -n 1 'ps aux | grep openlan-switch'

# 实时网络统计
iftop -i tap0 -n

# 网络吞吐量监控
sar -n DEV 1

# 使用 Prometheus 导出指标
# 配置 Prometheus 采集
```

---

## 安全最佳实践

### 密钥管理

#### 预共享密钥生成

```bash
# 生成安全的十六进制密钥
openssl rand -hex 6  # 12 字符（24 位）
openssl rand -hex 16 # 32 字符（64 位）

# 生成更强的密钥
python3 -c "import secrets; print(secrets.token_hex(16))"
```

#### 密钥存储

```bash
# 限制配置文件权限
chmod 600 /etc/openlan/switch/switch.yaml
chmod 600 /etc/openlan/access/*.yaml

# 使用密钥管理系统（如 Vault）
# 动态生成和轮换密钥
```

### 证书配置

#### 生成自签名证书

```bash
# 生成 CA 证书
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
  -subj "/CN=OpenLAN-CA"

# 生成服务器证书
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
  -subj "/CN=central.example.com"

# 签名证书
openssl x509 -req -days 365 -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt
```

#### 证书配置

```yaml
# /etc/openlan/switch/switch.yaml
http:
  listen: 0.0.0.0:10000
  cacert: /etc/openlan/certs/ca.crt
  cert: /etc/openlan/certs/server.crt
  key: /etc/openlan/certs/server.key
```

### 访问控制

#### 使用 ACL

```yaml
# /etc/openlan/switch/acl/example.yaml
name: example
rules:
  # 允许内部网络
  - name: allow-internal
    protocol: all
    srcPrefix: 192.168.0.0/16
    dstPrefix: 172.32.10.0/24
    action: accept
  
  # 拒绝所有其他流量
  - name: deny-all
    protocol: all
    srcPrefix: 0.0.0.0/0
    dstPrefix: 0.0.0.0/0
    action: drop
```

#### LDAP 集成

```yaml
# /etc/openlan/switch/ldap.yaml
ldap:
  enabled: true
  server: ldap.example.com:389
  bindDn: cn=admin,dc=example,dc=com
  bindPassword: admin-password
  baseDn: ou=users,dc=example,dc=com
  searchFilter: (uid=%s)
  groupFilter: (memberUid=%s)
  requireGroup: openlan-users
  groupBaseDn: ou=groups,dc=example,dc=com
```

### 传输安全

#### 使用 TLS 加密

```bash
# Switch 配置
protocol: tls
listen: 0.0.0.0:10002
```

#### 使用 IPSec 加密

```yaml
# /etc/openlan/switch/network/ipsec.yaml
name: ipsec
provider: ipsec
specifies:
  leftAddress: 203.0.113.1
  rightAddress: 198.51.100.1
  authAlgorithm: sha256
  encryptAlgorithm: aes256
```

### 监控和审计

#### 启用详细日志

```yaml
# /etc/openlan/switch/switch.yaml
log:
  file: /var/log/openlan/switch.log
  verbose: true

# /etc/openlan/access/example.yaml
log:
  file: /var/log/openlan/access.log
  verbose: true
```

#### 日志收集

```bash
# 使用 rsyslog 转发日志
echo "local3.* @@logserver.example.com:514" >> /etc/rsyslog.d/openlan.conf
systemctl restart rsyslog

# 使用 logrotate 管理日志
cat > /etc/logrotate.d/openlan << EOF
/var/log/openlan/*.log {
  daily
  rotate 7
  compress
  delaycompress
  missingok
  notifempty
  create 0600 root root
}
EOF
```

---

## 常见问题 FAQ

### 部署相关

**Q: OpenLAN 最低系统要求是什么？**

A: 最低配置为 512MB 内存、1 核 CPU 和 100MB 磁盘空间。对于生产环境建议使用 2GB+ 内存和 1GB+ 磁盘空间。

**Q: 是否支持在 NAT 后面部署？**

A: 是的。Access Point 可以在 NAT 后面，由于是由 Access Point 主动发起连接到 Central Switch，所以不需要映射入站端口。

**Q: 可以同时运行多个 Central Switch 吗？**

A: 可以。多个 Central Switch 可以通过 IPSec 隧道或 BGP 相互连接，形成分布式网络。

### 配置相关

**Q: 如何更改虚拟网络的 IP 段？**

A: 编辑对应网络的 YAML 配置文件中的 `bridge.address` 和 `subnet` 部分，然后重启 Switch 服务。

**Q: 支持哪些传输协议？**

A: 支持 TCP（推荐用于性能）、TLS（推荐用于安全）、UDP、KCP、WebSocket 和 WSS。

**Q: 如何启用零信任模式？**

A: 使用命令 `openlan ztrust --network <name> enable`，然后重启 Switch 服务。

### 网络相关

**Q: Access Point 之间能否直接通信？**

A: 可以。它们都连接到同一个虚拟网络后，可以像局域网主机一样直接通信。

**Q: 如何使用 OpenVPN 访问虚拟网络？**

A: 在 Switch 管理界面下载 OpenVPN 配置文件，在 OpenVPN 客户端导入后连接即可。

**Q: 可以限制带宽吗？**

A: 可以。通过 QoS 配置限制单个用户或链接的带宽。

### 故障排查

**Q: 连接后无法访问其他网络？**

A: 检查以下几点：
1. 虚拟网卡是否已配置 IP 地址
2. 是否添加了正确的转发路由
3. 防火墙是否阻止了 OpenLAN 流量
4. 是否启用了 SNAT

**Q: 经常断连怎么办？**

A: 尝试以下方法：
1. 增加连接超时时间
2. 配置故障转移（fallback）服务器
3. 检查网络稳定性
4. 更新到最新版本

**Q: 如何提升性能？**

A: 参考性能调优部分，包括：
1. 调整队列深度
2. 优化系统网络参数
3. 使用更快的协议（TCP 优于 TLS）
4. 增加硬件资源

### 安全相关

**Q: 如何修改预共享密钥？**

A: 编辑配置文件中的 `crypt.secret`，重启服务即可生效。建议在非生产环境先测试。

**Q: 支持多层加密吗？**

A: 支持。可以在 OpenLAN 加密的基础上，使用 TLS 协议获得额外加密。

**Q: 如何实现账户的单点登录（SSO）？**

A: 通过集成 LDAP 或其他目录服务实现，配置 `ldap.yaml` 文件即可。

### 其他问题

**Q: OpenLAN 和 WireGuard 有什么区别？**

A: 详见项目文档中的 openlan-vs-wireguard.md。简而言之，OpenLAN 更适合企业网络互联和多网络管理。

**Q: 是否有 GUI 管理界面？**

A: 目前提供 Web 管理接口（HTTPS），可通过浏览器访问 `https://<switch-ip>:10000`。

**Q: 如何获得技术支持？**

A: 可以通过以下方式：
1. GitHub Issues: https://github.com/luscis/openlan/issues
2. 官方文档: https://github.com/luscis/openlan/tree/master/docs
3. 社区讨论区

---

## 更新日志和版本说明

### 版本 v25.4.1（当前）

- 稳定版本
- 支持所有主要功能
- Docker 镜像标签：`luscis/openlan:v25.4.1.amd64.deb`

### 获取更新

```bash
# 检查最新版本
curl -s https://api.github.com/repos/luscis/openlan/releases/latest | grep tag_name

# 下载最新版本
wget https://github.com/luscis/openlan/releases/latest/download/openlan-*.tar.gz
```

---

**文档最后更新时间**：2025年2月27日  
**OpenLAN 版本**：v25.4.1+  
**维护者**：OpenLAN Community
