package commons

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/network"
)

// IPAddr returns the external IP address
func IPAddr() (net.IP, error) {
	ip, err := network.ExternalIP()
	if err != nil {
		return nil, errors.Wrap(err, "could not get IPv4 address")
	}
	return net.ParseIP(ip), nil
}

// CheckAddress checks that some address is accessible and returns error accordingly
func CheckAddress(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		return errors.Wrap(err, "IP address is not accessible")
	}
	if err := conn.Close(); err != nil {
		return errors.Wrap(err, "could not close connection")
	}
	return nil
}

// BuildMultiAddress creates a multiaddr from the given params
func BuildMultiAddress(ipAddr, protocol string, port uint, id peer.ID) (ma.Multiaddr, error) {
	parsedIP := net.ParseIP(ipAddr)
	if parsedIP.To4() == nil && parsedIP.To16() == nil {
		return nil, errors.Errorf("invalid ip address provided: %s", ipAddr)
	}
	maStr := fmt.Sprintf("/ip6/%s/%s/%d", ipAddr, protocol, port)
	if parsedIP.To4() != nil {
		maStr = fmt.Sprintf("/ip4/%s/%s/%d", ipAddr, protocol, port)
	}
	if len(id) > 0 {
		maStr = fmt.Sprintf("%s/p2p/%s", maStr, id.String())
	}
	return ma.NewMultiaddr(maStr)
}
