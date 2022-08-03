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

func Print(w io.Writer, vm map[string][]pkg.Node, netNsMap map[int]string, vpairs []pkg.Node) error {

	var content strings.Builder
	var vpair strings.Builder

	lw := list.NewWriter()
	lw.SetOutputMirror(&content)
	fmt.Fprintln(&content, lipgloss.NewStyle().
		Background(lipgloss.Color("#F25D94")).
		MarginLeft(10).
		MarginRight(10).
		Padding(0, 1).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB")).
		SetString("Bridge <---->veth pair"))
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
							fmt.Sprintf("%s\t%s", p.Veth, p.PeerNameInNetns)).String())
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

	lw.Reset()
	lw.SetOutputMirror(&vpair)

	fmt.Fprintln(&vpair, lipgloss.NewStyle().
		Background(lipgloss.Color("#F25D94")).
		MarginLeft(5).
		MarginRight(5).
		Padding(0, 1).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB")).
		SetString("unused veth pairs"))
	lw.SetStyle(list.StyleConnectedRounded)
	for _, veth := range vpairs {
		lw.AppendItem(fmt.Sprintf("%s%s%s", veth.Veth,
			lipgloss.NewStyle().
				MarginRight(1).
				MarginLeft(1).
				Padding(0, 1).
				Foreground(special).SetString("<----->"),
			veth.Peer))
	}
	lw.Render()
	fmt.Fprintln(w,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			mainStype.Render(content.String()),
			vethPairStyle.Render(vpair.String())))
	return nil
}
