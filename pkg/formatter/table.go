package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/TianZong48/iftree/pkg"
	"github.com/TianZong48/iftree/pkg/netutil"
)

var (
	tableNsColors = colorGrid(1, 5)
)

func Table(w io.Writer, m map[string][]pkg.Node) error {
	tbStr := strings.Builder{}
	t := table.NewWriter()
	t.SetOutputMirror(&tbStr)
	t.SetTitle(lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Italic(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		SetString("bridge <---> veth <---> veth-in container, GROUP BY NetNS").String())
	t.AppendHeader(table.Row{"bridge", "netns", "veth", "ifname(container)"})

	for bridge, v := range m {
		for _, vp := range v {
			id, err := netutil.NsidFromPath(vp.NetNsName)
			if err != nil {
				return err
			}
			c := lipgloss.Color(tableNsColors[id%len(tableNsColors)][0])
			t.AppendRow(table.Row{
				lipgloss.NewStyle().
					Bold(true).
					Padding(0, 1).
					Italic(true).
					Foreground(lipgloss.Color("#FFFFFF")).SetString(bridge),
				lipgloss.NewStyle().Bold(true).Foreground(c).SetString(vp.NetNsName),
				vp.Veth,
				vp.PeerNameInNetns})
			t.AppendSeparator()
		}
	}
	t.SetAutoIndex(true)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
	})
	t.SetStyle(table.StyleRounded)
	t.Render()

	fmt.Fprintln(w, tableStype.Render(tbStr.String()))
	return nil
}

func TableParis(w io.WriteCloser, vpairs []pkg.Node) error {
	tbStr := strings.Builder{}
	t := table.NewWriter()
	t.SetOutputMirror(&tbStr)
	t.SetTitle("unused veth pairs (experimental)")
	t.AppendHeader(table.Row{"veth", "pair"})
	for _, v := range vpairs {
		t.AppendRow(table.Row{v.Veth, v.Peer})
		t.AppendSeparator()
	}
	t.SetAutoIndex(true)
	t.SetStyle(table.StyleRounded)
	t.Render()
	fmt.Fprintln(w, tbStr.String())
	return nil
}
