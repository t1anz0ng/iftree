package formatter

import (
	"fmt"
	"net"
	"strings"

	"github.com/awalterschulze/gographviz"

	"github.com/TianZong48/iftree/pkg"
)

func Graph(m map[string][]pkg.Pair, vpairs []pkg.Pair, bm map[string]*net.IP) (string, error) {

	root := gographviz.NewEscape()
	if err := root.SetName("G"); err != nil {
		return "", err
	}
	root.AddAttr("G", "layout", "fdp")    //nolint:errcheck
	root.AddAttr("G", "splines", "ortho") //nolint:errcheck

	for bridge, v := range m {
		labels := []string{bridge}
		if ip, ok := bm[bridge]; ok {
			labels = append(labels, ip.String())
		}
		attr := map[string]string{
			"label":   strings.Join(labels, "\\n"),
			"nodesep": "4.0",
			"shape":   "octagon",
			"style":   "filled",
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
						"label":   fmt.Sprintf("NetNS\n%s", vp.NetNsName),
						"style":   "filled",
						"color":   "grey",
						"nodesep": "4.0",
						"shape":   "box",
					}

					if err := root.AddSubGraph("G", sub.Name, attr); err != nil {
						return "", err
					}
				}
				if err := root.AddNode("G", vp.Veth, map[string]string{
					"label": vp.Veth,
					"style": "filled",
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
					"shape": "oval",
					"style": "filled",
					"color": "#f0c674",
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

	visited := make(map[string]struct{})
	for _, vp := range vpairs {
		if _, ok := visited[vp.Veth]; !ok {
			root.AddNode("G", vp.Veth, //nolint:errcheck
				map[string]string{
					"label": vp.Veth,
					"style": "filled",
				})
			visited[vp.Veth] = struct{}{}
		}
		if _, ok := visited[vp.Peer]; !ok {
			root.AddNode("G", vp.Peer, //nolint:errcheck
				map[string]string{
					"label": vp.Peer,
					"style": "filled",
				})
			visited[vp.Peer] = struct{}{}
		}
		root.AddEdge(vp.Veth, vp.Peer, false, //nolint:errcheck
			map[string]string{
				"color": "black",
			})
	}

	return root.String(), nil
}
