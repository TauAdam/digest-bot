package bot

import (
	"context"
	"fmt"
	"github.com/TauAdam/digest-bot/internal/bot/markup"
	"github.com/TauAdam/digest-bot/internal/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
	"strings"
)

type sourceListProvider interface {
	ListSources(ctx context.Context) ([]model.Source, error)
}

func HandleCmdListSources(storage sourceListProvider) HandlerFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		sources, err := storage.ListSources(ctx)
		if err != nil {
			return err
		}

		formattedSources := lo.Map(sources, func(src model.Source, _ int) string {
			return prepareSources(src)
		})

		messageText := fmt.Sprintf(
			"List of sources \\(total: %d\\):\n\n%s",
			len(sources),
			strings.Join(formattedSources, "\n\n"),
		)

		if len(sources) == 0 {
			messageText = "No sources found\\."
		}

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)

		reply.ParseMode = "MarkdownV2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}

// prepareSources formats the source data to be displayed in the message
func prepareSources(src model.Source) string {
	return fmt.Sprintf(
		"📃 *%s*\nID: `%d`\nFeed URL: `%s`",
		markup.MarkdownEscape(src.Name),
		src.ID,
		markup.MarkdownEscape(src.FeedURL),
	)
}
