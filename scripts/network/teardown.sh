#! /bin/sh

# This is the teardown script for scripts/network/setup

sudo killall dnsmasq 2>/dev/null

sudo ip link delete br0 2>/dev/null
sudo ip link delete veth0 2>/dev/null

