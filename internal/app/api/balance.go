package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/joeljunstrom/go-luhn"
	"github.com/rs/zerolog/log"

	"github.com/hikjik/gophermart/internal/app/models"
)

func (rs *Resources) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID := ctx.Value(contextKeyUserID).(int)

	balance, err := rs.Storage.GetUserBalance(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(balance); err != nil {
		log.Warn().Err(err).Msg("Failed to encode balance")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rs *Resources) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	userID := ctx.Value(contextKeyUserID).(int)

	withdrawals, err := rs.Storage.ListWithdrawals(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get withdrawals")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(withdrawals); err != nil {
		log.Warn().Err(err).Msg("Failed to encode withdrawals")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rs *Resources) Withdraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(contextKeyUserID).(int)

	var transaction models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(transaction.OrderNum); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !luhn.Valid(transaction.OrderNum) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	balance, err := rs.Storage.GetUserBalance(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if balance.Current < transaction.Amount {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	transaction.UserID = userID
	transaction.Amount *= -1
	if err = rs.Storage.PutTransaction(ctx, &transaction); err != nil {
		log.Warn().Err(err).Msg("Failed to put Transaction")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
