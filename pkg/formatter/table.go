package formatter

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/TianZong48/iftree/pkg"
)

func Table(w io.Writer, m map[string][]pkg.Pair) error {
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

	return nil
}

func TableParis(w io.WriteCloser, vpairs []pkg.Pair) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetTitle("unused veth pairs (experimental)")
	t.AppendHeader(table.Row{"veth", "pair"})
	for _, v := range vpairs {
		t.AppendRow(table.Row{v.Veth, v.Peer})
		t.AppendSeparator()
	}
	t.SetAutoIndex(true)
	t.SetStyle(table.StyleRounded)
	t.Render()
	return nil
}
