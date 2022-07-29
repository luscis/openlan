
# --

sudo ip netns del net00
sudo ip link del veth-local

sudo ip link add veth-local type veth peer name veth-remote

sudo ip link set veth-local up
sudo ip addr add 192.168.100.1/24 dev veth-local
sudo ip route add 192.168.200.0/24 via 192.168.100.1


sudo ip netns add net00
sudo ip link set veth-remote netns net00

sudo ip netns exec net00 ip link set veth-remote up

sudo ip netns exec net00 ip addr add 192.168.100.141/24 dev veth-remote
sudo ip netns exec net00 ip route add 192.168.200.0/24 via 192.168.100.1
