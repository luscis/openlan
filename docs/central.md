# 🌿 Central Branch Example

This example follows the `tests/cases/access_success.sh` scenario.

It demonstrates a central switch with two access clients. Both clients authenticate into the same OpenLAN network and can reach the switch and each other.

## 🗺️ Topology

```text
           sw1(center) 100.100.0.241 / 192.11.0.1
                ^                    ^
                | tcp access          | udp access
        ac1 192.11.0.11       ac2 192.11.0.12
                both access clients join example network
```

- Docker management network: `100.100.0.0/24`
- Central switch: `sw1=100.100.0.241`
- OpenLAN network: `example=192.11.0.0/24`
- Gateway: `sw1=192.11.0.1`
- Access clients: `ac1=192.11.0.11`, `ac2=192.11.0.12`
- Crypt: `aes-128:ea64d5b0c96c`

## ⚙️ Configure the Central Switch

Create the switch configuration:

```bash
mkdir -p /opt/openlan/tests-sw1/etc/openlan/switch

cat > /opt/openlan/tests-sw1/etc/openlan/switch/switch.json <<'EOF'
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
```

Start the switch, then add the network and users:

```bash
openlan network --name example add --address 192.11.0.1/24

openlan user add --name t1@example --password 123456
openlan user add --name t2@example --password 123457
```

## 📡 Configure Access Client 1

`ac1` uses TCP access:

```yaml
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 100.100.0.241
username: t1@example
password: 123456
interface:
  address: 192.11.0.11/24
```

## 📡 Configure Access Client 2

`ac2` uses UDP access:

```yaml
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 100.100.0.241
username: t2@example
password: 123457
interface:
  address: 192.11.0.12/24
```

## ✅ Validate Access

The case validates successful authentication and reachability:

```bash
ping -c 3 192.11.0.1
ping -c 3 192.11.0.12
```

It also verifies crypt update behavior:

```bash
openlan crypt update --algorithm aes-128 --secret ea64d5b0c96d
openlan crypt ls
```

Clients using the old secret fail to reconnect, while clients updated to `ea64d5b0c96d` authenticate successfully.
