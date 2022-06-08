package pg

import (
	"context"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app"
	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/models"
)

func (s *StorageDB) PutOrder(ctx context.Context, order *models.Order) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT user_id FROM orders WHERE id = ($1)`, order.Number)
	if err != nil {
		return err
	}
	if rows.Next() {
		var userID int
		if err = rows.Scan(&userID); err != nil {
			return err
		}
		if userID != order.UserID {
			return app.ErrOrderUploadedByAnotherUser
		}
		return app.ErrOrderAlreadyUploaded
	}

	_, err = tx.Exec(
		ctx, `INSERT INTO orders (id, user_id, status) VALUES($1, $2, $3)`,
		order.Number, order.UserID, order.Status)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *StorageDB) ListOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(
		ctx,
		`SELECT id, status, accrual, uploaded_at FROM orders WHERE user_id = ($1) ORDER BY uploaded_at`,
		userID)
	if err != nil {
		return nil, err
	}
	orders := make([]*models.Order, 0)
	for rows.Next() {
		order := &models.Order{}
		if err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt.Time); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (s *StorageDB) UpdateOrder(ctx context.Context, order *models.Order) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(
		ctx,
		"UPDATE orders SET status = $1, accrual = $2 where id = $3;",
		order.Status, order.Accrual, order.Number)
	return err
}
