package fetcher

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/Zouizoui78/ovh-ddns/internal/model"
	"golang.org/x/sync/errgroup"
)

const PROVIDER = "https://ifconfig.me/ip"

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

func (i *Fetcher) FetchIps(parentCtx context.Context) (model.Ips, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	var ipv4, ipv6 net.IP

	eg.Go(func() error {
		var err error
		ipv4, err = fetchIp(ctx, i.v4Client)
		if err != nil {
			return fmt.Errorf("failed to get ipv4 address: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		ipv6, err = fetchIp(ctx, i.v6Client)
		if err != nil {
			return fmt.Errorf("failed to get ipv6 address: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return model.Ips{}, err
	}

	return model.Ips{
		V4: ipv4,
		V6: ipv6,
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

func fetchIp(ctx context.Context, c *http.Client) (net.IP, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", PROVIDER, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get ip: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("ip fetch: got status code %d. body: '%s'", res.StatusCode, body)
	}

	return net.ParseIP(string(body)), nil
}
