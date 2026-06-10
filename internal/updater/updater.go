package updater

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Zouizoui78/ovh-ddns/internal/fetcher"
	"github.com/Zouizoui78/ovh-ddns/internal/model"
	"github.com/Zouizoui78/ovh-ddns/internal/ovh"
	"golang.org/x/sync/errgroup"
)

type Updater struct {
	domains []string

	previousIps model.Ips

	fetcher *fetcher.Fetcher
	ovh     *ovh.Ovh

	t *time.Ticker
}

func New(f *fetcher.Fetcher, o *ovh.Ovh, domains []string) *Updater {
	return &Updater{
		domains: domains,
		fetcher: f,
		ovh:     o,
	}
}

func (u *Updater) Update(ctx context.Context) error {
	slog.Info("starting zones update")
	slog.Debug("fetching current public IPs")
	currentIps, err := u.fetcher.FetchIps(ctx)
	if err != nil {
		return err
	}
	slog.Debug("current public IPs", "v4", currentIps.V4, "v6", currentIps.V6)

	isPreviousIpsNil := u.previousIps.V4.IsUnspecified() && u.previousIps.V6.IsUnspecified()
	if isPreviousIpsNil && currentIps.Equal(u.previousIps) {
		slog.Info("public IPs have not changed, skipping dns update")
		return nil
	}
	u.previousIps = currentIps

	slog.Debug("fetching current dns zones")
	zones, err := u.ovh.GetDnsZones(ctx, u.domains)
	if err != nil {
		return err
	}

	eg, _ := errgroup.WithContext(ctx)
	for domain, zone := range zones {
		slog.Debug("current zone state", "A", *zone.A, "AAAA", *zone.AAAA)

		if zone.A != nil && zone.AAAA != nil {
			zoneIps := model.Ips{
				V4: zone.A.Target,
				V6: zone.AAAA.Target,
			}
			if currentIps.Equal(zoneIps) {
				slog.Info("up to date IPs in DNS, skipping update", "domain", domain)
				continue
			}
		}

		eg.Go(func() error {
			slog.Info("updating dns zone", "zone", domain, "ips", currentIps)
			u.updateZone(ctx, domain, currentIps, zone)
			return nil
		})
	}

	return eg.Wait()
}

func (u *Updater) updateZone(ctx context.Context, domain string, ips model.Ips, zone model.Zone) error {
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := u.updateRecord(
			egCtx,
			model.RecordTypeA,
			domain,
			ips.V4,
			zone.A,
		)
		if err != nil {
			return fmt.Errorf("failed to update A record: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		err := u.updateRecord(
			egCtx,
			model.RecordTypeAAAA,
			domain,
			ips.V6,
			zone.AAAA,
		)
		if err != nil {
			return fmt.Errorf("failed to update AAAA record: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("skipping '%s' zone refresh because of error during zone update: %w", domain, err)
	}

	return u.ovh.Refresh(ctx, domain)
}

func (u *Updater) updateRecord(ctx context.Context, recordType model.RecordType, domain string, ip net.IP, r *model.Record) error {
	if r == nil {
		slog.Debug(
			"posting new record",
			"zone", domain,
			"ip", ip,
			"record_type", recordType.String(),
		)

		var r model.Record
		var err error

		switch recordType {
		case model.RecordTypeA:
			r, err = u.ovh.PostARecord(
				ctx,
				domain,
				ip,
			)
		case model.RecordTypeAAAA:
			r, err = u.ovh.PostAAAARecord(
				ctx,
				domain,
				ip,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to post new record: %w", err)
		}

		slog.Info(
			"new record POST successful",
			"zone", domain,
			"ip", ip,
			"record_type", recordType.String(),
			"id", r.Id,
		)
	} else {
		slog.Info(
			"updating record",
			"zone", domain,
			"ip", ip,
			"record_type", recordType.String(),
		)

		r.Target = ip
		err := u.ovh.PutRecord(ctx, *r)
		if err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}

	return nil
}
