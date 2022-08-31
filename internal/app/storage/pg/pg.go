package pg

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/storage"
)

type StorageDB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (storage.Storage, error) {
	migrations, err := migrate.New("file://db/migrations", connString)
	if err != nil {
		return nil, err
	}
	err = migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &StorageDB{pool: pool}, nil
}
