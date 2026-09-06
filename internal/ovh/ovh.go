package ovh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
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
	PutWithContext(ctx context.Context, url string, reqBody, resType any) error
}

type Ovh struct {
	client ovhClient
	dryRun bool
}

func New(auth config.Auth, dryRun bool) (*Ovh, error) {
	client, err := ovhapi.NewClient(
		"ovh-eu",
		auth.AppKey,
		auth.AppSecret,
		auth.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ovh client: %w", err)
	}

	return NewFromClient(client, dryRun), nil
}

func NewFromClient(client ovhClient, dryRun bool) *Ovh {
	return &Ovh{
		client: client,
		dryRun: dryRun,
	}
}

func (ovh *Ovh) GetDnsZones(parentCtx context.Context, zoneNames []string) (map[string]model.Zone, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	ret := make(map[string]model.Zone, len(zoneNames))
	var mu sync.Mutex

	for _, zoneName := range zoneNames {
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
	d := dto.NewRecordPostDtoFromModel(r)

	if ovh.dryRun {
		json, _ := json.Marshal(d)
		slog.Warn(
			"dry run: skipping post",
			"zone", r.Zone,
			"recordType", r.RecordType,
			"target", r.Target,
			"dto", json,
		)
		r.Id = rand.Int()
		return r, nil
	}

	var record dto.Record
	err := ovh.client.PostWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record", r.Zone),
		d,
		&record,
	)
	if err != nil {
		return model.Record{}, fmt.Errorf("failed to POST record: %w", err)
	}

	return record.ToModel(), nil
}

func (ovh *Ovh) PostARecord(ctx context.Context, zoneName string, target net.IP) (model.Record, error) {
	return ovh.PostRecord(ctx, model.Record{
		Zone:       zoneName,
		RecordType: model.RecordTypeA,
		Target:     target,
	})
}

func (ovh *Ovh) PostAAAARecord(ctx context.Context, zoneName string, target net.IP) (model.Record, error) {
	return ovh.PostRecord(ctx, model.Record{
		Zone:       zoneName,
		RecordType: model.RecordTypeAAAA,
		Target:     target,
	})
}

func (ovh *Ovh) PutRecord(ctx context.Context, r model.Record) error {
	d := dto.NewRecordPutDtoFromModel(r)

	if ovh.dryRun {
		json, _ := json.Marshal(d)

		slog.Warn(
			"dry run: skipping put",
			"zone", r.Zone,
			"recordType", r.RecordType,
			"target", r.Target,
			"dto", json,
		)
		return nil
	}

	err := ovh.client.PutWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record/%d", r.Zone, r.Id),
		d,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to update dns record: %w", err)
	}

	return nil
}

func (ovh *Ovh) Refresh(ctx context.Context, zoneName string) error {
	if ovh.dryRun {
		slog.Warn(
			"dry run: skipping refresh",
			"zone", zoneName,
		)
		return nil
	}

	err := ovh.client.PostWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/refresh", zoneName),
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to refresh dns zone: %w", err)
	}

	return nil
}

// Helpers

func (ovh *Ovh) getDnsZone(parentCtx context.Context, zoneName string) (model.Zone, error) {
	eg, ctx := errgroup.WithContext(parentCtx)
	var a, aaaa *model.Record

	type recordGetter = func(ctx context.Context, zoneName string) (*model.Record, error)
	task := func(r **model.Record, g recordGetter) func() error {
		return func() error {
			var err error
			*r, err = g(ctx, zoneName)
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

func (ovh *Ovh) getRecord(ctx context.Context, zoneName string, recordType string) (*model.Record, error) {
	var ids []int
	err := ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record?fieldType=%s&subDomain=", zoneName, recordType),
		&ids,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get id of %s record for zone %s: %w",
			recordType,
			zoneName,
			err,
		)
	}

	if len(ids) == 0 {
		return nil, ErrNoRecord
	}
	if len(ids) > 1 {
		slog.Warn("multiple dns record, picking first one", "zone", zoneName, "type", recordType)
	}

	var dto dto.Record
	err = ovh.client.GetWithContext(
		ctx,
		fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, ids[0]),
		&dto,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s record for zone %s: %w",
			recordType,
			zoneName,
			err,
		)
	}

	ret := dto.ToModel()
	return &ret, nil
}

func (ovh *Ovh) getARecord(ctx context.Context, zoneName string) (*model.Record, error) {
	return ovh.getRecord(ctx, zoneName, "A")
}

func (ovh *Ovh) getAAAARecord(ctx context.Context, zoneName string) (*model.Record, error) {
	return ovh.getRecord(ctx, zoneName, "AAAA")
}
