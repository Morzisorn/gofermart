package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/morzisorn/gofermart/internal/models"
	"resty.dev/v3"
)

type HTTPClient struct {
	BaseURL string
	Client  *resty.Client
}

func (c *HTTPClient) CalculateBonuses(ctx context.Context, number string) (*models.LoyaltyOrder, error) {
	base := &url.URL{
		Scheme: "http",
		Host: c.BaseURL,
		Path: "api/orders/",
	}

	url := base.ResolveReference(&url.URL{Path: number})

	var order models.LoyaltyOrder
	resp, err := c.Client.R().
		SetResult(&order).
		Get(url.String())

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return &order, nil
}
