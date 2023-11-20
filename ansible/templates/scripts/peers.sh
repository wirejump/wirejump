#!/bin/bash
#
# this script updates public addresses of currently connected peers
# should be run by cron/systemd timer

# which interface to query for peers
INTERFACE="{{ wirejump.interfaces.downstream.name }}"

# local zone TTL in seconds
ZONETTL=60

# zone name WITHOUT trailing dot
ZONENAME="wjpeers"

# output file for unbound
HOSTSFILE=/etc/unbound/unbound.conf.d/local-peers.conf

# ensure interface is up and running
if ip a | grep -Eq ": ${INTERFACE}:.*UP"; then
    TEMPFILE=$(mktemp)
    PEERSFILE=$(mktemp)

    # each peer will get a DNS name corresponding to first 8 bytes of their public key
    for PEER in $(wg show "${INTERFACE}" peers); do
        ENDPOINT=$(wg show "${INTERFACE}" endpoints | grep "${PEER}")

        # during reconnects, peer can become stale; don't process it then
        if ! echo "${ENDPOINT}" | grep -q '(none)'; then
            ADDRESS=$(echo "${ENDPOINT}" | cut -f 2 -d ' ' | cut -f 1 -d ':')
            KEYPREFIX=$(echo "${PEER}" | base64 -d | hexdump -n 8 -e '1/1 "%02x"')

            printf "\tlocal-data: \"%s.%s. %s IN A %s\"\n" "${KEYPREFIX}" "${ZONENAME}" "${ZONETTL}" "${ADDRESS}" >> "${PEERSFILE}"
            printf "\tlocal-data-ptr: \"%s %s %s.%s\"\n\n" "${ADDRESS}" "${ZONETTL}" "${KEYPREFIX}" "${ZONENAME}" >> "${PEERSFILE}"
        fi
    done

    # if some peers have been processed, format final file
    if [[ $(wc -l < "${PEERSFILE}") -gt 0 ]]; then
        # shellcheck disable=SC2129
        printf "server:\n" >> "${TEMPFILE}"
        printf "\tlocal-zone: \"%s.\" static\n\n" "${ZONENAME}">> "${TEMPFILE}"
        cat "${PEERSFILE}" >> "${TEMPFILE}"
    fi

    rm "${PEERSFILE}"

    # check if new and old files differ and reload unbound if they do
    if ! cmp -s "${HOSTSFILE}" "${TEMPFILE}"; then
        rm -rf "${HOSTSFILE}"
        cp "${TEMPFILE}" "${HOSTSFILE}"

        # unbound requires precisely these permissions,
        # otherwise it will fail to start
        chmod 644 "${HOSTSFILE}"
        chown unbound:unbound "${HOSTSFILE}"

        systemctl reload unbound
    fi

    rm "${TEMPFILE}"
fi
