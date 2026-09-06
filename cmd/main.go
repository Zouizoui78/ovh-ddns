package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Zouizoui78/ovh-ddns/internal/config"
	"github.com/Zouizoui78/ovh-ddns/internal/fetcher"
	"github.com/Zouizoui78/ovh-ddns/internal/ovh"
	"github.com/Zouizoui78/ovh-ddns/internal/updater"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   config.PROG_NAME,
	Short: config.PROG_NAME,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	cfg, cfgErr := config.LoadConfig(cmd)
	if len(cfgErr) != 0 {
		for _, e := range cfgErr {
			slog.Error("validation error", "error", e)
		}
		os.Exit(1)
	}

	level := slog.LevelWarn
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	defer slog.Debug("exiting")

	if cfg.DryRun {
		slog.Info("dry run mode active")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := fetcher.New()
	ovh, err := ovh.New(cfg.Auth, cfg.DryRun)
	if err != nil {
		slog.Error("failed to create ovh client", "err", err)
	}
	updater := updater.New(fetcher, ovh, cfg.Domains)
	err = updater.Update(ctx)
	if err != nil {
		slog.Error(err.Error())
	}
}

func init() {
	cmd.PersistentFlags().String(config.LOG_LEVEL_FLAG, "warn", "Log level. Can be debug, info, warn or error")
	cmd.PersistentFlags().Bool(config.DRY_RUN_FLAG, false, "If enabled, DNS zones are not updated")
	cmd.PersistentFlags().String(config.DOMAINS_FLAG, "", "Domains for which to set the IP addresses")
	cmd.PersistentFlags().String(config.APP_KEY_FLAG, "", "OVH application key")
	cmd.PersistentFlags().String(config.APP_SECRET_FLAG, "", "OVH application secret")
	cmd.PersistentFlags().String(config.CONSUMER_KEY_FLAG, "", "OVH application consumer key")
	cmd.PersistentFlags().Duration(config.UPDATE_INTERVAL_FLAG, time.Minute, "Wait time between two updates. Must be >1min")
}

func main() {
	cmd.Execute()
}
