# Troubleshooting

# How do I...
Check out [FAQ](./faq.md) document.

# There's no Internet!

There can be multiple reasons to this, as there are multiple links in the system. Quickest way to determine where the problem lies is to check your routing from any device connected to your router (example for Mullvad):

```
$ traceroute 1.1.1.1
traceroute to 1.1.1.1 (1.1.1.1), 64 hops max, 52 byte packets
 1  192.168.1.1 (192.168.1.1)  0.924 ms  0.613 ms  0.473 ms
 2  172.16.1.1 (172.16.1.1)  16.056 ms  15.971 ms  15.884 ms
 3  10.64.0.1 (10.64.0.1)  27.467 ms  27.184 ms  27.386 ms
 4  something.31173.se (13.12.48.89)  28.237 ms  28.836 ms  28.148 ms
 ...
 8  one.one.one.one (1.1.1.1)  30.056 ms  30.174 ms  30.450 ms
```

Your particular issue can be identified by a failed hop number:
- First hop (`192.168.1.1`) is your home router. If you can't reach it, check your WiFi/cable and/or reboot your router.
- Second hop (`172.16.1.1`) is your WireJump server. Consult **Server reachability** and **Tunnel link establishment** sections.
- Third hop (`10.64.0.1`) is your VPN provider gateway. Consult **VPN reachability** section.
- All other hops are not important, because they describe your VPN provider routing and change when you change your VPN connection.

If you can reach `1.1.1.1`, check if you have DNS resolution working:
```
nslookup google.com
```
If this has failed, consult **DNS reachability** section.


## Server reachability
Start from checking your local Internet connection: ensure modem/router is on, connection is active and that you can reach your server via `downstream` interface. Run this from any local machine, connected to your router (or from the router itself):

```
$ ping 172.16.1.1
PING 172.16.1.1 (172.16.1.1): 56 data bytes
64 bytes from 172.16.1.1: icmp_seq=0 ttl=63 time=15.759 ms
64 bytes from 172.16.1.1: icmp_seq=1 ttl=63 time=16.219 ms
64 bytes from 172.16.1.1: icmp_seq=2 ttl=63 time=15.988 ms
^C
--- 172.16.1.1 ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 15.759/15.989/16.219/0.188 ms
```

If there's no reply, it means that the tunnel is either dead or your Internet connection is down. Ensure Internet is active by logging into your router and pinging server by its public IP (assuming your server address is `11.22.33.44`):

```
[root@router] > /ping 11.22.33.44
  SEQ HOST                                     SIZE TTL TIME       STATUS                                       
    0 11.22.33.44                                56  54 19ms918us
    1 11.22.33.44                                56  54 18ms361us
    2 11.22.33.44                                56  54 18ms371us
    3 11.22.33.44                                56  54 18ms477us
    sent=4 received=4 packet-loss=0% min-rtt=18ms361us avg-rtt=18ms782us max-rtt=19ms918us
```

If you can't reach your server, try reaching it from the Internet by its public address. If it's still not available, well, check it with your hosting management panel.

## Tunnel link establishment
First and foremost: **double check your client and server keys!**

Setting up WireGuard for a first time can be frustrating because there are no certs and passwords, and whole configuration depends on whether keys match or not. However, once you get that out of the picture, tunnel usually starts working by itself without any restarts.

If you have not changed default settings and followed client setup closely, then tunnel should work by itself. If it does not:
- You may want to reduce `PersistentKeepalive` setting value on your client. By default it's set to `25` secs, and it works quite okay with DSL ISPs, for example. You may want to reduce this value to `10`. Interface restart is required on both client and server.
- Another option to look for is `MTU`. Default setting of `1420` should fill most use cases (additional tunneling, DSL, mobile hotspot etc), but you may want to reduce it to `1380` or even less.

To change these settings value on the server, you'll need stop `downstream` interface, change the config (MTU setting is on top of the file) and start interface again:

```
# wg-quick stop downstream
# vim /opt/wirejump/scripts/downstream.sh
# wg-quick start downstream
```

Sometimes you need to reboot whole server, though usually it's not required. If you'll go with rebooting, it's probably good idea to run some `apt-get update && apt-get upgrade` before, to fetch new kernel, which may contain latest WireGuard fixes.

If you still can't connect, your ISP may be blocking your WireGuard connection. In that case, consider wrapping `downstream` into another protocol. For example:

- https://github.com/shadowsocks
- https://github.com/wangyu-/udp2raw
- https://github.com/Snawoot/wg-decoy
- https://github.com/Snawoot/dtlspipe
- https://github.com/infinet/xt_wgobfs

Detailed example is out the scope of this project. Contributions are welcome.

## VPN reachability
Login to your server and check connection status:
```
$ wjcli status
Connection
  Active:               true                   
  Provider:             mullvad                    
  Preferred location:   N/A                     
  Country:              Netherlands                     
  City:                 Rotterdam                     
Account
  Valid until:          Fri, 09 Feb 2023 19:55:00 UTC                     
  Valid until (UNIX):   1234567890
```

`wjcli` command is pretty verbose if something is wrong, so consult with its output. In any case, if your provider has been set up successfully and your account is not expired yet, restart your `upstream` connection by issuing

```
wjcli connect
```

## While at the server
Default install features lots of tools: nmap, mtr, tcpdump to name a few. They can be pretty useful for diagnosting any network-related issues. While you're at the server, you might want to inspect your WireGuard interfaces manually:
```
# wg
interface: downstream
  public key: xxxxxxxxxxxxx
  private key: (hidden)
  listening port: 12345

peer: xxxxxxxxxxxxx
  endpoint: 11.22.33.44:12345
  allowed ips: 172.16.1.78/32, 172.16.1.0/24
  latest handshake: 44 seconds ago
  transfer: 33.58 GiB received, 43.35 GiB sent

peer: xxxxxxxxxxxxx
  endpoint: 22.33.44.55:12345
  allowed ips: 172.16.1.103/32, 172.16.1.0/24
  latest handshake: 1 minute, 55 seconds ago
  transfer: 52.76 GiB received, 90.31 GiB sent

interface: upstream
  public key: xxxxxxxxxxxxx
  private key: (hidden)
  listening port: 16456

peer: xxxxxxxxxxxxx
  endpoint: 50.40.30.20:51820
  allowed ips: 0.0.0.0/0
  latest handshake: 56 seconds ago
  transfer: 475.87 MiB received, 33.77 MiB sent
```

Both `downstream` and `upstream` interfaces should be active. If `downstream` is not active, your clients will be unable to connect. Start it manually:
```
wg-quick up downstream
```

If `downstream` is active, but there's still no connection, check `latest handshake` field. If it's more than few mins, peer has probably went stale. Restart your ISP modem/router and try again. You can also try to restart the interface itself:
```
wg-quick down downstream && wg-quick up downstream
```
Your shell might freeze for few a seconds, but SSH session should survive. Also, you may consider using something like [mosh](https://mosh.org).

If you're connected to your server by SSH via `downstream`, two commands together are required, otherwise you'll be kicked out immediately. If you want to debug `downstream` in peace, connect to your server via its external IP (and NOT via your router :) ). 

If `upstream` interface is not present, it means that VPN connection has not been set up yet or has failed for some reason. Consult connection manager:
```
wjcli status
```

Frankly speaking, nothing stops you from doing stuff like 

```
wg-quick down upstream
wg-quick up upstream
```

However, you still will be connected to the same upstream (exit node), and if it's broken or not available, this will not help. That's why it's advised to manage VPN connection via `wjcli`, which will take care of VPN server selection.


## DNS reachability
Default install features `unbound` running on the server. While you're not required to use it for your downstream network, it may be beneficial if you want to establish direct connections between your peers, as there's a dedicated local zone for them. You also may want to put your ad-blocking lists there (if you're not using PiHole already).

Anyway, check your DNS from within your local network (here `192.168.1.1` is the address of your router) first:

```
$ nslookup google.com 192.168.1.1
Server:		192.168.1.1
Address:	192.168.1.1#53

Non-authoritative answer:
Name:	google.com
Address: 142.250.186.174
```

Your options here are: reconnect your device, restart your router, check further (if router is using WireJump DNS server as upstream, as per default setup):

```
nslookup google.com 172.16.1.1
```

If that fails, you may want to SSH into your server and restart unbound:

```
systemctl restart unbound
```

If that didn't help, try hardcoding something like `1.1.1.1` as upstream DNS in your router settings, and try again. If it still didn't help, your VPN provider may be blocking your DNS requests. Your options here are either fixing unbound (because it does not use VPN tunnel for outgoing connections), or using DoH/DoT on your clients.

## Everything seems to be ok, but the Internet is still down!
Some devices can display "There is no Internet" message when Internet is actually present. It usually happens due to some misconfiguration of your local network. If DNS is working normally, devices in question should reach their connection checking endpoints without any issue (99% of all messages like this are caused by lagging DNS). Reconnect/restart your devices and restart your router.

Also, modern browsers started to use DoH/DoT by default, and in this case they try to resolve DNS by themselves, bypassing your local DNS server, advertised by your router. Either disable that option or restart your browser.


# Some sites work, while others don't

That sounds like MTU issue. Ensure that MTU on both client and server match and that's the value is <1500. In some extreme cases values like `1280` can be required. Consult **Tunnel link establishment** section.

# Can't reach other peers
If you are getting `ping: sendmsg: Required key not available` or can't reach other peers, that means that you registered your device as `isolated` peer. Either change config manually by adding required network manually (**While at the server** section), or just re-register this peer without `--isolated` flag.
