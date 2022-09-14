package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/t1anz0ng/iftree/pkg/types"
)

func Print(w io.Writer, vm map[string][]types.Node, netNsMap map[int]string, vpairs []types.Node, all bool) {

	var contents []string

	if len(vm) > 0 {
		var content strings.Builder

		lw := list.NewWriter()
		lw.SetOutputMirror(&content)
		fmt.Fprintln(&content, titleHighlight.SetString("Bridge <----> veth pair"))
		for k, v := range vm {
			master, err := netlink.LinkByName(k)
			if err != nil {
				logrus.Fatal(err)
			}
			lw.AppendItem(
				bridgeStyle.SetString(
					fmt.Sprintf("%s\t%s", k, master.Attrs().OperState)))
			lw.Indent()
			for _, nsName := range netNsMap {
				f := false
				for _, p := range v {
					if nsName == p.NetNsName {
						if !f {
							lw.AppendItem(netNsStyle.SetString(nsName))
							f = true
							lw.Indent()
						}

						lw.AppendItem(
							vethStyle.SetString(
								fmt.Sprintf("%s\t%s",
									basicTextStyle.SetString(p.Veth),
									basicTextStyle.SetString(p.PeerNameInNetns))))
					}
				}
				if f {
					lw.UnIndent()
				}
			}
			lw.UnIndent()
		}

		lw.SetStyle(list.StyleConnectedRounded)
		lw.Render()

		contents = append(contents, mainStype.Render(content.String()))
	}

	if all && len(vpairs) > 0 {
		var vpair strings.Builder
		visited := make(map[string]struct{})

		fmt.Fprintln(&vpair, titleHighlight.SetString("not bridged veth pairs"))

		for _, veth := range vpairs {
			h := hashVethpair(veth.Veth, veth.Peer)
			if _, ok := visited[h]; ok {
				continue
			}

			fmt.Fprintf(&vpair, "%s%s%s\t%s\t%s\n",
				basicTextStyle.SetString(veth.Veth),
				textHighlight.SetString("<-->"),
				basicTextStyle.SetString(veth.Peer),
				basicTextStyle.SetString(veth.Route.String()),
				netNsStyle.SetString(netNsMap[veth.NetNsID]),
			)
			visited[h] = struct{}{}
		}

		contents = append(contents, vethPairStyle.Render(vpair.String()))
	}

	if len(contents) > 0 {
		fmt.Fprintln(w, lipgloss.JoinVertical(lipgloss.Top, contents...))
	}
}
