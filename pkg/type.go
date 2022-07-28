package pkg

type Pair struct {
	Veth        string
	Peer        string
	PeerInNetns string
	PeerId      int

	NetNsID   int
	NetNsName string
}
