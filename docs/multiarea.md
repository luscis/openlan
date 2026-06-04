# 🌐 Multi-Hop Route Example

This example follows the `tests/cases/switch_route3.sh` scenario.

It demonstrates three switches connected in a chain. `sw3` reaches loopback VIPs on `sw1` and `sw2` through OpenLAN outputs and static routes.

## 🗺️ Topology

```text
sw1 VIP 10.251.0.11
   ^
   | output
sw2 VIP 10.251.0.12
   ^
   | output + static routes
sw3 reaches sw1/sw2 loopback VIPs through nexthops
```

- Docker management network: `100.100.0.0/24`
- `sw1=100.100.0.241`, `sw2=100.100.0.242`, `sw3=100.100.0.243`
- OpenLAN network: `example=192.51.0.0/24`
- Service addresses: `sw1=192.51.0.1`, `sw2=192.51.0.2`, `sw3=192.51.0.3`
- Loopback VIPs: `sw1=10.251.0.11/32`, `sw2=10.251.0.12/32`
- Crypt: `aes-128:ea64d5b0c96c`

## ⚙️ Configure `sw1`

```bash
openlan network --name example add --address 192.51.0.1/24
openlan router address add --device lo --address 10.251.0.11/32
openlan user add --name edge1@example --password 123456
```

## ⚙️ Configure `sw2`

```bash
openlan network --name example add --address 192.51.0.2/24
openlan router address add --device lo --address 10.251.0.12/32
openlan user add --name edge2@example --password 123457

openlan network --name example output add \
  --remote 100.100.0.241 \
  --protocol tcp \
  --secret edge1@example:123456 \
  --crypt aes-128:ea64d5b0c96c
```

## ⚙️ Configure `sw3`

```bash
openlan network --name example add --address 192.51.0.3/24

openlan network --name example output add \
  --remote 100.100.0.242 \
  --protocol tcp \
  --secret edge2@example:123457 \
  --crypt aes-128:ea64d5b0c96c

openlan network --name example route add \
  --prefix 10.251.0.11/32 \
  --nexthop 192.51.0.1

openlan network --name example route add \
  --prefix 10.251.0.12/32 \
  --nexthop 192.51.0.2
```

## ✅ Validate Routes

`sw3` should have routes for both VIPs:

```bash
ip route show
# 10.251.0.11 ...
# 10.251.0.12 ...
```

Reachability checks from `sw3`:

```bash
ping -c 3 192.51.0.1
ping -c 3 192.51.0.2
ping -c 3 10.251.0.11
ping -c 3 10.251.0.12
```

The case also runs `openlan reload --save` on all three switches and verifies that output authentication and route reachability survive reload.
