# Setup

Please read whole setup guide thoroughly and make a backup of your router configuration before proceeding! Assuming you know how to manage your router (routes, interfaces etc), whole WireJump setup from start to finish should not take more than 30 mins.

## Overview

In a nutshell, there are three pieces of the puzzle:
- Server: the most important part. It will accept connections from your clients and will forward them to the VPN provider of your choice;
- Client (home router or another device): you will need to create a WireGuard tunnel to your server and setup some additional stuff if it's a router;
- VPN account. It should be active and it should be possible to add new keys/devices in the account.

## Index
- [Server setup](server.md)
- [Client setup](client.md)
- [VPN setup](upstream.md)
