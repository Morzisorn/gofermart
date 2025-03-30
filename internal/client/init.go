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
		BaseURL: "localhost:8080",
		Client: resty.New().
			SetBaseURL("localhost:8080"),
	}
}
