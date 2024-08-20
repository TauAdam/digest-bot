package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"runtime/debug"
	"time"
)

type HandlerFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error

type Bot struct {
	api       *tgbotapi.BotAPI
	cmdRouter map[string]HandlerFunc
}

func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot{
		api: api,
	}
}

// RegisterNewCommand add new command to the bot
func (b *Bot) RegisterNewCommand(cmd string, router HandlerFunc) {
	if b.cmdRouter == nil {
		b.cmdRouter = make(map[string]HandlerFunc)
	}

	b.cmdRouter[cmd] = router
}

// handleUpdate handles incoming updates
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in handleUpdate: %v\n%s", r, string(debug.Stack()))
		}
	}()

	var handler HandlerFunc

	// currently bot only handle commands
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	cmd := update.Message.Command()

	cmdHandler, ok := b.cmdRouter[cmd]
	if !ok {
		return
	}

	handler = cmdHandler

	if err := handler(ctx, b.api, update); err != nil {
		log.Printf("failed to handle command %s: %v", cmd, err)

		if _, err := b.api.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Something went wrong",
		)); err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}

}

// Run starts the bot and listens for updates
func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			updateContext, updateCancel := context.WithTimeout(ctx, 5*time.Second)
			b.handleUpdate(updateContext, update)
			updateCancel()
		}
	}
}
