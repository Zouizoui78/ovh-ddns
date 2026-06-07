package model

import "net"

type Ips struct {
	V4 net.IP
	V6 net.IP
}

func (i *Ips) Equal(other Ips) bool {
	return i.V4.Equal(other.V4) && i.V6.Equal(other.V6)
}
