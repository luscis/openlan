module github.com/luscis/openlan

go 1.22.0

toolchain go1.23.0

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.8.1
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/net => github.com/golang/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190209173611-3b5209105503
	golang.org/x/time => github.com/golang/time v0.0.0-20210220033141-f8bda1e9f3ba
)

exclude github.com/sirupsen/logrus v1.8.1

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/go-logr/logr v1.1.0
	github.com/go-logr/stdr v1.1.0
	github.com/gorilla/mux v1.8.0
	github.com/miekg/dns v1.1.65
	github.com/moby/libnetwork v0.5.6
	github.com/ovn-org/libovsdb v0.6.1-0.20220127023511-a619f0fd93be
	github.com/prometheus/client_golang v1.11.0
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/vishvananda/netlink v1.1.0
	github.com/xtaci/kcp-go/v5 v5.6.1
	golang.org/x/net v0.35.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/cenkalti/hub v1.0.1 // indirect
	github.com/cenkalti/rpc2 v0.0.0-20210604223624-c1acbc6ec984 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/chzyer/logex v1.2.0 // indirect
	github.com/chzyer/test v0.0.0-20210722231415-061457976a23 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/libnetwork v0.5.6 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/reedsolomon v1.9.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mmcloughlin/avo v0.0.0-20200803215136-443f81d77104 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/riobard/go-bloom v0.0.0-20200614022211-cdc8013cb5b3 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/templexxx/cpu v0.0.7 // indirect
	github.com/templexxx/xorsimd v0.4.1 // indirect
	github.com/tjfoc/gmsm v1.3.2 // indirect
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
