package notifier

import (
	"context"
	"github.com/TauAdam/digest-bot/internal/model"
	"time"
)

type ArticleProvider interface {
	UnsentArticles(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error)
	MarkPosted(ctx context.Context, id int64) error
}

type Notifier struct {
	articles         ArticleProvider
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	channelID        int64
}
