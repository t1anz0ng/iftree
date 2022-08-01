

<div align="center">

# ☘️ iftree

`iftree` command visulize local network interfaces.

intent for better understanding container networks :D

[![golangci-lint](https://github.com/TianZong48/iftree/actions/workflows/golangci-lint.yml/badge.svg?branch=main)](https://github.com/TianZong48/iftree/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/TianZong48/iftree/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/TianZong48/iftree/actions/workflows/codeql-analysis.yml)
[![Go Report](https://goreportcard.com/badge/github.com/TianZong48/iftree)](https://goreportcard.com/badge/github.com/TianZong48/iftree)

</div>



![networ-devices](./sample.jpg)

### features

- [x] visualize Veth/bridge connections
- [x] support graphviz
- [x] table output
- [ ] rich text
- [ ] ascii graph
- [ ] support more networking device

## usage

```
Usage:
  iftree [options]
    -d, --debug   print debug message
    -g, --graph   output in graphviz dot language(https://graphviz.org/doc/info/lang.html
    -t, --table   output in table
Help Options:
    -h, --help       Show this help message
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


### text

```
# sudo go run cmd/iftree/main.go

╭─ BRIDGE: br0    up
│  ╰─ /var/run/netns/netns0
│     ╰─ veth0    ceth0
├─ BRIDGE: docker0    up
│  ╰─ /var/run/docker/netns/415d70663520
│     ╰─ veth08e8cd7    eth0
├─ BRIDGE: cni_bridge0    up
│  ╰─ /var/run/netns/123456
│     ╰─ veth57e09f05    eth13
╰─ BRIDGE: cni_br    up
   ├─ /var/run/netns/321
   │  ╰─ veth6328d76d    eth1
   ├─ /var/run/docker/netns/415d70663520
   │  ╰─ veth319e1bda    eth22
   ╰─ /var/run/netns/123
      ├─ veth5e41415a    eth1
      ├─ veth90c9f5fa    eth2
      ╰─ veth385ac3bb    eth3
----------------------------------------------------
unused veth pairs
╭─ veth-tt1 <----> veth-tt
╰─ veth-tt <----> veth-tt1
```

### table

```
# sudo iftree --table
╭─────────────────────────────────────────────────────────────────────────────────────────╮
│ bridge <---> veth <---> veth-in container, GROUP BY NetNS                               │
├───┬─────────────┬────────────────────────────────────┬──────────────┬───────────────────┤
│   │ BRIDGE      │ NETNS                              │ VETH         │ IFNAME(CONTAINER) │
├───┼─────────────┼────────────────────────────────────┼──────────────┼───────────────────┤
│ 1 │ cni_bridge0 │ /var/run/netns/123456              │ veth57e09f05 │ eth13             │
├───┼─────────────┼────────────────────────────────────┼──────────────┼───────────────────┤
│ 2 │ cni_br      │ /var/run/netns/123                 │ veth5e41415a │ eth1              │
├───┤             │                                    ├──────────────┼───────────────────┤
│ 3 │             │                                    │ veth90c9f5fa │ eth2              │
├───┤             │                                    ├──────────────┼───────────────────┤
│ 4 │             │                                    │ veth385ac3bb │ eth3              │
├───┤             ├────────────────────────────────────┼──────────────┼───────────────────┤
│ 5 │             │ /var/run/netns/321                 │ veth6328d76d │ eth1              │
├───┤             ├────────────────────────────────────┼──────────────┼───────────────────┤
│ 6 │             │ /var/run/docker/netns/415d70663520 │ veth319e1bda │ eth22             │
├───┼─────────────┼────────────────────────────────────┼──────────────┼───────────────────┤
│ 7 │ br0         │ /var/run/netns/netns0              │ veth0        │ ceth0             │
├───┼─────────────┼────────────────────────────────────┼──────────────┼───────────────────┤
│ 8 │ docker0     │ /var/run/docker/netns/415d70663520 │ veth08e8cd7  │ eth0              │
╰───┴─────────────┴────────────────────────────────────┴──────────────┴───────────────────╯

╭─────────────────────────╮
│ unused veth pairs (expe │
│ rimental)               │
├───┬──────────┬──────────┤
│   │ VETH     │ PAIR     │
├───┼──────────┼──────────┤
│ 1 │ veth-tt1 │ veth-tt  │
├───┼──────────┼──────────┤
│ 2 │ veth-tt  │ veth-tt1 │
╰───┴──────────┴──────────╯
```
