package sources

import (
	"context"
	"github.com/SlyMarbo/rss"
	"github.com/TauAdam/digest-bot/internal/model"
	"github.com/samber/lo"
)

type RSSSource struct {
	URL        string
	SourceName string
	SourceID   int64
}

func NewRSSSource(m model.Source) RSSSource {
	return RSSSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

// processFeed fetches RSS feed from URL and returns it as a Feed struct
func (s RSSSource) processFeed(ctx context.Context, URL string) (*rss.Feed, error) {
	var (
		feedChannel  = make(chan *rss.Feed)
		errorsChanel = make(chan error)
	)

	go func() {
		feed, err := rss.Fetch(URL)
		if err != nil {
			errorsChanel <- err
			return
		}
		feedChannel <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errorsChanel:
		return nil, err
	case feed := <-feedChannel:
		return feed, nil
	}
}

func (s RSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	feed, err := s.processFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}

	// Map RSS items to model items
	return lo.Map(feed.Items, func(item *rss.Item, index int) model.Item {
		return model.Item{
			Title:      item.Title,
			Link:       item.Link,
			Categories: item.Categories,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.SourceName,
		}
	}), nil
}
