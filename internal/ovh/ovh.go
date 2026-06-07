package ovh

import (
	"context"
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

func (ovh *Ovh) GetDnsZones(parentCtx context.Context, zones []string) (map[string]model.Zone, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	ret := make(map[string]model.Zone, len(zones))
	var mu sync.Mutex

	for _, zoneName := range zones {
		eg.Go(func() error {
			zone, err := ovh.getDnsZone(ctx, zoneName)
			if err != nil {
				return err
			}
			mu.Lock()
			ret[zoneName] = zone
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (ovh *Ovh) PostRecord(ctx context.Context, r model.Record) (model.Record, error) {
	var record dto.Record
	err := ovh.client.PostWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record", r.Zone),
		dto.FromModel(r),
		&record,
	)
	if err != nil {
		return model.Record{}, fmt.Errorf("failed to POST record: %w", err)
	}

	return record.ToModel(), nil
}

func (ovh *Ovh) PostARecord(ctx context.Context, zone string, target net.IP) (model.Record, error) {
	return ovh.PostRecord(ctx, model.Record{
		Zone:       zone,
		RecordType: model.RecordTypeA,
		Target:     target,
	})
}

func (ovh *Ovh) PostAAAARecord(ctx context.Context, zone string, target net.IP) (model.Record, error) {
	return ovh.PostRecord(ctx, model.Record{
		Zone:       zone,
		RecordType: model.RecordTypeAAAA,
		Target:     target,
	})
}

// Helpers

func (ovh *Ovh) getDnsZone(parentCtx context.Context, zone string) (model.Zone, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	var a, aaaa *model.Record

	type recordGetter = func(ctx context.Context, zone string) (*model.Record, error)
	task := func(r **model.Record, g recordGetter) func() error {
		return func() error {
			var err error
			*r, err = g(ctx, zone)
			if err != nil && !errors.Is(err, ErrNoRecord) {
				return err
			}
			return nil
		}
	}

	eg.Go(task(&a, ovh.getARecord))
	eg.Go(task(&aaaa, ovh.getAAAARecord))

	if err := eg.Wait(); err != nil {
		return model.Zone{}, err
	}

	return model.Zone{
		A:    a,
		AAAA: aaaa,
	}, nil
}

func (ovh *Ovh) getRecord(ctx context.Context, zone string, recordType string) (*model.Record, error) {
	var ids []int
	err := ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record?fieldType=%s", zone, recordType),
		&ids,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get id of %s record for zone %s: %w",
			recordType,
			zone,
			err,
		)
	}

	if len(ids) == 0 {
		return nil, ErrNoRecord
	}
	if len(ids) > 1 {
		slog.Warn("multiple dns record, picking first one", "zone", zone, "type", recordType)
	}

	var dto dto.Record
	err = ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record/%d", zone, ids[0]),
		&dto,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s record for zone %s: %w",
			recordType,
			zone,
			err,
		)
	}

	ret := dto.ToModel()
	return &ret, nil
}

func (ovh *Ovh) getARecord(ctx context.Context, zone string) (*model.Record, error) {
	return ovh.getRecord(ctx, zone, "A")
}

func (ovh *Ovh) getAAAARecord(ctx context.Context, zone string) (*model.Record, error) {
	return ovh.getRecord(ctx, zone, "AAAA")
}
