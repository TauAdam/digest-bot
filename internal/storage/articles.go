package storage

import (
	"context"
	"database/sql"
	"github.com/TauAdam/digest-bot/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"time"
)

type ArticlePostgresStorage struct {
	db *sqlx.DB
}

func (s *ArticlePostgresStorage) Save(ctx context.Context, article model.Article) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO articles (source_id, title, link, summary, published_at)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT DO NOTHING`,
		article.SourceID,
		article.Title,
		article.Link,
		article.Summary,
		article.PublishedAt,
	); err != nil {
		return err
	}

	return nil
}

// UnsentArticles returns articles that have not been posted yet and were published after the specified time.
func (s *ArticlePostgresStorage) UnsentArticles(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []ArticlePostgres
	if err := conn.SelectContext(
		ctx,
		&articles,
		`SELECT * FROM articles
         WHERE posted_at IS NULL
        AND published_at >= $1::timestamp 
        ORDER BY published_at DESC 
        LIMIT $2`, since.UTC().Format(time.RFC3339), limit,
	); err != nil {
		return nil, err
	}

	return lo.Map(articles, func(a ArticlePostgres, _ int) model.Article {
		return model.Article{
			ID:          a.ID,
			SourceID:    a.SourceID,
			Title:       a.Title,
			Link:        a.Link,
			Summary:     a.Summary,
			PublishedAt: a.PublishedAt,
			PostedAt:    a.PostedAt.Time,
			CreatedAt:   a.CreatedAt,
		}
	}), nil
}

// MarkPosted marks the article as posted.
func (s *ArticlePostgresStorage) MarkPosted(ctx context.Context, id int64) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`UPDATE articles SET posted_at = $1 WHERE id = $2`,
		time.Now().UTC().Format(time.RFC3339), id,
	); err != nil {
		return err
	}

	return nil
}

func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{
		db: db,
	}
}

// ArticlePostgres represents an article in the Postgres database.
type ArticlePostgres struct {
	ID          int64        `db:"id"`
	SourceID    int64        `db:"source_id"`
	Title       string       `db:"title"`
	Link        string       `db:"link"`
	Summary     string       `db:"summary"`
	PublishedAt time.Time    `db:"published_at"`
	PostedAt    sql.NullTime `db:"posted_at"`
	CreatedAt   time.Time    `db:"created_at"`
}
