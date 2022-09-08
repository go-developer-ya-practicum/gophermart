package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/hikjik/gophermart/internal/app/provider"
	"github.com/hikjik/gophermart/internal/app/storage"
	"github.com/hikjik/gophermart/pkg/wpool"
)

type Resources struct {
	AuthKey    []byte
	Storage    storage.Storage
	Provider   provider.Provider
	WorkerPool *wpool.WorkerPool
}

func (rs *Resources) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/user", func(r chi.Router) {
		r.Use(middleware.Compress(5))
		r.Post("/register", rs.SignUp)
		r.Post("/login", rs.SignIn)

		r.Group(func(r chi.Router) {
			r.Use(rs.AuthMiddleware)

			r.Post("/orders", rs.UploadOrder)
			r.Get("/orders", rs.ListOrders)

			r.Get("/balance", rs.GetUserBalance)
			r.Post("/balance/withdraw", rs.Withdraw)
			r.Get("/balance/withdrawals", rs.ListWithdrawals)
		})
	})
	return r
}
