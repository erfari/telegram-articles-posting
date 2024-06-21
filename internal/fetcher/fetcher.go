package fetcher

import (
	"articles/telegram-articles-posting/internal/model"
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tomakado/containers/set" //переделать на сет
)

type ArticlesStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceProvider interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles ArticlesStorage
	sources  SourceProvider

	fetchInterval  time.Duration
	filterKeywords []string
}

func New(
	articlesStorage ArticlesStorage,
	sourceProvider SourceProvider,
	fetchInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articlesStorage,
		sources:        sourceProvider,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, source := range sources {
		wg.Add(1)
		rssSource := source.NewRSSSourceFromModel(src)
		go func(source Source) {
			defer wg.Done()
			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] Fetching items from source %s: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] Failed to process items from source %s: %v", source.Name(), err)
				return
			}
		}(src.NewRSSSourceFromModel(source))
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			continue
		}

		if err := f.articles.Store(ctx, model.Article{
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

func (f *Fetcher) itemShouldBeSkipped(item model.Item) bool { // тут переделать на сет
	categoriesSet := set.New(item.Categories...)

	for _, keyword := range f.filterKeywords {
		if categoriesSet.Contains(keyword) || strings.Contains(strings.ToLower(item.Title), keyword) {
			return true
		}
	}
	return false
}
