package config

import "net"

var IpBlackList = []net.IP{
	net.IPv4(94, 79, 55, 28),
}

func IsBlack(i net.IP) bool {
	for _, black := range IpBlackList {
		if black.Equal(i) {
			return true
		}
	}
	return false
}
