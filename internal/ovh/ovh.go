package ovh

import (
	"context"
	"fmt"
	"net"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
	"github.com/Zouizoui78/ovh-ddns/internal/ips"
	"github.com/Zouizoui78/ovh-ddns/internal/ovh/dto"
	ovhapi "github.com/ovh/go-ovh/ovh"
	"golang.org/x/sync/errgroup"
)

const ()

type Ovh struct {
	client *ovhapi.Client
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

	return &Ovh{
		client: client,
	}, nil
}

func (ovh *Ovh) GetDomainsIps(parentCtx context.Context, domains []string) (map[string]ips.Ips, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	ret := make(map[string]ips.Ips)

	for _, domain := range domains {
		eg.Go(func() error {
			ip, err := ovh.getDomainIps(ctx, domain)
			if err != nil {
				return err
			}
			ret[domain] = *ip
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

	eg.Go(func() error {
		var err error
		v4, err = ovh.getARecordTarget(ctx, domain)
		if err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		v6, err = ovh.getAAAARecordTarget(ctx, domain)
		if err != nil {
			return err
		}
		return nil
	})

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
		return nil, fmt.Errorf("there is no %s record for domain %s", recordType, domain)
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
