package bot

import (
	"context"
	"fmt"
	"github.com/TauAdam/digest-bot/internal/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type cmdArguments struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type sourceProvider interface {
	AddSource(ctx context.Context, source model.Source) (int64, error)
}

func HandleCmdAddSource(storage sourceProvider) HandlerFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := ParseJSONInput[cmdArguments](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		src := model.Source{
			Name:    args.Name,
			FeedURL: args.URL,
		}

		srcID, err := storage.AddSource(ctx, src)
		if err != nil {
			return err
		}

		var (
			// \\ is used to escape the backslash
			messageText = fmt.Sprintf("Source added with ID: %d\\. Use this id to manage source\\.", srcID)
			reply       = tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
		)

		// MarkdownV2 is used to let telegram know that the message is in Markdown format
		reply.ParseMode = "MarkdownV2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
