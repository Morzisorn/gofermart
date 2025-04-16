package processing

import (
	"context"
	"sync"

	"github.com/morzisorn/gofermart/config"
	"github.com/morzisorn/gofermart/internal/client"
	"github.com/morzisorn/gofermart/internal/logger"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/services/orders"
	"go.uber.org/zap"
)

type ProcessingService struct {
	service *orders.OrderService
	client  client.LoyaltyClient
}

func NewProcessingService(service *orders.OrderService, client client.LoyaltyClient) *ProcessingService {
	return &ProcessingService{
		service: service,
		client:  client,
	}
}

func (ps *ProcessingService) ProcessOrders(ctx context.Context) error {
	chIn := ps.ordersProducer(ctx)

	var wg sync.WaitGroup
	var loyaltyWg sync.WaitGroup

	rateLimit := config.GetConfig().RateLimit

	chLoyaltyUpdates := make(chan models.Order, 10)

	ps.runLoyaltyWorkers(ctx, chIn, chLoyaltyUpdates, &loyaltyWg, rateLimit)

	go func() {
		loyaltyWg.Wait()
		close(chLoyaltyUpdates)
	}()

	ps.runUpdateWorker(ctx, chLoyaltyUpdates, &wg, rateLimit)

	wg.Wait()

	return nil
}

func (ps *ProcessingService) ordersProducer(ctx context.Context) chan models.Order {
	orders, err := ps.service.GetUpprocessedOrders(ctx)
	if err != nil {
		logger.Log.Panic("Failed to get unprocessed orders")
	}

	ch := make(chan models.Order, len(*orders))

	go func() {
		defer close(ch)
		for _, o := range *orders {
			ch <- o
		}
	}()

	return ch
}

func (ps *ProcessingService) runLoyaltyWorkers(ctx context.Context, chIn chan models.Order, chOut chan models.Order, wg *sync.WaitGroup, rateLimit int) {
	for w := 0; w < rateLimit; w++ {
		wg.Add(1)
		go ps.loyaltyJob(ctx, chIn, chOut, wg)
	}
}

func (ps *ProcessingService) loyaltyJob(ctx context.Context, chIn chan models.Order, chOut chan models.Order, wg *sync.WaitGroup) {
	defer wg.Done()

	for o := range chIn {
		lo, err := ps.client.CalculateBonuses(ctx, o.Number)
		if err != nil {
			logger.Log.Error("Failed to calculate bonuses. ", zap.String("Order number: %s", o.Number))
			continue
		}

		switch lo.Status {
		case models.LoyaltyStatusREGISTERED:
			if o.Status == models.OrderStatusPROCESSING {
				continue
			}
			o.Status = models.OrderStatusPROCESSING
		case models.LoyaltyStatusPROCESSING:
			if o.Status == models.OrderStatusNEW {
				o.Status = models.OrderStatusPROCESSING
			}
		case models.LoyaltyStatusINVALID:
			o.Status = models.OrderStatusINVALID
		case models.OrderStatusPROCESSED:
			o.Accrual = lo.Accrual
			o.Status = models.OrderStatusPROCESSED
		}

		chOut <- o
	}
}

func (ps *ProcessingService) runUpdateWorker(ctx context.Context, chIn chan models.Order, wg *sync.WaitGroup, rateLimit int) {
	for w := 0; w < rateLimit; w++ {
		wg.Add(1)
		go ps.updateOrdersJob(ctx, chIn, wg)
	}
}

func (ps *ProcessingService) updateOrdersJob(ctx context.Context, chIn chan models.Order, wg *sync.WaitGroup) {
	defer wg.Done()
	for o := range chIn {
		switch o.Status {
		case models.OrderStatusPROCESSED:
			_ = ps.service.OrderProcessed(ctx, o)
		default:
			_ = ps.service.UpdateOrderStatus(ctx, o.Number, o.Status)
		}
	}
}
