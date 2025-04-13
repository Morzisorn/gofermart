package errs

import "errors"

var (
	//Order errors
	ErrIncorrectNumber         = errors.New("number validation failed")
	ErrOrderAlreadyExist       = errors.New("order number is already exist")
	ErrOrderBelongsAnotherUser = errors.New("belongs to another user")
	ErrNoData                  = errors.New("no data")
	
	//User errors
	ErrInsufficientBalance     = errors.New("insufficient balance")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyRegistered = errors.New("user is already registered")
	ErrIncorrectCredentials  = errors.New("incorrect login or password")
	
	//Other errors
	ErrInternalServerError   = errors.New("internal server error")
)
