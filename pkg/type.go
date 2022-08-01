package pkg

import (
	"net"
)

type NodeType int

const (
	VethType NodeType = iota
	BridgeType
	LoType
)

var (
	typeMap = map[NodeType]string{
		VethType:   "veth",
		BridgeType: "bridge",
		LoType:     "lo",
	}
)

type Node struct {
	ID   string
	Type NodeType

	// veth
	Veth            string
	Peer            string
	PeerNameInNetns string
	PeerId          int
	Orphaned        bool
	NetNsID         int
	NetNsName       string
	Master          *Bridge

	// general
	Status string
}

func (n *Node) String() string {
	return typeMap[n.Type]
}

type Bridge struct {
	Name string
	IP   *net.IP
}
