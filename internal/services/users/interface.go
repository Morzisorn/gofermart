package users

import (
	"context"

	"github.com/morzisorn/gofermart/internal/models"
)

type BalanceGetter interface {
	GetBalance(ctx context.Context, user *models.User) (*models.UserBalance, error)
}


