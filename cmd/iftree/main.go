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
		log.SetReportCaller(true)
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
	bm := make(map[string]*net.IP)
	for _, link := range ll {
		// log.Infof("index: %d\tname: %s\ttype: %s\tmasterid: %d", link.Attrs().Index, link.Attrs().Name, link.Type(), link.Attrs().MasterIndex)
		// if link.Attrs().Slave != nil {
		// 	log.Infof("\t%s", link.Attrs().Slave.SlaveType())
		// }
		// if veth, ok := link.(*netlink.Veth); ok {
		// 	peerIdx, _ := netlink.VethPeerIndex(veth)
		// 	log.Infof("veth, peer index %d, %d, %s", peerIdx, veth.NetNsID, netNsMap[veth.NetNsID])
		// 	peer, err := netlink.LinkByIndex(peerIdx)
		// 	if err != nil {
		// 		log.Warn(err)
		// 		continue
		// 	}
		// 	log.Infof("peer name: %s", peer.Attrs().Name)
		// }
		// log.Info("------------------------------")
		// continue
		veth, ok := link.(*netlink.Veth)
		if ok { // skip device not enslaved to any bridge

			origin, _ := netns.Get()
			defer origin.Close()

			peerIdx, err := netlink.VethPeerIndex(veth)
			if err != nil {
				log.Fatal(err)
			}
			if link.Attrs().MasterIndex == -1 || veth.MasterIndex == 0 {
				p := pkg.Pair{
					Veth:    veth.Name,
					Peer:    veth.PeerName,
					PeerId:  peerIdx,
					NetNsID: veth.NetNsID}
				vpairs = append(vpairs, p)
				continue
			}
			log.Debugf("%+v\n", veth)

			master, err := netlink.LinkByIndex(veth.Attrs().MasterIndex)
			if err != nil {
				log.Fatal(err)
			}

			// if master is not bridge
			if _, ok := master.(*netlink.Bridge); !ok {
				log.Debug("todo: not bridge")
			}
			bridge := master.Attrs().Name
			v, ok := vm[bridge]
			if !ok {
				vm[bridge] = []pkg.Pair{}
			}
			pair := pkg.Pair{
				Veth:    veth.Name,
				PeerId:  peerIdx,
				NetNsID: veth.NetNsID,
			}
			if peerNetNs, ok := netNsMap[veth.NetNsID]; ok {
				peerInNs, err := netutil.GetPeerInNs(peerNetNs, peerIdx, origin)
				if err != nil {
					log.Fatal(err)
				}
				pair.NetNsName = peerNetNs
				pair.PeerInNetns = peerInNs
			} else {
				pair.Orphaned = true
			}

			addrs, err := netlink.AddrList(master, syscall.AF_INET)
			if err != nil {
				log.Fatal(err)
			}
			if len(addrs) > 0 {
				pair.Master = &pkg.Bridge{
					Name: bridge,
					IP:   &addrs[0].IP,
				}
				bm[bridge] = &addrs[0].IP
			}
			vm[bridge] = append(v, pair)

		}

	}
	if *isGraph {
		output, err := graph.GenerateGraph(vm, vpairs, bm)
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
