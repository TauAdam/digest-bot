package storage

import (
	"context"
	"github.com/TauAdam/digest-bot/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"time"
)

type SourcePostgres struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	FeedURL   string    `db:"feed_url"`
	CreatedAt time.Time `db:"created_at"`
}

type SourcesPostgresStorage struct {
	db *sqlx.DB
}

func (s *SourcesPostgresStorage) ListSources(ctx context.Context) ([]model.Source, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var sources []SourcePostgres
	if err := conn.SelectContext(ctx, &sources, `SELECT * FROM sources`); err != nil {
		return nil, err
	}

	return lo.Map(sources, func(s SourcePostgres, _ int) model.Source {
		return model.Source(s)
	}), nil
}

func (s *SourcesPostgresStorage) SourceByID(ctx context.Context, id int64) (*model.Source, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var source SourcePostgres
	if err := conn.GetContext(ctx, &source, `SELECT * FROM sources WHERE id=$1`, id); err != nil {
		return nil, err
	}

	// Type conversion
	return (*model.Source)(&source), nil
}

func (s *SourcesPostgresStorage) AddSource(ctx context.Context, source model.Source) (int64, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var id int64

	row := conn.QueryRowxContext(
		ctx,
		`INSERT INTO sources (name, feed_url, created_at) VALUES ($1, $2, $3) RETURNING id`,
		source.Name,
		source.FeedURL,
		source.CreatedAt,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SourcesPostgresStorage) DeleteSource(ctx context.Context, id int64) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM sources WHERE id=$1`, id); err != nil {
		return err
	}

	return nil
}
