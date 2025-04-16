package client

import (
	"context"

	"github.com/morzisorn/gofermart/config"
	"github.com/morzisorn/gofermart/internal/models"
	"resty.dev/v3"
)

type LoyaltyClient interface {
	CalculateBonuses(ctx context.Context, number string) (*models.LoyaltyOrder, error)
}

func NewClient(cnfg *config.Config) LoyaltyClient {
	return &HTTPClient{
		BaseURL: cnfg.AccrualSystemAddress,
		Client: resty.New().
			SetBaseURL(cnfg.AccrualSystemAddress),
	}
}
