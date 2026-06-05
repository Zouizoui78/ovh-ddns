package updater

import (
	"context"
	"log/slog"
	"time"

	"github.com/Zouizoui78/ovh-ddns/internal/fetcher"
	"github.com/Zouizoui78/ovh-ddns/internal/ips"
	"github.com/Zouizoui78/ovh-ddns/internal/ovh"
	"golang.org/x/sync/errgroup"
)

type Updater struct {
	domains []string

	previousIps *ips.Ips

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

	if u.previousIps != nil && currentIps.Equal(u.previousIps) {
		slog.Info("public IPs have not changed, skipping dns update")
		return nil
	}
	u.previousIps = currentIps

	slog.Debug("fetching current dns zones IPs")
	ovhIps, err := u.ovh.GetDomainsIps(ctx, u.domains)
	if err != nil {
		return err
	}
	slog.Debug("current dns zones IPs", "ips", ovhIps)

	eg, _ := errgroup.WithContext(ctx)
	for domain, ips := range ovhIps {
		if ips.Equal(currentIps) {
			slog.Info("up to date IPs in DNS, skipping update", "domain", domain)
			continue
		}

		if dryRun {
			slog.Info("dry run: skipping dns zone update", "zone", domain)
			continue
		}

		eg.Go(func() error {
			slog.Info("updating dns zone", "zone", domain, "ips", ips)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
