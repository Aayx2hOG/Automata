package models

import "errors"

var (
	ErrorUserNotFound       = errors.New("User not found.")
	ErrorUserAlreadyExists  = errors.New("User already exists.")
	ErrorInvalidCredentials = errors.New("Invalid Email or Password.")
	ErrorInvalidToken       = errors.New("Invalid or expired Token.")
	ErrorTokenRevoked       = errors.New("Token has been revoked.")
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
