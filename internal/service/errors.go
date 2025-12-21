package service

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal error")
)
