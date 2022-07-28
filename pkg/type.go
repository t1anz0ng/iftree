package pkg

import "net"

type Pair struct {
	Veth        string
	Peer        string
	PeerInNetns string
	PeerId      int

	NetNsID   int
	NetNsName string

	Master *Bridge
}

type Bridge struct {
	Name string
	IP   []*net.IP
}
