package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/containerd/nerdctl/pkg/rootlessutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/t1anz0ng/iftree/pkg"
	"github.com/t1anz0ng/iftree/pkg/formatter"
	"github.com/t1anz0ng/iftree/pkg/netutil"
)

var (
	debug = pflag.BoolP("debug", "d", false, "print debug message")

	oNotBridgedVeths = pflag.BoolP("all", "a", true, "show all veths, including not bridged.")
	oGraph           = pflag.BoolP("graph", "g", false, "output in png by defaul")
	oGraphType       = pflag.StringP("gtype", "T", "dot", `graph output type, "jpg", "png", "svg", "dot"(graphviz dot language(https://graphviz.org/doc/info/lang.html)`)
	oGraphName       = pflag.StringP("output", "O", "output", "graph output name/path")

	oTable = pflag.BoolP("table", "t", false, "output in table")

	help = pflag.BoolP("help", "h", false, "")

	version = "unknown"

	defaultOutput = os.Stdout
)

func init() {
	pflag.Usage = func() {
		fmt.Printf(`Usage:
iftree [options]

Example:
  generate tree output
    # sudo iftree 
  generate png graph with name "output.png"
    # sudo iftree --graph -Tpng -Ooutput.png
  generate table output
    # sudo iftree --table

  -d, --debug		print debug message
  -t, --table		output in table
  -g, --graph		output in graphviz dot language(https://graphviz.org/doc/info/lang.html
    -O, --output	string   graph output name/path (default "output")
    -T, --gtype		string    graph output type, "jpg", "png", "svg", "dot" (graphviz dot language) default "png"
  -a, --all	 	show all veths, including unused.
Help Options:
  -h, --help       Show this help message

version: %s
`, version)
	}
}

func helper() error {
	if *oGraph && *oTable {
		return fmt.Errorf(`only one of "graph", or "table" can be set`)
	}
	if *help {
		pflag.Usage()
		os.Exit(0)
	}
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
	if len(netNsMap) == 0 {
		log.Warn("no netns found")
		os.Exit(0)
	}
	log.Debugf("net namespace id <-> name map:\n%+v\n", netNsMap)

	ll, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("net link list:\n%+v\n", ll)

	bridgeVethM := make(map[string][]pkg.Node) // map bridge <-> veth paris
	unBridgedVpairs := []pkg.Node{}
	bridgeIps := make(map[string]*net.IP) // bridge ip
	loS := []pkg.Node{}                   // loopback

	origin, err := netns.Get()
	if err != nil {
		log.Fatalf("failed get current netne, %v", err)
	}
	// FIXME: use undirected graph insted of array/map to store relations
	for _, link := range ll {
		veth, ok := link.(*netlink.Veth)
		if !ok {
			// skip device not enslaved to any bridge
			log.Debugf("skip %s, type: %s", link.Attrs().Name, link.Type())
			continue
		}
		log.Debugf("found veth device: %s", veth.Name)

		peerIdx, err := netlink.VethPeerIndex(veth)
		if err != nil {
			log.Fatal(err)
		}
		if link.Attrs().MasterIndex == -1 || veth.MasterIndex == 0 {
			log.Debugf("%s not has a bridge as master, MasterIndex: %d", veth.Name, link.Attrs().MasterIndex)
			if veth.PeerName == "" {
				p, err := netlink.LinkByIndex(peerIdx)
				if err != nil {
					log.Fatal(err)
				}
				veth.PeerName = p.Attrs().Name
			}
			routes, err := netlink.RouteList(link, 4)
			if err != nil {
				log.Fatal(err)
			}
			node := pkg.Node{
				Type:    pkg.VethType,
				Veth:    veth.Name,
				Peer:    veth.PeerName,
				PeerId:  peerIdx,
				NetNsID: veth.NetNsID,
			}
			if len(routes) > 0 {
				// TODO: more than one IP?
				node.Route = routes[0].Dst.IP
			}
			unBridgedVpairs = append(unBridgedVpairs, node)
			continue
		}

		master, err := netlink.LinkByIndex(veth.Attrs().MasterIndex)
		if err != nil {
			log.Fatal(err)
		}

		// if master is not bridge
		if _, ok := master.(*netlink.Bridge); !ok {
			// TODO: what if master is not bridge?
			continue
		}
		bridge := master.Attrs().Name
		v, ok := bridgeVethM[bridge]
		if !ok {
			bridgeVethM[bridge] = []pkg.Node{}
		}
		pair := pkg.Node{
			Type:    pkg.VethType,
			Veth:    veth.Name,
			PeerId:  peerIdx,
			NetNsID: veth.NetNsID,
		}
		if peerNetNs, ok := netNsMap[veth.NetNsID]; ok {
			peerInNs, err := netutil.GetPeerInNs(peerNetNs, origin, peerIdx)
			if err != nil {
				log.Fatal(err)
			}
			pair.NetNsName = peerNetNs
			pair.PeerNameInNetns = peerInNs.Attrs().Name
			pair.Status = peerInNs.Attrs().OperState.String()

			lo, err := netutil.GetLoInNs(peerNetNs, origin)
			if err == nil && lo != nil {
				loS = append(loS, pkg.Node{
					Type:      pkg.LoType,
					NetNsName: peerNetNs,
					Status:    lo.Attrs().OperState.String(),
				})
			}
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
			bridgeIps[bridge] = &addrs[0].IP
		}
		bridgeVethM[bridge] = append(v, pair)
	}
	log.Debugf("bridgeVethMap: %+v", bridgeVethM)

	switch {
	case *oGraph:
		buf := bytes.Buffer{}
		output, err := formatter.GraphInDOT(bridgeVethM, unBridgedVpairs, loS, bridgeIps)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(&buf, output)
		gType := strings.ToLower(*oGraphType)

		switch gType {
		case "dot":
			_, err = io.Copy(defaultOutput, &buf)
		case "jpg", "png", "svg":
			if !pflag.CommandLine.Changed("output") && !pflag.CommandLine.Changed("gtype") {
				log.Warn(`default output dst file: "output.png"`)
			}
			err = formatter.GenImage(buf.Bytes(), oGraphName, gType)
		default:
			log.Fatal("invalid graph type")
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	case *oTable:
		if len(bridgeVethM) > 0 {
			err := formatter.Table(defaultOutput, bridgeVethM)
			if err != nil {
				log.Fatal(err)
			}
		}
		if *oNotBridgedVeths && len(unBridgedVpairs) > 0 {
			formatter.TableParis(defaultOutput, unBridgedVpairs)
		}
		return
	default:
		formatter.Print(defaultOutput, bridgeVethM, netNsMap, unBridgedVpairs, *oNotBridgedVeths)
	}
}
