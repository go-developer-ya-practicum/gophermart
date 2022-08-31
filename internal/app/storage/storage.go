package storage

import (
	"context"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/models"
)

type Storage interface {
	PutUser(ctx context.Context, user *models.User) (userID int, err error)
	GetUser(ctx context.Context, user *models.User) (userID int, err error)

	PutOrder(ctx context.Context, order *models.Order) error
	UpdateOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context, userID int) ([]*models.Order, error)

	GetUserBalance(ctx context.Context, userID int) (balance *models.UserBalance, err error)
	ListWithdrawals(ctx context.Context, userID int) ([]*models.Transaction, error)
	PutTransaction(ctx context.Context, transaction *models.Transaction) error
}
