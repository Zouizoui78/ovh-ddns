package main

import (
	"log/slog"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "ovh-ddns",
	Short: "ovh-ddns",
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

func main() {
	cmd.Execute()
}
