package main

import (
	"log/slog"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
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
	}

	slog.Info(
		"config",
		"domains", cfg.Domains,
		"app_key", cfg.Auth.AppKey,
		"app_secret", cfg.Auth.AppSecret,
		"consumer_key", cfg.Auth.ConsumerKey,
	)
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
