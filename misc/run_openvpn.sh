#!/bin/bash


mkdir -p /dev/net
if [ ! -c /dev/net/tun ]; then
    mknod /dev/net/tun c 10 200
fi

iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -o eth0 -j MASQUERADE

/build/capture-server -i=eth0 -v=2 -logtostderr=true  &
exec openvpn --config /etc/openvpn/server.conf