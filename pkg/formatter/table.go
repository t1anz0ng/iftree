package formatter

import (
	"fmt"
	"io"

	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/t1anz0ng/iftree/pkg/netutil"
	"github.com/t1anz0ng/iftree/pkg/types"
)

var (
	tableNsColors = colorGrid(1, 5)
)

func Table(w io.Writer, m map[string][]types.Node) error {

	tbStr := strings.Builder{}
	t := table.NewWriter()
	t.SetOutputMirror(&tbStr)

	t.SetTitle(boldItalicTextStyle.SetString("bridge <---> veth <---> veth in container, GROUP BY NetNS").String())
	t.AppendHeader(table.Row{"bridge", "netns", "veth", "ifname"})

	for bridge, v := range m {
		for _, vp := range v {
			if vp.NetNsName == "" {
				logrus.Warnf("no netns name for %+v", vp)
				continue
			}
			id, err := netutil.NsidFromPath(vp.NetNsName)
			if err != nil {
				return errors.Wrapf(err, "get nsid from path `%s`", vp.NetNsName)
			}
			c := lipgloss.Color(tableNsColors[id%len(tableNsColors)][0])
			t.AppendRow(table.Row{
				bridgeStyle.SetString(bridge),
				netNsStyle.Copy().Foreground(c).SetString(vp.NetNsName),
				basicTextStyle.SetString(vp.Veth),
				basicTextStyle.SetString(vp.PeerNameInNetns)})
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

func TableParis(w io.Writer, vpairs []types.Node) {

	if len(vpairs) == 0 {
		return
	}
	tbStr := strings.Builder{}
	t := table.NewWriter()
	t.SetOutputMirror(&tbStr)

	// (experimental)
	t.SetTitle("not bridged veth pairs")
	t.AppendHeader(table.Row{"veth", "route", "pair(In NetNS)"})
	visited := make(map[string]struct{})
	for _, v := range vpairs {
		h := hashVethpair(v.Veth, v.Peer)
		if _, ok := visited[h]; ok {
			continue
		}
		t.AppendRow(table.Row{
			basicTextStyle.SetString(v.Veth),
			basicTextStyle.SetString(v.Route.String()),
			basicTextStyle.SetString(v.Peer),
		})
		t.AppendSeparator()
		visited[h] = struct{}{}
	}
	t.SetAutoIndex(true)
	t.SetStyle(table.StyleRounded)
	t.Render()
	fmt.Fprintln(w, tbStr.String())
}

func hashVethpair(a, b string) string {
	// FIXME, replace with a better hash function
	res := []rune(a + b)
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return string(res)
}
