package app

import "errors"

var (
	ErrLoginIsAlreadyInUse = errors.New("login is already in use")
	ErrInvalidCredentials  = errors.New("invalid request credentials")
)
