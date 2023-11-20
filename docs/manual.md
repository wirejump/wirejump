# User manual

This document assumes you have successfully installed the server and configured a client. It will guide on how to use your WireJump server. 

## Server management

Server installation features a CLI control utility – `wjcli`. It should be present in default PATH, but if it's not, it's located at `/opt/wirejump/bin/wjcli` by default.

It's strongly recommended to manage your server via this utility.

```
Usage:  wjcli [OPTIONS] COMMAND

WireJump server management CLI

Options:
  -h, --help                  Display this help             
  -v, --version               Display program version       

Commands:
  peer                        Manage downstream peers           
  list                        List available providers          
  setup                       Setup upstream provider           
  servers                     Manage available server locations 
  connect                     Manage upstream connection        
  status                      Get current connection status     
  disconnect                  Disconnect upstream               
  reset                       Reset upstream state              
  version                     Get server daemon version
```

Each command has extensive help, which can be accessed via `-h/--help` flag. Your supposed workflow is the following:

- setup a peer with `wjcli peer` (part of client installation)
- get a list of all available VPN providers via `wjcli list`
- setup a VPN connection via `wjcli setup`
- setup a preferred exit server location via `wjcli servers`
- establish a connection via `wjcli connect`
- check connection status via `wjcli status`
- enjoy!

You can reconnect by running `wjcli connect` again (will connect to a different exit node within the same preferred country, if it's set; if no preference is available, will select a random country and server) or disconnect via `wjcli disconnect`. If you want to reset current VPN provider state (and connection, if it's active), run `wjcli reset`. Reset command will also try to delete your last used public key from provider account before resetting the state.

Every `wjcli` command supports JSON output, if you want to integrate this tool into your workflow. This mode can be triggered by providing a `-j/--json` flag. For example (usual output):

```
$ wjcli status
Upstream connection
  Online:               yes                           
  Active since:         Fri, 10 Mar 2000 19:59:59 UTC 
  Country:              France                       
  City:                 Paris                     
Provider
  Name:                 mullvad                       
  Preferred location:   France                       
  Account expires:      Fri, 10 Mar 2000 19:59:59 UTC                      
```

JSON output:
```
$ wjcli status -j
{"error":false,"message":{"upstream":{"online":true,"active_since":952714799,"country":"France","city":"Paris"},"provider":{"name":"mullvad","preferred":"France","expires":952714799}}}
```

In JSON mode, all time fields are returned as UNIX timestamps of integer type.

If command supports data entry, you can trigger interactive input via `-i/--interactive:`

```
$ wjcli servers -i
Interactive mode enabled
Preferred location : 
```

In this case, you can enter sensitive data (like your VPN provider credentials) without them being saved in your shell command history.

## Design choices

- `wjcli setup` is the only command which will trigger interactive mode if you don't provide required data via command-line options. All other commands will display an error if required data is missing;
- For Mullvad specifically, available servers are filtered down to only include those which are owned by Mullvad, that's why countries list is less than on their website;
- All data which is entered into the tool is kept in memory and is never written to disk. That's why you have to setup provider again if you reboot your server;
- There's a cron job which updates servers (`wjcli servers`) every hour, so you don't have to do it manually (but you still can, if you want).

## Automation

Server installation creates an additional user account (`manager` by default), which uses a special shell and is restricted to `wjcli` command only. This can be useful in various automation scenarios. For example, you may want to schedule a script to reconnect daily or reset your connection after some time. It's recommended to add a public key of your device to the server for passwordless login from a scheduler/cron script.

### MikroTik note

Despite latest ROS version (7.12 at the moment of writing this) claims to support ED25519 keys completely, it's still impossible to import such key type, so you have to stick with RSA. Generate keys manually:

```
ssh-keygen -b 2048 -t rsa -m PEM -f id_router
```

And add value of `id_router.pub` to the file `/home/manager/.ssh/authorized_keys` on the server. Then, upload private key to the router and import it for your current user:

```
/user/ssh-keys/private import user=admin private-key-file=id_router
```

If you want to execute the script from another user, adjust `user` field accordingly.

Next, add a [reconnect script](../scripts/reconnect.rsc) to your router' scripts directory. Finally, add this script to the scheduler. For example, this will reconnect your network daily at 5 PM (script is saved as `wirejump-reconnect`):

```
/system/scheduler add interval=1d name=reconnect on-event=wirejump-reconnect policy=read,write start-time=17:00:00
```

## Accessing home network from the Internet

Assuming your home router is configured correctly and is accepting SSH connections from `downstream`, you can forward any ports from your home network or just access it directly. It works best when you have setup a public-key SSH authentication for your home router and target machine in your home network.

In this example:
- `machine` is the target machine in your home network. For example, `pi@192.168.1.24`
- `wirejump` is your deployed WireJump server. For example, `root@11.22.33.44`
- `router` is your home router, accessed by its address in the `downstream` network (configured at [client setup](client.md)). For example, `root@172.16.1.128`
- `5900` is the local port on your local machine
- `5900` is the port on your home network

Finally, let the magic of SSH JumpHost do its thing. To access target machine:
```
ssh machine -J wirejump -J router
```

To access service running at your home network locally:
```
ssh -L5900:machine:5900 -J wirejump router -N
```

If that's not enough, how about you add this command to your local `~/.ssh/config` file:

```
Host home-raspberry-vnc
	User root
	HostName 172.16.1.128
	LocalForward 5900 192.168.1.24:5900
	ProxyJump root@11.22.33.44
```

Now, you can access your Raspberry Pi remote desktop from anywhere only by running
```
ssh home-raspberry-vnc -N
```
in your terminal. This will work even after your home ISP reconnects your router and issues a new public IP (DSL reconnects in the middle of the night, for example). Isn't that cool?

To know more about SSH forwarding, check out some examples: https://help.ubuntu.com/community/SSH/OpenSSH/PortForwarding

## P2P communication

Server install features `unbound` server, which is configured to provide DNS service to downstream network. One of the features is a special DNS zone (`.wjpeers` by default) which allows your peers to communicate directly between each other. Consider a following scenario:

You're sharing your WireJump server with a friend and want to access service running in their network (or vice versa).

You setup port forwards and firewall rules and everything is working great (you're communicating via `downstream` network). However, at some point you discover that your WireJump server has metered traffic (1TB, for example), and you have to pay extra for everything over that limit. What to do?

You can establish a _direct_ WireGuard connection between you and your friend, using your public addresses, omitting the WireJump server. But wait – what if you both are having dynamic IPs? In this case, each peer' public IP is represented by a special name in `.wjpeers` zone. Name consists of first few characters of peer public key and is persistent unless key changes. Server has a dedicated updater script, which is being run by cron every minute, so peer hostnames are being updated as soon as the addresses change.

For MikroTik users, this is essentially the same as `IP -> Cloud -> DDNS Enabled`: you're also getting DDNS, but in this case it's tied to your WireGuard interface public key.

To setup such configuration, log in to your server first and inspect all peers:
```
# wg show downstream
interface: downstream
  public key: kkkkkkkkkkkkk
  private key: (hidden)
  listening port: 12345

peer: xxxxxxxxxxxxx
  endpoint: 11.22.33.44:12345
  allowed ips: 172.16.1.78/32, 172.16.1.0/24
  latest handshake: 44 seconds ago
  transfer: 33.58 GiB received, 43.35 GiB sent

peer: yyyyyyyyyyyyy
  endpoint: 22.33.44.55:12345
  allowed ips: 172.16.1.103/32, 172.16.1.0/24
  latest handshake: 1 minute, 55 seconds ago
  transfer: 52.76 GiB received, 90.31 GiB sent
```

Let's say your desired peer is `yyyyyyyyyyyyy`, and it's connected from `22.33.44.55`. Inspect unbound zone file:
```
# grep 22.33.44.55 /etc/unbound/unbound.conf.d/local-peers.conf 
local-data: "123456789abc.wjpeers. 60 IN A 22.33.44.55"
local-data-ptr: "22.33.44.55 60 123456789abc.wjpeers"
```

Thus, your desired hostname is `123456789abc.wjpeers`. It will resolve to a different IP address once it changes, but hostname will depend only on peer's public key. WireGuard has the ability to resolve hostname in endpoints, and that will come in handy for establishing such connection.

Setup new WireGuard interfaces on both peers. Select a peer which should be a server: it will require a port opened to the Internet. You may need to setup port forwarding on your ISP's router. On another peer, which will connect to the server, add hostname in `wjpeers` zone as Endpoint address.

If you're using MikroTik, you will need a script which will resolve these addresses and update endpoint settings for peers accordingly. Take a look at [scripts/direct-peering.rsc](../scripts/direct-peering.rsc). Depending on your routing configuration, you will need to either add a route or an address list entry (but not both).
