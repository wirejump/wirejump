#!/bin/sh
#
# iptables firewall with essential bits,
# feel free to edit to your liking

PATH="/sbin:/usr/sbin:/bin:/usr/bin"

firewall_start() {
    # create dedicated *-allowed chains
    iptables -N tcp-allowed
    iptables -N udp-allowed
    iptables -N forward-allowed

    # create SSHGuard chain and block whatever it says
    iptables -N sshguard
    iptables -A INPUT -j sshguard

    # accept existing connections and localhost
    iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
    iptables -A INPUT -i lo -j ACCEPT

    # accept allowed TCP/UDP & ICMP
    iptables -A INPUT -p udp -m conntrack --ctstate NEW -j udp-allowed
    iptables -A INPUT -p tcp --syn -m conntrack --ctstate NEW -j tcp-allowed
    iptables -A INPUT -p icmp --icmp-type 8 -m conntrack --ctstate NEW -j ACCEPT

    # drop invalid packets, reject everything which is not explicitly allowed
    iptables -A INPUT -m conntrack --ctstate INVALID -j DROP
    iptables -A INPUT -p udp -j REJECT --reject-with icmp-port-unreachable
    iptables -A INPUT -p tcp -j REJECT --reject-with tcp-reset
    iptables -A INPUT -j REJECT --reject-with icmp-proto-unreachable

    # adjust MSS, it doesn't seem to make a difference to move this to mangle/POSTROUTING;
    # also even the manual uses FORWARD chain:
    # https://manpages.debian.org/bookworm/iptables/iptables-extensions.8.en.html#TCPMSS
    iptables -A FORWARD -p tcp --tcp-flags SYN,RST SYN --j TCPMSS --clamp-mss-to-pmtu

    # conntrack, forward only what's allowed
    iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
    iptables -A FORWARD -j forward-allowed 
    iptables -A FORWARD -j REJECT --reject-with icmp-host-unreachable

    # allow SSH (after SSHGuard checked it)
    iptables -A tcp-allowed -p tcp --dport 22 -j ACCEPT

    # drop everything else
    iptables -P INPUT DROP
    iptables -P FORWARD DROP
}

# clear iptables configuration
firewall_stop() {
    iptables -F
    iptables -X
    iptables -P INPUT   ACCEPT
    iptables -P FORWARD ACCEPT
}

# process params
case "$1" in
    start|restart)
        echo "Starting firewall"
        firewall_stop
        firewall_start
        ;;
    stop)
        echo "Stopping firewall"
        firewall_stop
        ;;
esac
