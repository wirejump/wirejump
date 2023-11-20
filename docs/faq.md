# FAQ

## What about end-to-end encryption?

Please refer to [architecture page](architecture.md) to see the route your data will travel. In short: connection to your server is E2E encrypted, as the connection between your server and a VPN provider of your choice. After data leaves your home network, it can be accessed _at least_ twice: at your server, and at the exit node of your VPN provider.

## How secure is this?

As secure as you want to make it. If you ABSOLUTELY wanna be sure that _they_ won't get you, go and touch some grass right now. On a more serious note, default install is pretty strict: firewall is minimal and by default only SSH port is opened to the outside (and it's guarded by `sshguard`). Once `downstream` interface gets activated, outside port for incoming WireGuard connections is opened as well, so peers can connect to your server. Also, DNS ports 53 (tcp and udp) are opened for `downstream` peers (but not for everyone). `OUTPUT` chain is empty, since stuff is routed manually, and your default server internet connection is not affected. `unbound` also uses default server internet connection to reach upstream DNS servers, and is configured to use DoT only. `upstream` interface is not firewalled since it's up to your VPN provider and your particular needs.

## How private is this?

Main question here is: whom you can trust? If you're using a VPN you're already trusting your VPN provider and giving them your data (literally, they route it). If you gonna host this in a cloud, you gonna trust that cloud provider: technically, they can pause your VM and dump all machine memory. Even more, since this is NOT FULLY E2E ENCRYPTED (assuming you've read the first answer), they can run `tcpdump` inside your VM and just dump all decrypted data as it comes from `downstream` before it's encrypted again and sent to your VPN provider. Good news is, most of the traffic today is already encrypted, so they won't get much info from that. However, if you're using default `unbound` DNS server, it can be inspected for all the hostnames you're accessing; however they would need either to dump machine memory or change server configuration for that (default install has remote control disabled and has zero logs). WireJump connection manager daemon generates only minimal logs and your VPN provider data is never saved to disk. That's why after your reboot your server you have to setup your connection again.

## How do I change server settings after installation?

### Overview

It's much easier to deploy the server with your required settings right from the start. However, if you feel adventurous and want to get familiar with the project architecture, go on.

Server installation is organized into different folders:

- `/etc/wireguard`: system WireGuard directory
- `/opt/wirejump/config`: interface and daemon config files
- `/opt/wirejump/scripts`: contains interface scripts and auxillary scripts
- `/opt/wirejump/bin`: wirejump daemon and cli binaries

### Add/remove downstream peers

You can do it via `wjcli peer` command (recommended), as outlined in the [client setup](./client.md). If you want to do it manually, change `/opt/wirejump/config/downstream.conf` and restart `downstream` interface after.

### Change interface names

If you want to change `downstream` interface name, again, it's MUCH easier to change playbook and redeploy the server.  
Actual interface names are defined by symlink names in `/etc/wireguard`, and those are registered with systemd to be picked up by `wg-quick` when interface needs to be restarted. Also, WireJump daemon restart (or server reboot) will be required for the change to take effect. For example, to change `downstream` interface name to `my-downstream-wg0`:

```
wg-quick down downstream
systemctl stop wirejumpd
systemctl disable wg-quick@downstream
cd /etc/wireguard
mv downstream.conf my-downstream-wg0.conf
systemctl enable wg-quick@my-downstream-wg0
sed -i 's/downstream/my-downstream-wg0' /opt/wirejump/config/wirejumpd.conf
sed -i 's/downstream/my-downstream-wg0' /etc/systemd/system/wirejump.service
systemctl daemon-reload
systemctl start wirejumpd
wg-quick up my-downstream-wg0
```

### Change downstream port, MTU or network

If you want to change some of these settings, you'll need to stop the interface first (so old settings are deleted), make changes and start the interface again:

```
wg-quick down downstream
vim /opt/wirejump/scripts/downstream.sh
vim /opt/wirejump/config/downstream.conf # required if you need to change downstream network address
wg-quick up downstream
```

## Open/close firewall port

Firewall is located at `/opt/wirejump/scripts/firewall.sh`. It's advised to use dedicated chains like `tcp-allowed` and `udp-allowed`. For example, once `downstream` is brought up, interface script adds an entry in `udp-allowed` to allow you to connect to the server from outside; it also opens DNS ports but they are accessible from `downstream` only (and from the server itself).

## DNS settings

Unbound main configuration file is `/etc/unbound/unbound.conf.d/main.conf`. There's a special script which creates local DNS zone for all connected peers, so they can communicate with each other without knowing their addresses. Zone name is defined at `/etc/wirejump/scripts/peers.sh`. This script is scheduled by cron to run every minute. Adjust crontab to change that interval if needed.
