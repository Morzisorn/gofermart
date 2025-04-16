package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/morzisorn/gofermart/internal/errs"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/repositories"
	"github.com/morzisorn/gofermart/internal/services/users"
)

type OrderService struct {
	repo repositories.Repository
	user users.BalanceGetter
}

func NewOrderService(repo repositories.Repository, user users.BalanceGetter) *OrderService {
	return &OrderService{
		repo: repo,
		user: user,
	}
}

func (os *OrderService) UploadOrder(ctx context.Context, login, number string) error {
	if !isNumberValid(number) {
		return fmt.Errorf("failed to upload order: %w", errs.ErrIncorrectNumber)
	}

	_, err := os.repo.UploadOrder(ctx, login, number)
	if err != nil {
		return fmt.Errorf("failed to upload order: %w", err)
	}

	return nil
}

func (os *OrderService) GetUserOrders(ctx context.Context, login string) (*[]models.Order, error) {
	orders, err := os.repo.GetUserOrders(ctx, login)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("get user orders error: %w", errs.ErrNoData)
	case err != nil:
		return nil, fmt.Errorf("get user orders error: %w", err)
	}
	return orders, nil
}

func (os *OrderService) GetUpprocessedOrders(ctx context.Context) (*[]models.Order, error) {
	return os.repo.GetUpprocessedOrders(ctx)
}

func (os *OrderService) GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdrawal, error) {
	withdrawals, err := os.repo.GetUserWithdrawals(ctx, login)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("get user withdrawals error: %w", errs.ErrNoData)
	case err != nil:
		return nil, fmt.Errorf("get user withdrawals error: %w", err)
	}
	return withdrawals, nil
}

func (os *OrderService) Withdraw(ctx context.Context, login string, w *models.Withdrawal) error {
	balance, err := os.user.GetBalance(ctx, &models.User{Login: login})
	if err != nil {
		return fmt.Errorf("withdrawal error: %w", err)
	}

	if balance.Current < w.Sum {
		return fmt.Errorf("withdrawal error: %w", errs.ErrInsufficientBalance)
	}

	if !isNumberValid(w.Number) {
		return fmt.Errorf("withdrawal error: %w", errs.ErrIncorrectNumber)
	}

	return os.repo.Withdraw(ctx, login, w.Number, w.Sum)
}

func (os *OrderService) UpdateOrderStatus(ctx context.Context, number, newStatus string) error {
	err := os.repo.UpdateOrderStatus(ctx, number, models.OrderStatusPROCESSING)
	if err != nil {
		return fmt.Errorf("update order status error: %w", err)
	}
	return nil
}

func (os *OrderService) OrderProcessed(ctx context.Context, order models.Order) error {
	err := os.repo.OrderProcessed(ctx, order.UserLogin, order.Number, order.Accrual)
	if err != nil {
		return fmt.Errorf("finish order processing error: %w", err)
	}
	return nil
}

// Luhn algorithm
func isNumberValid(number string) bool {
	numR := []rune(number)

	var checkEven bool

	if len(numR)%2 == 1 {
		checkEven = true
	}

	var sum int

	for i := 0; i < len(numR); i++ {
		n := int(numR[i]) - int('0')
		if i%2 == 1 && checkEven || i%2 == 0 && !checkEven {
			if n*2 > 9 {
				sum += n*2 - 9
			} else {
				sum += n * 2
			}
		} else {
			sum += n
		}
	}

	return sum%10 == 0
}
