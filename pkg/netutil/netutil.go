package netutil

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// https://github.com/shemminger/iproute2/blob/main/ip/ipnetns.c#L432
// https://github.com/shemminger/iproute2/blob/main/ip/ipnetns.c#L106
func NetNsMap() (map[int]string, error) {
	nsArr, err := getNetNs()
	if err != nil {
		return nil, err
	}
	m := make(map[int]string)
	for _, ns := range nsArr {
		netnsFd, err := netns.GetFromName(ns)
		if err != nil {
			return nil, err
		}

		id, err := netlink.GetNetNsIdByFd(int(netnsFd))
		if err != nil {
			return nil, err
		}
		m[id] = ns
	}
	return m, nil
}

func getNetNs() ([]string, error) {
	// https://man7.org/linux/man-pages/man8/ip-netns.8.html
	path := "/var/run/netns"
	es, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var ns []string
	for _, e := range es {
		ns = append(ns, e.Name())
	}
	return ns, nil
}

func GetPeerInNs(ns string, peerIdx int, origin netns.NsHandle) (string, error) {
	hd, err := netns.GetFromName(ns)
	if err != nil {
		log.Fatal(err)
	}
	if err := netns.Set(hd); err != nil {
		return "", err
	}
	defer netns.Set(origin) //nolint: defer errcheck

	peerInNs, err := netlink.LinkByIndex(peerIdx)
	if err != nil {
		return "", err
	}
	// Switch back to the original namespace
	return peerInNs.Attrs().Name, nil
}
