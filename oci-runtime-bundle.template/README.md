# Example OCI bundle for FILEgrain

Usage:
- Run `sudo ./prepare.sh` to create directories (`rootfs`, `volumes/{root, home}`).
- Run `sudo filegrain mount /tmp/your-filegrain-image $(pwd)/rootfs` to mount `rootfs`.
- Run `sudo runc run foo` to run a container.

Differences from the runc v1.0.0-rc3 default `config.json`:
- netns: host
- hostname: "filegrain"
- tmpfs: `/tmp`, `/run`. and `/var/log`
- bind-mount: `/etc/{hosts, hostname, resolv.conf}` and `/tmp/.X11-unix`
- persistent volumes: `volumes/{root, home}` on `/root` and `/home`
