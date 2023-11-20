# WireJump connection manager

This folder contains the code of core WireJump components – connection manager daemon (`wirejumpd`) and CLI client (`wjcli`). Daemon manages upstream VPN connection and CLI allows user to interact with it. Daemon and CLI communicate over local UNIX socket.

Go 1.18 or later required.


## Supported VPN providers:
- Mullvad
