package formatter

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/list"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/TianZong48/iftree/pkg"
)

func Print(w io.Writer, vm map[string][]pkg.Pair, netNsMap map[int]string, vpairs []pkg.Pair) error {
	lw := list.NewWriter()
	lw.SetOutputMirror(w)
	for k, v := range vm {
		master, err := netlink.LinkByName(k)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Fprintln(w, "----------------------------------------------------")
		// lw.AppendItem(("BRIDGE: %s\t%s", k, master.Attrs().OperState))
		lw.AppendItem(fmt.Sprintf("%s\t%s",
			color.New(color.Bold, color.FgYellow).Sprint(k),
			color.New(color.Bold, color.FgYellow).Sprint((master.Attrs().OperState))))
		lw.Indent()
		for _, nsName := range netNsMap {
			f := false
			for _, p := range v {
				if nsName == p.NetNsName {
					if !f {
						lw.AppendItem(color.New(color.FgHiMagenta).Sprint(nsName))
						f = true
						lw.Indent()
					}
					lw.AppendItem(fmt.Sprintf("%s\t%s",
						color.New(color.FgBlue).Sprint(p.Veth),
						color.New(color.FgBlue).Sprint(p.PeerInNetns)))
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

	fmt.Fprintln(w, "----------------------------------------------------")
	fmt.Fprintln(w, "unused veth pairs")
	lw.Reset()
	lw.SetStyle(list.StyleConnectedRounded)
	for _, veth := range vpairs {
		lw.AppendItem(fmt.Sprintf("%s <----> %s", veth.Veth, veth.Peer))
	}
	lw.Render()

	return nil
}
