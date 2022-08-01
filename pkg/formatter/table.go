package formatter

import (
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/TianZong48/iftree/pkg"
)

func Table(w io.Writer, m map[string][]pkg.Pair, vpairs []pkg.Pair) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetTitle("bridge <---> veth <---> veth-in container, GROUP BY NetNS")
	t.AppendHeader(table.Row{"bridge", "netns", "veth", "ifname(container)"})
	for bridge, v := range m {
		for _, vp := range v {
			t.AppendRow(table.Row{bridge, vp.NetNsName, vp.Veth, vp.PeerInNetns})
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

	fmt.Fprintln(w, "") //nolint:errcheck

	t2 := table.NewWriter()
	t2.SetOutputMirror(w)
	t2.SetTitle("unused veth pairs (experimental)")
	t2.AppendHeader(table.Row{"veth", "pair"})
	for _, v := range vpairs {
		t2.AppendRow(table.Row{v.Veth, v.Peer})
		t2.AppendSeparator()
	}
	t2.SetAutoIndex(true)
	t2.SetStyle(table.StyleRounded)
	t2.Render()

	return nil
}
