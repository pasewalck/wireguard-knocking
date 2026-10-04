# WireGuard Knocking

Basically a daemon that watches a WireGuard interface. It executes commands for peers connecting or losing connection. This is simply done by polling `wg show <someiterface> endpoints`, thus deriving which peers are connected via which IPs. This is inspired by the way [knockd](https://manpages.debian.org/stretch/knockd/knockd.1.en.html) works, essentially building on the idea but instead of listening to port knocking it listens to specific WireGuard connections.

This software is designed primarily an unconventional means of implicit authorization. A WireGuard peer is treated as an authorization signal to perform a task – e.g. open some ports or forward traffic for the given address.