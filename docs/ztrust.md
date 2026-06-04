# 🔒 Zero Trust Example

This example follows the `tests/cases/switch_ztrust.sh` scenario.

## 🧭 Topology

```text
vpn1 10.93.0.10
      |
      v OpenVPN tcp/1194
sw1 192.59.0.1:8081
      ZTrust guest + knock gates service access
```

- Docker management network: `100.100.0.0/24`
- Switch: `sw1=100.100.0.241`
- OpenLAN network: `example=192.59.0.0/24`, `sw1=192.59.0.1`
- OpenVPN subnet: `10.93.0.0/24`, `vpn1@example=10.93.0.10`
- Protected service: `http://192.59.0.1:8081`

## ✅ Enable ZTrust

Before ZTrust is enabled, the OpenVPN client can access the service:

```bash
wget -qO- -T 3 -t 1 http://192.59.0.1:8081
# ztrust-8081
```

Enable ZTrust on the network:

```bash
openlan ztrust --network example enable
```

After enabling, OpenLAN installs a Zero Trust mangle chain for the network. Existing related/established traffic is accepted, while new unmatched traffic is denied:

```bash
iptables -t mangle -S TT_pre-example
# ... Goto Zero Trust ...

iptables -t mangle -S ZT_example
# ... ZTrust Deny All ...
```

At this point, `vpn1` can no longer access `192.59.0.1:8081` until it is added as a guest and creates a knock rule.

## 👤 Add Guests

An admin can add a guest explicitly:

```bash
openlan ztrust --network example guest add \
  --user vpn1 \
  --address 10.93.0.10
```

An authenticated user token can also derive the user and network automatically:

```bash
openlan --token vpn2@example:123457 ztrust guest add \
  --address 10.93.0.11
```

List guests:

```bash
openlan ztrust --network example guest ls
# total 2
# username                 address
# vpn1@example             10.93.0.10
# vpn2@example             10.93.0.11
```

If a token user does not have a known address, guest creation fails with `can't find address`.

## 🚪 Knock a Service

A user must first be registered as a guest before adding knock rules. A non-guest knock is rejected:

```bash
openlan --token vpn3@example:123458 ztrust knock add \
  --protocol tcp \
  --socket 192.59.0.1:8081 \
  --age 120
# fails: guest not found
```

The registered guest can open temporary access:

```bash
openlan --token vpn1@example:123456 ztrust knock add \
  --protocol tcp \
  --socket 192.59.0.1:8081 \
  --age 120
```

List knocks as an admin for a specific user:

```bash
openlan ztrust --network example knock ls --user vpn1
# total 1
# username                 protocol socket                   age  createAt
# vpn1@example             tcp      192.59.0.1:8081          ...  ...
```

List knocks as the authenticated user:

```bash
openlan --token vpn1@example:123456 ztrust knock ls
```

Other users cannot see or use `vpn1`'s knock:

```bash
openlan --token vpn2@example:123457 ztrust knock ls
# total 0
```

## 🔗 Access the Protected Service

After the guest and knock rule exist, `vpn1` can access the service again:

```bash
wget -qO- -T 3 -t 1 http://192.59.0.1:8081
# ztrust-8081
```

Knock rules are temporary. The `--age` value is the allowed lifetime in seconds, and expired rules are cleaned automatically.

## ♻️ Reload and Disable

The scenario verifies that enabled ZTrust state survives reload:

```bash
openlan reload --save
iptables -t mangle -S TT_pre-example
iptables -t mangle -S ZT_example
```

Remove the guest and disable ZTrust:

```bash
openlan ztrust --network example guest rm --user vpn1
openlan ztrust --network example disable
```

After disabling, access is restored without a guest or knock:

```bash
wget -qO- -T 3 -t 1 http://192.59.0.1:8081
# ztrust-8081
```
