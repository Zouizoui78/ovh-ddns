package model

import (
	"net"
	"testing"
)

func TestIpsEqual(t *testing.T) {
	a := Ips{
		V4: net.ParseIP("123.456.789.123"),
		V6: net.ParseIP("2001:db8::68"),
	}
	b := a

	c := Ips{
		V4: net.ParseIP("987.654.321.987"),
		V6: net.ParseIP("2002:db8::68"),
	}

	if !a.Equal(b) {
		t.Fatal("IPs should be equal")
	}

	if a.Equal(c) {
		t.Fatal("IPs should not be equal")
	}
}
