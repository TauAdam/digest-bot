package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"runtime/debug"
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
func (b *Bot) RegisterNewCommand(cmd string, router HandlerFunc) {
	if b.cmdRouter == nil {
		b.cmdRouter = make(map[string]HandlerFunc)
	}

	b.cmdRouter[cmd] = router
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in handleUpdate: %v\n%s", r, string(debug.Stack()))
		}
	}()

	var handler HandlerFunc

	if !update.Message.IsCommand() {
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
