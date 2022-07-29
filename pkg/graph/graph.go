package graph

import (
	"fmt"
	"net"
	"strings"

	"github.com/awalterschulze/gographviz"

	"github.com/TianZong48/iftree/pkg"
)

func GenerateGraph(m map[string][]pkg.Pair, vpairs []pkg.Pair, bm map[string]*net.IP) (string, error) {

	root := gographviz.NewEscape()
	if err := root.SetName("G"); err != nil {
		return "", err
	}

	for bridge, v := range m {
		labels := []string{bridge}
		if ip, ok := bm[bridge]; ok {
			labels = append(labels, ip.String())
		}
		attr := map[string]string{
			"label":   strings.Join(labels, "\\n"),
			"nodesep": "4.0",
		}
		if err := root.AddNode("G", bridge, attr); err != nil {
			return "", err
		}
		m := make(map[string]*gographviz.SubGraph)
		for i, vp := range v {
			// group by vp.NetNsName
			if vp.NetNsName != "" {
				sub, ok := m[vp.NetNsName]
				if !ok {
					// init subgraph for netns
					sub = gographviz.NewSubGraph(fmt.Sprintf("cluster%s%c", bridge, 'A'+i))
					m[vp.NetNsName] = sub
					attr := map[string]string{
						"label":   fmt.Sprintf("NetNS: %s", vp.NetNsName),
						"style":   "filled",
						"nodesep": "4.0",
					}

					if err := root.AddSubGraph("G", sub.Name, attr); err != nil {
						return "", err
					}
				}
				if err := root.AddNode("G", vp.Veth, map[string]string{
					"label": vp.Veth,
				}); err != nil {
					return "", err
				}
				if err := root.AddEdge(vp.Veth, bridge, false, map[string]string{
					"color": "black",
				}); err != nil {
					return "", err
				}
				vethInNsName := fmt.Sprintf("%s_%d", vp.PeerInNetns, i)
				if err := root.AddNode(sub.Name, vethInNsName, map[string]string{
					"label": vp.PeerInNetns,
				}); err != nil {
					return "", err
				}
				if err := root.AddEdge(vp.Veth, vethInNsName, false, map[string]string{
					"color":     "red",
					"fontcolor": "red",
				}); err != nil {
					return "", err
				}
			} else {
				attr := map[string]string{
					"label": vp.Veth,
				}
				if vp.Orphaned {
					attr["label"] += "\n(orphaned)"
				}
				if err := root.AddNode("G", vp.Veth, attr); err != nil {
					return "", err
				}
				if err := root.AddEdge(vp.Veth, bridge, false, map[string]string{
					"color": "black",
				}); err != nil {
					return "", err
				}
			}
		}
	}

	return root.String(), nil
}
