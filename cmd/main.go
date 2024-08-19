package main

import (
	"context"
	"errors"
	"github.com/TauAdam/digest-bot/internal/aggregator"
	"github.com/TauAdam/digest-bot/internal/config"
	"github.com/TauAdam/digest-bot/internal/notifier"
	"github.com/TauAdam/digest-bot/internal/storage"
	"github.com/TauAdam/digest-bot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("failed to create bot: %v", err)
		return
	}

	db, err := openDB()
	if err != nil {
		log.Printf("failed to connect to db: %v", err)
		return
	}

	var (
		articleRepository = storage.NewArticleStorage(db)
		sourcesRepository = storage.NewSourcesStorage(db)
		aggregatorService = aggregator.New(
			articleRepository,
			sourcesRepository,
			config.Get().UpdateInterval,
			config.Get().IgnoredKeywords,
		)

		notifierService = notifier.NewNotifier(
			articleRepository,
			summary.NewOpenAISummarizer(
				config.Get().OpenAIAPIKey,
				config.Get().OpenAIPrompt,
			),
			botAPI,
			config.Get().NotificationInterval,
			config.Get().UpdateInterval,
			config.Get().TelegramChannelID,
		)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		if err := aggregatorService.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to run aggregatorService: %v", err)
				return
			}

			log.Printf("aggregatorService stopped")
		}
	}(ctx)

	if err := notifierService.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("failed to run aggregatorService: %v", err)
			return
		}

		log.Printf("aggregatorService stopped")
	}
}

func openDB() (*sqlx.DB, error) {
	// Connect to a database and verify with a ping.
	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db, nil
}
