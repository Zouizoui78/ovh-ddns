package model

import "net"

type Record struct {
	Id         int
	RecordType RecordType
	SubDomain  string
	Target     net.IP
	Ttl        int
	Zone       string
}
