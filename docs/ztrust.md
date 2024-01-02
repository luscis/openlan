# Zero Trust 

## Enable ztrust on a network
```
$ cat /etc/openlan/switch/network/example.json
{
	...
	"ztrust": "enable"
}
$
```

## Add yourself to ztrust
```
$ export TOKEN="daniel@example:g4nlzmk5nxek1hbcqsbr"
$ export URL="https://your-central-switch-address:10000"
$ openlan guest add
$ openlan guest ls
# total 1
username                 address
daniel@internal          169.254.15.6
$
```

## Knock a host service
```
$ openlan knock add --protocol icmp --socket 192.168.20.10
$ openlan knock add --protocol tcp --socket 192.168.20.10:22
$ openlan knock ls
# total 2
username                 protocol socket                   age  createAt
daniel@internal          tcp      192.168.20.10:22         57   2024-01-02 12:42:06 +0000 UTC
daniel@internal          icmp     192.168.20.10:           46   2024-01-02 12:41:55 +0000 UTC
$
```

## Connect to a host service
```
$ ssh root@192.168.20.10 who
root     pts/0        2024-01-02 06:49 (100.66.88.1)
$
```