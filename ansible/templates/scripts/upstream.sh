#!/bin/bash
#
# this script is run for upstream interface

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
UPSTREAM_GW=$(get_upstream_gw)
TABLE="wirejump_table"

# valid ip is required
if [[ "$UPSTREAM_GW" != "" ]]; then
    if [[ "$INTERFACE" != "" && "$OPERATION" != "" ]]; then
        if [[ "$OPERATION" == "up" ]]; then
            # add link route to default gw (main table)
            ip -4 route add "$UPSTREAM_GW" scope link dev "$INTERFACE"

            # add default gw for upstream (upstream table)
            ip -4 route add 0.0.0.0/0 via "$UPSTREAM_GW" table "$TABLE"

            # use wgtable for fwmarked traffic [coming from downstream]
            ip rule add from all fwmark "$FWMARK" lookup "$TABLE"

            # masquerade everything for upstream
            iptables -t nat -A POSTROUTING -o "$INTERFACE" -j MASQUERADE

            info "$1 brought up"
        elif [[ "$OPERATION" == "down" ]]; then

            # remove masquerade
            iptables -t nat -D POSTROUTING -o "$INTERFACE" -j MASQUERADE

            # no need to route traffic [from downstream] anymore
            ip rule del from all fwmark "$FWMARK" lookup "$TABLE" || info "rule already deleted"

            # remove default gw for upstream (upstream table)
            ip -4 route del 0.0.0.0/0 via "$UPSTREAM_GW" table "$TABLE" || info "def route already deleted"

            # remove link route to default gw (main table)
            ip -4 route del "$UPSTREAM_GW" scope link dev "$INTERFACE" || info "gw route already deleted"

            info "$1 will be brought down"
        else
            fail "invalid operation"
        fi
    else
        fail "invalid params"
    fi
else
    fail "invalid upstream gateway address"
fi
