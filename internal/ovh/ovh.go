package ovh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
	"github.com/Zouizoui78/ovh-ddns/internal/model"
	"github.com/Zouizoui78/ovh-ddns/internal/ovh/dto"
	ovhapi "github.com/ovh/go-ovh/ovh"
	"golang.org/x/sync/errgroup"
)

type ovhClient interface {
	GetWithContext(ctx context.Context, url string, resType any) error
	PostWithContext(ctx context.Context, url string, reqBody, resType any) error
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

func (ovh *Ovh) GetDnsZones(parentCtx context.Context, domains []string) (map[string]model.Zone, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	ret := make(map[string]model.Zone)
	var mu sync.Mutex

	for _, domain := range domains {
		eg.Go(func() error {
			zone, err := ovh.getDnsZone(ctx, domain)
			if err != nil {
				return err
			}
			mu.Lock()
			ret[domain] = zone
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (ovh *Ovh) getDnsZone(parentCtx context.Context, domain string) (model.Zone, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	var a, aaaa model.Record

	type recordGetter = func(ctx context.Context, domain string) (model.Record, error)
	task := func(r *model.Record, g recordGetter) func() error {
		return func() error {
			var err error
			*r, err = g(ctx, domain)
			if err != nil && !errors.Is(err, ErrNoRecord) {
				return err
			}
			return nil
		}
	}

	eg.Go(task(&a, ovh.getARecordTarget))
	eg.Go(task(&aaaa, ovh.getAAAARecordTarget))

	if err := eg.Wait(); err != nil {
		return model.Zone{}, err
	}

	return model.Zone{
		A:    a,
		AAAA: aaaa,
	}, nil
}

func (ovh *Ovh) getRecord(ctx context.Context, domain string, recordType string) (model.Record, error) {
	var ids []int
	err := ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record?fieldType=%s", domain, recordType),
		&ids,
	)
	if err != nil {
		return model.Record{}, fmt.Errorf(
			"failed to get id of %s record for domain %s: %w",
			recordType,
			domain,
			err,
		)
	}

	if len(ids) == 0 {
		return model.Record{}, ErrNoRecord
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
		return model.Record{}, fmt.Errorf(
			"failed to get %s record for domain %s: %w",
			recordType,
			domain,
			err,
		)
	}

	return record.ToModel(), nil
}

func (ovh *Ovh) postRecord(ctx context.Context, domain string, recordType string, target net.IP) error {
	json, err := json.Marshal(dto.RecordPost{
		FieldType: recordType,
		Target:    target,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal POST record dto: %w", err)
	}

	var record dto.Record
	err = ovh.client.PostWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record", domain),
		json,
		&record,
	)
	if err != nil {
		return fmt.Errorf("failed to POST record: %w", err)
	}

	return nil
}

func (ovh *Ovh) getARecordTarget(ctx context.Context, domain string) (model.Record, error) {
	return ovh.getRecord(ctx, domain, "A")
}

func (ovh *Ovh) getAAAARecordTarget(ctx context.Context, domain string) (model.Record, error) {
	return ovh.getRecord(ctx, domain, "AAAA")
}

func (ovh *Ovh) postARecord(ctx context.Context, domain string, target net.IP) error {
	return ovh.postRecord(ctx, domain, "A", target)
}

func (ovh *Ovh) postAAAARecord(ctx context.Context, domain string, target net.IP) error {
	return ovh.postRecord(ctx, domain, "AAAA", target)
}
