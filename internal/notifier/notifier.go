package notifier

import (
	"context"
	"fmt"
	"github.com/TauAdam/digest-bot/internal/bot/markup"
	"github.com/TauAdam/digest-bot/internal/model"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ArticleProvider interface {
	UnsentArticles(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error)
	MarkPosted(ctx context.Context, id int64) error
}

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type Notifier struct {
	articles         ArticleProvider
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	channelID        int64
	bot              *tgbotapi.BotAPI
	summarizer       Summarizer
}

func (n *Notifier) SelectAndPost(ctx context.Context) error {
	topArticles, err := n.articles.UnsentArticles(ctx, time.Now().Add(-n.lookupTimeWindow), 1)
	if err != nil {
		return err
	}

	if len(topArticles) == 0 {
		return nil
	}

	article := topArticles[0]

	summary, err := n.extractSummary(ctx, article)
	if err != nil {
		return err
	}

	if err := n.sendArticle(article, summary); err != nil {
		return err
	}

	return n.articles.MarkPosted(ctx, article.ID)
}

func (n *Notifier) extractSummary(ctx context.Context, article model.Article) (string, error) {
	var r io.Reader

	if article.Summary != "" {
		r = strings.NewReader(article.Summary)
	} else {
		resp, err := http.Get(article.Link)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		r = resp.Body
	}

	document, err := readability.FromReader(r, nil)
	if err != nil {
		return "", err
	}

	summary, err := n.summarizer.Summarize(ctx, removeRedundantNewLines(document.TextContent))
	if err != nil {
		return "", err
	}

	return "\n\n" + summary, nil
}

func (n *Notifier) sendArticle(article model.Article, summary string) error {
	const messageFormat = "*%s*%s\n\n%s"

	msg := tgbotapi.NewMessage(n.channelID, fmt.Sprintf(
		messageFormat,
		markup.MarkdownEscape(article.Title),
		markup.MarkdownEscape(summary),
		markup.MarkdownEscape(article.Link),
	))

	_, err := n.bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func removeRedundantNewLines(str string) string {
	return regexp.MustCompile(`\n{3,}`).ReplaceAllString(str, "\n")
}

func NewNotifier(
	articles ArticleProvider,
	sendInterval time.Duration,
	lookupTimeWindow time.Duration,
	channelID int64,
	bot *tgbotapi.BotAPI,
	summarizer Summarizer,
) *Notifier {
	return &Notifier{
		articles:         articles,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		channelID:        channelID,
		bot:              bot,
		summarizer:       summarizer,
	}
}
