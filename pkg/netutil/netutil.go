package netutil

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
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
		netnsFd, err := netns.GetFromPath(path)
		if err != nil {
			return nil, errors.Wrapf(err, "fail get netns from path %s", path)
		}

		id, err := netlink.GetNetNsIdByFd(int(netnsFd))
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

func listNetNsPath() ([]string, error) {
	// https://man7.org/linux/man-pages/man8/ip-netns.8.html
	path := "/var/run/netns"
	es, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var ns []string
	for _, e := range es {
		ns = append(ns, filepath.Join(path, e.Name()))
	}
	// default docker dir
	dockerPath := "/var/run/docker/netns"
	dEs, err := os.ReadDir(dockerPath)
	if err != nil {
		return nil, err
	}
	for _, e := range dEs {
		ns = append(ns, filepath.Join(dockerPath, e.Name()))
	}
	return ns, nil
}

func GetPeerInNs(ns string, peerIdx int, origin netns.NsHandle) (string, error) {
	hd, err := netns.GetFromPath(ns)
	if err != nil {
		log.Fatal(err)
	}
	if err := netns.Set(hd); err != nil {
		return "", err
	}
	defer netns.Set(origin) //nolint: errcheck

	peerInNs, err := netlink.LinkByIndex(peerIdx)
	if err != nil {
		return "", err
	}
	// Switch back to the original namespace
	return peerInNs.Attrs().Name, nil
}
