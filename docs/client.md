# Client setup

## Overview

This document outlines client (your home router) setup. It's supposed that the server is already installed. Whole process is fairly straightforward:
- Creating the keys (optional)
- Setting up the interface
- Making connection primary

Essentially, you will be creating a WireGuard tunnel between your router (client) and your newly setup WireJump server, and then making some routing changes, so that this tunnel is used as the main Internet connection.

Before you begin: BACKUP YOUR CURRENT ROUTER SETTINGS! If you mess up your config, you can always go back and restore your Internet connection.

WireGuard setup can vary between routers, but overall principle is the same. If your guide is missing, you can always use MikroTik one as a base:

- [Linux](./linux.md)
- [MikroTik](./mikrotik.md)
- [OpenWRT](./openwrt.md)
