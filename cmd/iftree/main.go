package main

import (
	"errors"
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
	"github.com/TianZong48/iftree/pkg/netutil"
)

var (
	debug    = pflag.BoolP("debug", "d", false, "print debug message")
	oGraph   = pflag.BoolP("graph", "g", false, "output in graphviz dot language(https://graphviz.org/doc/info/lang.html")
	oTable   = pflag.BoolP("table", "t", false, "output in table")
	richText = pflag.BoolP("plain", "", true, "output colorful text, default is true")
)

func init() {
	pflag.Usage = func() {
		fmt.Println(`Usage:
  iftree [options]
    -d, --debug   print debug message
    -g, --graph   output in graphviz dot language(https://graphviz.org/doc/info/lang.html
    -t, --table   output in table
Help Options:
    -h, --help       Show this help message`)
	}
	pflag.ErrHelp = errors.New(``)
}

func helper() error {
	if *oGraph && *oTable {
		return fmt.Errorf(`only one of "graph", or "table" can be set`)
	}
	_ = richText
	return nil
}
func main() {
	pflag.Parse()
	if err := helper(); err != nil {
		log.Fatal(err)
	}
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
	// map bridge <-> veth paris
	vm := make(map[string][]pkg.Pair)
	// unused veth paris
	vpairs := []pkg.Pair{}
	// bridge ip
	bm := make(map[string]*net.IP)

	for _, link := range ll {
		veth, ok := link.(*netlink.Veth)
		if ok { // skip device not enslaved to any bridge

			origin, _ := netns.Get()
			defer origin.Close()

			peerIdx, err := netlink.VethPeerIndex(veth)
			if err != nil {
				log.Fatal(err)
			}
			if link.Attrs().MasterIndex == -1 || veth.MasterIndex == 0 {
				if veth.PeerName == "" {
					p, err := netlink.LinkByIndex(peerIdx)
					if err != nil {
						log.Fatal(err)
					}
					veth.PeerName = p.Attrs().Name
				}
				p := pkg.Pair{
					Veth:    veth.Name,
					Peer:    veth.PeerName,
					PeerId:  peerIdx,
					NetNsID: veth.NetNsID}
				vpairs = append(vpairs, p)
				continue
			}

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
	if *oGraph {
		output, err := formatter.Graph(vm, vpairs, bm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(os.Stdout, output)
		return
	}
	if *oTable {
		err := formatter.Table(os.Stdout, vm, vpairs)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 4, ' ', 0)
	if err := formatter.Print(w, vm, netNsMap, vpairs); err != nil {
		log.Fatal(err)
	}
	w.Flush()

}
