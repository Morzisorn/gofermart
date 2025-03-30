package orders

import (
	"context"
	"database/sql"
	"errors"

	"github.com/morzisorn/gofermart/internal/logger"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/repositories"
	"github.com/morzisorn/gofermart/internal/services/users"
	"go.uber.org/zap"
)

var (
	ErrIncorrectNumber         = errors.New("number validation failed")
	ErrOrderAlreadyExist       = errors.New("order number is already exist")
	ErrOrderBelongsAnotherUser = errors.New("belongs to another user")
	ErrNoData                  = errors.New("no data")
	ErrInsufficientBalance     = errors.New("insufficient balance")
)

type OrderService struct {
	repo repositories.Repository
	user *users.UserService
}

func NewOrderService(repo repositories.Repository, user *users.UserService) *OrderService {
	return &OrderService{
		repo: repo,
		user: user,
	}
}

func (os *OrderService) UploadOrder(ctx context.Context, login, number string) error {
	if !isNumberValid(number) {
		return ErrIncorrectNumber
	}

	dbLogin, err := os.repo.UploadOrder(ctx, login, number)
	switch {
	case errors.Is(err, ErrOrderAlreadyExist):
		if login == dbLogin {
			return ErrOrderAlreadyExist
		}
		return ErrOrderBelongsAnotherUser
	case err != nil:
		return err
	}

	//go processOrder(ctx, number)

	return nil
}

func (os *OrderService) GetUserOrders(ctx context.Context, login string) (*[]models.Order, error) {
	orders, err := os.repo.GetUserOrders(ctx, login)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNoData
	case err != nil:
		return nil, err
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
		return nil, ErrNoData
	case err != nil:
		return nil, err
	}
	return withdrawals, nil
}

func (os *OrderService) Withdraw(ctx context.Context, login string, w *models.Withdrawal) error {
	balance, err := os.user.GetBalance(ctx, &models.User{Login: login})
	if err != nil {
		return err
	}

	if balance.Current < w.Sum {
		return ErrInsufficientBalance
	}

	if !isNumberValid(w.Number) {
		return ErrIncorrectNumber
	}

	return os.repo.Withdraw(ctx, login, w.Number, w.Sum)
}

func (os *OrderService) UpdateOrderStatus(ctx context.Context, number, newStatus string) error {
	err := os.repo.UpdateOrderStatus(ctx, number, models.OrderStatusPROCESSING)
	if err != nil {
		logger.Log.Error("Failed to change order status. ",
			zap.String("Order number: %s", number),
			zap.String("New status: %s", models.OrderStatusPROCESSING),
		)
	}
	return err
}

func (os *OrderService) OrderProcessed(ctx context.Context, order models.Order) error {
	return os.repo.OrderProcessed(ctx, order.UserLogin, order.Number, order.Accrual)
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
