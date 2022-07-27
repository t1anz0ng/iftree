# iftree

CLI, easy way to illustrate local network interface.

The intent is for understanding container networks :D

```
# go run main.go

{bridge}: cni_bridge0
netnsName    veth    peer    netnsID
|----123456
     |____veth57e09f05    enp5s0    0

{bridge}: cni_br
netnsName    veth    peer    netnsID
|----321
     |____veth6328d76d    enp5s0    3
|----123
     |____veth5e41415a    enp5s0     2
     |____veth90c9f5fa    wlp4s0     2
     |____veth385ac3bb    docker0    2
```