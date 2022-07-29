# TODO
To Implement OpenLAN prototype by C.
To Implement OpenLAN prototype by C++.

# Golang
v5.2.10: 2 vcpu/ 1G memory
* prototype:          54MiB / 57MiB
* openlan-no-crypt:   32MiB / 57MiB
* openlan-xor-crypt:  21MiB / 57MiB

v5.2.12: 2 vcpu / 1G memory
* openlan-no-trace-no-crypt:    42MiB / 57MiB
* openlan-no-trace-xor-crypt:   41MiB / 57MiB
* openlan-with-trace-xor-crypt: 30MiB / 57MiB

# Protocol
tcp > ws > tls > wss > udp > kcp
