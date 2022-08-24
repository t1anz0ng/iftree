

<div align="center">

# ☘️ iftree

`iftree` command visulize local network interfaces.

intent for better understanding container networks :D

[![golangci-lint](https://github.com/t1anz0ng/iftree/actions/workflows/golangci-lint.yml/badge.svg?branch=main)](https://github.com/t1anz0ng/iftree/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/t1anz0ng/iftree/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/t1anz0ng/iftree/actions/workflows/codeql-analysis.yml)
[![Go Report](https://goreportcard.com/badge/github.com/t1anz0ng/iftree)](https://goreportcard.com/badge/github.com/t1anz0ng/iftree)
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

- [x] **visualize** Veth/bridge connections
- [x] **table** output
- [x] **rich** text
- [x] rendering **image**
- [x] output **graphviz DOT** language


## usage

```
iftree [options]

Example:
  generate tree output
    # sudo iftree 
  generate png graph with name "output.png"
    # sudo iftree --graph -Tpng -Ooutput.png
  generate image with dot
    # sudo iftree --graph -Tdot | dot -Tpng  > output.png
  generate table output
    # sudo iftree --table
```

### text

```shell
sudo iftree
```

### graph

support `jpg`, `svg`, `png`

```shell
sudo iftree --graph -Tpng
```

Or create an ouput image with any [graphviz](https://www.graphviz.org/) compatible renderer.
e.g: online editor: https://dreampuf.github.io/GraphvizOnline

```shell
sudo iftree --graph -Tdot
```

generate image using `dot`(http://www.graphviz.org/download/#executable-packages)

```shell
sudo iftree --graph -Tdot | dot -Tpng  > output.png
```

### table

```shell
sudo iftree --table
```

![table](./asset/sample-table.png)

