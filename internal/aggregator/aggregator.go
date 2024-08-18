package aggregator

import (
	"context"
	"github.com/TauAdam/digest-bot/internal/model"
	"github.com/TauAdam/digest-bot/internal/rss-sources"
	"log"
	"sync"
	"time"
)

type ArticleRepo interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceRepo interface {
	ListSources(ctx context.Context) ([]model.Source, error)
}

type Aggregator struct {
	articles ArticleRepo
	sources  SourceRepo

	updateInterval time.Duration
	whitelist      []string
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
		articles:       articleRepo,
		sources:        sourceRepo,
		updateInterval: updateInterval,
		whitelist:      whitelist,
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
