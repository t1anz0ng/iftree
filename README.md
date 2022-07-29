[![golangci-lint](https://github.com/TianZong48/iftree/actions/workflows/golangci-lint.yml/badge.svg?branch=main)](https://github.com/TianZong48/iftree/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/TianZong48/iftree/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/TianZong48/iftree/actions/workflows/codeql-analysis.yml)

# iftree

CLI, easy way to illustrate local network interface.

The intent is for understanding container networks :D

![networ-devices](./sample.jpg)

## usage

### text

```
# sudo go run main.go

----------------------------------------------------
BRIDGE: br0    up
netnsName      veth    peerInNetns    netnsID
└────/var/run/netns/netns0
     ├────veth0    ceth0    4

----------------------------------------------------
BRIDGE: docker0    up
netnsName          veth    peerInNetns    netnsID
└────/var/run/docker/netns/883628ab52b7
     ├────veth4f13cd2    eth0    5

----------------------------------------------------
BRIDGE: cni_bridge0    up
netnsName              veth    peerInNetns    netnsID
└────/var/run/netns/123456
     ├────veth57e09f05    eth13    0
└────/var/run/docker/netns/0de88faa84ac
     ├────veth31bc095b    eth0    1
     ├────veth12d98148    eth1    1

----------------------------------------------------
BRIDGE: cni_br    up
netnsName         veth    peerInNetns    netnsID
└────/var/run/netns/321
     ├────veth6328d76d    eth1    3
└────/var/run/netns/123
     ├────veth5e41415a    eth1    2
     ├────veth90c9f5fa    eth2    2
     ├────veth385ac3bb    eth3    2

----------------------------------------------------
unused veth pairs
VETH        PEER        NETNSID
veth-tt1    veth-tt     -1
veth-tt     veth-tt1    -1
```

### graph

Create an ouput image with [graphviz](https://www.graphviz.org/) compatible renderer.
e.g: online editor: https://dreampuf.github.io/GraphvizOnline

```
# sudo go run cmd/iftree/main.go --graph 
```

generate image using `dot`(http://www.graphviz.org/download/#executable-packages)

```
# sudo go run cmd/iftree/main.go --graph | dot -Tpng  > output.png
```

---

### roadmap

- [x] show peer name in container
- [x] graphviz
- [ ] rich text
- [ ] topo relation in ascii graph
- [ ] support more networking device
