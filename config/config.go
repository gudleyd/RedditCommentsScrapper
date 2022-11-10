package config

import (
	"github.com/caarlos0/env/v6"
)

type Configuration struct {
	DatabaseConnectionString string `env:"DATABASE_CONNECTION_STRING"`
	BotUserAgent string `env:"BOT_USER_AGENT"`
	ClientID string `env:"CLIENT_ID"`
	ClientSecret string `env:"CLIENT_SECRET"`
	RedditUsername string `env:"REDDIT_USERNAME"`
	RedditPassword string `env:"REDDIT_PASSWORD"`
	MaxPostsPreload int `env:"MAX_POSTS_PRELOAD" envDefault:"10"`
	MaxErrors int `env:"MAX_ERRORS" envDefault:"25"`
}

func GetConfig() (Configuration, error) {
	config := Configuration{}
	if err := env.Parse(&config); err != nil {
		return config, err
	}
	return config, nil
}
