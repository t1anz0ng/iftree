package formatter

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/TianZong48/iftree/pkg"
)

func Print(w io.Writer, vm map[string][]pkg.Pair, netNsMap map[int]string, vpairs []pkg.Pair) error {
	for k, v := range vm {
		master, err := netlink.LinkByName(k)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(w, "----------------------------------------------------")
		fmt.Fprintf(w, "BRIDGE: %s\t%s\n", k, master.Attrs().OperState)
		fmt.Fprintf(w, "netnsName\tveth\tpeer\tpeerInNetns\tnetnsID\n")
		for _, nsName := range netNsMap {
			f := false
			for _, p := range v {
				if nsName == p.NetNsName {
					if !f {
						fmt.Fprintf(w, "|____%s\n", nsName)
						f = true
					}
					fmt.Fprintf(w, "     |----%s\t%s\t%s\t%d\n", p.Veth, p.Peer, p.PeerInNetns, p.NetNsID)
				}
			}
		}
		fmt.Fprintf(w, "\n")

	}
	fmt.Fprintln(w, "----------------------------------------------------")
	fmt.Fprintln(w, "unused veth pairs")
	fmt.Fprintf(w, "veth\tpeer\tnetnsID\n")
	for _, p := range vpairs {
		fmt.Fprintf(w, "%s\t%s\t%d\n", p.Veth, p.Peer, p.NetNsID)
	}

	return nil
}
