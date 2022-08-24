package pkg

import (
	"net"
	"strings"
)

type NodeType int

const (
	VethType NodeType = iota
	BridgeType
	LoType
)

var (
	typeMap = map[NodeType]string{ //nolint
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
	Route           net.IP

	// general
	Name   string
	Status string
}

func (n *Node) Label() string {
	//FIXME: deprecate VETH
	if n.Name == "" {
		n.Name = n.Veth
	}
	switch n.Type {
	case LoType:
		return "lo"
	default:
		return strings.TrimSpace(n.Name)
	}
}

type Bridge struct {
	Name string
	IP   *net.IP
}
