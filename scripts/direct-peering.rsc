# This script updates direct peer endpoint and route (if needed)
# Add this script to the scheduler and execute regularly, depending on your connection.
# 1 hour should be good enough for most cases.

# WireGuard interface name to use for direct peering
:local INTERFACE "wireguard5"

# Remote peer hostname
:local WJPEER "123abc.wjpeer"

# Whether to add a dedicated route, and via which gateway
:local ADDROUTE false
:local GATEWAY "192.168.1.1"

# Whether to add resolved address to the address list
:local ADDTOLIST true
:local LISTNAME "WireJumpExclude"

##############################
##############################
:local CURRENT
:local RESOLVED
:local FAIL false

:do {
    :set CURRENT [/interface/wireguard/peers get value-name=endpoint-address [find interface=$INTERFACE]]
    :set RESOLVED [:resolve $WJPEER]
} on-error={
    :set FAIL true
    :log info "WireJump: failed to resolve direct peer address"
}

:if ($FAIL != true) do={
    # upgrade endpoint address
    :if ($CURRENT != $RESOLVED) do={
        /interface/wireguard/peers set endpoint-address=$RESOLVED [find interface=$INTERFACE]

        :log info "WireJump: updated direct peer endpoint"

        # update static route if needed
        :if ($ADDROUTE = true) do={
            # remove old route first
            :if ([/ip/route find dst-address="$CURRENT/32"] != "") do={
                /ip/route remove [find dst-address="$CURRENT/32"]
            }

            # check for the new one
            :if ([/ip/route find dst-address="$RESOLVED/32"] = "") do={
                /ip/route add dst-address="$RESOLVED/32" gateway=$GATEWAY comment="WireJump: direct peer"

                :log info "WireJump: updated direct peer route"
            }
        }

        # update address list if needed
        :if ($ADDTOLIST = true) do={
            # remove old entry if present
            :if (:len [/ip/firewall/address-list find address="$CURRENT" list="$LISTNAME"] != 0) do={
                /ip/firewall/address-list remove [find address="$CURRENT" list="$LISTNAME"]
            }

            # check for the new one
            :if (:len [/ip/firewall/address-list find address="$RESOLVED" list="$LISTNAME"] != 0) do={
                /ip/firewall/address-list add address="$RESOLVED" list="$LISTNAME" comment="WireJump: direct peer"

                :log info "WireJump: updated direct peer address list entry"
            }
        }
    }
}
