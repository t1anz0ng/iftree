package netutil

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	// https://man7.org/linux/man-pages/man8/ip-netns.8.html
	netNsPath = "/var/run/netns"
	// default docker dir
	docNetNSkerPath = "/var/run/docker/netns"
)

// https://github.com/shemminger/iproute2/blob/main/ip/ipnetns.c#L432
// https://github.com/shemminger/iproute2/blob/main/ip/ipnetns.c#L106
func NetNsMap() (map[int]string, error) {
	nsArr, err := listNetNsPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed list netns")
	}

	m := make(map[int]string)
	for _, path := range nsArr {
		id, err := NsidFromPath(path)
		if err != nil {
			return nil, err
		}
		// -1 if the namespace does not have an ID set.
		if id != -1 {
			m[id] = path
		}
	}
	return m, nil
}
func NsidFromPath(path string) (int, error) {
	netnsFd, err := netns.GetFromPath(path)
	if err != nil {
		return 0, errors.Wrapf(err, "fail get netns from path %s", path)
	}

	id, err := netlink.GetNetNsIdByFd(int(netnsFd))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func listNetNsPath() ([]string, error) {
	var ns []string

	es, err := os.ReadDir(netNsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else {
		for _, e := range es {
			ns = append(ns, filepath.Join(netNsPath, e.Name()))
		}
	}
	dEs, err := os.ReadDir(docNetNSkerPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else {
		for _, e := range dEs {
			ns = append(ns, filepath.Join(docNetNSkerPath, e.Name()))
		}
	}

	return ns, nil
}

// GetPeerInNs enter target netns to get veth peer's name
// root needed
func GetPeerInNs(ns string, origin netns.NsHandle, peerIdx int) (netlink.Link, error) {
	return netnsGetName(ns, origin, func() (netlink.Link, error) {
		return netlink.LinkByIndex(peerIdx)
	})
}

func GetLoInNs(ns string, origin netns.NsHandle) (netlink.Link, error) {
	return netnsGetName(ns, origin, func() (netlink.Link, error) {
		return netlink.LinkByName("lo")
	})
}

func netnsGetName(ns string, origin netns.NsHandle, fn func() (netlink.Link, error)) (link netlink.Link, err error) {
	// Switch back to the original namespace
	defer netns.Set(origin) //nolint: errcheck

	hd, err := netns.GetFromPath(ns)
	if err != nil {
		return nil, err
	}
	if err := netns.Set(hd); err != nil {
		return nil, err
	}
	return fn()
}
