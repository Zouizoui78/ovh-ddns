package config

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	PROG_NAME = "ovh-ddns"

	CONFIG_PATH_ENV_VAR = "OVH_DDNS_CONFIG_PATH"

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
	Domains []string `mapstructure:"domains"`
	Auth    Auth     `mapstructure:"auth"`
}

func configInit(cmd *cobra.Command) {
	viper.SetConfigName(PROG_NAME)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.Getenv(CONFIG_PATH_ENV_VAR))
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("OVH_DDNS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindPFlag("domains", cmd.PersistentFlags().Lookup(DOMAINS_FLAG))
	viper.BindPFlag("auth.app_key", cmd.PersistentFlags().Lookup(APP_KEY_FLAG))
	viper.BindPFlag("auth.app_secret", cmd.PersistentFlags().Lookup(APP_SECRET_FLAG))
	viper.BindPFlag("auth.consumer_key", cmd.PersistentFlags().Lookup(CONSUMER_KEY_FLAG))
}

func LoadConfig(cmd *cobra.Command) (*Config, error) {
	configInit(cmd)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, err
		}
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
