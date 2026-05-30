package fetcher

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

const PROVIDER = "https://ifconfig.me/ip"

type Ips struct {
	Ipv4 net.IP
	Ipv6 net.IP
}

type Fetcher struct {
	v4Client *http.Client
	v6Client *http.Client
}

func New() *Fetcher {
	return &Fetcher{
		v4Client: newTransportSpecificClient("tcp4"),
		v6Client: newTransportSpecificClient("tcp6"),
	}
}

func (i *Fetcher) FetchIps() (*Ips, error) {
	ipv4, err := fetchIp(i.v4Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get ipv4 address: %s", err)
	}

	ipv6, err := fetchIp(i.v6Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get ipv6 address: %s", err)
	}

	return &Ips{
		Ipv4: ipv4,
		Ipv6: ipv6,
	}, nil
}

func newTransportSpecificClient(transport string) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, transport, addr)
	}
	client := *http.DefaultClient
	client.Transport = t
	return &client
}

func fetchIp(c *http.Client) (net.IP, error) {
	resp, err := c.Get(PROVIDER)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %s", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ipv4 fetch: got status code %d. body: '%s'", resp.StatusCode, body)
	}

	return net.ParseIP(string(body)), nil
}
