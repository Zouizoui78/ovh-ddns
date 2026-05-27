package config

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
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
	cmd.PersistentFlags().String(DOMAINS_FLAG, "", "Domains for which to set the IP addresses")
	cmd.PersistentFlags().String(APP_KEY_FLAG, "", "OVH application key")
	cmd.PersistentFlags().String(APP_SECRET_FLAG, "", "OVH application secret")
	cmd.PersistentFlags().String(CONSUMER_KEY_FLAG, "", "OVH application consumer key")

	viper.BindPFlag("domains", cmd.PersistentFlags().Lookup(DOMAINS_FLAG))
	viper.BindPFlag("auth.app_key", cmd.PersistentFlags().Lookup(APP_KEY_FLAG))
	viper.BindPFlag("auth.app_secret", cmd.PersistentFlags().Lookup(APP_SECRET_FLAG))
	viper.BindPFlag("auth.consumer_key", cmd.PersistentFlags().Lookup(CONSUMER_KEY_FLAG))
}

func LoadConfig(cmd *cobra.Command) (*Config, error) {
	configInit(cmd)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.Getenv(CONFIG_PATH_ENV_VAR))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, err
		}
	}

	viper.SetEnvPrefix("OVH_DDNS")
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
