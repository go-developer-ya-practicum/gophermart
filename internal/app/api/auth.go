package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"

	"github.com/hikjik/gophermart/internal/app"
	"github.com/hikjik/gophermart/internal/app/models"
)

type Claims struct {
	jwt.StandardClaims
	UserID int `json:"user_id"`
}

func (rs *Resources) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if user.Login == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash, err := computeHash(user.Password, rs.AuthKey)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to compute hash")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.Hash = hash
	id, err := rs.Storage.PutUser(r.Context(), &user)
	if err != nil {
		if errors.Is(err, app.ErrLoginIsAlreadyInUse) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Warn().Err(err).Msg("Failed to put user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(id, rs.AuthKey)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: tokenString,
	})
}

func (rs *Resources) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Login == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash, err := computeHash(user.Password, rs.AuthKey)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to compute hash")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.Hash = hash
	id, err := rs.Storage.GetUser(r.Context(), &user)
	if err != nil {
		if errors.Is(err, app.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Warn().Err(err).Msg("Failed to get user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(id, rs.AuthKey)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: tokenString,
	})
}

func generateToken(id int, key []byte) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(24 * time.Hour).Unix(),
			IssuedAt:  now.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func computeHash(data string, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write([]byte(data)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
