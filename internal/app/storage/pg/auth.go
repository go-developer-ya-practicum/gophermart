package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app"
	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/models"
)

func (s *StorageDB) PutUser(ctx context.Context, user *models.User) (int, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var id int
	err = conn.QueryRow(
		ctx, `INSERT INTO users (login, hash) VALUES($1, $2) ON CONFLICT DO NOTHING RETURNING id`,
		user.Login, user.Hash,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return id, app.ErrLoginIsAlreadyInUse
	}
	return id, err
}

func (s *StorageDB) GetUser(ctx context.Context, user *models.User) (int, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var id int
	err = conn.QueryRow(
		ctx, `SELECT id FROM users WHERE login = ($1) AND hash = ($2)`,
		user.Login, user.Hash,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return id, app.ErrInvalidCredentials
	}
	return id, err
}
