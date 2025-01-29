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
	Database struct {
		Driver string `mapstructure:"driver"`
		DSN    string `mapstructure:"dsn"`
	} `mapstructure:"db"`
	Channels      []ChannelConfig              `mapstructure:"channels"`
	ReactionCache map[string]map[string]string // channelID -> reaction -> category

}

type ChannelConfig struct {
	Name           string       `mapstructure:"name"`
	Rules          []RuleConfig `mapstructure:"rules"`
	BeaconReaction string       `mapstructure:"beacon_reaction"`
	ID             string       `mapstructure:"id"`
}

type RuleConfig struct {
	Reaction string `mapstructure:"reaction"`
	Category string `mapstructure:"category"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Add environment variable support
	viper.SetEnvPrefix("SLACK") // Prefix all env variables with SLACK_
	viper.AutomaticEnv()

	// Bind specific environment variables manually
	_ = viper.BindEnv("slack.app_token", "SLACK_APP_TOKEN")
	_ = viper.BindEnv("slack.bot_token", "SLACK_BOT_TOKEN")
	_ = viper.BindEnv("slack.signing_secret", "SLACK_SIGNING_SECRET")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *Config) BuildReactionCache() {
	config.ReactionCache = make(map[string]map[string]string)

	for _, channel := range config.Channels {
		reactionMap := make(map[string]string)
		for _, rule := range channel.Rules {
			reactionMap[rule.Reaction] = rule.Category
		}
		config.ReactionCache[channel.ID] = reactionMap
	}
}
