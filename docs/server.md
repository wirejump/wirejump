# Server setup

## Prerequisites

- Linux machine running Debian-based distro. This has been tested on Debian 10, Ubuntu 20.04 LTS and 22.04 LTS. Ubuntu 23.10 has some weird network bug and does not work properly. Raspian may work okay, but wasn't tested.
- Ansible 2.12+. Other versions should work without any issue.

## Performance considerations

Using generic $5/month VPS instance (1 vCPU core, 1GB RAM) it's possible to reach ~500 Mb/s download, two cores and more will certainly allow gigabit+ speeds. Keep in mind that usually your Internet connection is the limiting factor. Most home routers can easily saturate 100 Mb/s links. 

There's no hard RAM requirement for the server, but 1GB is probably more than enough for your home network. Real world RAM usage for 10 users is ~200 MB for the whole system. 

## Setup

1. Clone this repo locally.
2. Skim through the comments in `ansible/playbook.yml` file: network options are specified there. They surely can be adjusted after installation, but it's much easier to do at this stage.
3. If you want to compile connection manager by yourself, you'll need Go 1.18 or later. Go to `wirejump` directory and run `make`, it should produce two binaries in the `build` folder. Another option is to use provided Docker file:
```
cd <repo folder> && docker build --output build .
```

Ansible will check for binaries in `build` folder before proceeding and will grab latest GitHub available release automatically if the folder is empty.

4. Run Ansible targeting your server machine:

```
$ ansible-galaxy collection install ansible.utils community.crypto
$ ansible-playbook -i user@your.server.address, ansible/playbook.yml
```

If you need to use a dedicated keyfile, this can be done like so:
```
$ ansible-playbook --private-key your_keyfile -i user@your.server.address, ansible/playbook.yml 
```

5. Verify that the server is up and running:
```
$ ssh user@your.server.address
$ wjcli version
```

This should print server software version.

## Troubleshooting

Sometimes Ansible can fail because playbook options are incompatible or because some local files are not found. Very rarely apt cannot lock files and installation dies early. It's safe to run the same Ansible command again, as all steps are only designed to run once and will be skipped if they are already performed (though, files will be copied again). Another option is to just recreate VM and start from the beginning. Whole installation takes about 5 mins.

Don't be ashamed to open a GitHub issue!
