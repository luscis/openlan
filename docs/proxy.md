# 🔀 Ceci Proxy Examples

These examples follow the proxy scenarios in `tests/cases`:

- `proxy_http.sh`: HTTP forward proxy to an HTTP target.
- `proxy_tcp.sh`: TCP proxy to a fixed target.
- `proxy_name_backends.sh`: DNS/name proxy with domain-matched backends.

All examples use Docker management network `100.100.0.0/24` and OpenLAN output links between switches.

## 🌐 HTTP Proxy

Topology:

```text
sw1 proxy client 192.52.0.1
       | wget via local Ceci HTTP proxy
       v
sw1 openceci(http) -- output --> sw2 192.52.0.2:18081
```

Configure `sw1`:

```bash
openlan network --name example add --address 192.52.0.1/24
openlan user add --name t1@example --password 123456
openlan ceci proxy add --mode http --listen 127.0.0.1:11082
```

Configure `sw2` and connect it back to `sw1`:

```bash
openlan network --name example add --address 192.52.0.2/24

openlan network --name example output add \
  --remote 100.100.0.241 \
  --protocol tcp \
  --secret t1@example:123456 \
  --crypt aes-128:ea64d5b0c96c
```

Validate from `sw1`:

```bash
wget -q -O- \
  -e use_proxy=yes \
  -e http_proxy=http://127.0.0.1:11082 \
  http://192.52.0.2:18081/
```

The case expects the response body `proxy-http-ok` and a log entry matching `HttpProxy.ServeHTTP`.

## 🔁 TCP Proxy

Topology:

```text
sw1 proxy client 192.53.0.1
       | wget via local Ceci TCP proxy
       v
sw1 openceci(tcp) -- output --> sw2 192.53.0.2:18082
```

Configure `sw1`:

```bash
openlan network --name example add --address 192.53.0.1/24
openlan user add --name t1@example --password 123456

openlan ceci proxy add \
  --mode tcp \
  --listen 127.0.0.1:12082 \
  --target 192.53.0.2:18082
```

Configure `sw2` and connect it back to `sw1`:

```bash
openlan network --name example add --address 192.53.0.2/24

openlan network --name example output add \
  --remote 100.100.0.241 \
  --protocol tcp \
  --secret t1@example:123456 \
  --crypt aes-128:ea64d5b0c96c
```

Validate from `sw1`:

```bash
wget -q -O- http://127.0.0.1:12082/
```

The case expects the response body `proxy-tcp-ok` and a log entry matching `TcpProxy.tunnel`.

## 🧭 Name Proxy with Matched Backends

Topology:

```text
sw1 openceci(name)
   ^                 ^
   | domain A         | domain B
sw2 dnsmasq        sw3 dnsmasq
192.55.0.2         192.55.0.3
```

- `sw1=192.55.0.1`
- `sw2=192.55.0.2`, upstream DNS A: `192.55.0.2:5353`
- `sw3=192.55.0.3`, upstream DNS B: `192.55.0.3:5353`
- Domain A: `proxy-name-a.test -> 192.55.0.2`
- Domain B: `proxy-name-b.test -> 192.55.0.3`

Create the name proxy configuration on `sw1`:

```yaml
listen: 127.0.0.1:1054
nameto: 8.8.8.8
backends:
  - server: 192.55.0.2
    match:
      - proxy-name-a.test
    nameto: 192.55.0.2:5353
  - server: 192.55.0.3
    match:
      - proxy-name-b.test
    nameto: 192.55.0.3:5353
```

Start `openceci` in name mode:

```bash
openceci \
  -mode name \
  -conf /var/openlan/ceci/127.0.0.1:1054.yaml \
  -log:file /var/openlan/ceci/127.0.0.1:1054.log
```

Validate DNS routing:

```bash
nslookup -port=1054 proxy-name-a.test 127.0.0.1
# Address: 192.55.0.2

nslookup -port=1054 proxy-name-b.test 127.0.0.1
# Address: 192.55.0.3
```

The case also checks that routes are installed toward matched backend answers and that `openlan reload --save` can restart the proxy path.
