package ovh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
	"github.com/Zouizoui78/ovh-ddns/internal/ips"
	"github.com/Zouizoui78/ovh-ddns/internal/ovh/dto"
	ovhapi "github.com/ovh/go-ovh/ovh"
	"golang.org/x/sync/errgroup"
)

type ovhClient interface {
	GetWithContext(ctx context.Context, url string, resType any) error
}

type Ovh struct {
	client ovhClient
}

func New(auth config.Auth) (*Ovh, error) {
	client, err := ovhapi.NewClient(
		"ovh-eu",
		auth.AppKey,
		auth.AppSecret,
		auth.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ovh client: %w", err)
	}

	return NewFromClient(client), nil
}

func NewFromClient(client ovhClient) *Ovh {
	return &Ovh{
		client: client,
	}
}

func (ovh *Ovh) GetDomainsIps(parentCtx context.Context, domains []string) (map[string]ips.Ips, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	ret := make(map[string]ips.Ips)
	var mu sync.Mutex

	for _, domain := range domains {
		eg.Go(func() error {
			ip, err := ovh.getDomainIps(ctx, domain)
			if err != nil {
				return err
			}
			mu.Lock()
			ret[domain] = *ip
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (ovh *Ovh) getDomainIps(parentCtx context.Context, domain string) (*ips.Ips, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	var v4, v6 net.IP

	type recordGetter = func(ctx context.Context, domain string) (net.IP, error)
	task := func(ip *net.IP, g recordGetter) func() error {
		return func() error {
			var err error
			*ip, err = g(ctx, domain)
			if err != nil && !errors.Is(err, ErrNoRecord) {
				return err
			}
			return nil
		}
	}

	eg.Go(task(&v4, ovh.getARecordTarget))
	eg.Go(task(&v6, ovh.getAAAARecordTarget))

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &ips.Ips{
		V4: v4,
		V6: v6,
	}, nil
}

func (ovh *Ovh) getRecordTarget(ctx context.Context, domain string, recordType string) (net.IP, error) {
	var ids []int
	err := ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record?fieldType=%s", domain, recordType),
		&ids,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get id of %s record for domain %s: %w",
			recordType,
			domain,
			err,
		)
	}

	if len(ids) == 0 {
		return nil, ErrNoRecord
	}
	if len(ids) > 1 {
		slog.Warn("multiple dns record, picking first one", "domain", domain, "type", recordType)
	}

	var record dto.Record
	err = ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record/%d", domain, ids[0]),
		&record,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s record for domain %s: %w",
			recordType,
			domain,
			err,
		)
	}

	return net.ParseIP(record.Target), nil
}

func (ovh *Ovh) getARecordTarget(ctx context.Context, domain string) (net.IP, error) {
	return ovh.getRecordTarget(ctx, domain, "A")
}

func (ovh *Ovh) getAAAARecordTarget(ctx context.Context, domain string) (net.IP, error) {
	return ovh.getRecordTarget(ctx, domain, "AAAA")
}
