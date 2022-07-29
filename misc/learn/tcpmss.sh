#!/bin/bash

# modprobe ipt_LOG
# iptables -A OUTPUT -t raw -p icmp -j LOG
# iptables -A PREROUTING -t raw -p icmp -j LOG

BR="br-vxlan"

iptables -t mangle -A FORWARD -i $BR -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --set-mss 1332


