#!/bin/bash
#
# this script provides some shared data
# and allows reading upstream gw address from a file

# this fwmark will be used to mark all traffic coming from downstream
export FWMARK=33

function get_upstream_gw() {
    local UPSTREAM_GW=""

    if [[ -f "{{ wirejump.basedir }}/config/upstream_gateway" ]]; then
        UPSTREAM_GW=$(sed -e 's/^[ \t\n]*//' "{{ wirejump.basedir }}/config/upstream_gateway")
    fi

    if [[ "$UPSTREAM_GW" != "" ]]; then
        # check address
        ipcalc "$UPSTREAM_GW" 2>&1 | grep -q -i invalid

        if [[ $? -ne 0 ]]; then
            echo "$UPSTREAM_GW"
        fi
    fi

    echo ""
}

function fail() {
    if [[ "$1" != "" ]]; then
        echo "[!] $1"
    fi

    exit 1
}

function info() {
    if [[ "$1" != "" ]]; then
        echo "[#] $1"
    fi
}