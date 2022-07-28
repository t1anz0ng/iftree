package main

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/TianZong48/iftree/pkg/netutil"
)

type Pair struct {
	Veth   string
	Peer   string
	PeerId int

	NetNsID   int
	NetNsName string
}

var (
	debug = pflag.BoolP("debug", "d", false, "print debug message")
)

func main() {
	pflag.Parse()
	log.SetLevel(log.ErrorLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	netNsMap, err := netutil.NetNsMap()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("net namespace id <-> name map:\n%+v\n", netNsMap)
	ll, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}
	vm := make(map[string][]Pair)

	for _, link := range ll {
		if veth, ok := link.(*netlink.Veth); ok {
			master, err := netlink.LinkByIndex(veth.Attrs().MasterIndex)
			if err != nil {
				log.Fatal(err)
			}

			orign, _ := netns.Get()
			defer orign.Close()

			v, ok := vm[master.Attrs().Name]
			if !ok {
				vm[master.Attrs().Name] = []Pair{}
			}
			peerIdx, err := netlink.VethPeerIndex(veth)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugf("%s %d", veth.Name, peerIdx)

			peerNetNs, ok := netNsMap[veth.NetNsID]
			if !ok {
				log.Debugf("can't found namespace name for %d", veth.NetNsID)
				continue
			}
			hd, err := netns.GetFromName(peerNetNs)
			if err != nil {
				log.Fatal(err)
			}
			log.Debug(hd)
			if err := netns.Set(hd); err != nil {
				log.Fatalf("can't set current netns to %s, err: %s", peerNetNs, err)
			}
			peerInNs, err := netlink.LinkByIndex(peerIdx)
			if err != nil {
				log.Fatal(err)
			}
			// Switch back to the original namespace
			if err := netns.Set(orign); err != nil {
				log.Fatal(err)
			}
			vm[master.Attrs().Name] = append(v, Pair{
				Veth:      veth.Name,
				Peer:      peerInNs.Attrs().Name,
				PeerId:    peerIdx,
				NetNsID:   veth.NetNsID,
				NetNsName: peerNetNs})
		}
	}
	for k, v := range vm {
		w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)

		master, err := netlink.LinkByName(k)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(w, "----------------------------------------------------")
		fmt.Fprintf(w, "BRIDGE: %s\t%s\n", k, master.Attrs().OperState)
		fmt.Fprintf(w, "netnsName\tveth\tpeer\tnetnsID\n")
		for _, nsName := range netNsMap {

			f := false
			for _, p := range v {
				if nsName == p.NetNsName {
					if !f {
						fmt.Fprintf(w, "|____%s\n", nsName)
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
