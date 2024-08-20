package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sourceDeleteProvider interface {
	DeleteSource(ctx context.Context, id int64) error
}
type cmdDeleteArguments struct {
	ID int64 `json:"id"`
}

func HandleCmdDeleteSource(storage sourceDeleteProvider) HandlerFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := ParseJSONInput[cmdDeleteArguments](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		srcID := args.ID

		if err := storage.DeleteSource(ctx, srcID); err != nil {
			return err
		}

		messageText := fmt.Sprintf("Source with ID: %d\\ deleted\\.", srcID)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
		reply.ParseMode = "MarkdownV2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
