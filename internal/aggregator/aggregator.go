package aggregator

import (
	"context"
	"github.com/TauAdam/digest-bot/internal/model"
	"github.com/TauAdam/digest-bot/internal/rss-sources"
	"github.com/TauAdam/digest-bot/pkg/set"
	"log"
	"strings"
	"sync"
	"time"
)

type ArticleRepo interface {
	Save(ctx context.Context, article model.Article) error
}

type SourceRepo interface {
	ListSources(ctx context.Context) ([]model.Source, error)
}

type Aggregator struct {
	articles ArticleRepo
	sources  SourceRepo

	updateInterval  time.Duration
	ignoredKeywords []string
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

func New(
	articleRepo ArticleRepo,
	sourceRepo SourceRepo,
	updateInterval time.Duration,
	whitelist []string,
) *Aggregator {
	return &Aggregator{
		articles:        articleRepo,
		sources:         sourceRepo,
		updateInterval:  updateInterval,
		ignoredKeywords: whitelist,
	}
}

func (a *Aggregator) Aggregate(ctx context.Context) error {
	sourcesList, err := a.sources.ListSources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, src := range sourcesList {
		wg.Add(1)

		rssSource := rss_sources.NewRSSSource(src)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("failed to fetch items from source %s: %v", source.Name(), err)
				return
			}

			if err := a.processItems(ctx, source, items); err != nil {
				log.Printf("failed to process items from source %s: %v", source.Name(), err)
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (a *Aggregator) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if a.isItemIrrelevant(item) {
			continue
		}

		if err := a.articles.Save(ctx, model.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *Aggregator) isItemIrrelevant(item model.Item) bool {
	categories := set.New(item.Categories...)

	for _, keyword := range a.ignoredKeywords {
		titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword))

		if categories.Includes(keyword) || titleContainsKeyword {
			return true
		}
	}

	return false
}

func (a *Aggregator) Start(ctx context.Context) error {
	ticker := time.NewTicker(a.updateInterval)
	defer ticker.Stop()

	if err := a.Aggregate(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.Aggregate(ctx); err != nil {
				return err
			}
		}
	}
}
