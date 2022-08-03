package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/list"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/t1anz0ng/iftree/pkg"
)

func Print(w io.Writer, vm map[string][]pkg.Node, netNsMap map[int]string, vpairs []pkg.Node, all bool) {

	var content strings.Builder
	var vpair strings.Builder

	lw := list.NewWriter()
	lw.SetOutputMirror(&content)
	fmt.Fprintln(&content, TitleHighlight.SetString("Bridge <---->veth pair"))
	for k, v := range vm {
		master, err := netlink.LinkByName(k)
		if err != nil {
			log.Fatal(err)
		}
		lw.AppendItem(
			bridgeStyle.SetString(
				fmt.Sprintf("%s\t%s", k, master.Attrs().OperState)).String())
		lw.Indent()
		for _, nsName := range netNsMap {
			f := false
			for _, p := range v {
				if nsName == p.NetNsName {
					if !f {
						lw.AppendItem(netNsStyle.SetString(nsName).String())
						f = true
						lw.Indent()
					}
					lw.AppendItem(
						vethStyle.SetString(
							fmt.Sprintf("%s\t%s",
								basicTextStyle.SetString(p.Veth),
								basicTextStyle.SetString(p.PeerNameInNetns))).String())
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

	var contents []string
	contents = append(contents, mainStype.Render(content.String()))
	if all {
		lw.Reset()
		lw.SetOutputMirror(&vpair)

		fmt.Fprintln(&vpair, unusedVethStyle.SetString("unused veth pairs"))
		lw.SetStyle(list.StyleConnectedLight)
		for _, veth := range vpairs {
			lw.AppendItem(fmt.Sprintf("%s%s%s",
				basicTextStyle.SetString(veth.Veth),
				textHighlight.SetString("<----->"),
				basicTextStyle.SetString(veth.Peer)))
		}
		lw.Render()
		contents = append(contents, vethPairStyle.Render(vpair.String()))
	}
	fmt.Fprintln(w, lipgloss.JoinVertical(lipgloss.Top, contents...))
}
