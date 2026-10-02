#! /bin/sh

# This is the script I use to have a working network interface with PCSX2 emulator
#  This setup might not be mandatory for the average user, it's just what works for me

sudo ip link add br0 type bridge
sudo ip addr add 10.0.0.1/24 dev br0  # This IP does nothing except keep PCSX2 happy

# Use a virtual cable to keep br0 up without using a dummy interface
#  This somehow works better than using a dummy for some reason... This is not my cup of tea tbh
sudo ip link add veth0 type veth peer name veth1
sudo ip link set veth0 master br0

sudo ip addr add 192.168.137.1/24 dev veth1

sudo ip link set veth1 up
sudo ip link set veth0 up
sudo ip link set br0 up

# Disable checksumming and offloading to make sure I have no issue with the packet capture
sudo ethtool -K veth0 tso off gso off gro off tx off rx off 2>/dev/null
sudo ethtool -K veth1 tso off gso off gro off tx off rx off 2>/dev/null
# Same-ish thing on L3, again this is not my cup of tea
sudo iptables -t mangle -A POSTROUTING -p udp --dport 68 -j CHECKSUM --checksum-fill

sudo firewall-cmd --zone=trusted --add-interface=br0
sudo firewall-cmd --zone=trusted --add-interface=veth0
sudo firewall-cmd --zone=trusted --add-interface=veth1

# Setup dnsmasq as permissive as possible on the receiving end of the bridge, this step is not mandatory since PCSX2 allows configuring this at emulator level
sudo dnsmasq -d -i veth1 --dhcp-range=192.168.137.10,192.168.137.50,12h --dhcp-broadcast --no-ping --dhcp-authoritative --address=/#/192.168.137.1
