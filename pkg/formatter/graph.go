package formatter

import (
	"fmt"
	"net"
	"os"
	"strings"

	graph "github.com/awalterschulze/gographviz"
	graphviz "github.com/goccy/go-graphviz"
	"github.com/pkg/errors"

	"github.com/t1anz0ng/iftree/pkg/types"
)

const (
	label = "label"
)

var (
	nodeAttr = map[string]string{
		"style": "filled",
	}
	netNsAttr = map[string]string{
		"style":   "filled",
		"color":   "grey",
		"nodesep": "4.0",
		"shape":   "box",
	}
	bridgeAttr = map[string]string{
		"nodesep":  "4.0",
		"shape":    "octagon",
		"style":    "filled",
		"fontsize": "16pt",
	}
	loAttr = map[string]string{
		"shape": "oval",
		"style": "filled",
		"color": "#f0c674",
	}
	edgeAttr = map[string]string{
		"color": "black",
	}
	redEdgeAttr = map[string]string{
		"color":     "red",
		"fontcolor": "red",
	}
	inNetNsAttr = map[string]string{
		"shape": "oval",
		"style": "filled",
		"color": "#f0c674",
	}
)

func GraphInDOT(brVethsM map[string][]types.Node, vpairs, los []types.Node, bridgeIps map[string]*net.IP) (string, error) {

	graphName := "G"
	root := graph.NewEscape()
	if err := root.SetName(graphName); err != nil {
		return "", err
	}
	root.AddAttr(graphName, "layout", "fdp")    //nolint:errcheck
	root.AddAttr(graphName, "splines", "ortho") //nolint:errcheck
	root.AddAttr(graphName, "ratio", "0.7")     //nolint:errcheck

	subGraphM := make(map[string]*graph.SubGraph)

	for bridge, v := range brVethsM {
		labels := []string{bridge}
		if ip, ok := bridgeIps[bridge]; ok {
			labels = append(labels, ip.String())
		}
		attr := generateAttr(bridgeAttr, strings.Join(labels, "\\n"))

		if err := root.AddNode(graphName, bridge, attr); err != nil {
			return "", errors.Wrapf(err, "create bridge node %s", bridge)
		}
		for i, vp := range v {
			// group by vp.NetNsName
			if vp.NetNsName != "" {
				sub, ok := subGraphM[vp.NetNsName]
				if !ok {
					// init subgraph for netns
					sub = graph.NewSubGraph(fmt.Sprintf("cluster-%s%c", bridge, 'A'+i))
					subGraphM[vp.NetNsName] = sub

					err := root.AddSubGraph(graphName, sub.Name, generateAttr(netNsAttr, fmt.Sprintf("NetNS\n%s", vp.NetNsName)))
					if err != nil {
						return "", errors.Wrapf(err, "create sub graph [%s] from %+v", sub, vp)
					}
				}
				// host veth
				err := root.AddNode(graphName, vp.Veth, generateAttr(nodeAttr, vp.Label()))
				if err != nil {
					return "", errors.Wrapf(err, "create veth node [%s]", vp.Veth)
				}
				// bridege <-> veth edge
				err = root.AddEdge(vp.Veth, bridge, false, edgeAttr)
				if err != nil {
					return "", errors.Wrapf(err, "create edge between [%s] and [%s]", vp.Veth, bridge)
				}

				// netns veth
				vethInNsName := fmt.Sprintf("%s_%d", bridge, i)
				err = root.AddNode(sub.Name, vethInNsName, generateAttr(inNetNsAttr, vp.PeerNameInNetns))
				if err != nil {
					return "", errors.Wrapf(err, "create veth node [%s]", vethInNsName)
				}
				// veths edge
				err = root.AddEdge(vp.Veth, vethInNsName, false, redEdgeAttr)
				if err != nil {
					return "", errors.Wrapf(err, "create edge between [%s] and [%s]", vp.Veth, vethInNsName)
				}
			} else {
				attr := map[string]string{
					"label": vp.Veth,
				}
				if vp.Orphaned {
					attr["label"] += "\n(orphaned)"
				}
				if err := root.AddNode(graphName, vp.Veth, attr); err != nil {
					return "", errors.Wrapf(err, "create veth node [%s]", vp.Veth)
				}
				if err := root.AddEdge(vp.Veth, bridge, false, edgeAttr); err != nil {
					return "", errors.Wrapf(err, "create edge between [%s] and [%s]", vp.Veth, bridge)
				}
			}
		}
	}

	// loopbacks
	for _, lo := range los {
		if lo.Status == "" {
			continue
		}
		if sub, ok := subGraphM[lo.NetNsName]; ok {
			err := root.AddNode(sub.Name, fmt.Sprintf("%s-lo", sub.Name), generateAttr(loAttr, lo.Label()))
			if err != nil {
				return "", errors.Wrapf(err, "create loopback in netns [%s]", lo.NetNsName)
			}
		}
	}

	visited := make(map[string]struct{})
	for _, vp := range vpairs {
		if _, ok := visited[vp.Veth]; !ok {
			root.AddNode(graphName, vp.Veth, generateAttr(nodeAttr, vp.Veth)) //nolint:errcheck
			visited[vp.Veth] = struct{}{}
		}
		if _, ok := visited[vp.Peer]; !ok {
			root.AddNode(graphName, vp.Peer, generateAttr(nodeAttr, vp.Peer)) //nolint:errcheck
			visited[vp.Peer] = struct{}{}
		}
		root.AddEdge(vp.Veth, vp.Peer, false, generateAttr(edgeAttr, "")) //nolint:errcheck
	}
	return root.String(), nil
}

func GenImage(data []byte, oGraphName *string, gType string) (fn string, err error) {

	g := graphviz.New()

	graph, err := graphviz.ParseBytes(data)
	if err != nil {
		return "", errors.Wrap(err, "parse dot bytes")
	}
	fn = fmt.Sprintf("%s.%s", *oGraphName, gType)
	f, err := os.Create(fn)
	if err != nil {
		return fn, errors.Wrapf(err, "create file `%s`", fn)
	}
	defer f.Close()

	switch gType {
	case "jpg":
		err = g.Render(graph, graphviz.JPG, f)
	case "png":
		err = g.Render(graph, graphviz.PNG, f)
	case "svg":
		err = g.Render(graph, graphviz.SVG, f)
	default:
		return fn, fmt.Errorf("unknown graph type %s", gType)
	}
	if err != nil {
		return fn, errors.Wrap(err, "render image")
	}
	return fn, nil
}

func generateAttr(base map[string]string, lb string) map[string]string {
	m := make(map[string]string)
	for k, v := range base {
		m[k] = v
	}
	if lb != "" {
		m[label] = lb
	}
	return m
}
