# Linux router setup

This guide assumes your router runs some distribution of Linux and is dedicated for advanced users.

## Creating the keys

Let's generate keys according to [the official documentation](https://www.wireguard.com/quickstart/#key-generation):

```
$ wg genkey | tee privatekey | wg pubkey > publickey
```

That gives (example output):
- Private key: `wPYATBX8gRn8tz+rn6HDhGan1eg+4HkEw6MGAt2W7Gk=`
- Public key: `B8uaRHLiYZDmfErZjAdtVpzPKa0szFnGGL2U7Wl+rR0=`

## Creating the interface

It's time to create a new WireGuard interface. On Linux, you have to create interface configuration file in `/etc/wireguard` folder if you want to be able to use `wg-quick` command to manage it later (so, `/etc/wireguard/wg0.conf` -> `wg-quick up wg0`).

Create interface file (`/etc/wireguard/wirejump.conf`):

```
[Interface]
PrivateKey = wPYATBX8gRn8tz+rn6HDhGan1eg+4HkEw6MGAt2W7Gk=
```

## Connecting to the server

In order to communicate with the server, you need to query your IP. You need to login to your server and query free IP:

```
$ wjcli peer --add --pubkey B8uaRHLiYZDmfErZjAdtVpzPKa0szFnGGL2U7Wl+rR0=
Peer
  IPv4 Address:         172.16.1.128/24
  Isolated:             no
```

This command will return the IP address which server has allocated for a particular public key. By default, all peers are NOT isolated and can communicate with each other. If such configuration is not desirable, pass `--isolated` flag.

> Don't worry about `/24` netmask. It's mainly needed in order to reach the server (`172.16.1.1` by default), but will allow you to access other peers (and them to access you). Since all communication between peers over `172.16.1.0` network will go through the server anyway, in case you've got an isolated address, server just won't forward any packets to or from you. Of course, you can specify something like `/31` here.

Adjust your interface file:

```
[Interface]
PrivateKey = wPYATBX8gRn8tz+rn6HDhGan1eg+4HkEw6MGAt2W7Gk=
Address = 172.16.1.128/24
DNS = 172.16.1.1
```

Now, it's time to add server peer:

```
[Peer]
PublicKey = abc
Endpoint = 1.2.3.4:51820
AllowedIPs = 0.0.0.0/0
Table = off
```

In this example:
- `1.2.3.4` is your server public address;
- `51820` is the default WireGuard port;
- `abc` is the public key of your server;
- `0.0.0.0/0` specifies which IP addresses are allowed to be sent/received through this peer. To use this as your new Internet interface, you need to specify `0.0.0.0/0` here (all possible addresses). This guide focuses on allowing all here and then restricting access via firewall. If you have other plans in mind, setup this field accordingly.
- `Table = off` means that `wg-quick` command should NOT create a routing table and add a new default route for you. If you omit this setting, running `wg-quick up wirejump` will create a routing table and force all your trafic to go into the tunnel (easiest configuration). It's recommended to leave this as is and configure routing later manually.

At this point, if you run `wg-quick up wirejump` you should be able to ping your server via it's internal address:

```
$ ping 172.16.1.1
PING 172.16.1.1 (172.16.1.1): 56 data bytes
64 bytes from 172.16.1.1: icmp_seq=0 ttl=63 time=17.382 ms
64 bytes from 172.16.1.1: icmp_seq=1 ttl=63 time=17.480 ms
```

If ping doesn't go through, check out **Tunnel link establishment** section of  [troubleshooting guide](./troubleshooting.md)

## Routing

This is up to you. Easiest way is to remove `Table = off` line from interface configuration file, in order for `wg-quick` to create routing table and default route for you.

If you would like to add route and firewall rules manually, check out MikroTik guide for some inspiration (commands are almost exact match).

## Killswitch

Classic killswitch is to use `PostUp` and `PreDown` items of your interface configuration file:

```
[Interface]
PostUp = iptables -I OUTPUT ! -o %i -m mark ! --mark $(wg show %i fwmark) -m addrtype ! --dst-type LOCAL -j REJECT
PreDown = iptables -D OUTPUT ! -o %i -m mark ! --mark $(wg show  %i fwmark) -m addrtype ! --dst-type LOCAL -j REJECT
```

Keep in mind, that this uses `OUTPUT` chain, which means traffic coming from this machine. For the proper router setup, you need to use `FORWARD` chain here.