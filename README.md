

<div align="center">

# ☘️ iftree

`iftree` command visulize local network interfaces.

intent for better understanding container networks :D

[![golangci-lint](https://github.com/TianZong48/iftree/actions/workflows/golangci-lint.yml/badge.svg?branch=main)](https://github.com/TianZong48/iftree/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/TianZong48/iftree/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/TianZong48/iftree/actions/workflows/codeql-analysis.yml)
[![Go Report](https://goreportcard.com/badge/github.com/TianZong48/iftree)](https://goreportcard.com/badge/github.com/TianZong48/iftree)
[![Github All Releases](https://img.shields.io/github/downloads/t1anz0ng/iftree/total.svg)](https://img.shields.io/github/downloads/t1anz0ng/iftree/total.svg)
</div>

---

<img
  src="./asset/sample.jpg"
  alt="iftree --graph"
  width="60%"
  align="right"
/>

<img
  src="./asset/sample-term.png"
  alt="iftree"
  width="60%"
  align="right"
/>

**Features**

- [x] visualize Veth/bridge connections
- [x] support graphviz
- [x] table output
- [x] rich text
- [ ] ascii graph
- [ ] support more networking device

## usage

```
Usage:
  iftree [options]
    -d, --debug   print debug message
    -g, --graph   output in graphviz dot language(https://graphviz.org/doc/info/lang.html
    -t, --table   output in table
	--no-color    disable color output
Help Options:
    -h, --help       Show this help message
```

### text

```shell
sudo go run cmd/iftree/main.go
```

### graph

Create an ouput image with [graphviz](https://www.graphviz.org/) compatible renderer.
e.g: online editor: https://dreampuf.github.io/GraphvizOnline

```shell
sudo go run cmd/iftree/main.go --graph 
```

generate image using `dot`(http://www.graphviz.org/download/#executable-packages)

```shell
sudo go run cmd/iftree/main.go --graph | dot -Tpng  > output.png
```

### table

```shell
sudo iftree --table
```

![table](./asset/sample-table.png)

