package app

import "errors"

var (
	ErrLoginIsAlreadyInUse        = errors.New("login is already in use")
	ErrInvalidCredentials         = errors.New("invalid request credentials")
	ErrOrderAlreadyUploaded       = errors.New("order already uploaded")
	ErrOrderUploadedByAnotherUser = errors.New("order was uploaded by another user")
)
