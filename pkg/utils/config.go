package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	Slack struct {
		AppToken      string `mapstructure:"app_token"`
		BotToken      string `mapstructure:"bot_token"`
		SigningSecret string `mapstructure:"signing_secret"`
	} `mapstructure:"slack"`
	Bot struct {
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"bot"`
	Channels []ChannelConfig `mapstructure:"channels"`
}

type ChannelConfig struct {
	Name  string       `mapstructure:"name"`
	Rules []RuleConfig `mapstructure:"rules"`
}

type RuleConfig struct {
	Reaction string `mapstructure:"reaction"`
	Category string `mapstructure:"category"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
