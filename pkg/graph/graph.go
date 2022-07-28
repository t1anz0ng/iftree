package graph

import (
	"fmt"

	"github.com/awalterschulze/gographviz"

	"github.com/TianZong48/iftree/pkg"
)

func GenerateGraph(m map[string][]pkg.Pair) (string, error) {

	root := gographviz.NewEscape()
	root.SetName("G")

	for bridge, v := range m {
		root.AddNode("G", bridge, map[string]string{
			"nodesep": "4.0",
		})
		m := make(map[string]*gographviz.SubGraph)
		for i, vp := range v {
			// group by vp.NetNsName
			sub, ok := m[vp.NetNsName]
			if !ok {
				sub = gographviz.NewSubGraph(fmt.Sprintf("cluster%s%c", bridge, 'A'+i))
				m[vp.NetNsName] = sub
				_ = root.AddSubGraph("G", sub.Name,
					map[string]string{
						"label":   "NetNS: " + vp.NetNsName,
						"style":   "filled",
						"nodesep": "4.0",
					})
			}
			if err := root.AddNode("G", vp.Veth, nil); err != nil {
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
				"label":     vp.Peer,
				"color":     "red",
				"fontcolor": "red",
			}); err != nil {
				return "", err
			}
		}
	}

	return root.String(), nil
}
