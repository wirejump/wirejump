#!/bin/bash
#
# this script processes downstream interface

# MTU for downstream interface
MTU="{{ wirejump.interfaces.downstream.mtu }}"

# WireGuard port for connecting from outside
PORT="{{ wirejump.interfaces.downstream.port }}"

# downstream interface CIDR
CIDR="{{ wirejump_downstream_cidr }}"

# upstream interface name (needed for killswitch)
UPSTREAM="{{ wirejump.interfaces.upstream.name }}"

if [[ -z "$THISDIR" ]]; then
    THIS=$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || echo "$0")
    THISDIR=$(dirname "${THIS}")
fi

if [[ -f "$THISDIR/gateway.sh" ]]; then
    # shellcheck source=gateway.sh
    source "$THISDIR/gateway.sh"
else
    echo "[!] failed to find gateway.sh script"
    exit 1
fi

INTERFACE="$1"
OPERATION="$2"

if [[ "$INTERFACE" != "" && "$OPERATION" != "" ]]; then
    if [[ "$OPERATION" == "up" ]]; then
        # mark all outside-destined traffic coming from this interface
        iptables -A PREROUTING -t mangle -i "$INTERFACE" ! -d "$CIDR" -j MARK --set-mark "$FWMARK"

        # killswitch: reject all outgoing non-local downstream traffic trying to leave the server not via upstream; allow everything else
        iptables -A forward-allowed -i "$INTERFACE" ! -o "$UPSTREAM" -m mark --mark "$FWMARK" -j REJECT --reject-with icmp-host-unreachable
        iptables -A forward-allowed -i "$INTERFACE" -j ACCEPT

        # allow DNS requests from downstream
        iptables -A tcp-allowed -p tcp -i "$INTERFACE" --dport 53 -j ACCEPT
        iptables -A udp-allowed -p udp -i "$INTERFACE" --dport 53 -j ACCEPT

        # finally, allow connection of downstream WireGuard peers from outside;
        # explicitly disallow upstream, even though it's probably filtered already
        iptables -A udp-allowed ! -i "$UPSTREAM" -p udp --dport "$PORT" -j ACCEPT

        # adjust MTU
        /sbin/ip link set dev "$INTERFACE" mtu "$MTU"

        info "$1 brought up"
    elif [[ "$OPERATION" == "down" ]]; then
        iptables -D udp-allowed ! -i "$UPSTREAM" -p udp --dport "$PORT" -j ACCEPT
        iptables -D udp-allowed -p udp -i "$INTERFACE" --dport 53 -j ACCEPT
        iptables -D tcp-allowed -p tcp -i "$INTERFACE" --dport 53 -j ACCEPT

        iptables -D forward-allowed -i "$INTERFACE" -j ACCEPT
        iptables -D forward-allowed -i "$INTERFACE" ! -o "$UPSTREAM" -m mark --mark "$FWMARK" -j REJECT --reject-with icmp-host-unreachable

        iptables -D PREROUTING -t mangle -i "$INTERFACE" ! -d "$CIDR" -j MARK --set-mark "$FWMARK"

        info "$2 brought down"
    else
        fail "invalid interface operation"
    fi
else
    fail "invalid params"
fi
