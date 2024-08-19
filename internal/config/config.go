package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
	"log"
	"sync"
	"time"
)

type Config struct {
	DatabaseDSN          string        `hcl:"database_dsn" env:"DATABASE_DSN" default:"postgres://postgres:password@localhost:5432/digest/?sslmode=disable"`
	UpdateInterval       time.Duration `hcl:"update_interval" env:"UPDATE_INTERVAL" default:"5m"`
	IgnoredKeywords      []string      `hcl:"ignored_keywords" env:"IGNORED_KEYWORDS"`
	NotificationInterval time.Duration `hcl:"notification_interval" env:"NOTIFICATION_INTERVAL" default:"1m"`

	TelegramBotToken  string `hcl:"telegram_bot_token" env:"TELEGRAM_BOT_TOKEN" required:"true"`
	TelegramChannelID int64  `hcl:"telegram_channel_id" env:"TELEGRAM_CHANNEL_ID" required:"true"`

	OpenAIAPIKey string `hcl:"openai_api_key" env:"OPENAI_API_KEY"`
	OpenAIPrompt string `hcl:"openai_prompt" env:"OPENAI_PROMPT"`
}

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	once.Do(func() {
		loader := aconfig.LoaderFor(&cfg, aconfig.Config{
			EnvPrefix: "DIGEST",
			Files:     []string{"./config.hcl", "./config.local.hcl"},
			FileDecoders: map[string]aconfig.FileDecoder{
				".hcl": aconfighcl.New(),
			},
		})

		if err := loader.Load(); err != nil {
			log.Printf("failed to load config: %v", err)
		}
	})

	return cfg
}
