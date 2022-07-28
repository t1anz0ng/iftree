package main

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"

	"github.com/containerd/nerdctl/pkg/rootlessutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/TianZong48/iftree/pkg/netutil"
)

type Pair struct {
	Veth        string
	Peer        string
	PeerInNetns string
	PeerId      int

	NetNsID   int
	NetNsName string
}

var (
	debug = pflag.BoolP("debug", "d", false, "print debug message")
)

func main() {
	pflag.Parse()
	if rootlessutil.IsRootless() {
		log.Error("iftree must be run as root to enter ns")
		os.Exit(1)
	}
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
	// master link
	vm := make(map[string][]Pair)
	vpairs := []Pair{}

	for _, link := range ll {
		if veth, ok := link.(*netlink.Veth); ok {

			orign, _ := netns.Get()
			defer orign.Close()

			peerIdx, err := netlink.VethPeerIndex(veth)
			if err != nil {
				log.Fatal(err)
			}

			peer, err := netlink.LinkByIndex(peerIdx)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugf("found pair: %s <-> %s", veth.Name, peer.Attrs().Name)
			peerNetNs, ok := netNsMap[veth.NetNsID]
			if !ok || veth.MasterIndex == 0 {
				p := Pair{
					Veth:    veth.Name,
					Peer:    peer.Attrs().Name,
					PeerId:  peerIdx,
					NetNsID: veth.NetNsID}
				vpairs = append(vpairs, p)
				log.Debugf("unused veth device %+v\n", p)
				continue
			}
			hd, err := netns.GetFromName(peerNetNs)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugf("peer netns fd: %d\n", hd)
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

			master, err := netlink.LinkByIndex(veth.Attrs().MasterIndex)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugf("%s\t%+v\n", veth.Name, *master.Attrs())
			// if master is not bridge
			if _, ok := master.(*netlink.Bridge); ok {
				v, ok := vm[master.Attrs().Name]
				if !ok {
					vm[master.Attrs().Name] = []Pair{}
				}
				vm[master.Attrs().Name] = append(v, Pair{
					Veth:        veth.Name,
					Peer:        peer.Attrs().Name,
					PeerInNetns: peerInNs.Attrs().Name,
					PeerId:      peerIdx,
					NetNsID:     veth.NetNsID,
					NetNsName:   peerNetNs})
			}
		}

	}
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)

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
					fmt.Fprintf(w, "     |____%s\t%s\t%s\t%d\n", p.Veth, p.Peer, p.PeerInNetns, p.NetNsID)
				}
			}
		}
		fmt.Fprintf(w, "\n")
		w.Flush()
	}
	fmt.Fprintln(w, "----------------------------------------------------")
	fmt.Fprintln(w, "unused veth pair without ")
	fmt.Fprintf(w, "veth\tpeer\tnetnsID\n")
	for _, p := range vpairs {
		fmt.Fprintf(w, "%s\t%s\t%d\n", p.Veth, p.Peer, p.NetNsID)
	}
	w.Flush()
}
