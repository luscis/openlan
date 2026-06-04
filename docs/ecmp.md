# ⚖️ ECMP / FindHop Example

This example follows the `tests/cases/switch_findhop.sh` scenario.

OpenLAN uses `findhop` to bind a route to multiple candidate nexthops. The same mechanism supports active-backup failover and load-balance style ECMP routes.

## 🗺️ Topology

```text
                   sw0 VIP 10.243.0.10
                 ^                       ^
                 | network a             | network b
             sw1.0 ------------------- sw1.1
                 ^                       ^
                 +--------- sw2 ---------+
               findhop chooses nexthop path
```

- Docker management network: `100.100.0.0/24`
- `sw0=100.100.0.240`, `sw1.0=100.100.0.241`, `sw1.1=100.100.0.242`, `sw2=100.100.0.243`
- Network `a`: `sw0=192.53.0.1`, `sw1.0=192.53.0.2`, `sw1.1=192.53.0.4`, `sw2=192.53.0.3`
- Network `b`: `sw0=192.54.0.1`, `sw1.1=192.54.0.2`, `sw2=192.54.0.3`
- VIP on `sw0`: `10.243.0.10/32`

## ✅ Baseline Routes

The scenario first verifies that each path can reach the VIP independently.

Via network `a`:

```bash
openlan network --name a route add \
  --prefix 10.243.0.10/32 \
  --nexthop 192.53.0.1

ping -c 3 10.243.0.10

openlan network --name a route rm \
  --prefix 10.243.0.10/32
```

Via network `b`:

```bash
openlan network --name b route add \
  --prefix 10.243.0.10/32 \
  --nexthop 192.54.0.1

ping -c 3 10.243.0.10

openlan network --name b route rm \
  --prefix 10.243.0.10/32
```

## 🔁 Active-Backup FindHop

Create a `findhop` checker on `sw2` with two candidate nexthops and active-backup mode:

```bash
openlan network --name a findhop add \
  --findhop sw0-hop \
  --nexthop 192.53.0.1,192.54.0.1 \
  --check ping \
  --mode active-backup
```

Bind the VIP route to the checker instead of a fixed nexthop:

```bash
openlan network --name a route add \
  --prefix 10.243.0.10/32 \
  --findhop sw0-hop
```

Verify that Linux installs a route through one reachable nexthop and the VIP is reachable:

```bash
ip r get 10.243.0.10
ping -c 3 10.243.0.10
```

The case then validates failover:

- Stop `sw1.0`: traffic moves to `192.54.0.1`.
- Stop `sw1.1` and recover `sw1.0`: traffic moves back to `192.53.0.1`.
- Run `openlan reload --save`: findhop and route state remain valid after reload.

## 🛡️ Remove Guard

A bound findhop cannot be removed while a route still references it:

```bash
openlan network --name a findhop rm --findhop sw0-hop
# checker has route
```

Remove the route first, then remove the findhop:

```bash
openlan network --name a route rm \
  --prefix 10.243.0.10/32

openlan network --name a findhop rm \
  --findhop sw0-hop
```

## ⚖️ Load-Balance FindHop

The same nexthops can be used in load-balance mode:

```bash
openlan network --name a findhop add \
  --findhop sw0-hop-lb \
  --nexthop 192.53.0.1,192.54.0.1 \
  --check ping \
  --mode load-balance

openlan network --name a route add \
  --prefix 10.243.0.10/32 \
  --findhop sw0-hop-lb
```

The resulting Linux route contains both nexthops:

```bash
ip route show
# ... nexthop via 192.53.0.1 ...
# ... nexthop via 192.54.0.1 ...
```

The VIP remains reachable:

```bash
ping -c 3 10.243.0.10
```

Cleanup:

```bash
openlan network --name a route rm \
  --prefix 10.243.0.10/32

openlan network --name a findhop rm \
  --findhop sw0-hop-lb
```
