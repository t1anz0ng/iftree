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
	graphviz "github.com/goccy/go-graphviz"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/TianZong48/iftree/pkg"
	"github.com/TianZong48/iftree/pkg/formatter"
	"github.com/TianZong48/iftree/pkg/netutil"
)

var (
	debug = pflag.BoolP("debug", "d", false, "print debug message")

	oGraph     = pflag.BoolP("graph", "g", false, "output in png by defaul")
	oGraphType = pflag.StringP("gtype", "T", "png", `graph output type, "jpg", "png", "svg", "dot"(graphviz dot language(https://graphviz.org/doc/info/lang.html)`)
	oGraphName = pflag.StringP("output", "O", "output", "graph output name/path")

	oTable = pflag.BoolP("table", "t", false, "output in table")

	help = pflag.BoolP("help", "h", false, "")
)

func init() {
	pflag.Usage = func() {
		fmt.Println(`Usage:
iftree [options]

Example:
  generate tree output
    # sudo iftree 
  generate png graph with name "output.png"
    # sudo iftree --graph -Tpng -Ooutput.png
  generate table output
    # sudo iftree --table

  -d, --debug   print debug message
  -t, --table   output in table
  -g, --graph   output in graphviz dot language(https://graphviz.org/doc/info/lang.html
    -O, --output string   graph output name/path (default "output")
    -T, --gtype string    graph output type, "jpg", "png", "svg", "dot" (graphviz dot language) default "png"
Help Options:
  -h, --help       Show this help message`)
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
	log.Debugf("net namespace id <-> name map:\n%+v\n", netNsMap)
	ll, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}

	vm := make(map[string][]pkg.Node) // map bridge <-> veth paris
	vpairs := []pkg.Node{}            // unused veth paris
	bm := make(map[string]*net.IP)    // bridge ip
	los := []pkg.Node{}               // loopback

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

				vpairs = append(vpairs,
					pkg.Node{
						Type:    pkg.VethType,
						Veth:    veth.Name,
						Peer:    veth.PeerName,
						PeerId:  peerIdx,
						NetNsID: veth.NetNsID})
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
				vm[bridge] = []pkg.Node{}
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
					los = append(los, pkg.Node{
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
				bm[bridge] = &addrs[0].IP
			}
			vm[bridge] = append(v, pair)
		}

	}

	if *oGraph {
		buf := bytes.Buffer{}
		output, err := formatter.Graph(vm, vpairs, los, bm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(&buf, output)
		gType := strings.ToLower(*oGraphType)

		switch gType {
		case "dot":
			_, err = io.Copy(os.Stdout, &buf)
		case "jpg", "png", "svg":
			graph, errG := graphviz.ParseBytes(buf.Bytes())
			if errG != nil {
				log.Fatal(errG)
			}
			g := graphviz.New()
			fn := fmt.Sprintf("%s.%s", *oGraphName, gType)
			f, errF := os.Create(fn)
			if errF != nil {
				log.Fatal(errF)
			}
			defer f.Close()
			switch gType {
			case "jpg":
				err = g.Render(graph, graphviz.JPG, f)
			case "png":
				err = g.Render(graph, graphviz.PNG, f)
			case "svg":
				err = g.Render(graph, graphviz.SVG, f)
			}
		default:
			log.Fatal("invalid graph type")
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if *oTable {
		err := formatter.Table(os.Stdout, vm)
		if err != nil {
			log.Fatal(err)
		}
		if err := formatter.TableParis(os.Stdout, vpairs); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := formatter.Print(os.Stdout, vm, netNsMap, vpairs); err != nil {
		log.Fatal(err)
	}

}
