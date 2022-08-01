package formatter

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/TianZong48/iftree/pkg"
)

func Table(w io.Writer, m map[string][]pkg.Pair, vpairs []pkg.Pair) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
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
	t.SetStyle(table.StyleBold)
	t.Render()

	// t2 := gtable.NewWriter()
	// t2.SetOutputMirror(w)
	return nil
}
