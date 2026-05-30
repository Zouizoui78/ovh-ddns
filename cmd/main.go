package main

import (
	"log/slog"
	"os"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
	"github.com/Zouizoui78/ovh-ddns/internal/fetcher"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   config.PROG_NAME,
	Short: config.PROG_NAME,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(cmd)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("cfg", "domains", cfg.Domains)

	fetcher := fetcher.New()
	ips, err := fetcher.FetchIps()
	if err != nil {
		slog.Error("failed to fetch ips", "err", err)
	}

	slog.Info("addr", "ipv4", ips.Ipv4, "ipv6", ips.Ipv6)
}

func init() {
	cmd.PersistentFlags().String(config.DOMAINS_FLAG, "", "Domains for which to set the IP addresses")
	cmd.PersistentFlags().String(config.APP_KEY_FLAG, "", "OVH application key")
	cmd.PersistentFlags().String(config.APP_SECRET_FLAG, "", "OVH application secret")
	cmd.PersistentFlags().String(config.CONSUMER_KEY_FLAG, "", "OVH application consumer key")
}

func main() {
	cmd.Execute()
}
