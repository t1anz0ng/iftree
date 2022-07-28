# iftree

CLI, easy way to illustrate local network interface.

The intent is for understanding container networks :D

```
# sudo go run main.go

----------------------------------------------------
BRIDGE: cni_bridge0    up
netnsName              veth    peer    netnsID
|____123456
     |____veth57e09f05    eth13    0

----------------------------------------------------
BRIDGE: cni_br    up
netnsName         veth    peer    netnsID
|____123
     |____veth5e41415a    eth1    2
     |____veth90c9f5fa    eth2    2
     |____veth385ac3bb    eth3    2
|____321
     |____veth6328d76d    eth1    3
```

### roadmap

- [x] show peer name in container
- [ ] rich text
- [ ] topo relation in ascii graph
- [ ] support more networking device