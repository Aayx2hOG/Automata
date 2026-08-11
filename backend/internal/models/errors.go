package models

import "errors"

var (
	ErrorUserNotFound       = errors.New("user not found")
	ErrorUserAlreadyExists  = errors.New("user already exists")
	ErrorInvalidCredentials = errors.New("invalid email or password")
	ErrorInvalidToken       = errors.New("invalid or expired token")
	ErrorTokenRevoked       = errors.New("token has been revoked")
)

type AppError struct {
	Status  int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(status int, message string, err error) *AppError {
	return &AppError{Status: status, Message: message, Err: err}
}
