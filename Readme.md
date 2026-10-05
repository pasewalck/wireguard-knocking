# WireGuard Knocking

Basically a daemon that watches a WireGuard interface. It executes commands for peers connecting or losing connection. This is simply done by polling `wg show <someiterface> dumps`, thus deriving which peers are connected via which IPs. This is inspired by the way [knockd](https://manpages.debian.org/stretch/knockd/knockd.1.en.html) works, essentially building on the idea but instead of listening to port knocking it listens to specific WireGuard connections.

This software is designed primarily an unconventional means of implicit authorization. A WireGuard peer is treated as an authorization signal to perform a task – e.g. open some ports or forward traffic for the given address.

Read more on the [context of this software's creation](https://jpasewalck.org/2026/10/wireguarding-my-homelab-and-overcoming-androids-limitations/).

## Install

1. [Install go](https://go.dev/doc/install)
2. Clone this repo and open it: `git clone https://github.com/pasewalck/wireguard-knocking.git && cd wireguard-knocking`
3. Install: `sudo make install`
4. Configure: `sudo vi /etc/wireguard-knocking/config.toml`
5. Start Timed Service: `sudo make start`
6. Enable Timed Service: `sudo make enable`

## Uninstall

`sudo make uninstall`

## Example Usage

### Traffic Forwarding

An example use case is to use this to forward any traffic (e.g. WireGuard traffic on different Port to separate machine).

```
add_ip_cmds = [
    "sudo iptables -t nat -A PREROUTING -s <ip> -p udp --dport 51822 -j DNAT --to-destination 10.200.200.2:51822",
    "sudo iptables -t nat -A POSTROUTING -s <ip> -d 10.200.200.2 -p udp --dport 51822 -j MASQUERADE"
    ]
remove_ip_cmds = [
    "sudo iptables -t nat -D PREROUTING -s <ip> -p udp --dport 51822 -j DNAT --to-destination 10.200.200.2:51822",
    "sudo iptables -t nat -D POSTROUTING -s <ip> -d 10.200.200.2 -p udp --dport 51822 -j MASQUERADE"
    ]
```

To make this work also make sure to forward traffic using `sudo make enable-forwarding`.
