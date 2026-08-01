package config

import (
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	PROG_NAME = "ovh-ddns"

	CONFIG_PATH_ENV_VAR = "OVH_DDNS_CONFIG_PATH"

	LOG_LEVEL_FLAG    = "log-level"
	DRY_RUN_FLAG      = "dry-run"
	DOMAINS_FLAG      = "domains"
	APP_KEY_FLAG      = "app-key"
	APP_SECRET_FLAG   = "app-secret"
	CONSUMER_KEY_FLAG = "consumer-key"
)

type Auth struct {
	AppKey      string `mapstructure:"app_key"`
	AppSecret   string `mapstructure:"app_secret"`
	ConsumerKey string `mapstructure:"consumer_key"`
}

type Config struct {
	LogLevel string   `mapstructure:"log_level"`
	DryRun   bool     `mapstructure:"dry_run"`
	Domains  []string `mapstructure:"domains"`
	Auth     Auth     `mapstructure:"auth"`
}

func configInit(cmd *cobra.Command) {
	viper.SetConfigName(PROG_NAME)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.Getenv(CONFIG_PATH_ENV_VAR))
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("OVH_DDNS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindPFlag("log_level", cmd.PersistentFlags().Lookup(LOG_LEVEL_FLAG))
	viper.BindPFlag("dry_run", cmd.PersistentFlags().Lookup(DRY_RUN_FLAG))
	viper.BindPFlag("domains", cmd.PersistentFlags().Lookup(DOMAINS_FLAG))
	viper.BindPFlag("auth.app_key", cmd.PersistentFlags().Lookup(APP_KEY_FLAG))
	viper.BindPFlag("auth.app_secret", cmd.PersistentFlags().Lookup(APP_SECRET_FLAG))
	viper.BindPFlag("auth.consumer_key", cmd.PersistentFlags().Lookup(CONSUMER_KEY_FLAG))
}

func LoadConfig(cmd *cobra.Command) (*Config, []error) {
	configInit(cmd)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, []error{err}
		}
	}

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, []error{err}
	}

	if err := validate(&cfg); len(err) != 0 {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) []error {
	var err []error

	err = validateLogLevel(cfg.LogLevel, err)
	err = validateDomains(cfg.Domains, err)
	err = validateAuth(cfg.Auth, err)

	return err
}

func validateLogLevel(logLevel string, err []error) []error {
	possibleLogLevels := []string{"debug", "info", "warn", "error"}
	if !slices.Contains(possibleLogLevels, logLevel) {
		err = append(err, errors.New("log level must be either 'debug', 'info', 'warn' or 'error'"))
	}
	return err
}

func validateDomains(domains []string, err []error) []error {
	if len(domains) == 0 {
		err = append(err, errors.New("need at least one configured domain"))
	}
	return err
}

func validateAuth(auth Auth, err []error) []error {
	if auth.AppKey == "" {
		err = append(err, errors.New("app key cannot be empty"))
	}

	if auth.AppSecret == "" {
		err = append(err, errors.New("app secret cannot be empty"))
	}

	if auth.ConsumerKey == "" {
		err = append(err, errors.New("consumer key cannot be empty"))
	}

	return err
}
