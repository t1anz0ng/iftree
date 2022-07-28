package main

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"

	"github.com/vishvananda/netlink"

	"github.com/TianZong48/iftree/pkg/netutil"
)

type Pair struct {
	Veth string
	Peer string

	NetNsID   int
	NetNsName string
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	netNsMap, err := netutil.NetNsMap()
	if err != nil {
		panic(err)
	}

	ll, err := netlink.LinkList()
	if err != nil {
		panic(err)
	}
	vm := make(map[string][]Pair)

	for _, link := range ll {
		if veth, ok := link.(*netlink.Veth); ok {
			master, err := netlink.LinkByIndex(veth.Attrs().MasterIndex)
			if err != nil {
				panic(err)
			}

			v, ok := vm[master.Attrs().Name]
			if !ok {
				vm[master.Attrs().Name] = []Pair{}
			}
			peerIdx, err := netlink.VethPeerIndex(veth)
			if err != nil {
				panic(err)
			}
			peer, err := netlink.LinkByIndex(peerIdx)
			if err != nil {
				panic(err)
			}

			vm[master.Attrs().Name] = append(v, Pair{
				Veth:      veth.Name,
				Peer:      peer.Attrs().Name,
				NetNsID:   veth.NetNsID,
				NetNsName: netNsMap[veth.NetNsID]})
		}
	}
	for k, v := range vm {
		w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)

		master, err := netlink.LinkByName(k)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "BRIDGE: %s\t%s\n", k, master.Attrs().OperState)
		fmt.Fprintf(w, "netnsName\tveth\tpeer\tnetnsID\n")
		for _, nsName := range netNsMap {

			f := false
			for _, p := range v {
				if nsName == p.NetNsName {
					if !f {
						fmt.Fprintf(w, "|----%s\n", nsName)
						f = true
					}
					fmt.Fprintf(w, "     |____%s\t%s\t%d\n", p.Veth, p.Peer, p.NetNsID)
				}
			}
			w.Flush()
		}

		fmt.Fprintf(w, "\n")
		w.Flush()
	}
}
