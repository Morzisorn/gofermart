package errs

import "errors"

var (
	ErrIncorrectNumber         = errors.New("number validation failed")
	ErrOrderAlreadyExist       = errors.New("order number is already exist")
	ErrOrderBelongsAnotherUser = errors.New("belongs to another user")
	ErrNoData                  = errors.New("no data")
	ErrInsufficientBalance     = errors.New("insufficient balance")
)
