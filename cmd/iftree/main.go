package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
	"text/tabwriter"

	"github.com/containerd/nerdctl/pkg/rootlessutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/TianZong48/iftree/pkg"
	"github.com/TianZong48/iftree/pkg/formatter"
	"github.com/TianZong48/iftree/pkg/graph"
	"github.com/TianZong48/iftree/pkg/netutil"
)

var (
	debug   = pflag.BoolP("debug", "d", false, "print debug message")
	isGraph = pflag.BoolP("graph", "g", false, "output in graphviz dot language(https://graphviz.org/doc/info/lang.html")
)

func main() {
	pflag.Parse()
	if rootlessutil.IsRootless() {
		log.Error("iftree must be run as root to enter ns")
		os.Exit(1)
	}
	log.SetLevel(log.InfoLevel)
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
	vm := make(map[string][]pkg.Pair)
	vpairs := []pkg.Pair{}

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
			peerNetNs, ok := netNsMap[veth.NetNsID]
			if !ok || veth.MasterIndex == 0 {
				p := pkg.Pair{
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

			// if master is not bridge
			if master, ok := master.(*netlink.Bridge); ok {
				v, ok := vm[master.Attrs().Name]
				if !ok {
					vm[master.Attrs().Name] = []pkg.Pair{}
				}

				pair := pkg.Pair{
					Veth:        veth.Name,
					Peer:        peer.Attrs().Name,
					PeerInNetns: peerInNs.Attrs().Name,
					PeerId:      peerIdx,
					NetNsID:     veth.NetNsID,
					NetNsName:   peerNetNs,
				}
				addrs, err := netlink.AddrList(master, syscall.AF_INET)
				if err != nil {
					log.Fatal(err)
				}
				if len(addrs) > 0 {
					pair.Master = &pkg.Bridge{
						Name: master.Name,
						IP:   []*net.IP{&addrs[0].IP},
					}
				}

				vm[master.Attrs().Name] = append(v, pair)
			}
		}

	}
	if *isGraph {
		output, err := graph.GenerateGraph(vm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(os.Stdout, output)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)
	if err := formatter.Print(w, vm, netNsMap, vpairs); err != nil {
		log.Fatal(err)
	}
	w.Flush()

}
