package storage

import (
	"context"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/models"
)

type Storage interface {
	PutUser(ctx context.Context, user *models.User) (userID int, err error)
	GetUser(ctx context.Context, user *models.User) (userID int, err error)
}
