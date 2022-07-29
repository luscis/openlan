
# ping on moon

## host2host

ping 192.168.200.130 -s 1500

## net2net

ip netns exec net00 ping 192.168.200.130 -s 1500


# tcpdump on moon

tcpdump -i ens33 -p esp -nne
