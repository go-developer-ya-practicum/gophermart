package api

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/joeljunstrom/go-luhn"
	"github.com/rs/zerolog/log"

	"github.com/hikjik/gophermart/internal/app"
	"github.com/hikjik/gophermart/internal/app/models"
)

func (rs *Resources) UploadOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)
	if _, err = strconv.Atoi(orderNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !luhn.Valid(orderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	userID := ctx.Value(contextKeyUserID).(int)

	order := &models.Order{
		UserID: userID,
		Number: orderNumber,
		Status: models.OrderStatusNew,
	}
	err = rs.Storage.PutOrder(ctx, order)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrOrderUploadedByAnotherUser):
			w.WriteHeader(http.StatusConflict)
			return
		case errors.Is(err, app.ErrOrderAlreadyUploaded):
			w.WriteHeader(http.StatusOK)
			return
		default:
			log.Warn().Err(err).Msg("Failed to put order")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	rs.WorkerPool.Do(func(ctx context.Context) {
		rs.processOrder(ctx, order)
	})
	w.WriteHeader(http.StatusAccepted)
}

func (rs *Resources) ListOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID := ctx.Value(contextKeyUserID).(int)

	orders, err := rs.Storage.ListOrders(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(orders); err != nil {
		log.Warn().Err(err).Msg("Failed to list orders")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rs *Resources) processOrder(ctx context.Context, order *models.Order) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			updatedOrder, err := rs.Provider.GetOrderAccrual(order.Number)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to get order accrual")
				continue
			}

			if updatedOrder.Status != order.Status {
				err = rs.Storage.UpdateOrder(ctx, updatedOrder)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to to update order info")
					return
				}
			}

			if updatedOrder.Status == models.OrderStatusProcessed {
				transaction := &models.Transaction{
					UserID:   order.UserID,
					OrderNum: order.Number,
					Amount:   updatedOrder.Accrual,
				}
				if err = rs.Storage.PutTransaction(ctx, transaction); err != nil {
					log.Warn().Err(err).Msg("Failed to store transaction")
				}
			}

			if updatedOrder.Status == models.OrderStatusProcessed || updatedOrder.Status == models.OrderStatusInvalid {
				return
			}
		}
	}
}
