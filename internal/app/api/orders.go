package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/joeljunstrom/go-luhn"
	"github.com/rs/zerolog/log"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app"
	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/models"
)

func (rs *Resources) UploadOrder(w http.ResponseWriter, r *http.Request) {
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
		case errors.Is(err, app.ErrOrderAlreadyUploaded):
			w.WriteHeader(http.StatusOK)
		default:
			log.Warn().Err(err).Msg("Failed to put order")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
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
