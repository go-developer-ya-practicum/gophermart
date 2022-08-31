package pg

import (
	"context"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/models"
)

func (s *StorageDB) GetUserBalance(ctx context.Context, userID int) (*models.UserBalance, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var balance models.UserBalance
	err = conn.QueryRow(
		ctx, `SELECT balance, withdrawn FROM users WHERE id = ($1)`,
		userID,
	).Scan(&balance.Current, &balance.Withdrawn)
	return &balance, err
}

func (s *StorageDB) ListWithdrawals(ctx context.Context, userID int) ([]*models.Transaction, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(
		ctx,
		`SELECT order_id, amount, created_at FROM transactions WHERE user_id = ($1) ORDER BY created_at`,
		userID)
	if err != nil {
		return nil, err
	}
	transactions := make([]*models.Transaction, 0)
	for rows.Next() {
		transaction := &models.Transaction{}
		if err = rows.Scan(&transaction.OrderNum, &transaction.Amount, &transaction.ProcessedAt.Time); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (s *StorageDB) PutTransaction(ctx context.Context, transaction *models.Transaction) error {
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

	_, err = tx.Exec(ctx,
		"INSERT INTO transactions (user_id, order_id, amount) VALUES ($1, $2, $3)",
		transaction.UserID, transaction.OrderNum, transaction.Amount)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		"UPDATE users SET balance = balance + $1 WHERE id = $2",
		transaction.Amount, transaction.UserID)
	if err != nil {
		return err
	}

	if transaction.Amount < 0 {
		_, err = tx.Exec(ctx,
			"UPDATE users SET withdrawn = withdrawn - $1 WHERE id = $2",
			transaction.Amount, transaction.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
