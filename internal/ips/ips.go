package ips

import "net"

type Ips struct {
	V4 net.IP
	V6 net.IP
}
