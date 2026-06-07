package updater

import (
	"context"
	"log/slog"
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

func (u *Updater) Update(ctx context.Context, dryRun bool) error {
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
	slog.Debug("current dns zones", "zone", zones)

	eg, _ := errgroup.WithContext(ctx)
	for domain, zone := range zones {
		zoneIps := model.Ips{
			V4: zone.A.Target,
			V6: zone.AAAA.Target,
		}
		if currentIps.Equal(zoneIps) {
			slog.Info("up to date IPs in DNS, skipping update", "domain", domain)
			continue
		}

		if dryRun {
			slog.Info("dry run: skipping dns zone update", "zone", domain)
			continue
		}

		eg.Go(func() error {
			slog.Info("updating dns zone", "zone", domain, "ips", zoneIps)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
