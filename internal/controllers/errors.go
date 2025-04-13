package controllers

import (
	"errors"
	"net/http"

	"github.com/morzisorn/gofermart/internal/errs"
)

func statusFromError(err error) int {
	switch {
	case errors.Is(err, errs.ErrIncorrectNumber):
		return http.StatusUnprocessableEntity
	case errors.Is(err, errs.ErrOrderAlreadyExist):
		return http.StatusOK
	case errors.Is(err, errs.ErrOrderBelongsAnotherUser):
		return http.StatusConflict
	case errors.Is(err, errs.ErrNoData):
		return http.StatusNoContent
	case errors.Is(err, errs.ErrInsufficientBalance):
		return http.StatusPaymentRequired
	case errors.Is(err, errs.ErrUserAlreadyRegistered):
		return http.StatusConflict
	case errors.Is(err, errs.ErrIncorrectCredentials):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
